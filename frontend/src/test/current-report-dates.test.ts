import {it,expect} from 'vitest';
import {currentReportDates,readFilters} from '../lib/api/filters';
it('defaults to the last 30 dates in WIB and preserves user-selected dates',()=>{
 expect(currentReportDates(new Date('2026-10-21T18:00:00Z'))).toEqual({from:'2026-09-23',to:'2026-10-22'});
 const f=readFilters(new URLSearchParams('p_from=2026-10-20&p_to=2026-10-21&p_scope=room'),false,'zwitsal');
 expect(f.date_from).toBe('2026-10-20');expect(f.date_to).toBe('2026-10-21');expect(f.scope).toBe('room');
});
