import {ExternalActionLink} from '../components/ui/Interactions';
import {Notifications} from '../components/ui/Notifications';
import {lazy,Suspense} from 'react';
import {Loading} from '../components/ui/Primitives';
import {PermissionRoute} from './PermissionRoute';
import {BrowserRouter,Routes,Route,Navigate} from 'react-router-dom';
import {QueryClientProvider} from '@tanstack/react-query';
import {queryClient} from '../lib/query/client';
import {Guard,useWorkspace} from './Shell';
import {Login,Reset} from '../features/auth';
const DesignSystem=lazy(()=>import('../features/designSystem').then(m=>({default:m.DesignSystem})));
const Dashboard=lazy(()=>import('../features/dashboard').then(m=>({default:m.Dashboard})));
const Sessions=lazy(()=>import('../features/sessions').then(m=>({default:m.Sessions})));
const ReportDetail=lazy(()=>import('../features/sessions').then(m=>({default:m.ReportDetail})));
const Upload=lazy(()=>import('../features/upload').then(m=>({default:m.Upload})));
const Leaderboard=lazy(()=>import('../features/leaderboard').then(m=>({default:m.Leaderboard})));
const HostProfile=lazy(()=>import('../features/leaderboard').then(m=>({default:m.HostProfile})));
const Operations=lazy(()=>import('../features/admin').then(m=>({default:m.Operations})));
const Masters=lazy(()=>import('../features/admin').then(m=>({default:m.Masters})));
const Users=lazy(()=>import('../features/admin').then(m=>({default:m.Users})));
const Config=lazy(()=>import('../features/admin').then(m=>({default:m.Config})));
const Insights=lazy(()=>import('../features/admin').then(m=>({default:m.Insights})));
function Home(){const me=useWorkspace();return me.user.role==='host'?<Navigate to="/admin/upload" replace/>:<Dashboard/>}
export default function App(){return <QueryClientProvider client={queryClient}><BrowserRouter><Notifications/><Suspense fallback={<Loading/>}><Routes><Route path="/login" element={<Login/>}/><Route path="/reset" element={<Reset/>}/><Route element={<Guard/>}><Route index element={<Home/>}/><Route path="design-system" element={<DesignSystem/>}/><Route path="sessions" element={<PermissionRoute permission="read_brand_report"><Sessions/></PermissionRoute>}/><Route path="sessions/:id" element={<PermissionRoute permission="read_brand_report"><ReportDetail/></PermissionRoute>}/><Route path="leaderboard" element={<PermissionRoute permission="read_global_host_summary"><Leaderboard/></PermissionRoute>}/><Route path="leaderboard/:id" element={<PermissionRoute permission="read_global_host_summary"><HostProfile/></PermissionRoute>}/><Route path="insights" element={<PermissionRoute permission="use_insights"><Insights/></PermissionRoute>}/><Route path="admin/upload" element={<PermissionRoute permission="upload_report"><Upload/></PermissionRoute>}/><Route path="admin/reports" element={<PermissionRoute permission="review_report"><Sessions admin/></PermissionRoute>}/><Route path="admin/operations" element={<PermissionRoute permission="review_report"><Operations/></PermissionRoute>}/><Route path="admin/masters" element={<PermissionRoute permission="manage_host"><Masters/></PermissionRoute>}/><Route path="admin/users" element={<PermissionRoute permission="manage_brand_users"><Users/></PermissionRoute>}/><Route path="admin/config" element={<PermissionRoute permission="configure_brand_dashboard"><Config/></PermissionRoute>}/><Route path="*" element={<section className="card"><h1>Tidak ditemukan</h1><ExternalActionLink href="/">Kembali ke Ringkasan</ExternalActionLink></section>}/></Route><Route path="/auth" element={<Navigate to="/login"/>}/></Routes></Suspense></BrowserRouter></QueryClientProvider>}
