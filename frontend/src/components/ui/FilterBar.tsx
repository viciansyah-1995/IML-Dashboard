import {notify} from '../../lib/feedback';
import {relevantChoices,availableFilters,type FilterOption} from '../../lib/filter-options';
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
export function FilterBar({global=false,optional=false}:{global?:boolean;optional?:boolean}){
 const {f,update,reset}=useFilters(global),masters=useMasters(),me=useWorkspace(),location=useLocation();
 const [draft,setDraft]=useState(f),[mobile,setMobile]=useState(false);
 const serialized=JSON.stringify(f);useEffect(()=>setDraft(JSON.parse(serialized)),[serialized]);
 const includeDrafts=location.pathname==='/admin/reports';
 const params=new URLSearchParams({date_from:draft.date_from,date_to:draft.date_to,channel:draft.channel,include_drafts:includeDrafts?'1':'0'});
 if(!global)params.set('brand_id',draft.brand_id);
 if(optional)params.set('published_all','1');
 const query=params.toString();
 const options=useQuery({queryKey:key(me.user,global?'global':'private','filter-options',query),queryFn:({signal})=>api<{items:FilterOption[]}>(`${global?'global':'reports'}/filter-options?${query}`,{signal}),enabled:(optional||!!draft.date_from&&!!draft.date_to&&draft.date_from<=draft.date_to)&&(global||!!draft.brand_id)});
 const items=options.data?.items||[],m=masters.data;
 const known=options.isSuccess,selection=known?availableFilters(draft,items,global,optional):draft;
 const configuredBasis=selection.channel==='tiktok'?['Pesanan Teratribusi']:selection.channel==='shopee'?['Pesanan Dibuat','Pesanan Selesai']:['Pesanan Teratribusi','Pesanan Dibuat','Pesanan Selesai'];
 const choice=(actual:(string|undefined)[],fallback:string[],selected:string)=>known?[...new Set(actual.filter((value):value is string=>!!value))]:relevantChoices(actual,fallback,selected);
 const scopes=choice(items.map(x=>x.scope).filter(x=>x==='host_slot'||x==='room'),['host_slot','room'],selection.scope);
 const scoped=items.filter(x=>(!selection.channel||!x.channel||x.channel===selection.channel)&&(!selection.scope||x.scope===selection.scope));
 const bases=choice(scoped.map(x=>x.basis),configuredBasis,selection.basis);
 const selected=scoped.filter(x=>!selection.basis||x.basis===selection.basis);
 const definitions=choice(selected.map(x=>x.definition_version),m?.mappings.filter(x=>!selection.channel||x.channel_id===selection.channel).map(x=>x.id)||[],selection.definition_version);
 const withinDefinition=selected.filter(x=>!selection.definition_version||x.definition_version===selection.definition_version);
 const studios=choice(withinDefinition.map(x=>x.studio_id),m?.studios.map(x=>x.id)||[],selection.studio_id);
 const hosts=choice(withinDefinition.filter(x=>!selection.studio_id||x.studio_id===selection.studio_id).map(x=>x.host_id),m?.hosts.map(x=>x.id)||[],selection.host_id);
 const changed=JSON.stringify(selection)!==serialized;
 const name=(rows:Master[]|undefined,id:string)=>rows?.find(x=>x.id===id)?.name||id;
 const controls=<form className="filter-form" onSubmit={e=>{
  e.preventDefault();if(selection.date_to&&selection.date_from&&selection.date_to<selection.date_from){notify('Tanggal akhir harus sama atau setelah tanggal awal. Periksa kembali periode laporan.','error');return}
  if(!optional&&(new Date(selection.date_to).getTime()-new Date(selection.date_from).getTime())/86400000>92){notify('Rentang laporan maksimal 93 hari. Persempit periode sebelum menerapkan filter.','error');return}
  if(optional)update({...selection,host_id:selection.scope==='room'?'':selection.host_id,page:'1'},true);else update({...selection,host_id:selection.scope==='room'?'':selection.host_id,page:'1'});setMobile(false);notify('Filter diterapkan. Data dimuat sesuai periode dan cakupan yang dipilih.');
 }}>
 <Field>Dari<input type="date" required={!optional} value={selection.date_from} onChange={e=>setDraft({...selection,date_from:e.target.value})}/></Field>
 <Field>Sampai<input type="date" required={!optional} value={selection.date_to} onChange={e=>setDraft({...selection,date_to:e.target.value})}/></Field>
 <Field>Channel<select value={selection.channel} onChange={e=>setDraft({...selection,channel:e.target.value,definition_version:'',studio_id:'',host_id:'',basis:optional?'':e.target.value==='shopee'?'Pesanan Dibuat':'Pesanan Teratribusi'})} required={!optional}>{optional&&<option value="">Semua channel</option>}<option value="tiktok">TikTok</option><option value="shopee">Shopee</option></select></Field>
 {(optional||bases.length>0)&&<Field>Basis pesanan<select value={selection.basis} onChange={e=>setDraft({...selection,basis:e.target.value,definition_version:'',studio_id:'',host_id:''})} required={!optional}>{optional&&<option value="">Semua basis pesanan</option>}{bases.map(v=><option key={v} value={v}>{v}</option>)}</select></Field>}
 {(optional||definitions.length>1)&&<Field>Versi definisi<select value={selection.definition_version} onChange={e=>setDraft({...selection,definition_version:e.target.value,studio_id:'',host_id:''})}><option value="">Semua definisi yang kompatibel</option>{definitions.map(id=><option key={id} value={id}>{m?.mappings.find(x=>x.id===id)?.label||id}</option>)}</select></Field>}
 {global?<Field>Urutkan menurut<select value={selection.metric} onChange={e=>setDraft({...selection,metric:e.target.value})}>{Object.entries(labels).map(([id,label])=><option key={id} value={id}>{label}</option>)}</select></Field>:<>
 {(optional||scopes.length>0)&&<Field>Cakupan<select value={selection.scope} onChange={e=>setDraft({...selection,scope:e.target.value,definition_version:'',studio_id:'',host_id:''})}>{optional&&<option value="">Semua cakupan</option>}{scopes.map(v=><option key={v} value={v}>{v==='room'?'Sepanjang LIVE':'Per slot host'}</option>)}</select></Field>}
 {(optional||studios.length>1)&&<Field>Studio<select value={selection.studio_id} onChange={e=>setDraft({...selection,studio_id:e.target.value,host_id:''})}><option value="">Semua studio</option>{studios.map(id=><option key={id} value={id}>{name(m?.studios,id)}</option>)}</select></Field>}
 {selection.scope!=='room'&&(optional||hosts.length>1)&&<Field>Host<select value={selection.host_id} onChange={e=>setDraft({...selection,host_id:e.target.value})}><option value="">Semua host</option>{hosts.map(id=><option key={id} value={id}>{name(m?.hosts,id)}</option>)}</select></Field>}
 </>}
 <Button className="primary" type="submit" disabled={options.isFetching}>Terapkan</Button>{optional&&<Button type="button" onClick={()=>{reset();setMobile(false)}}>Reset filter</Button>}</form>;
 return <><div className="card filter-desktop">{controls}</div><Button className="filter-mobile" onClick={()=>setMobile(true)}><Filter size={17}/>Filter laporan<ChevronDown size={17}/></Button><div className="filter-caption"><span>{f.date_from||'Semua tanggal'}{f.date_to?' – '+f.date_to:''}</span><span>{f.channel||'Semua channel'} · {f.basis||'Semua basis'} · IDR</span><span>{global?'Global · tidak dibatasi brand':!f.scope?'Semua cakupan':f.scope==='room'?'Sepanjang LIVE · terpisah dari slot':'Per slot host'}</span></div>{changed&&<p className="note" role="status">Pilihan filter berubah. Klik Terapkan untuk memperbarui hasil.</p>}{known&&!items.length&&<p className="note" role="status">{includeDrafts?'Belum ada laporan':'Belum ada laporan yang dipublikasikan'} untuk channel dan periode ini. Coba channel atau periode lain.</p>}{!global&&selection.scope==='room'&&<p className="field-hint">Sepanjang LIVE memakai tanggal mulai sumber (GMT+7). Hasil ditampilkan per LIVE tanpa filter host.</p>}{options.isError&&<p className="note" role="status">Pilihan filter belum dapat dimuat. Seluruh pilihan konfigurasi tetap tersedia.</p>}{mobile&&<Dialog title="Filter laporan" drawer onClose={()=>setMobile(false)}>{controls}</Dialog>}</>;
}
export {labels as metricLabels};
