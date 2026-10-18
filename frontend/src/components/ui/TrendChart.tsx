import {Pagination} from './Interactions';
import {useLocalPagination} from '../../lib/pagination';
import {useEffect,useRef} from 'react';
import {Chart,registerables} from 'chart.js';
import type {Daily} from '../../lib/api/client';
import {num} from '../../lib/api/client';
Chart.register(...registerables);
const categories={views:'views',interactions:'interactions',sales:'sales',slots:'activity'} as const;
const markers={views:'circle',interactions:'triangle',sales:'rectRot',slots:'rect'} as const;
export function TrendChart({title,rows,metric,from,to,bar=false}:{title:string;rows:Daily[];metric:keyof typeof categories;from:string;to:string;bar?:boolean}){
 const ref=useRef<HTMLCanvasElement>(null),days:string[]=[];
 const start=new Date(from+'T00:00:00Z'),end=new Date(to+'T00:00:00Z');
 for(let date=start;date<=end&&days.length<93;date=new Date(date.getTime()+86400000))days.push(date.toISOString().slice(0,10));
 const paging=useLocalPagination(days,from+'|'+to,10),values=days.map(d=>rows.find(r=>r.date===d)?.[metric]??null);
 useEffect(()=>{
  if(!ref.current)return;
  // Canvas needs resolved colors; CSS var() strings are not canvas colors.
  const style=getComputedStyle(document.documentElement),color=(name:string)=>style.getPropertyValue('--color-'+name).trim();
  const family=getComputedStyle(document.body).fontFamily;
  const category=categories[metric],stroke=color('chart-'+category),fill=color('chart-'+category+'-fill');
  const chart=new Chart(ref.current,{type:bar?'bar':'line',data:{labels:days.map(d=>d.slice(5)),datasets:[{label:title,data:values.map(v=>v==null?null:Number(v)),borderColor:stroke,backgroundColor:fill,pointBackgroundColor:stroke,pointBorderColor:stroke,pointStyle:markers[metric],pointRadius:values.filter(v=>v!=null).length===1?5:3,tension:0,spanGaps:false,fill:!bar,borderWidth:2}]},options:{responsive:true,maintainAspectRatio:false,animation:window.matchMedia('(prefers-reduced-motion: reduce)').matches?false:undefined,plugins:{legend:{display:false},tooltip:{bodyFont:{family,weight:400},titleFont:{family,weight:500},backgroundColor:color('surface'),titleColor:color('text'),bodyColor:color('text'),borderColor:color('border-control'),borderWidth:1,displayColors:true,callbacks:{label:ctx=>`${title}: ${num(ctx.raw as number)} ${metric==='sales'?'IDR':''}`}}},scales:{x:{grid:{display:false},border:{color:color('chart-grid')},ticks:{font:{family,weight:400},color:color('chart-label'),maxTicksLimit:7}},y:{beginAtZero:true,grid:{color:color('chart-grid')},border:{color:color('chart-grid')},ticks:{font:{family,weight:400},color:color('chart-label')}}}}});
  let alive=true;void document.fonts?.ready.then(()=>{if(alive){chart.resize();chart.update('none')}});
  return()=>{alive=false;chart.destroy()};
 },[rows,metric,from,to,title,bar]);
 return <section className="card chart" data-metric={metric}><h2><span className="chart-key" aria-hidden="true"/>{title}</h2><p>{metric==='sales'?'IDR sesuai basis sumber':metric==='slots'?'Slot unik final':'Bukan orang unik lintas sesi'}</p><div className="canvas-wrap"><canvas ref={ref} role="img" aria-label={`${title}, ${from} sampai ${to}`}/></div>{values.filter(v=>v!=null).length<2&&<small>Butuh minimal 2 tanggal untuk melihat tren</small>}<details><summary>Lihat data grafik</summary><div className="table-wrap"><table><thead><tr><th>Tanggal</th><th>{title}</th></tr></thead><tbody>{paging.items.map(d=><tr key={d}><td>{d}</td><td>{num(values[days.indexOf(d)])}</td></tr>)}</tbody></table></div><Pagination label="tanggal" total={paging.total} page={paging.page} pageSize={paging.pageSize} onChange={paging.onChange}/></details></section>;
}
