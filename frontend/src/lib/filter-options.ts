import type {Filters} from './api/filters';
export type FilterOption={channel?:string;scope:string;basis:string;definition_version:string;studio_id?:string;host_id?:string};
// Keep explicit URL selections reachable, even when no data currently matches.
export function relevantChoices(actual:(string|undefined)[],fallback:string[],selected:string){
 const available=[...new Set(actual.filter((v):v is string=>!!v))];
 const values=available.length?available:[...new Set(fallback)];
 return selected&&!values.includes(selected)?[...values,selected]:values;
}

// Resolve dependent choices against reports in the selected channel and period.
export function availableFilters(filters:Filters,items:FilterOption[],global=false,optional=false):Filters{
 const next={...filters};
 if(optional){
  const channelRows=items.filter(row=>!next.channel||!row.channel||row.channel===next.channel);
  if(next.scope&&!channelRows.some(row=>row.scope===next.scope))next.scope='';
  const scoped=channelRows.filter(row=>!next.scope||row.scope===next.scope);
  if(next.basis&&!scoped.some(row=>row.basis===next.basis))next.basis='';
  const based=scoped.filter(row=>!next.basis||row.basis===next.basis);
  if(!based.some(row=>row.definition_version===next.definition_version))next.definition_version='';
  const defined=based.filter(row=>!next.definition_version||row.definition_version===next.definition_version);
  if(!defined.some(row=>row.studio_id===next.studio_id))next.studio_id='';
  if(next.scope==='room'||!defined.some(row=>(!next.studio_id||row.studio_id===next.studio_id)&&row.host_id===next.host_id))next.host_id='';
  return next;
 }
 if(!items.length)return {...next,definition_version:'',studio_id:'',host_id:''};
 const scopeRows=items.filter(row=>global?row.scope==='host_slot':row.scope==='host_slot'||row.scope==='room');
 if(!scopeRows.length)return {...next,definition_version:'',studio_id:'',host_id:''};
 if(!scopeRows.some(row=>row.scope===next.scope))next.scope=scopeRows[0]?.scope||'host_slot';
 const scoped=scopeRows.filter(row=>row.scope===next.scope);
 if(!scoped.some(row=>row.basis===next.basis))next.basis=scoped[0]?.basis||next.basis;
 const based=scoped.filter(row=>row.basis===next.basis);
 if(!based.some(row=>row.definition_version===next.definition_version))next.definition_version='';
 const defined=based.filter(row=>!next.definition_version||row.definition_version===next.definition_version);
 if(!defined.some(row=>row.studio_id===next.studio_id))next.studio_id='';
 if(next.scope==='room'||global||!defined.some(row=>(!next.studio_id||row.studio_id===next.studio_id)&&row.host_id===next.host_id))next.host_id='';
 return next;
}
