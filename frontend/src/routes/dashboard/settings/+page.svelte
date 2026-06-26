<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Input from '$components/ui/input.svelte';
	import Button from '$components/ui/button.svelte';
	import Checkbox from '$components/ui/checkbox.svelte';

	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	// Capture server-provided values once; user edits them locally
	const initialUser = data.user;
	const initialPrefs = data.preferences;

	// Profile section
	let profileName = $state(initialUser?.name ?? '');
	let profileCurrency = $state(initialUser?.preferredCurrency ?? 'USD');
	let profileImage = $state(initialUser?.image ?? '');
	let profileLoading = $state(false);

	// Notifications section
	let emailAlerts = $state(initialPrefs.emailAlerts);
	let weeklySummary = $state(initialPrefs.weeklySummary);
	let prefsLoading = $state(false);

	// Security section
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordLoading = $state(false);

	// Feedback derived from form action result
	const profileSuccess = $derived(form?.action === 'updateProfile' && (form as { success?: boolean })?.success);
	const profileError = $derived(form?.action === 'updateProfile' ? (form as { error?: string })?.error ?? '' : '');
	const prefsSuccess = $derived(form?.action === 'updatePreferences' && (form as { success?: boolean })?.success);
	const prefsError = $derived(form?.action === 'updatePreferences' ? (form as { error?: string })?.error ?? '' : '');
	const passwordSuccess = $derived(form?.action === 'changePassword' && (form as { success?: boolean })?.success);
	const passwordError = $derived(form?.action === 'changePassword' ? (form as { error?: string })?.error ?? '' : '');

	$effect(() => {
		if (passwordSuccess) {
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		}
	});
</script>

<svelte:head>
	<title>Configuración - FINEXIA</title>
	<meta name="description" content="Preferencias de cuenta y notificaciones de FINEXIA" />
</svelte:head>

<PageHeader
	title="Configuración"
	subtitle="Ajusta preferencias de experiencia, alertas y seguridad de cuenta."
/>

<div class="settings-layout">
	<!-- Profile -->
	<Card variant="elevated" padding="none">
		<div class="section">
			<h2 class="section-title">Perfil</h2>
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
					<Input
						label="Correo electrónico"
						name="email"
						value={data.user?.email ?? ''}
						disabled
					/>
					<Input
						label="Moneda preferida"
						name="preferredCurrency"
						bind:value={profileCurrency}
						placeholder="USD"
					/>
					<Input
						label="URL de avatar"
						name="image"
						bind:value={profileImage}
						placeholder="https://..."
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
		</div>
	</Card>

	<!-- Notifications -->
	<Card variant="elevated" padding="none">
		<div class="section">
			<h2 class="section-title">Notificaciones</h2>
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
						<div>
							<p class="toggle-label">Alertas por correo</p>
							<p class="toggle-hint">Recibe notificaciones de actividad importante</p>
						</div>
						<Checkbox name="emailAlerts" bind:checked={emailAlerts} />
					</div>
					<div class="toggle-row">
						<div>
							<p class="toggle-label">Resumen semanal</p>
							<p class="toggle-hint">Un resumen de tu portafolio cada semana</p>
						</div>
						<Checkbox name="weeklySummary" bind:checked={weeklySummary} />
					</div>
				</div>
				{#if prefsError}
					<p class="feedback error">{prefsError}</p>
				{/if}
				{#if prefsSuccess}
					<p class="feedback success">Preferencias guardadas.</p>
				{/if}
				<div class="form-actions">
					<Button type="submit" loading={prefsLoading}>Guardar preferencias</Button>
				</div>
			</form>
		</div>
	</Card>

	<!-- Appearance -->
	<Card variant="elevated" padding="none">
		<div class="section">
			<h2 class="section-title">Apariencia</h2>
			<div class="appearance-info">
				<div class="theme-badge">
					<span class="theme-dot"></span>
					<span>Tema oscuro FINEXIA</span>
				</div>
				<p class="hint">
					La aplicación utiliza exclusivamente el tema premium oscuro con acentos dorados. No hay
					temas adicionales disponibles actualmente.
				</p>
			</div>
		</div>
	</Card>

	<!-- Security -->
	<Card variant="elevated" padding="none">
		<div class="section">
			<h2 class="section-title">Seguridad</h2>
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
		</div>
	</Card>
</div>

<style>
	.settings-layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	.section {
		padding: 1.5rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.form-fields {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-actions {
		margin-top: 1.5rem;
		display: flex;
		justify-content: flex-end;
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
		border-bottom: 1px solid rgba(212, 145, 42, 0.1);
	}

	.toggle-row:last-child {
		border-bottom: none;
	}

	.toggle-label {
		margin: 0 0 0.25rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
	}

	.toggle-hint {
		margin: 0;
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.appearance-info {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.theme-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.875rem;
		border-radius: 20px;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.25);
		font-size: 0.875rem;
		color: var(--amber);
		font-weight: 500;
		width: fit-content;
	}

	.theme-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--amber);
		flex-shrink: 0;
	}

	.hint {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
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
		.settings-layout {
			grid-template-columns: 1fr;
		}
	}
</style>
