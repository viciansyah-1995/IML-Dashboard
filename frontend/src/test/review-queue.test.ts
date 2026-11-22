import {describe,it,expect} from 'vitest';
import {readFilters,reviewQueueQuery,publishedQuery,readPublishedFilters,optionalPublishedQuery} from '../lib/api/filters';
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

describe('published default',()=>{it('ignores inherited date/channel/host filters and excludes drafts',()=>{const q=new URLSearchParams(publishedQuery(readFilters(new URLSearchParams('p_channel=tiktok&p_scope=host_slot&p_from=2099-01-01'),false,'brand-a')));expect(q.get('published_all')).toBe('1');expect(q.get('brand_id')).toBe('brand-a');for(const k of ['date_from','channel','scope','basis','include_drafts','review_queue'])expect(q.has(k)).toBe(false)})});

describe('optional published filters',()=>{
 it('starts with no active constraints even when old default filter values are in the URL',()=>{const f=readPublishedFilters(new URLSearchParams('dashboard_view=filtered&p_channel=tiktok&p_from=2026-10-24'),'a');expect(f.channel).toBe('');expect(f.date_from).toBe('');expect(f.scope).toBe('')});
 it('sends only explicitly applied filters',()=>{const f=readPublishedFilters(new URLSearchParams('p_applied=1&p_channel=shopee&p_scope=room'),'a');const q=new URLSearchParams(optionalPublishedQuery(f));expect(q.get('channel')).toBe('shopee');expect(q.get('scope')).toBe('room');expect(q.has('date_from')).toBe(false);expect(q.has('basis')).toBe(false)});
});
