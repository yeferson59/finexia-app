import { describe, it, expect } from 'vitest';
import { cn } from './utils';

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
