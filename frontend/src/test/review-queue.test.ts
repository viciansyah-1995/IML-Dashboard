import {describe,it,expect} from 'vitest';
import {readFilters,reviewQueueQuery} from '../lib/api/filters';
describe('review inbox',()=>{
 it('does not hide Shopee room submissions behind inherited dashboard filters',()=>{
  const f=readFilters(new URLSearchParams('p_channel=tiktok&p_scope=host_slot&p_basis=Pesanan+Teratribusi&p_brand=brand-a&p_from=2099-01-01&p_host=other-host&p_definition=other-mapping'),false);
  const query=new URLSearchParams(reviewQueueQuery(f));
  expect(query.get('review_queue')).toBe('1');
  expect(query.get('brand_id')).toBe('brand-a');
  expect(query.get('include_drafts')).toBe('1');
  for(const name of ['channel','scope','basis','date_from','date_to','host_id','definition_version'])expect(query.has(name)).toBe(false);
 });
 it('preserves search and pagination',()=>{
  const query=new URLSearchParams(reviewQueueQuery({...readFilters(new URLSearchParams(),false,'a'),search:'261101-00001',page:'2',page_size:'50'}));
  expect(query.get('search')).toBe('261101-00001');expect(query.get('page')).toBe('2');expect(query.get('page_size')).toBe('50');
 });
});
