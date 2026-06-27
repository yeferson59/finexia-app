<script lang="ts">
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Button from '$components/ui/button.svelte';
	import Checkbox from '$components/ui/checkbox.svelte';

	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	let emailAlerts = $state(untrack(() => data.preferences.emailAlerts));
	let weeklySummary = $state(untrack(() => data.preferences.weeklySummary));
	let prefsLoading = $state(false);

	const prefsSuccess = $derived(
		form?.action === 'updatePreferences' && (form as { success?: boolean })?.success
	);
	const prefsError = $derived(
		form?.action === 'updatePreferences' ? ((form as { error?: string })?.error ?? '') : ''
	);
</script>

<svelte:head>
	<title>Notificaciones - FINEXIA</title>
	<meta name="description" content="Configura tus preferencias de notificaciones en FINEXIA" />
</svelte:head>

<PageHeader
	title="Notificaciones"
	subtitle="Elige cómo y cuándo quieres recibir alertas sobre tu portafolio."
/>

<div class="notifications-layout">
	<Card variant="elevated" padding="none">
		<div class="section">
			<div class="section-header">
				<div class="section-icon">
					<svg
						width="18"
						height="18"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"
						></path>
						<polyline points="22,6 12,13 2,6"></polyline>
					</svg>
				</div>
				<div>
					<h2 class="section-title">Correo electrónico</h2>
					<p class="section-desc">Notificaciones enviadas a tu cuenta de correo</p>
				</div>
			</div>

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
		</div>
	</Card>

	<Card variant="elevated" padding="none">
		<div class="section coming-soon-section">
			<div class="section-header">
				<div class="section-icon">
					<svg
						width="18"
						height="18"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
						<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
					</svg>
				</div>
				<div>
					<h2 class="section-title">Alertas en la app</h2>
					<p class="section-desc">Notificaciones en tiempo real dentro del dashboard</p>
				</div>
			</div>
			<div class="coming-soon-badge">
				<span class="badge-dot"></span>
				<span>Próximamente</span>
			</div>
			<p class="coming-soon-hint">
				Las alertas en tiempo real para cambios de precio y eventos de portafolio estarán
				disponibles pronto.
			</p>
		</div>
	</Card>
</div>

<style>
	.notifications-layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
		align-items: start;
	}

	.section {
		padding: 1.5rem;
	}

	.section-header {
		display: flex;
		align-items: flex-start;
		gap: 0.875rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.25rem;
		border-bottom: 1px solid rgba(212, 145, 42, 0.1);
	}

	.section-icon {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.2);
		color: var(--amber);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.section-title {
		margin: 0 0 0.2rem;
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
	}

	.section-desc {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
	}

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

	.form-actions {
		margin-top: 1.5rem;
		display: flex;
		justify-content: flex-end;
	}

	.coming-soon-section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.coming-soon-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.4rem 0.875rem;
		border-radius: 20px;
		background: rgba(212, 145, 42, 0.07);
		border: 1px solid rgba(212, 145, 42, 0.2);
		font-size: 0.8rem;
		color: var(--amber);
		font-weight: 500;
		width: fit-content;
	}

	.badge-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--amber);
		flex-shrink: 0;
	}

	.coming-soon-hint {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.45);
		line-height: 1.65;
	}

	.feedback {
		margin: 0.875rem 0 0;
		font-size: 0.835rem;
		padding: 0.6rem 0.875rem;
		border-radius: 6px;
	}

	.feedback.success {
		background: rgba(74, 222, 128, 0.08);
		border: 1px solid rgba(74, 222, 128, 0.25);
		color: #4ade80;
	}

	.feedback.error {
		background: rgba(224, 90, 90, 0.08);
		border: 1px solid rgba(224, 90, 90, 0.25);
		color: var(--red, #e05a5a);
	}

	@media (max-width: 1024px) {
		.notifications-layout {
			grid-template-columns: 1fr;
		}
	}
</style>
