export type OperatorReport={channel:'tiktok'|'shopee';date:string;start:string;end:string;studio:string;brand:string;host:string;idLine:string};
const months=['jan','feb','mar','apr','mei','jun','jul','agu','sep','okt','nov','des'];
export const operatorName=(s:string)=>s.toLowerCase().replace(/[^a-z0-9]/g,'');
export function parseOperatorReport(text:string):OperatorReport {
 const fields:Record<string,string>={};for(const line of text.split(/\r?\n/)){const split=line.indexOf(':');if(split>=0)fields[operatorName(line.slice(0,split))]=line.slice(split+1).trim()}
 const raw=fields.tanggal||'',m=raw.toLowerCase().match(/^(\d{1,2})\s*([a-z]+|\d{1,2})[\s/-]+(\d{4})$/);
 if(!m)throw new Error('Tanggal harus lengkap, misalnya 20 okt 2026.');
 const month=/^\d+$/.test(m[2])?Number(m[2])-1:months.findIndex(v=>m[2].startsWith(v));
 const year=Number(m[3]),day=Number(m[1]),date=new Date(Date.UTC(year,month,day));
 if(month<0||month>11||date.getUTCFullYear()!==year||date.getUTCMonth()!==month||date.getUTCDate()!==day)throw new Error('Tanggal laporan tidak valid.');
 const channel=operatorName(fields.channel||'');if(channel!=='tiktok'&&channel!=='shopee')throw new Error('Channel harus TikTok atau Shopee.');
 const times=(fields.jamstartend||'').match(/^(\d{1,2})[.:](\d{2})\s*(?:[-–—]|\s)\s*(\d{1,2})[.:](\d{2})$/);
 if(!times||Number(times[1])>23||Number(times[3])>23||Number(times[2])>59||Number(times[4])>59)throw new Error('Jam Start-End harus valid, misalnya 15:00-18:00 atau 7.00 10.00.');
 const a=Number(times[1])*60+Number(times[2]),b=Number(times[3])*60+Number(times[4]);if(a===b)throw new Error('Jam mulai dan akhir tidak boleh sama.');
 const end=new Date(date);if(b<a)end.setUTCDate(end.getUTCDate()+1);
 const stamp=(d:Date,h:string,min:string)=>`${d.toISOString().slice(0,10)}T${h.padStart(2,'0')}:${min}`;
 for(const key of ['studio','brand','host','idline'])if(!fields[key])throw new Error(`${key} belum tercantum pada teks operator.`);
 if(!/^\d{6}-\d{5}$/.test(fields.idline))throw new Error('ID-Line harus berformat YYMMDD-00001.');
 const businessDate=date.toISOString().slice(0,10);if(fields.idline.slice(0,6)!==businessDate.slice(2).replaceAll('-',''))throw new Error('Tanggal pada ID-Line berbeda dari tanggal laporan.');
 return {channel,date:businessDate,start:stamp(date,times[1],times[2]),end:stamp(end,times[3],times[4]),studio:fields.studio,brand:fields.brand,host:fields.host,idLine:fields.idline};
}
export function resolveOperatorMaster<T extends {id:string;name?:string;active?:boolean}>(items:T[],name:string):string {
 const matches=items.filter(x=>x.active!==false&&operatorName(x.name||'')===operatorName(name));
 return matches.length===1?matches[0].id:'';
}
