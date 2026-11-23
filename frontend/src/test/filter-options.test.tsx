import {render,screen,fireEvent,waitFor} from '@testing-library/react';
import {describe,it,expect,vi,beforeEach} from 'vitest';
import {MemoryRouter} from 'react-router-dom';
import {QueryClient,QueryClientProvider} from '@tanstack/react-query';
import {readFilters} from '../lib/api/filters';
import {relevantChoices,availableFilters,type FilterOption} from '../lib/filter-options';
import {FilterBar} from '../components/ui/FilterBar';
const state=vi.hoisted(()=>({items:[] as FilterOption[],update:vi.fn(),reset:vi.fn(),filter:{} as ReturnType<typeof readFilters>}));
vi.mock('../lib/api/client',()=>({api:vi.fn(async()=>({items:state.items}))}));
vi.mock('../app/Shell',()=>({useFilters:()=>({f:state.filter,update:state.update,reset:state.reset}),useWorkspace:()=>({user:{id:'test',access_version:1}}),useMasters:()=>({data:{mappings:[{id:'def-1',channel_id:'shopee',label:'Mapping 1'},{id:'def-2',channel_id:'shopee',label:'Mapping 2'}],studios:[{id:'studio-1',name:'Studio 1'},{id:'studio-9',name:'Studio 9'}],hosts:[{id:'seli',name:'Seli'},{id:'nurul',name:'Nurul'}]}})}));
const mount=(path='/',optional=false)=>render(<MemoryRouter initialEntries={[path]}><QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false}}})}><FilterBar optional={optional}/></QueryClientProvider></MemoryRouter>);
beforeEach(()=>{state.update.mockClear();state.filter=readFilters(new URLSearchParams('p_from=2026-10-20&p_to=2026-10-21&p_brand=harvest&p_scope=room&p_channel=shopee&p_basis=Pesanan+Dibuat'),false);state.items=[{scope:'room',basis:'Pesanan Dibuat',definition_version:'def-1',studio_id:'studio-9',host_id:'seli'}]});
describe('relevant report filters',()=>{
 it('hides singleton filters and unused Host, preserving the effective request',async()=>{
  mount();await waitFor(()=>expect(screen.queryByText('Versi definisi',{exact:true})).not.toBeInTheDocument());
  for(const label of ['Studio','Host'])expect(screen.queryByText(label,{exact:true})).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole('button',{name:'Terapkan'}));
  expect(state.update).toHaveBeenCalledWith(expect.objectContaining({scope:'room',basis:'Pesanan Dibuat',definition_version:'',studio_id:'',host_id:''}));
 });
 it('restores filters when more than one actual choice exists',async()=>{
  state.items.push({scope:'room',basis:'Pesanan Dibuat',definition_version:'def-2',studio_id:'studio-1',host_id:'nurul'},{scope:'host_slot',basis:'Pesanan Selesai',definition_version:'def-1',studio_id:'studio-1',host_id:'nurul'});
  mount();await screen.findByText('Cakupan',{exact:true});
  expect(screen.getByText('Versi definisi',{exact:true})).toBeVisible();expect(screen.getByText('Studio',{exact:true})).toBeVisible();expect(screen.getByText('Basis pesanan',{exact:true})).toBeVisible();
 });
 it('keeps the actual unread basis available for review',async()=>{
  state.filter={...state.filter,basis:'Belum terbaca'};
  state.items=[{...state.items[0],basis:'Belum terbaca'}];
  mount('/admin/reports');
  const basis=await screen.findByRole('combobox',{name:'Basis pesanan'});
  await waitFor(()=>expect(basis).toHaveValue('Belum terbaca'));
  fireEvent.click(basis);
  expect(screen.getByRole('option',{name:'Belum terbaca'})).toBeVisible();
 });
 it('removes incompatible filters when switching to whole LIVE data',()=>{
  const f={...state.filter,scope:'host_slot',basis:'Pesanan Selesai',definition_version:'old',studio_id:'old',host_id:'old'};
  expect(availableFilters(f,state.items)).toEqual(expect.objectContaining({scope:'room',basis:'Pesanan Dibuat',definition_version:'',studio_id:'',host_id:''}));
 });
 it('does not display master dropdown choices when no reports are available',async()=>{
  state.items=[];
  mount();
  await screen.findByText(/Belum ada laporan yang dipublikasikan/);
  for(const label of ['Studio','Host','Versi definisi','Cakupan','Basis pesanan'])expect(screen.queryByText(label,{exact:true})).not.toBeInTheDocument();
 });
 it('uses the channel default outside the review list',()=>{
  const params=new URLSearchParams('p_channel=shopee&p_basis=Belum+terbaca');
  expect(readFilters(params,false).basis).toBe('Pesanan Dibuat');
  expect(readFilters(params,false,'',true).basis).toBe('Belum terbaca');
 });
 it('retains stale explicit selections and offers configured choices with no data',()=>{
  expect(relevantChoices(['room'],['room','host_slot'],'host_slot')).toEqual(['room','host_slot']);
  expect(relevantChoices([],['def-1','def-2'],'')).toEqual(['def-1','def-2']);
 });
});

describe('optional filter controls',()=>{
 it('keeps the same form visible and applies only on submit',async()=>{
 state.filter={...state.filter,date_from:'',date_to:'',channel:'',basis:'',scope:'',host_id:'',studio_id:''};
 mount('/',true);expect(screen.getByRole('button',{name:'Terapkan'})).toBeVisible();expect(screen.getByRole('combobox',{name:'Channel'})).toBeVisible();expect(state.update).not.toHaveBeenCalled();
 await waitFor(()=>expect(screen.getByRole('button',{name:'Terapkan'})).toBeEnabled());
 fireEvent.click(screen.getByRole('button',{name:'Terapkan'}));expect(state.update).toHaveBeenCalledWith(expect.objectContaining({date_from:'',date_to:'',channel:'',scope:''}),true);
 fireEvent.click(screen.getByRole('button',{name:'Reset filter'}));expect(state.reset).toHaveBeenCalled();
 });
});

it('preserves all-source optional filters while clearing host for whole LIVE',()=>{
 const all={...state.filter,date_from:'',date_to:'',channel:'',basis:'',scope:'',definition_version:'',host_id:'',studio_id:''};
 expect(availableFilters(all,state.items,false,true)).toEqual(all);
 expect(availableFilters({...all,scope:'room',host_id:'seli'},state.items,false,true).host_id).toBe('');
});
