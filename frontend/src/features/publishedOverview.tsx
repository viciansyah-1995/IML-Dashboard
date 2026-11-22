import {useQuery} from '@tanstack/react-query';
import {useFilters,useWorkspace,PageHeader} from '../app/Shell';
import {api,num,money,type Summary,type Report} from '../lib/api/client';
import {publishedQuery} from '../lib/api/filters';
import {key} from '../lib/query/client';
import {ActionLink,Pagination} from '../components/ui/Interactions';
import {Loading,ErrorState,MetricCard,EmptyState,StatusBadge} from '../components/ui/Primitives';

type Group={channel:string;basis:string;scope:string;definition_version:string;currency:string;date_from:string;date_to:string;count:number;summary:Summary};
const scopeName=(scope:string)=>scope==='room'?'Sepanjang LIVE':scope==='snapshot'?'Snapshot':'Per slot host';
export function PublishedOverview(){
 const me=useWorkspace(),{f,update}=useFilters();
 const overview=useQuery({queryKey:key(me.user,'private','published-overview',f.brand_id),queryFn:({signal})=>api<{groups:Group[];total:number}>(`reports/overview?brand_id=${encodeURIComponent(f.brand_id)}`,{signal})});
 const query=publishedQuery(f);
 const list=useQuery({queryKey:key(me.user,'private','published-overview-list',query),queryFn:({signal})=>api<{items:Report[];total:number;page:number;page_size:number}>(`sessions?${query}`,{signal})});
 return <><PageHeader title="Ringkasan Livestream" subtitle="Seluruh laporan yang sudah dipublikasikan untuk brand aktif. Gunakan Filter laporan untuk melihat data tertentu."/>
 {overview.isPending?<Loading/>:overview.error?<ErrorState error={overview.error} retry={()=>overview.refetch()}/>:<>
 <div className="kpi-grid"><MetricCard label="Laporan published" value={overview.data.total} description="Seluruh tanggal dan channel pada brand aktif."/><MetricCard label="Kelompok sumber" value={overview.data.groups.length} description="Metrik dipisahkan menurut channel, basis pesanan, cakupan, dan versi definisi."/></div>
 {overview.data.groups.length===0&&<section className="card"><EmptyState title="Belum ada laporan published" message="Laporan akan tampil setelah admin menyelesaikan review dan publikasi."/></section>}
 {overview.data.groups.map(g=><section className="card" key={[g.channel,g.basis,g.scope,g.definition_version,g.currency].join('|')}><h2>{g.channel==='tiktok'?'TikTok':'Shopee'} · {scopeName(g.scope)}</h2><p>{g.basis} · {g.definition_version} · {g.date_from} – {g.date_to}</p><div className="kpi-grid"><MetricCard label="Laporan published" value={g.count}/><MetricCard label="Views sumber" value={g.summary.metrics.views} coverage={g.summary.coverage.views}/><MetricCard label="Interaksi sumber" value={g.summary.metrics.interactions} coverage={g.summary.coverage.interactions}/><MetricCard label="GMV sumber" value={g.summary.metrics.sales} unit="IDR" coverage={g.summary.coverage.sales}/></div></section>)}
 </>}
 <section className="card"><h2>Semua laporan published</h2>{list.isPending?<Loading/>:list.error?<ErrorState error={list.error} retry={()=>list.refetch()}/>:<><div className="table-wrap"><table><thead><tr><th>Tanggal / ID Line</th><th>Host / akun</th><th>Channel / cakupan</th><th>Views</th><th>Penjualan</th><th>Status</th><th>Detail</th></tr></thead><tbody>{list.data.items.map(r=><tr key={r.id}><td>{r.date}<small>{r.id_line}</small></td><td>{r.host}<small>{r.account}</small></td><td>{r.channel}<small>{scopeName(r.scope)} · {r.basis}</small></td><td>{num(r.metrics.views)}</td><td>{money(r.metrics.sales)}</td><td><StatusBadge status={r.status}/></td><td><ActionLink to={`/sessions/${r.id}`}>Lihat detail →</ActionLink></td></tr>)}</tbody></table></div><Pagination label="laporan published" total={list.data.total} page={list.data.page} pageSize={list.data.page_size} onChange={page=>update({...f,page:String(page)})}/></>}</section>
 </>;
}
