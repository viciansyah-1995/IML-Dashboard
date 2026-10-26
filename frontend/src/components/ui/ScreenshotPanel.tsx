import {UnavailableValue,unavailableLabel} from './UnavailableValue';
import {money,num} from '../../lib/api/client';
const formulaNumber=(v:string|null,u:string)=>v==null?unavailableLabel:u==='IDR'?money(v):num(v)+(u==='percent'?'%':u==='minutes'?' menit':u==='per_minute'?' / menit':'');
import type {ScreenshotResult} from '../../lib/api/client';

const labels:Record<string,string>={buyers:'Pembeli',views:'Tayangan/Dilihat',product_clicks:'Klik produk absolut',new_followers:'Pengikut baru',impressions:'Impresi',orders:'Pesanan',add_to_cart:'Tambah ke keranjang',sales:'Penjualan/Perolehan GMV',estimated_sales:'Perkiraan GMV',comments:'Komentar',source_minutes:'Durasi sumber terverifikasi',watch_duration:'Durasi tonton'};
export function ScreenshotPanel({result}:{result?:ScreenshotResult}){
 if(!result)return null;
 return <section className="card formula-details" aria-label="Perhitungan screenshot"><h2>Perhitungan dari screenshot</h2><p>Nilai dihitung dari label sumber. Angka K/M dibulatkan; hasilnya dapat berbeda sedikit dari rasio platform. Bagian ini tidak memberi atribusi host atau mengganti skor V2.1.</p>{result.warnings.map(w=><p className="note" key={w}>{w}</p>)}<div className="table-wrap"><table><thead><tr><th>Metrik</th><th>Perhitungan</th><th>Hasil</th><th>Kelengkapan</th></tr></thead><tbody>{result.metrics.map(m=><tr key={m.key}><th scope="row">{m.label}</th><td>{m.formula}</td><td>{m.value==null?<UnavailableValue reason={m.missing.length?`Input belum tersedia atau belum valid: ${m.missing.map(k=>labels[k]||k).join(', ')}.`:'Belum ada input sumber yang lengkap dengan penyebut lebih dari nol.'}/>:formulaNumber(m.value,m.unit)}</td><td>{m.value==null?'Perlu dilengkapi atau diverifikasi':'Tersedia dari subset sumber'}</td></tr>)}</tbody></table></div></section>;
}
