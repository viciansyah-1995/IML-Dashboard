import {useState} from 'react';
import {render,screen,fireEvent} from '@testing-library/react';
import {describe,it,expect,vi} from 'vitest';
import {Avatar,AssignmentPicker,initials,PhotoPicker,StatusToggle} from '../components/ui/Profiles';
import {MasterForm,UserForm,type ManagedUser} from '../features/profiles';
import type {Masters} from '../lib/api/client';
const masters:Masters={brands:[{id:'a',name:'Brand A'},{id:'b',name:'Brand B'}],studios:[],hosts:[],channels:[],accounts:[],mappings:[]};
const user:ManagedUser={id:'user-1',email:'client@example.com',name:'Nadia Putri',role:'client',active:true,access_version:1,joined_on:'2026-10-01',brand_ids:['a','b'],studio_ids:[]};
describe('host and user profiles',()=>{
 it('uses a keyboard accessible switch and records the actual active value in forms',()=>{function Toggle(){const [active,setActive]=useState(true);return <StatusToggle name="active" label="Status aktif pengguna" value={active} onChange={setActive}/>};render(<Toggle/>);fireEvent.click(screen.getByRole('switch'));expect(screen.getByRole('switch')).toHaveAttribute('aria-checked','false');expect(document.querySelector('input[name=active]')).toHaveValue('off')});
 it('uses name initials when photo is absent or fails to load',()=>{expect(initials(' Nadia Putri ')).toBe('NP');render(<Avatar name="Nadia Putri" src="/missing.png"/>);fireEvent.error(screen.getByRole('img'));expect(screen.getByText('NP')).toBeVisible()});
 it('selects and clears every brand while preserving initial assignments',()=>{function Picker(){const [ids,setIds]=useState(['a']);return <AssignmentPicker items={masters.brands} selected={ids} onChange={setIds}/>};render(<Picker/>);expect(screen.getByRole('checkbox',{name:'Brand A'})).toBeChecked();fireEvent.click(screen.getByRole('button',{name:'Pilih semua'}));expect(screen.getByRole('checkbox',{name:'Brand B'})).toBeChecked();fireEvent.click(screen.getByRole('button',{name:'Batalkan semua'}));expect(screen.getByRole('checkbox',{name:'Brand A'})).not.toBeChecked()});
 it('shows every role for Superadmin, a mandatory join date, and retains multiple assignments on edit',()=>{render(<UserForm editing={user} masters={masters} superadmin pending={false} onSave={vi.fn()}/>);fireEvent.click(screen.getByRole('combobox',{name:'Role'}));expect(screen.getByRole('option',{name:'Superadmin'})).toBeInTheDocument();expect(screen.getByRole('option',{name:'Admin'})).toBeInTheDocument();expect(screen.getByRole('option',{name:'Host'})).toBeInTheDocument();expect(screen.getByLabelText('Tanggal bergabung')).toBeRequired();expect(screen.getByRole('checkbox',{name:'Brand A'})).toBeChecked();expect(screen.getByRole('checkbox',{name:'Brand B'})).toBeChecked()});
 it('limits the Admin form to Client creation',()=>{render(<UserForm editing={{...user,id:''}} masters={masters} superadmin={false} pending={false} onSave={vi.fn()}/>);fireEvent.click(screen.getByRole('combobox',{name:'Role'}));expect(screen.getByRole('option',{name:'Client'})).toBeVisible();expect(screen.queryByRole('option',{name:'Superadmin'})).toBeNull();expect(screen.queryByRole('option',{name:'Admin'})).toBeNull()});
 it('keeps photo optional and displays host demographic and reporting fields',()=>{render(<MasterForm editing={{id:'h',name:'Host Nadia',full_name:'Nadia Putri',active:true,brand_ids:['a']}} kind="hosts" masters={masters} superadmin pending={false} onSave={vi.fn()}/>);expect(screen.getByLabelText('Nama lengkap host')).toBeRequired();expect(screen.getByLabelText('Email host')).not.toBeRequired();expect(screen.getByLabelText('Foto profil (opsional)')).not.toBeRequired();expect(screen.getByLabelText('Tanggal lahir')).toHaveAttribute('type','date');expect(screen.getByRole('combobox',{name:'Jenis kelamin'})).toHaveValue('');expect(screen.getByLabelText('Spesialisasi konten / produk')).toBeInTheDocument()});
 it('removes an existing photo explicitly so the server can restore initials',()=>{render(<PhotoPicker name="Nadia Putri" src="/photo.png"/>);fireEvent.click(screen.getByRole('button',{name:'Hapus foto'}));expect(screen.getByText('NP')).toBeVisible();expect(document.querySelector('input[name=remove_photo]')).toHaveValue('true')});
 it('submits a linked host account with assigned brands and studios',()=>{
  const onSave=vi.fn();
  render(<UserForm editing={{...user,id:'',role:'host',host_id:'h',studio_ids:['studio']}} masters={{...masters,hosts:[{id:'h',name:'Dea',active:true}],studios:[{id:'studio',name:'Studio 1',active:true}]}} superadmin pending={false} onSave={onSave}/>);
  expect(screen.getByRole('combobox',{name:'Profil host'})).toBeVisible();
  expect(screen.getByRole('checkbox',{name:'Studio 1'})).toBeChecked();
  fireEvent.submit(screen.getByRole('button',{name:'Simpan pengguna'}).closest('form')!);
  const body=onSave.mock.calls[0][0] as FormData;
  expect(JSON.parse(String(body.get('metadata')))).toMatchObject({role:'host',host_id:'h',brand_ids:['a','b'],studio_ids:['studio'],active:true});
 });

});
