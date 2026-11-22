import {formatDateTime} from '../lib/datetime';
import {it,expect} from 'vitest';
import {currentReportDates,readFilters} from '../lib/api/filters';
it('defaults to the last 30 dates in WIB and preserves user-selected dates',()=>{
 expect(currentReportDates(new Date('2026-10-21T18:00:00Z'))).toEqual({from:'2026-09-23',to:'2026-10-22'});
 const f=readFilters(new URLSearchParams('p_from=2026-10-20&p_to=2026-10-21&p_scope=room'),false,'zwitsal');
 expect(f.date_from).toBe('2026-10-20');expect(f.date_to).toBe('2026-10-21');expect(f.scope).toBe('room');
});

it('formats timestamps in GMT+7 regardless of the original offset',()=>{
 const utc=formatDateTime('2026-11-21T18:00:00Z');
 expect(utc).toContain('22/11/2026');
 expect(utc).toContain('01.00.00');
 expect(utc).toContain('GMT+7');
 expect(formatDateTime('2026-11-22T01:00:00+07:00')).toBe(utc);
 expect(formatDateTime(null)).toBe('Belum diketahui');
});
