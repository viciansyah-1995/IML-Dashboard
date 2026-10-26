export type FilterOption={scope:string;basis:string;definition_version:string;studio_id?:string;host_id?:string};
// Keep explicit URL selections reachable, even when no data currently matches.
export function relevantChoices(actual:(string|undefined)[],fallback:string[],selected:string){
 const available=[...new Set(actual.filter((v):v is string=>!!v))];
 const values=available.length?available:[...new Set(fallback)];
 return selected&&!values.includes(selected)?[...values,selected]:values;
}
