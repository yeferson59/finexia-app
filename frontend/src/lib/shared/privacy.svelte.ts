/**
 * Hidden mode ("modo oculto"): masks monetary amounts across the dashboard so
 * the app can be used in public without exposing balances. Toggled from the
 * header's eye button and persisted per browser.
 */
const STORAGE_KEY = 'finexia:hidden-mode';
const MASK = '••••••';

class PrivacyStore {
	hidden = $state(false);

	constructor() {
		if (typeof localStorage !== 'undefined') {
			this.hidden = localStorage.getItem(STORAGE_KEY) === '1';
		}
	}

	toggle(): void {
		this.hidden = !this.hidden;
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem(STORAGE_KEY, this.hidden ? '1' : '0');
		}
	}

	/**
	 * Returns the formatted monetary value as-is, or a fixed mask while hidden
	 * mode is on. Reading `hidden` here makes every call site reactive to the
	 * toggle without each component subscribing explicitly.
	 */
	money(formatted: string): string {
		return this.hidden ? MASK : formatted;
	}
}

export const privacy = new PrivacyStore();
