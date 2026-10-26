import {render,screen} from '@testing-library/react';
import {it,expect} from 'vitest';
import {ScreenshotPanel} from '../components/ui/ScreenshotPanel';
it('shows partial source calculations without presenting them as a host score',()=>{
 render(<ScreenshotPanel result={{version:'screenshot-v1',scope:'room',warnings:['Periode sumber berbeda dari slot.'],metrics:[{key:'ctor',label:'CTOR dihitung ulang',formula:'Orders / clicks',unit:'percent',value:null,missing:['product_clicks']},{key:'cart',label:'Cart → Order',formula:'Orders / ATC',unit:'percent',value:'10.38961',missing:[]}]}}/>);
 expect(screen.getByText('Data belum tersedia')).toBeVisible();expect(screen.getByText('Input belum tersedia atau belum valid: Klik produk absolut.')).toBeVisible();expect(screen.getByText('10,39%')).toBeVisible();expect(screen.getByText(/tidak memberi atribusi host/)).toBeVisible();
});
