import type {ReactNode} from 'react';
import {can} from '../lib/api/client';
import {useWorkspace} from './Shell';
export function PermissionRoute({permission,children}:{permission:string;children:ReactNode}){const me=useWorkspace();return can(me.user,permission)?children:<section className="card"><h1>Akses tidak diizinkan</h1><p>Halaman ini memerlukan permission yang sesuai. Tidak ada data privat dimuat.</p></section>}
