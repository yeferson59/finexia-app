<script lang="ts">
	/** Datos de perfil: foto, nombre, correo (solo lectura) y moneda preferida. */
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import SettingsSection from './settings-section.svelte';
	import AvatarUploader from './avatar-uploader.svelte';
	import { actionError, actionSucceeded, type SettingsForm } from '../settings';

	interface Props {
		user: App.Locals['user'];
		form: SettingsForm;
	}

	let { user, form }: Props = $props();

	// Seeded from the server once; the user edits locally from there.
	let profileName = $state(untrack(() => user?.name ?? ''));
	let profileCurrency = $state(untrack(() => user?.preferredCurrency ?? 'USD'));
	let profileLoading = $state(false);

	const profileSuccess = $derived(actionSucceeded(form, 'updateProfile'));
	const profileError = $derived(actionError(form, 'updateProfile'));
</script>

<SettingsSection title="Perfil">
	<AvatarUploader {user} {form} />

	<form
		method="POST"
		action="?/updateProfile"
		use:enhance={() => {
			profileLoading = true;
			return async ({ update }) => {
				await update();
				profileLoading = false;
			};
		}}
	>
		<div class="form-fields">
			<Input label="Nombre" name="name" bind:value={profileName} required />
			<Input label="Correo electrónico" name="email" value={user?.email ?? ''} disabled />
			<Input
				label="Moneda preferida"
				name="preferredCurrency"
				bind:value={profileCurrency}
				placeholder="USD"
			/>
		</div>
		{#if profileError}
			<p class="feedback error">{profileError}</p>
		{/if}
		{#if profileSuccess}
			<p class="feedback success">Perfil actualizado correctamente.</p>
		{/if}
		<div class="form-actions">
			<Button type="submit" loading={profileLoading}>Guardar perfil</Button>
		</div>
	</form>
</SettingsSection>
