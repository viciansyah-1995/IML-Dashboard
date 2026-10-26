import {describe,it,expect} from 'vitest';
import {parseOperatorReport,resolveOperatorMaster} from '../lib/operator-report';
const text=(date:string,channel:string,time:string,id:string)=>`Tanggal: ${date}\nStudio: Studio 1\nChannel: ${channel}\nBrand: zwitsal\nID-Line:${id}\nHost: Nurul\nJam Start-End:${time}`;
describe('operator report text',()=>{
 it('reads the TikTok format without creating a source period',()=>{
  expect(parseOperatorReport(text('20 okt 2026','TikTok','15:00-18:00','261020-00001'))).toMatchObject({date:'2026-10-20',channel:'tiktok',start:'2026-10-20T15:00',end:'2026-10-20T18:00',idLine:'261020-00001'});
 });
 it('accepts the Shopee dot/space format and overnight slots',()=>{
  expect(parseOperatorReport(text('21okt 2026','shopee','7.00 10.00','261021-00014')).end).toBe('2026-10-21T10:00');
  expect(parseOperatorReport(text('21 okt 2026','shopee','23.00-02.00','261021-00014')).end).toBe('2026-10-22T02:00');
 });
 it('rejects invalid dates, times and conflicting ID dates',()=>{
  expect(()=>parseOperatorReport(text('31 feb 2026','TikTok','15:00-18:00','260231-00001'))).toThrow();
  expect(()=>parseOperatorReport(text('20 okt 2026','TikTok','25:00-18:00','261020-00001'))).toThrow();
  expect(()=>parseOperatorReport(text('20 okt 2026','TikTok','15:00-18:00','261021-00001'))).toThrow();
 });
 it('does not guess an inactive, missing, or ambiguous master',()=>{
  expect(resolveOperatorMaster([{id:'a',name:'HarvestIndonesia.id'}],'harvestindonesia.id')).toBe('a');
  expect(resolveOperatorMaster([{id:'a',name:'Nurul',active:false}],'Nurul')).toBe('');
  expect(resolveOperatorMaster([{id:'a',name:'Nurul'},{id:'b',name:'Nurul'}],'Nurul')).toBe('');
 });
});
