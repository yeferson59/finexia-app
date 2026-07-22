import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import InviteOnlyNotice from './invite-only-notice.svelte';

describe('invite-only-notice.svelte', () => {
	it('renders the invite-only copy and the waitlist call to action', async () => {
		render(InviteOnlyNotice, { onSwitchToLogin: () => {} });

		await expect.element(page.getByText('Registro por invitación')).toBeInTheDocument();
		await expect
			.element(page.getByRole('link', { name: 'Unirme a la lista de espera' }))
			.toBeInTheDocument();
	});
});
