<script lang="ts">
	/** Cambio de contraseña. */
	import { enhance } from '$app/forms';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import SettingsSection from './settings-section.svelte';
	import { actionError, actionSucceeded, type SettingsForm } from '../settings';

	interface Props {
		form: SettingsForm;
	}

	let { form }: Props = $props();

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordLoading = $state(false);

	const passwordSuccess = $derived(actionSucceeded(form, 'changePassword'));
	const passwordError = $derived(actionError(form, 'changePassword'));

	$effect(() => {
		if (passwordSuccess) {
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		}
	});
</script>

<SettingsSection title="Seguridad">
	<form
		method="POST"
		action="?/changePassword"
		use:enhance={() => {
			passwordLoading = true;
			return async ({ update }) => {
				await update();
				passwordLoading = false;
			};
		}}
	>
		<div class="form-fields">
			<Input
				label="Contraseña actual"
				type="password"
				name="currentPassword"
				bind:value={currentPassword}
				required
			/>
			<Input
				label="Nueva contraseña"
				type="password"
				name="newPassword"
				bind:value={newPassword}
				required
			/>
			<Input
				label="Confirmar nueva contraseña"
				type="password"
				name="confirmPassword"
				bind:value={confirmPassword}
				required
				error={confirmPassword && confirmPassword !== newPassword
					? 'Las contraseñas no coinciden'
					: ''}
			/>
		</div>
		{#if passwordError}
			<p class="feedback error">{passwordError}</p>
		{/if}
		{#if passwordSuccess}
			<p class="feedback success">Contraseña actualizada correctamente.</p>
		{/if}
		<div class="form-actions">
			<Button type="submit" loading={passwordLoading}>Cambiar contraseña</Button>
		</div>
	</form>
</SettingsSection>
