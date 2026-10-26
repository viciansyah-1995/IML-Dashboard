import {describe,it,expect,vi} from 'vitest';
import {render,screen,fireEvent} from '@testing-library/react';
import {MemoryRouter,Routes,Route} from 'react-router-dom';
import {QueryClient,QueryClientProvider} from '@tanstack/react-query';
import {ReportDetail} from '../features/sessions';
vi.mock('../app/Shell',()=>({PageHeader:()=>null,useWorkspace:()=>({user:{id:'reviewer',role:'superadmin',permissions:['review_report']}})}));
vi.mock('../lib/api/client',async importOriginal=>({...await importOriginal<typeof import('../lib/api/client')>(),api:vi.fn(async()=>({report:{id:'r',status:'review',values:[{key:'sales',value:'0',evidence_text:'GMV 0',validation_flags:[]}],start_at:'2026-10-18T09:00:00+07:00',end_at:'2026-10-18T12:00:00+07:00'},definitions:[{key:'sales',label:'GMV',unit:'IDR',required:true,canonical_confirmed:true}],attachments:[],history:[],blockers:[]}))}));
describe('report preview',()=>{
 it('shows mapping and values in a table with correction controls closed by default',async()=>{
  render(<MemoryRouter initialEntries={['/sessions/r']}><QueryClientProvider client={new QueryClient()}><Routes><Route path="/sessions/:id" element={<ReportDetail/>}/></Routes></QueryClientProvider></MemoryRouter>);
  expect(await screen.findByRole('columnheader',{name:'Mapping field'})).toBeVisible();
  expect(screen.getByText('sales')).toBeVisible();
  expect(screen.queryAllByRole('textbox')).toHaveLength(0);
  expect(screen.queryByRole('button',{name:'Simpan review'})).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole('button',{name:'Buka koreksi reviewer'}));
  expect(screen.getByLabelText('Alasan / catatan review')).toBeVisible();
 });
});
