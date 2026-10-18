import React from 'react';
import {render,screen} from '@testing-library/react';
import {describe,it,expect,vi} from 'vitest';
import {PermissionRoute} from '../app/PermissionRoute';
vi.mock('../app/Shell',()=>({useWorkspace:()=>({user:{role:'client',permissions:['read_brand_report']}})}));
describe('permission route',()=>{it('does not mount an admin feature or its requests for a client',()=>{const mounted=vi.fn();function Admin(){mounted();return <p>Private Admin</p>};render(<PermissionRoute permission="upload_report"><Admin/></PermissionRoute>);expect(screen.getByText('Akses tidak diizinkan')).toBeVisible();expect(mounted).not.toHaveBeenCalled();expect(screen.queryByText('Private Admin')).not.toBeInTheDocument()})});
