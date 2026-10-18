import {useEffect,useRef,useState} from 'react';
import {Check,ImagePlus,Trash2} from 'lucide-react';
import {Button,Field} from './Interactions';
import {notify} from '../../lib/feedback';
import type {Master} from '../../lib/api/client';

export function initials(name:string){return name.trim().split(/\s+/).filter(Boolean).slice(0,2).map(word=>Array.from(word)[0]).join('').toUpperCase()||'?'}
export function Avatar({name,src='',large=false}:{name:string;src?:string;large?:boolean}){
 const [failed,setFailed]=useState('');
 return <span className={`person-avatar${large?' person-avatar-large':''}`} aria-label={`Foto ${name}`}>
  {src&&failed!==src?<img src={src} alt={name} onError={()=>setFailed(src)}/>:<span aria-hidden="true">{initials(name)}</span>}
 </span>;
}
export const todayWIB=()=>new Intl.DateTimeFormat('en-CA',{timeZone:'Asia/Jakarta',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date());
export function PhotoPicker({name,src}:{name:string;src?:string}){
 const [file,setFile]=useState<File|null>(null),[preview,setPreview]=useState(''),[remove,setRemove]=useState(false),input=useRef<HTMLInputElement>(null);
 useEffect(()=>{if(!file){return}const url=URL.createObjectURL(file);setPreview(url);return()=>URL.revokeObjectURL(url)},[file]);
 const clear=()=>{setFile(null);setPreview('');setRemove(true);if(input.current)input.current.value=''};
 return <section className="photo-picker"><Avatar name={name} src={file?preview:remove?'':src} large/><div><Field>Foto profil (opsional)<input ref={input} type="file" name="photo" accept="image/png,image/jpeg" onChange={e=>{const next=e.target.files?.[0];if(!next)return;if(!['image/png','image/jpeg'].includes(next.type)||next.size>2*1024*1024){notify('Gunakan foto PNG atau JPEG, maksimal 2 MB.','error');e.target.value='';setFile(null);setPreview('');return}setFile(next);setRemove(false)}}/></Field><small>PNG/JPEG · maksimal 2 MB. Tanpa foto, inisial nama tampil otomatis.</small>{(file||src&&!remove)&&<Button onClick={clear}><Trash2 size={15}/>Hapus foto</Button>}<input type="hidden" name="remove_photo" value={remove?'true':'false'}/></div></section>;
}
export function AssignmentPicker({items,name='brands',label='Assignment brand',selected,onChange,required=false}:{items:Master[];name?:string;label?:string;selected:string[];onChange:(ids:string[])=>void;required?:boolean}){
 const all=items.length>0&&items.every(item=>selected.includes(item.id));
 return <fieldset className="assignment-picker"><legend><Check size={16}/>{label}{required&&<span className="required-mark" aria-hidden="true"> *</span>}</legend><div className="assignment-heading"><small>{selected.length} dari {items.length} dipilih</small><Button disabled={!items.length} disabledReason="Belum ada brand yang dapat diassign." onClick={()=>onChange(all?[]:items.map(item=>item.id))}><ImagePlus size={15}/>{all?'Batalkan semua':'Pilih semua'}</Button></div>{items.length?<div className="assignment-options">{items.map(item=><Field className="check" key={item.id}><input type="checkbox" name={name} value={item.id} checked={selected.includes(item.id)} onChange={e=>onChange(e.target.checked?[...selected,item.id]:selected.filter(id=>id!==item.id))}/>{item.name}{item.active===false?' (nonaktif)':''}</Field>)}</div>:<p>Belum ada assignment yang tersedia.</p>}</fieldset>;
}
export function profileBody(form:HTMLFormElement,metadata:Record<string,unknown>){const fields=new FormData(form),body=new FormData();body.set('metadata',JSON.stringify({...metadata,remove_photo:fields.get('remove_photo')==='true'}));const photo=fields.get('photo');if(photo instanceof File&&photo.size)body.set('photo',photo);return body}

export function StatusToggle({value,onChange,label,disabled=false,title='',name}:{value:boolean;onChange:(value:boolean)=>void;label:string;disabled?:boolean;title?:string;name?:string}){return <div className="user-status"><button type="button" className="status-switch" role="switch" aria-checked={value} aria-label={label} disabled={disabled} title={title} onClick={()=>onChange(!value)}><span/></button><span>{value?'Aktif':'Nonaktif'}</span>{name&&<input name={name} type="hidden" value={value?'on':'off'}/>}</div>}
