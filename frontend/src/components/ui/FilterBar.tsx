import {notify} from '../../lib/feedback';
import {relevantChoices,type FilterOption} from '../../lib/filter-options';
import {api,type Master} from '../../lib/api/client';
import {key} from '../../lib/query/client';
import {Field,Button} from './Interactions';
import {useState,useEffect} from 'react';
import {useLocation} from 'react-router-dom';
import {useQuery} from '@tanstack/react-query';
import {Filter,ChevronDown} from 'lucide-react';
import {useFilters,useMasters,useWorkspace} from '../../app/Shell';
import {Dialog} from './Primitives';
const labels:Record<string,string>={formula_score:'Skor V2.1',cross_platform_score:'Skor lintas platform',gpm:'GPM',conversion:'Conversion Rate',ctr:'CTR',watch:'Watch Time',ctor:'CTOR',follow:'Follow Rate',hook:'Hook Rate (proxy)',comment:'Comment Velocity',views:'Total Views',interactions:'Interaksi',views_per_session:'Views/sesi',sales:'Penjualan',sales_per_hour:'Penjualan/jam',slots:'Slot valid'};
export function FilterBar({global=false}:{global?:boolean}){
 const {f,update}=useFilters(global),masters=useMasters(),me=useWorkspace(),location=useLocation();
 const [draft,setDraft]=useState(f),[mobile,setMobile]=useState(false);
 const serialized=JSON.stringify(f);useEffect(()=>setDraft(JSON.parse(serialized)),[serialized]);
 const includeDrafts=location.pathname==='/admin/reports';
 const params=new URLSearchParams({date_from:draft.date_from,date_to:draft.date_to,channel:draft.channel,include_drafts:includeDrafts?'1':'0'});
 if(!global)params.set('brand_id',draft.brand_id);
 const query=params.toString();
 const options=useQuery({queryKey:key(me.user,global?'global':'private','filter-options',query),queryFn:({signal})=>api<{items:FilterOption[]}>(`${global?'global':'reports'}/filter-options?${query}`,{signal}),enabled:!!draft.date_from&&!!draft.date_to&&draft.date_from<=draft.date_to&&(global||!!draft.brand_id)});
 const items=options.data?.items||[],m=masters.data;
 const configuredBasis=draft.channel==='tiktok'?['Pesanan Teratribusi']:['Pesanan Dibuat','Pesanan Selesai'];
 const bases=relevantChoices(items.map(x=>x.basis),configuredBasis,draft.basis);
 const scopes=relevantChoices(items.map(x=>x.scope).filter(x=>x==='host_slot'||x==='room'),['host_slot','room'],draft.scope);
 const selected=items.filter(x=>x.scope===draft.scope&&x.basis===draft.basis);
 const definitions=relevantChoices(selected.map(x=>x.definition_version),m?.mappings.filter(x=>x.channel_id===draft.channel).map(x=>x.id)||[],draft.definition_version);
 const withinDefinition=selected.filter(x=>!draft.definition_version||x.definition_version===draft.definition_version);
 const studios=relevantChoices(withinDefinition.map(x=>x.studio_id),m?.studios.map(x=>x.id)||[],draft.studio_id);
 const hosts=relevantChoices(withinDefinition.filter(x=>!draft.studio_id||x.studio_id===draft.studio_id).map(x=>x.host_id),m?.hosts.map(x=>x.id)||[],draft.host_id);
 const name=(rows:Master[]|undefined,id:string)=>rows?.find(x=>x.id===id)?.name||id;
 const controls=<form className="filter-form" onSubmit={e=>{
  e.preventDefault();if(draft.date_to<draft.date_from){notify('Tanggal akhir harus sama atau setelah tanggal awal. Periksa kembali periode laporan.','error');return}
  if((new Date(draft.date_to).getTime()-new Date(draft.date_from).getTime())/86400000>92){notify('Rentang laporan maksimal 93 hari. Persempit periode sebelum menerapkan filter.','error');return}
  update({...draft,host_id:draft.scope==='room'?'':draft.host_id,page:'1'});setMobile(false);notify('Filter diterapkan. Data dimuat sesuai periode dan cakupan yang dipilih.');
 }}>
 <Field>Dari<input type="date" required value={draft.date_from} onChange={e=>setDraft({...draft,date_from:e.target.value})}/></Field>
 <Field>Sampai<input type="date" required value={draft.date_to} onChange={e=>setDraft({...draft,date_to:e.target.value})}/></Field>
 <Field>Channel<select value={draft.channel} onChange={e=>setDraft({...draft,channel:e.target.value,definition_version:'',studio_id:'',host_id:'',basis:e.target.value==='shopee'?'Pesanan Dibuat':'Pesanan Teratribusi'})} required><option value="tiktok">TikTok</option><option value="shopee">Shopee</option></select></Field>
 {bases.length>1&&<Field>Basis pesanan<select value={draft.basis} onChange={e=>setDraft({...draft,basis:e.target.value,definition_version:''})} required>{bases.map(v=><option key={v} value={v}>{v}</option>)}</select></Field>}
 {definitions.length>1&&<Field>Versi definisi<select value={draft.definition_version} onChange={e=>setDraft({...draft,definition_version:e.target.value})}><option value="">Semua definisi yang kompatibel</option>{definitions.map(id=><option key={id} value={id}>{m?.mappings.find(x=>x.id===id)?.label||id}</option>)}</select></Field>}
 {global?<Field>Urutkan menurut<select value={draft.metric} onChange={e=>setDraft({...draft,metric:e.target.value})}>{Object.entries(labels).map(([id,label])=><option key={id} value={id}>{label}</option>)}</select></Field>:<>
 {scopes.length>1&&<Field>Cakupan<select value={draft.scope} onChange={e=>setDraft({...draft,scope:e.target.value,host_id:''})}>{scopes.map(v=><option key={v} value={v}>{v==='room'?'Sepanjang LIVE':'Per slot host'}</option>)}</select></Field>}
 {studios.length>1&&<Field>Studio<select value={draft.studio_id} onChange={e=>setDraft({...draft,studio_id:e.target.value,host_id:''})}><option value="">Semua studio</option>{studios.map(id=><option key={id} value={id}>{name(m?.studios,id)}</option>)}</select></Field>}
 {draft.scope==='host_slot'&&hosts.length>1&&<Field>Host<select value={draft.host_id} onChange={e=>setDraft({...draft,host_id:e.target.value})}><option value="">Semua host</option>{hosts.map(id=><option key={id} value={id}>{name(m?.hosts,id)}</option>)}</select></Field>}
 </>}
 <Button className="primary" type="submit">Terapkan</Button></form>;
 return <><div className="card filter-desktop">{controls}</div><Button className="filter-mobile" onClick={()=>setMobile(true)}><Filter size={17}/>Filter laporan<ChevronDown size={17}/></Button><div className="filter-caption"><span>{f.date_from} – {f.date_to}</span><span>{f.channel} · {f.basis} · IDR</span><span>{global?'Global · tidak dibatasi brand':f.scope==='room'?'Sepanjang LIVE · terpisah dari slot':'Per slot host'}</span></div>{options.isError&&<p className="note" role="status">Pilihan filter belum dapat dimuat. Seluruh pilihan konfigurasi tetap tersedia.</p>}{mobile&&<Dialog title="Filter laporan" drawer onClose={()=>setMobile(false)}>{controls}</Dialog>}</>;
}
export {labels as metricLabels};
