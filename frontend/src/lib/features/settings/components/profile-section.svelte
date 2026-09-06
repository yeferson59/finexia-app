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

<style>
	/*
	 * El desplegable de moneda es el único de toda la página de ajustes, y su
	 * aspecto era una copia a mano de `ui/input.svelte` —lo decía el comentario
	 * que había aquí— con medio punto más de letra y otro tamaño de relleno.
	 * Ahora pide el mismo control que el resto del panel y solo añade lo suyo:
	 * la flecha, porque un `<select>` sin `appearance: none` pinta la del
	 * sistema, que en un panel oscuro se ve blanca.
	 *
	 * Doble clase para ganarle en especificidad al `background` de
	 * `.field-control`, que sin ella borraría la flecha.
	 */
	.field-control.field-select {
		padding-right: 2.5rem;
		cursor: pointer;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.9rem center;
	}
</style>
