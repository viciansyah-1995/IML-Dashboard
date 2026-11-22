export const timeZoneLabel='GMT+7';
export function formatDateTime(value?:string|null){
 if(!value)return 'Belum diketahui';
 const date=new Date(value);
 if(Number.isNaN(date.getTime()))return 'Belum diketahui';
 return new Intl.DateTimeFormat('id-ID',{timeZone:'Asia/Jakarta',day:'2-digit',month:'2-digit',year:'numeric',hour:'2-digit',minute:'2-digit',second:'2-digit',hourCycle:'h23'}).format(date)+' '+timeZoneLabel;
}
