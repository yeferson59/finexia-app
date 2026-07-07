import { describe, it, expect } from 'vitest';
import { cn, formatCalendarDate, todayLocalDateString } from './utils';

describe('cn', () => {
	it('joins plain string class names with a single space', () => {
		expect(cn('a', 'b', 'c')).toBe('a b c');
	});

	it('drops falsy values (undefined, null, empty string)', () => {
		expect(cn('a', undefined, null, '', 'b')).toBe('a b');
	});

	it('keeps only the keys whose value is truthy in a record', () => {
		expect(cn({ active: true, disabled: false, hidden: true })).toBe('active hidden');
	});

	it('mixes strings and conditional records', () => {
		expect(cn('base', { active: true, muted: false }, 'trailing')).toBe('base active trailing');
	});

	it('returns an empty string when nothing is truthy', () => {
		expect(cn(undefined, null, '', { off: false })).toBe('');
	});
});

describe('formatCalendarDate', () => {
	it('keeps the calendar day for a UTC-midnight ISO timestamp regardless of local timezone', () => {
		expect(
			formatCalendarDate('2026-07-07T00:00:00Z', {
				year: 'numeric',
				month: '2-digit',
				day: '2-digit'
			})
		).toBe('07/07/2026');
	});

	it('keeps the calendar day for a plain date-only string', () => {
		expect(
			formatCalendarDate('2026-01-31', { year: 'numeric', month: '2-digit', day: '2-digit' })
		).toBe('31/01/2026');
	});
});

describe('todayLocalDateString', () => {
	it('matches the local Y-M-D components of the current time', () => {
		const now = new Date();
		const expected = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
		expect(todayLocalDateString()).toBe(expected);
	});
});
