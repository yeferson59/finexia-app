<script lang="ts">
	/** Preferencias de correo: alertas de actividad y resumen semanal. */
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import Button from '$lib/ui/button.svelte';
	import Checkbox from '$lib/ui/checkbox.svelte';
	import NotificationSection from './notification-section.svelte';
	import { actionError, actionSucceeded, type ActionForm } from '$lib/shared/form';
	import type { UserPreferences } from '$lib/api/types';

	interface Props {
		preferences: UserPreferences;
		form: ActionForm;
	}

	let { preferences, form }: Props = $props();

	// Seeded from the server once; the user toggles locally from there.
	let emailAlerts = $state(untrack(() => preferences.emailAlerts));
	let weeklySummary = $state(untrack(() => preferences.weeklySummary));
	let prefsLoading = $state(false);

	const prefsSuccess = $derived(actionSucceeded(form, 'updatePreferences'));
	const prefsError = $derived(actionError(form, 'updatePreferences'));
</script>

<NotificationSection
	title="Correo electrónico"
	description="Notificaciones enviadas a tu cuenta de correo"
>
	{#snippet icon()}
		<svg
			width="18"
			height="18"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
		>
			<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
			<polyline points="22,6 12,13 2,6"></polyline>
		</svg>
	{/snippet}

	<form
		method="POST"
		action="?/updatePreferences"
		use:enhance={() => {
			prefsLoading = true;
			return async ({ update }) => {
				await update();
				prefsLoading = false;
			};
		}}
	>
		<div class="toggle-list">
			<div class="toggle-row">
				<div class="toggle-info">
					<p class="toggle-label">Alertas de actividad</p>
					<p class="toggle-hint">
						Recibe un correo cuando haya movimientos importantes en tu portafolio
					</p>
				</div>
				<Checkbox name="emailAlerts" bind:checked={emailAlerts} />
			</div>
			<div class="toggle-row">
				<div class="toggle-info">
					<p class="toggle-label">Resumen semanal</p>
					<p class="toggle-hint">Un resumen con el desempeño de tu portafolio cada semana</p>
				</div>
				<Checkbox name="weeklySummary" bind:checked={weeklySummary} />
			</div>
		</div>

		{#if prefsError}
			<p class="feedback error">{prefsError}</p>
		{/if}
		{#if prefsSuccess}
			<p class="feedback success">Preferencias guardadas correctamente.</p>
		{/if}

		<div class="form-actions">
			<Button type="submit" loading={prefsLoading}>
				{prefsLoading ? 'Guardando…' : 'Guardar preferencias'}
			</Button>
		</div>
	</form>
</NotificationSection>

<style>
	.toggle-list {
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.toggle-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.875rem 0;
		border-bottom: 1px solid rgba(212, 145, 42, 0.08);
	}

	.toggle-row:last-child {
		border-bottom: none;
	}

	.toggle-info {
		flex: 1;
		padding-right: 1rem;
	}

	.toggle-label {
		margin: 0 0 0.25rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
	}

	.toggle-hint {
		margin: 0;
		font-size: 0.775rem;
		color: rgba(236, 234, 229, 0.45);
		line-height: 1.55;
	}
</style>
