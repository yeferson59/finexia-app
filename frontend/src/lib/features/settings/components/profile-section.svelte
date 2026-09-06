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
					class="field-select"
					bind:value={profileCurrency}
				>
					{#each SUPPORTED_CURRENCIES as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
				<p class="field-hint">
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
	/* Mismo aspecto que `ui/input.svelte`, que es el vecino de este campo en el
	   formulario. No se factoriza al chrome de `settings-section` porque ninguna
	   otra sección de ajustes tiene un desplegable. */
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.field-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.field-select {
		width: 100%;
		padding: 0.875rem 1rem;
		padding-right: 2.5rem;
		border-radius: 8px;
		border: 1px solid rgba(212, 145, 42, 0.2);
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-size: 0.95rem;
		font-family: var(--font-body);
		transition: all 0.25s ease;
		outline: none;
		box-sizing: border-box;
		cursor: pointer;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.9rem center;
	}

	.field-select:hover {
		border-color: rgba(212, 145, 42, 0.35);
	}

	.field-select:focus {
		border-color: var(--amber);
		box-shadow: 0 0 0 3px var(--border);
	}

	.field-select option {
		background: var(--bg);
		color: var(--text);
	}

	.field-hint {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
		line-height: 1.5;
	}
</style>
