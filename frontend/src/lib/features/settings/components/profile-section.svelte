<script lang="ts">
	/** Datos de perfil: foto, nombre, correo (solo lectura) y moneda preferida. */
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import SettingsSection from './settings-section.svelte';
	import AvatarUploader from './avatar-uploader.svelte';
	import { SUPPORTED_CURRENCIES, resolveDisplayCurrency } from '$lib/shared/currency';
	import { actionError, actionSucceeded, type SettingsForm } from '../settings';

	interface Props {
		user: App.Locals['user'];
		form: SettingsForm;
	}

	let { user, form }: Props = $props();

	// Seeded from the server once; the user edits locally from there. La moneda
	// pasa por resolveDisplayCurrency para que una preferencia guardada antes de
	// que se validara el campo no deje el selector sin ninguna opción marcada.
	let profileName = $state(untrack(() => user?.name ?? ''));
	let profileCurrency = $state(untrack(() => resolveDisplayCurrency(user?.preferredCurrency)));
	let profileLoading = $state(false);

	const profileSuccess = $derived(actionSucceeded(form, 'updateProfile'));
	const profileError = $derived(actionError(form, 'updateProfile'));
</script>

<SettingsSection
	title="Foto, nombre y moneda"
	description="Tu foto sale en la barra de arriba, tu nombre en el saludo del panel y la moneda es la de sus totales."
>
	{#snippet aside()}
		<AvatarUploader {user} {form} />
	{/snippet}

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
			<!-- El correo no está: era un campo desactivado, que no se envía ni se
			     edita. Lo dice el resumen del grupo, que es donde se lee. -->
			<div class="field">
				<label class="field-label" for="preferredCurrency">Moneda preferida</label>
				<select
					id="preferredCurrency"
					name="preferredCurrency"
					class="field-control field-select"
					bind:value={profileCurrency}
				>
					{#each SUPPORTED_CURRENCIES as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
				<p class="hint">
					En esta moneda se muestran los totales del panel. Solo aparecen las que la app puede
					convertir.
				</p>
			</div>
		</div>
		{#if profileError}
			<p class="feedback error">{profileError}</p>
		{/if}
		{#if profileSuccess}
			<p class="feedback success">Perfil actualizado correctamente.</p>
		{/if}
		<div class="form-actions">
			<Button type="submit" size="sm" loading={profileLoading}>Guardar cambios</Button>
		</div>
	</form>
</SettingsSection>

<!-- La flecha del desplegable de moneda la pone `select.field-control` en
     `routes/layout.css`, la misma de todos los desplegables del panel. -->
