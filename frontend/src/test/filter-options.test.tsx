import {render,screen,fireEvent,waitFor} from '@testing-library/react';
import {describe,it,expect,vi,beforeEach} from 'vitest';
import {MemoryRouter} from 'react-router-dom';
import {QueryClient,QueryClientProvider} from '@tanstack/react-query';
import {readFilters} from '../lib/api/filters';
import {relevantChoices,type FilterOption} from '../lib/filter-options';
import {FilterBar} from '../components/ui/FilterBar';
const state=vi.hoisted(()=>({items:[] as FilterOption[],update:vi.fn(),filter:{} as ReturnType<typeof readFilters>}));
vi.mock('../lib/api/client',()=>({api:vi.fn(async()=>({items:state.items}))}));
vi.mock('../app/Shell',()=>({useFilters:()=>({f:state.filter,update:state.update}),useWorkspace:()=>({user:{id:'test',access_version:1}}),useMasters:()=>({data:{mappings:[{id:'def-1',channel_id:'shopee',label:'Mapping 1'},{id:'def-2',channel_id:'shopee',label:'Mapping 2'}],studios:[{id:'studio-1',name:'Studio 1'},{id:'studio-9',name:'Studio 9'}],hosts:[{id:'seli',name:'Seli'},{id:'nurul',name:'Nurul'}]}})}));
const mount=()=>render(<MemoryRouter><QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false}}})}><FilterBar/></QueryClientProvider></MemoryRouter>);
beforeEach(()=>{state.update.mockClear();state.filter=readFilters(new URLSearchParams('p_from=2026-10-20&p_to=2026-10-21&p_brand=harvest&p_scope=room&p_channel=shopee&p_basis=Pesanan+Dibuat'),false);state.items=[{scope:'room',basis:'Pesanan Dibuat',definition_version:'def-1',studio_id:'studio-9',host_id:'seli'}]});
describe('relevant report filters',()=>{
 it('hides singleton filters and unused Host, preserving the effective request',async()=>{
  mount();await waitFor(()=>expect(screen.queryByText('Versi definisi',{exact:true})).not.toBeInTheDocument());
  for(const label of ['Cakupan','Studio','Host','Basis pesanan'])expect(screen.queryByText(label,{exact:true})).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole('button',{name:'Terapkan'}));
  expect(state.update).toHaveBeenCalledWith(expect.objectContaining({scope:'room',basis:'Pesanan Dibuat',definition_version:'',studio_id:'',host_id:''}));
 });
 it('restores filters when more than one actual choice exists',async()=>{
  state.items.push({scope:'room',basis:'Pesanan Dibuat',definition_version:'def-2',studio_id:'studio-1',host_id:'nurul'},{scope:'host_slot',basis:'Pesanan Selesai',definition_version:'def-1',studio_id:'studio-1',host_id:'nurul'});
  mount();await screen.findByText('Cakupan',{exact:true});
  expect(screen.getByText('Versi definisi',{exact:true})).toBeVisible();expect(screen.getByText('Studio',{exact:true})).toBeVisible();expect(screen.getByText('Basis pesanan',{exact:true})).toBeVisible();
 });
 it('retains stale explicit selections and offers configured choices with no data',()=>{
  expect(relevantChoices(['room'],['room','host_slot'],'host_slot')).toEqual(['room','host_slot']);
  expect(relevantChoices([],['def-1','def-2'],'')).toEqual(['def-1','def-2']);
 });
});
