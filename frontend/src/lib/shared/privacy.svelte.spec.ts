import { describe, it, expect } from 'vitest';
import { privacy } from './privacy.svelte';

describe('privacy store (hidden mode)', () => {
	it('shows monetary values as-is by default', () => {
		expect(privacy.hidden).toBe(false);
		expect(privacy.money('$1.234,56')).toBe('$1.234,56');
	});

	it('masks values while hidden mode is on and persists the choice', () => {
		privacy.toggle();

		expect(privacy.hidden).toBe(true);
		expect(privacy.money('$1.234,56')).toBe('••••••');
		expect(localStorage.getItem('finexia:hidden-mode')).toBe('1');

		privacy.toggle();

		expect(privacy.hidden).toBe(false);
		expect(privacy.money('$1.234,56')).toBe('$1.234,56');
		expect(localStorage.getItem('finexia:hidden-mode')).toBe('0');
	});
});
