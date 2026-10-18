import {notify,actionCopy,errorCopy} from '../feedback';
export class ApiError extends Error {constructor(public status:number, public code:string, message:string){super(message)}}
let csrf='';
export const setCSRF=(token:string)=>{csrf=token};
export async function api<T>(path:string,options:RequestInit={}):Promise<T>{
 const headers=new Headers(options.headers),mutation=!!options.method&&options.method!=='GET',copy=actionCopy(path,options.body);
 if(options.body&&!(options.body instanceof FormData))headers.set('Content-Type','application/json');if(mutation)headers.set('X-CSRF-Token',csrf);
 const notice=mutation?notify(copy[0],'info',undefined,true):'';
 try{const res=await fetch(`/api/${path}`,{...options,headers,credentials:'same-origin'});if(!res.ok){const data=await res.json().catch(()=>({code:'NETWORK',message:'Permintaan gagal'}));throw new ApiError(res.status,data.code,errorCopy(data.code,data.message))}
 const data=await res.json() as T;if(mutation)notify(copy[1],'success',notice);return data;
 }catch(error){if(error instanceof DOMException&&error.name==='AbortError')throw error;const issue=error instanceof ApiError?error:new ApiError(0,'NETWORK',errorCopy('NETWORK',''));if(mutation)notify(issue.message,'error',notice);throw issue}
}
export const post=<T>(path:string,body:unknown)=>api<T>(path,{method:'POST',body:JSON.stringify(body)});
export type User={id:string;name:string;email:string;role:'superadmin'|'admin'|'client';permissions:string[];access_version:number;csrf_token:string;photo_url?:string};
export type Brand={id:string;name:string;is_demo:boolean};
export type Me={user:User;brands:Brand[];demo_mode:boolean};
export type Master={photo_url?:string;full_name?:string;email?:string;birth_date?:string|null;gender?:string;joined_on?:string|null;city?:string;languages?:string;specialties?:string;brand_ids?:string[];profile_note?:string;global_visible?:boolean;id:string;name?:string;label?:string;active?:boolean;brand_id?:string;channel_id?:string;is_demo?:boolean;confirmed?:boolean};
export type Masters={brands:Master[];accounts:Master[];studios:Master[];hosts:Master[];channels:Master[];mappings:Master[]};
export type Metrics=Record<string,string|null>;
export type Coverage={eligible_slots:number;total_slots:number;missing_slots:number;unit:string;basis:string;definition_version:string;limitations:string[]};
export type Daily={date:string;views:string|null;interactions:string|null;sales:string|null;slots:number;hours:string};
export type FormulaResult={version:string;score:string|null;observed_score:string|null;metric_scores:Metrics;values:Metrics;benchmarks:Metrics;benchmark_counts:Record<string,number>;benchmark_scope:string;window_from:string;window_to:string;eligible_sessions:number;total_sessions:number;history_sessions:number;shrunk:boolean;flags:string[];missing_inputs:string[]};
export type Summary={formula?:FormulaResult;metrics:Metrics;coverage:Record<string,Coverage>;slots:number;room_count:number;hours:string;hosts:number;updated_at:string|null;revision:string;trend:Daily[]};
export type MetricValue={key:string;value:string|null;raw_label:string;raw_value:string;unit:string;evidence_text:string;attachment_id:string|null;page_number:number|null;rounded_display:boolean;definition_status:string;validation_flags:string[]};
export type Report={duplicate_evidence:boolean;id:string;source_id:string;slot_id:string;room_id:string;brand_id:string;brand:string;id_line:string;account:string;channel:string;studio:string;host_id:string;host:string;date:string;start_at:string;end_at:string;period_start:string|null;period_end:string|null;scope:string;basis:string;currency:string;definition_version:string;revision:number;number:number;status:string;active:boolean;is_demo:boolean;duration_seconds:number;completeness_status:string;validity_status:string;compatibility_status:string;reason:string;published_at:string|null;metrics:Metrics;values:MetricValue[]};
export type Definition={key:string;label:string;unit:string;kind:string;required:boolean;canonical_confirmed:boolean};
export type Evidence={id:string;name:string;content_type:string;bytes:number};
export type Detail={report:Report;definitions:Definition[];attachments:Evidence[];history:{id:string;number:number;status:string;active:boolean;reason:string;published_at:string|null}[];blockers:string[]};
export type CrossPlatform={platform_scores:Metrics;base_score:string|null;consistency:string|null;bonus:string;score:string|null;percentile:string|null;z_score:string|null;pool_size:number;flags:string[]};
export type Host={cross_platform?:CrossPlatform;photo_url?:string;host_id:string;name:string;active:boolean;profile_note:string;rank:number|null;sample_limited:boolean;summary:Summary;sales_contribution:string|null};
export const can=(u:User,p:string)=>u.role==='superadmin'||u.permissions.includes(p)||(u.role==='client'&&p==='read_global_host_summary');
export const num=(v:string|number|null|undefined)=>{if(v==null)return 'N/A';if(typeof v==='number')return new Intl.NumberFormat('id-ID',{maximumFractionDigits:2}).format(v);if(!/^-?\d+(\.\d+)?$/.test(v))return 'N/A';const [whole,fraction='']=v.split('.');let scaled=BigInt(whole.replace('-',''))*100n+BigInt((fraction+'00').slice(0,2));if(Number(fraction[2]||'0')>=5)scaled+=1n;const digits=(scaled/100n).toLocaleString('id-ID');const decimals=String(scaled%100n).padStart(2,'0').replace(/0+$/,'');return (whole.startsWith('-')?'-':'')+digits+(decimals?','+decimals:'')};
export const money=(v:string|null|undefined)=>v==null?'N/A':`Rp${num(v)}`;
