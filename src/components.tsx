import { DatasetContext } from './context.ts';
import { useEffect, useRef, useContext, type ReactNode } from 'react';
import { Info, X, Inbox, ArrowUpRight } from 'lucide-react';
import { definitions, metricValue, dateLabel, datasetLabel, type Dataset, type Metrics, type Report, number } from './data.ts';
export function Hint({ text }: {
    text: string;
}) { const dataset = useContext(DatasetContext); const description = `${text} · ${datasetLabel(dataset)}`; return <span className="hint" tabIndex={0} aria-label={description}><Info size={14}/><span role="tooltip">{description}</span></span>; }
export function Metric({ name, value, note, icon }: {
    name: keyof Metrics;
    value: number | null;
    note?: string;
    icon?: ReactNode;
}) { return <div className="metric card"><div className="metric-label"><span>{definitions[name].label}</span><Hint text={definitions[name].info}/>{icon && <span className="metric-icon">{icon}</span>}</div><strong>{metricValue(name, value)}</strong><small>{value == null ? 'Belum tersedia' : note || 'Sesuai cakupan laporan'}</small></div>; }
export function Empty({ title, description, action, onAction }: {
    title: string;
    description?: string;
    action?: string;
    onAction?: () => void;
}) { return <div className="empty"><div className="empty-icon"><Inbox size={25}/></div><h3>{title}</h3>{description && <p>{description}</p>}{action && <button className="button secondary" onClick={onAction}>{action}<ArrowUpRight size={15}/></button>}</div>; }
export function Overlay({ title, children, onClose, drawer = false }: {
    title: string;
    children: ReactNode;
    onClose: () => void;
    drawer?: boolean;
}) {
    const ref = useRef<HTMLDivElement>(null), close = useRef(onClose);
    close.current = onClose;
    useEffect(() => { const prev = document.activeElement as HTMLElement; const bodyOverflow = document.body.style.overflow; document.body.style.overflow = 'hidden'; ref.current?.querySelector<HTMLElement>('button,input,select')?.focus(); const listener = (e: KeyboardEvent) => { if (e.key === 'Escape')
        close.current(); if (e.key === 'Tab') {
        const nodes = Array.from(ref.current?.querySelectorAll<HTMLElement>('button:not(:disabled),input:not(:disabled),select:not(:disabled),a[href],[tabindex="0"]') ?? []);
        const first = nodes[0], last = nodes[nodes.length - 1];
        if (e.shiftKey && document.activeElement === first) {
            e.preventDefault();
            last?.focus();
        }
        else if (!e.shiftKey && document.activeElement === last) {
            e.preventDefault();
            first?.focus();
        }
    } }; document.addEventListener('keydown', listener); return () => { document.removeEventListener('keydown', listener); document.body.style.overflow = bodyOverflow; prev?.focus(); }; }, []);
    return <div className={'overlay ' + (drawer ? 'drawer-overlay' : '')} onMouseDown={e => { if (e.target === e.currentTarget)
        onClose(); }}><div ref={ref} className={drawer ? 'drawer' : 'modal'} role="dialog" aria-modal="true" aria-label={title}><div className="modal-head"><h2>{title}</h2><button className="icon-button" aria-label="Tutup" onClick={onClose}><X size={20}/></button></div>{children}</div></div>;
}
export function TrendChart({ rows, metric, dataset }: {
    rows: Report[];
    metric: 'sourceGmv' | 'attributedOrders' | 'productClicks';
    dataset: Dataset;
}) {
    const daily = [...new Set(rows.map(r => r.date))].sort().map(date => ({ date, value: rows.filter(r => r.date === date).reduce((n, r) => n + (r.metrics?.[metric] ?? 0), 0) }));
    const max = Math.max(...daily.map(d => d.value), 1) * 1.2, w = 700, h = 220, p = 55, py = 20, bottom = 180;
    if (!daily.length)
        return <Empty title="Belum ada data grafik" description="Hasil belum tersedia untuk cakupan dan filter ini."/>;
    const points = daily.map((d, i) => ({ x: daily.length === 1 ? w / 2 : p + i * (w - p - 20) / (daily.length - 1), y: bottom - d.value / max * (bottom - py), ...d }));
    return <><div className="chart-wrap"><svg viewBox={`0 0 ${w} ${h}`} role="img" aria-label={`${definitions[metric].label}, ${datasetLabel(dataset)}. ${daily.map(d => `${dateLabel(d.date)}: ${metricValue(metric, d.value)}`).join('; ')}`}>
 {[0, 1, 2, 3].map(i => <g key={i}><line x1={p} x2={w - 15} y1={py + i * (bottom - py) / 3} y2={py + i * (bottom - py) / 3} stroke="#e8edef" strokeDasharray="3 5"/><text x={p - 10} y={py + i * (bottom - py) / 3 + 4} textAnchor="end" fill="#84909b" fontSize="11">{metric === 'sourceGmv' ? `${number(max * (3 - i) / 3 / 1e6)} jt` : number(Math.round(max * (3 - i) / 3))}</text></g>)}
 {metric === 'sourceGmv' && points.length > 1 && <><path d={`M${points[0].x},${bottom} ${points.map(pt => `L${pt.x},${pt.y}`).join(' ')} L${points[points.length - 1].x},${bottom} Z`} fill="#e9f4f1"/><polyline points={points.map(pt => `${pt.x},${pt.y}`).join(' ')} fill="none" stroke="#0f766e" strokeWidth="2.5"/></>}
 {points.map((pt, i) => <g key={pt.date} tabIndex={0}><title>{`${definitions[metric].label} · ${dateLabel(pt.date)} · ${metricValue(metric, pt.value)} · ${datasetLabel(dataset)}`}</title>{metric === 'sourceGmv' ? <><circle cx={pt.x} cy={pt.y} r="7" fill="#fff" stroke="#0f766e" strokeWidth="3"/>{points.length === 1 && <text x={pt.x} y={pt.y - 17} textAnchor="middle" fill="#0f766e" fontWeight="600" fontSize="13">{metricValue(metric, pt.value)}</text>}</> : <rect x={pt.x - 9} y={pt.y} width="18" height={bottom - pt.y} rx="4" fill="#5271a3"/>}{(i === 0 || i === points.length - 1 || i % 3 === 0) && <text x={pt.x} y={bottom + 25} fill="#84909b" textAnchor="middle" fontSize="11">{pt.date.slice(8)} Sep</text>}</g>)}
 </svg></div><div className="chart-foot"><span><i className="dot"/> {definitions[metric].label} <span className="muted">· {datasetLabel(dataset)}</span></span><span>{daily.length < 2 ? 'Butuh minimal 2 tanggal untuk melihat tren' : `${daily.length} tanggal simulasi · tanpa pembanding`}</span></div></>;
}
export function Bars({ items }: {
    items: {
        label: string;
        value: number;
        display: string;
        sub?: string;
    }[];
}) { const max = Math.max(...items.map(i => i.value), 1); return <div className="bars">{items.map(i => <div className="bar-row" key={i.label}><div className="bar-top"><span>{i.label}</span><strong>{i.display}</strong></div><div className="bar-track"><div style={{ width: `${Math.max(i.value / max * 100, i.value > 0 ? 1 : 0)}%` }}/></div>{i.sub && <small>{i.sub}</small>}</div>)}</div>; }
