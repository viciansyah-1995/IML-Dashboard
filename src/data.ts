export type Scope = 'room' | 'slot';
export type Dataset = 'source' | 'simulation';
export type Metrics = {
    sourceGmv: number | null;
    estimatedGmv: number | null;
    impressions: number | null;
    views: number | null;
    attributedOrders: number | null;
    buyers: number | null;
    averageViewingSeconds: number | null;
    currentViewers: number | null;
    sourceAovIdr: number | null;
    productClicks: number | null;
    newFollowers: number | null;
    likes: number | null;
    comments: number | null;
    shares: number | null;
    productsSold: number | null;
    sourceLiveCtr: number | null;
    sourceCtor: number | null;
    sourceGmvMaxRoi: number | null;
    adSpend: number | null;
};
export type Report = {
    id: string;
    date: string;
    channel: string;
    studio: string | null;
    scope: Scope;
    sourceTimeLabel: string;
    durationMinutes: number;
    capturedAt: null;
    metrics: Metrics | null;
    host?: string;
    idLine?: string;
    relatedSourceReportId?: string;
    valid: boolean;
    simulation: boolean;
};
export type Filters = {
    from: string;
    to: string;
    scope: Scope;
    channel: string;
    studio: string;
    host: string;
};
export const brand = { name: 'Zwitsal', account: 'zwitsal_id' };
export const sourceRoom: Report = { id: 'demo-room-report-20260926', date: '2026-09-26', channel: 'TikTok', studio: null, scope: 'room', sourceTimeLabel: 'Sekitar 06.00–18.00 WIB', durationMinutes: 43192 / 60, capturedAt: null, valid: false, simulation: false, metrics: { sourceGmv: 15484967, estimatedGmv: 12240000, impressions: 2420000, views: 61520, attributedOrders: 222, buyers: 201, averageViewingSeconds: 15, currentViewers: 128, sourceAovIdr: 69750, productClicks: 3820, newFollowers: 59, likes: 5220, comments: 392, shares: 56, productsSold: 270, sourceLiveCtr: 0.0621, sourceCtor: 0.0581, sourceGmvMaxRoi: 5.03, adSpend: 3340000 } };
export const sourceSlot: Report = { id: 'demo-slot-260926-00002', date: '2026-09-26', channel: 'TikTok', studio: 'Studio 1', scope: 'slot', sourceTimeLabel: '16.00–18.00 WIB', durationMinutes: 120, capturedAt: null, host: 'Linda', idLine: '260926-00002', relatedSourceReportId: sourceRoom.id, valid: false, simulation: false, metrics: null };
const additive = new Set<keyof Metrics>(['sourceGmv', 'estimatedGmv', 'impressions', 'views', 'attributedOrders', 'productClicks', 'newFollowers', 'likes', 'comments', 'shares', 'productsSold', 'adSpend']);
const syntheticSlots: Report[] = Array.from({ length: 28 }, (_, i) => {
    const day = Math.floor(i / 2), host = ['Host A', 'Host B', 'Host C'][i % 3], gmv = 1250000 + ((i * 791) % 2400) * 1000, orders = 25 + (i * 7) % 45;
    return { id: `sim-slot-${i + 1}`, date: `2026-09-${String(13 + day).padStart(2, '0')}`, channel: 'TikTok', studio: i % 2 ? 'Studio 2' : 'Studio 1', scope: 'slot', sourceTimeLabel: i % 2 ? '14.00–16.00 WIB' : '10.00–12.00 WIB', durationMinutes: 120, capturedAt: null, host, idLine: `SIM-2609-${String(i + 1).padStart(3, '0')}`, valid: true, simulation: true, metrics: { sourceGmv: gmv, estimatedGmv: null, impressions: 120000 + (i * 7223) % 200000, views: 5000 + (i * 923) % 8000, attributedOrders: orders, buyers: null, averageViewingSeconds: null, currentViewers: null, sourceAovIdr: null, productClicks: 450 + (i * 143) % 850, newFollowers: 10 + (i * 3) % 40, likes: 900 + (i * 177) % 2000, comments: 70 + (i * 23) % 160, shares: 12 + (i * 7) % 40, productsSold: orders + 15, sourceLiveCtr: null, sourceCtor: null, sourceGmvMaxRoi: null, adSpend: null } };
});
const syntheticRooms: Report[] = Array.from({ length: 14 }, (_, i) => {
    const slots = syntheticSlots.slice(i * 2, i * 2 + 2);
    return { id: `sim-room-${i + 1}`, date: slots[0].date, channel: 'TikTok', studio: null, scope: 'room', sourceTimeLabel: '10.00–16.00 WIB (simulasi)', durationMinutes: 360, capturedAt: null, valid: true, simulation: true, metrics: aggregate(slots) };
});
export const datasets = { source: { rooms: [sourceRoom], slots: [sourceSlot] }, simulation: { rooms: syntheticRooms, slots: syntheticSlots } };
export function defaults(dataset: Dataset): Filters { return { from: dataset === 'source' ? '2026-09-26' : '2026-09-13', to: '2026-09-26', scope: 'room', channel: 'all', studio: 'all', host: 'all' }; }
;
export function filterReports(rows: Report[], f: Filters) { return rows.filter(r => r.date >= f.from && r.date <= f.to && (f.channel === 'all' || r.channel === f.channel) && (f.studio === 'all' || r.studio === f.studio) && (f.scope === 'room' || f.host === 'all' || r.host === f.host)); }
export function aggregate(rows: Report[]): Metrics { const keys = Object.keys(sourceRoom.metrics!) as (keyof Metrics)[]; return Object.fromEntries(keys.map(key => { const values = rows.map(r => r.metrics?.[key]).filter((v): v is number => v != null); return [key, values.length === 0 ? null : additive.has(key) ? values.reduce((a, b) => a + b, 0) : rows.length === 1 ? values[0] : null]; })) as Metrics; }
export function leaderboard(rows: Report[], sort: 'gmv' | 'hour' = 'gmv') {
    const unique = [...new Map(rows.map(r => [r.id, r])).values()];
    const names = [...new Set(unique.map(r => r.host).filter((h): h is string => !!h))];
    const all = names.map(name => { const slots = unique.filter(r => r.host === name); const valid = slots.filter(r => r.scope === 'slot' && r.valid && r.metrics?.sourceGmv != null); const gmv = valid.reduce((s, r) => s + r.metrics!.sourceGmv!, 0), hours = valid.reduce((s, r) => s + r.durationMinutes / 60, 0); return { name, slots, valid, gmv: valid.length ? gmv : null, hours, perHour: hours ? gmv / hours : null, rank: null as number | null, contribution: null as number | null }; });
    const total = all.reduce((s, h) => s + (h.gmv ?? 0), 0);
    const ranked = all.filter(h => h.gmv != null).sort((a, b) => (sort === 'gmv' ? b.gmv! - a.gmv! : (b.perHour ?? 0) - (a.perHour ?? 0)) || a.name.localeCompare(b.name));
    ranked.forEach((h, i) => { const prev = ranked[i - 1]; h.rank = prev && (sort === 'gmv' ? prev.gmv === h.gmv : prev.perHour === h.perHour) ? prev.rank : i + 1; h.contribution = total > 0 ? h.gmv! / total * 100 : null; });
    return { ranked, missing: all.filter(h => h.gmv == null), total };
}
export const number = (n: number | null | undefined) => n == null ? '—' : new Intl.NumberFormat('id-ID', { maximumFractionDigits: 2 }).format(n);
export const rupiah = (n: number | null | undefined) => n == null ? '—' : `Rp${number(Math.round(n))}`;
export const dateLabel = (s: string) => new Date(s + 'T00:00:00+07:00').toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'Asia/Jakarta' });
export const periodLabel = (f: Filters) => f.from === f.to ? dateLabel(f.from) : `${dateLabel(f.from)} – ${dateLabel(f.to)}`;
export const scopeLabel = (s: Scope) => s === 'room' ? 'Sepanjang LIVE' : 'Per slot host';
export const datasetLabel = (d: Dataset) => d === 'source' ? 'Pratinjau data contoh' : 'Data simulasi';
export const definitions: Record<keyof Metrics, {
    label: string;
    kind: 'money' | 'number' | 'seconds' | 'percent' | 'ratio';
    info: string;
    raw?: string;
}> = {
    sourceGmv: { label: 'Perolehan GMV', kind: 'money', info: 'Kolom ledger P: Direct GMV. Definisi final masih perlu dikonfirmasi. Nilai sepanjang LIVE tidak dapat diatribusikan ke slot host.' },
    estimatedGmv: { label: 'Perkiraan GMV', kind: 'money', info: 'Nilai sumber terpisah; tidak dijumlahkan dengan Perolehan GMV dan belum didefinisikan sebagai Indirect GMV.', raw: '12.24M' },
    impressions: { label: 'Impresi', kind: 'number', info: 'Label screenshot: Impressions. Label PV pada ledger masih perlu dikonfirmasi. Nilai tampilan dibulatkan.', raw: '2.42M' },
    views: { label: 'Tayangan', kind: 'number', info: 'Label screenshot: Views. Definisi PV ledger masih perlu dikonfirmasi; bukan unique viewers.', raw: '61.52K' },
    attributedOrders: { label: 'Pesanan teratribusi', kind: 'number', info: 'Pesanan sesuai cakupan sumber; berbeda dengan produk terjual.' },
    buyers: { label: 'Pembeli', kind: 'number', info: 'Pembeli pada laporan sumber. Tidak diklaim sebagai unique buyers lintas periode.' },
    averageViewingSeconds: { label: 'Durasi tonton rata-rata', kind: 'seconds', info: 'Rata-rata durasi tonton dari sumber, dalam detik.' },
    currentViewers: { label: 'Penonton saat screenshot', kind: 'number', info: 'Snapshot 128 penonton; waktu pengambilan belum tersedia. Bukan total atau unique viewers. Tidak diagregasikan.' },
    sourceAovIdr: { label: 'AOV sumber', kind: 'money', info: 'Nilai screenshot, dibulatkan. Definisi Avg. Price belum disepakati; bukan rata-rata harga produk.', raw: '69.75K' },
    productClicks: { label: 'Klik produk', kind: 'number', info: 'Product clicks pada screenshot. Nilai tampilan dibulatkan.', raw: '3.82K' },
    newFollowers: { label: 'Pengikut baru', kind: 'number', info: 'New followers dalam cakupan laporan.' }, likes: { label: 'Suka', kind: 'number', info: 'Likes pada sumber; nilai tampilan dibulatkan.', raw: '5.22K' }, comments: { label: 'Komentar', kind: 'number', info: 'Comments dalam cakupan laporan.' }, shares: { label: 'Dibagikan', kind: 'number', info: 'Shares dalam cakupan laporan.' }, productsSold: { label: 'Produk terjual', kind: 'number', info: '270 produk pada screenshot; bukan jumlah pesanan.' },
    sourceLiveCtr: { label: 'CTR LIVE sumber', kind: 'percent', info: 'Rasio sumber 6,21%, tidak dihitung ulang oleh aplikasi.' }, sourceCtor: { label: 'CTOR sumber', kind: 'percent', info: 'Rasio sumber 5,81%, tidak dihitung ulang oleh aplikasi.' }, sourceGmvMaxRoi: { label: 'ROI GMV Max sumber', kind: 'ratio', info: 'Nilai sumber 5,03. Basis belum diketahui; bukan ROAS hitungan aplikasi.' }, adSpend: { label: 'Biaya iklan', kind: 'money', info: 'Nilai sumber dibulatkan.', raw: '3.34M' }
};
export function metricValue(key: keyof Metrics, value: number | null | undefined) { if (value == null)
    return '—'; const kind = definitions[key].kind; return kind === 'money' ? rupiah(value) : kind === 'seconds' ? `${number(value)} detik` : kind === 'percent' ? `${number(value * 100)}%` : number(value); }
export function csvCell(value: unknown) { const s = String(value ?? ''); return '"' + (/^[=+@\-\t\r]/.test(s) ? "'" + s : s).replaceAll('"', '""') + '"'; }
export function reportCsv(rows: Report[], f: Filters, d: Dataset, mode: 'summary' | 'detail' | 'hosts') {
    const metadata = [['Dataset', datasetLabel(d)], ['Periode', periodLabel(f)], ['Cakupan', mode === 'hosts' ? 'Per slot host' : scopeLabel(f.scope)], ['Channel', f.channel], ['Studio', f.studio], ['Host', f.host], ['Basis omzet', d === 'simulation' ? 'demo_gross_gmv' : 'Direct GMV sumber; definisi perlu dikonfirmasi'], []];
    let body: unknown[][];
    if (mode === 'hosts') {
        body = [['Peringkat', 'Host', 'Total omzet live IDR', 'Kontribusi %', 'Slot valid', 'Jam live', 'Omzet per jam IDR', 'Cakupan data', 'ID Line tercakup'], ...leaderboard(rows).ranked.map(h => [h.rank, h.name, h.gmv, h.contribution, h.valid.length, h.hours, h.perHour, `${h.valid.length}/${h.slots.length}`, h.valid.map(s => s.idLine).join('; ')]), ...leaderboard(rows).missing.map(h => ['', h.name, '', '', '', '', '', `0/${h.slots.length}`, 'Hasil belum tersedia'])];
    }
    else {
        const keys = Object.keys(definitions) as (keyof Metrics)[];
        body = mode === 'summary' ? [['Metrik', 'Nilai', 'Satuan / definisi'], ...keys.map(k => [definitions[k].label, aggregate(rows)[k], definitions[k].info])] : [['ID laporan', 'Tanggal', 'Akun', 'Channel', 'Studio', 'Cakupan', 'Host', 'ID Line', 'Rentang sumber', 'Durasi menit', 'Status', ...keys.map(k => definitions[k].label)], ...rows.map(r => [r.id, r.date, brand.account, r.channel, r.studio, scopeLabel(r.scope), r.host, r.idLine, r.sourceTimeLabel, r.durationMinutes, r.simulation ? 'Data simulasi' : r.scope === 'slot' ? 'Hasil slot belum tersedia' : 'Pratinjau sumber belum terverifikasi', ...keys.map(k => r.metrics?.[k] ?? null)])];
    }
    return '\ufeff' + [...metadata, ...body].map(row => row.map(csvCell).join(',')).join('\r\n');
}
