import {describe,it,expect,vi,beforeEach} from 'vitest';
import {render,screen,fireEvent,waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import {QueryClient,QueryClientProvider} from '@tanstack/react-query';
import {Upload,slotDefaults,validateFiles} from '../features/upload';
const {api,post,state}=vi.hoisted(()=>({api:vi.fn(),post:vi.fn(),state:{draft:null as any}}));
vi.mock('../lib/api/client',async importOriginal=>({...await importOriginal<typeof import('../lib/api/client')>(),api,post}));
vi.mock('../lib/query/client',()=>({invalidateReports:vi.fn()}));
vi.mock('../app/Shell',()=>({
 PageHeader:()=>null,
 useWorkspace:()=>({brands:[{id:'brand',name:'Brand'}],user:{id:'u',name:'Reviewer',role:'admin'},demo_mode:true}),
 useMasters:()=>({data:{accounts:[{id:'account',name:'Account',brand_id:'brand',channel_id:'tiktok',active:true}],studios:[{id:'studio',name:'Studio 1',active:true}],hosts:[{id:'host',name:'Dea',brand_ids:['brand'],active:true}]}}),
}));
function setup(url='/admin/upload'){const queryClient=new QueryClient({defaultOptions:{queries:{retry:false},mutations:{retry:false}}});render(<MemoryRouter initialEntries={[url]}><QueryClientProvider client={queryClient}><Upload/></QueryClientProvider></MemoryRouter>);return queryClient}
function fixture(){return {id:'draft',channel:'tiktok',status:'extracted',step:2,revision:2,metadata:{},corrections:[],attachments:[],definitions:[{key:'sales',label:'GMV',unit:'IDR',required:true}],result:{values:[{key:'sales',value:'2000',raw_value:'2000',raw_label:'GMV',evidence_text:'GMV 2000',validation_flags:[]}],flags:[],observation:{period_start:'2026-10-18T06:00:00+07:00',period_end:'2026-10-18T12:00:00+07:00'}},last_error:''}}
beforeEach(()=>{
 state.draft=fixture();api.mockReset();post.mockReset();
 api.mockImplementation(async(path:string)=>path.startsWith('upload-intakes?')?{items:[],total:0,page:1,page_size:10}:state.draft);
 post.mockImplementation(async(path:string,body:any)=>{
  if(path.endsWith('/check'))state.draft={...state.draft,step:3,revision:3};
  if(path.endsWith('/metadata'))state.draft={...state.draft,metadata:body.metadata,step:4,revision:4};
  if(path.endsWith('/submit'))state.draft={...state.draft,status:'submitted',report_id:'report'};
  if(path.endsWith('/corrections'))state.draft={...state.draft,corrections:body.corrections,revision:3};
  return {id:'draft'};
 });
});
describe('screenshot-first wizard',()=>{
 it('uses WIB and handles midnight rollover',()=>{expect(slotDefaults(new Date('2026-10-18T16:45:59Z'))).toEqual({start:'2026-10-18T23:45',end:'2026-10-19T01:45'})});
 it('starts with screenshots rather than operator metadata and rejects PDFs',async()=>{
  setup();expect(screen.getByLabelText('Screenshot laporan')).toBeInTheDocument();expect(screen.queryByLabelText('Host')).not.toBeInTheDocument();expect(screen.queryByLabelText('Mulai slot (WIB)')).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole('button',{name:'Baca screenshot'}));await screen.findByText('Pilih minimal satu screenshot.');
  expect(post).not.toHaveBeenCalled();expect(validateFiles([new File(['pdf'],'report.pdf',{type:'application/pdf'})])).toContain('PNG/JPEG');
 });
 it('restores AI results read-only and sends metadata only after user checks the result',async()=>{
  setup('/admin/upload?intake=draft');await screen.findByRole('heading',{name:'2. Periksa hasil AI'});
  expect(screen.queryAllByRole('textbox')).toHaveLength(0);expect(screen.getByText('Terbaca')).toBeVisible();expect(screen.queryByLabelText('Host')).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole('button',{name:'Hasil sudah diperiksa, lanjutkan'}));
  await screen.findByRole('heading',{name:'3. Lengkapi metadata'});
  fireEvent.change(screen.getByLabelText('Mulai slot (WIB)'),{target:{value:'2026-10-18T09:00'}});
  fireEvent.change(screen.getByLabelText('Akhir slot (WIB)'),{target:{value:'2026-10-18T12:00'}});
  fireEvent.submit(screen.getByRole('button',{name:'Periksa ringkasan'}).closest('form')!);
  await screen.findByRole('heading',{name:'4. Ringkasan pengajuan'});
  expect(screen.queryAllByRole('textbox')).toHaveLength(0);
  expect(screen.getByText(/berbeda dari jadwal host/)).toBeVisible();
  expect(post.mock.calls.some(([path])=>path.endsWith('/submit'))).toBe(false);
  fireEvent.click(screen.getByRole('button',{name:'Ajukan ke admin'}));await screen.findByText('Laporan berhasil diajukan');
  expect(post.mock.calls.some(([path])=>path.endsWith('/extract'))).toBe(false);
 });
 it('stores correction reasons without overwriting the original AI value',async()=>{
  setup('/admin/upload?intake=draft');await screen.findByRole('button',{name:'Koreksi'});
  fireEvent.click(screen.getByRole('button',{name:'Koreksi'}));
  fireEvent.change(screen.getByLabelText('Nilai koreksi'),{target:{value:'2500'}});
  fireEvent.change(screen.getByLabelText('Alasan koreksi'),{target:{value:'Terlihat jelas di gambar'}});
  fireEvent.submit(screen.getByRole('button',{name:'Simpan koreksi'}).closest('form')!);
  await screen.findByText('Dikoreksi user');expect(state.draft.result.values[0].value).toBe('2000');
  expect(state.draft.corrections[0].value).toBe('2500');
 });
 it('retains the uploaded draft when OpenClaw is unavailable',async()=>{
  state.draft={...fixture(),status:'uploaded',step:1,result:null};post.mockRejectedValue(new Error('OpenClaw belum dikonfigurasi. Draft tersimpan.'));
  setup('/admin/upload?intake=draft');await screen.findByRole('button',{name:'Coba baca screenshot'});
  fireEvent.click(screen.getByRole('button',{name:'Coba baca screenshot'}));await screen.findByText('OpenClaw belum dikonfigurasi. Draft tersimpan.');
  expect(screen.queryByLabelText('Host')).not.toBeInTheDocument();
  await waitFor(()=>expect(api.mock.calls.some(([path])=>path==='upload-intakes/draft')).toBe(true));
 });
});
