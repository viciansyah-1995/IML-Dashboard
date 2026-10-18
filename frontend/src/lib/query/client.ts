import {QueryClient} from '@tanstack/react-query';
import type {User} from '../api/client';
export const queryClient=new QueryClient({defaultOptions:{queries:{retry:1,staleTime:10000,refetchOnWindowFocus:true}}});
export const key=(u:User,scope:'private'|'global'|'operations',feature:string,filter:unknown)=>[u.id,u.access_version,scope,feature,filter];
export async function invalidateReports(){await queryClient.invalidateQueries({predicate:q=>q.queryKey[2]==='private'||q.queryKey[2]==='global'||q.queryKey[2]==='operations'})}
