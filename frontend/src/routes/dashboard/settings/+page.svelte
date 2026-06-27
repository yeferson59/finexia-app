<script lang="ts">
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Input from '$components/ui/input.svelte';
	import Button from '$components/ui/button.svelte';

	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	// Profile section — seeded from server once; user edits locally
	let profileName = $state(untrack(() => data.user?.name ?? ''));
	let profileCurrency = $state(untrack(() => data.user?.preferredCurrency ?? 'USD'));
	let profileLoading = $state(false);

	// Avatar section
	let avatarPreview = $state<string | null>(null);
	let avatarFile = $state<File | null>(null);
	let avatarLoading = $state(false);
	let avatarFileInput = $state<HTMLInputElement | null>(null);

	// Security section
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordLoading = $state(false);

	// Form action feedback
	const profileSuccess = $derived(
		form?.action === 'updateProfile' && (form as { success?: boolean })?.success
	);
	const profileError = $derived(
		form?.action === 'updateProfile' ? ((form as { error?: string })?.error ?? '') : ''
	);
	const avatarSuccess = $derived(
		form?.action === 'uploadAvatar' && (form as { success?: boolean })?.success
	);
	const avatarError = $derived(
		form?.action === 'uploadAvatar' ? ((form as { error?: string })?.error ?? '') : ''
	);
	const passwordSuccess = $derived(
		form?.action === 'changePassword' && (form as { success?: boolean })?.success
	);
	const passwordError = $derived(
		form?.action === 'changePassword' ? ((form as { error?: string })?.error ?? '') : ''
	);

	// Avatar URL: prefer the uploaded URL returned by the server action, then the stored image
	const savedAvatarUrl = $derived(
		avatarSuccess
			? ((form as { imageUrl?: string })?.imageUrl ?? data.user?.image ?? '')
			: (data.user?.image ?? '')
	);

	const displayAvatar = $derived(
		avatarPreview ?? (savedAvatarUrl && savedAvatarUrl !== 'avatar.png' ? savedAvatarUrl : null)
	);

	const userInitial = $derived((data.user?.name ?? '').trim().charAt(0).toUpperCase());

	$effect(() => {
		if (passwordSuccess) {
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		}
	});

	$effect(() => {
		if (avatarSuccess) {
			avatarPreview = null;
			avatarFile = null;
		}
	});

	function compressImageForAvatar(file: File): Promise<File> {
		const MAX = 512;
		const QUALITY = 0.85;
		const outputType = file.type === 'image/webp' ? 'image/webp' : 'image/jpeg';

		return new Promise((resolve) => {
			const img = new Image();
			const url = URL.createObjectURL(file);
			img.onload = () => {
				URL.revokeObjectURL(url);
				let { width, height } = img;
				if (width > MAX || height > MAX) {
					const ratio = Math.min(MAX / width, MAX / height);
					width = Math.round(width * ratio);
					height = Math.round(height * ratio);
				}
				const canvas = document.createElement('canvas');
				canvas.width = width;
				canvas.height = height;
				canvas.getContext('2d')!.drawImage(img, 0, 0, width, height);
				canvas.toBlob(
					(blob) => {
						if (!blob) return resolve(file);
						const ext = outputType === 'image/webp' ? '.webp' : '.jpg';
						resolve(new File([blob], 'avatar' + ext, { type: outputType }));
					},
					outputType,
					QUALITY
				);
			};
			img.onerror = () => {
				URL.revokeObjectURL(url);
				resolve(file);
			};
			img.src = url;
		});
	}

	async function onFileChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		if (avatarPreview) URL.revokeObjectURL(avatarPreview);
		const compressed = await compressImageForAvatar(file);
		// Replace input.files so the form submission sends the compressed file
		const dt = new DataTransfer();
		dt.items.add(compressed);
		input.files = dt.files;
		avatarFile = compressed;
		avatarPreview = URL.createObjectURL(compressed);
	}
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

			<!-- Avatar upload -->
			<div class="avatar-section">
				<div class="avatar-display">
					{#if displayAvatar}
						<img src={displayAvatar} alt="Avatar de usuario" class="avatar-img" />
					{:else}
						<div class="avatar-initials" aria-hidden="true">{userInitial}</div>
					{/if}
				</div>
				<div class="avatar-controls">
					<form
						method="POST"
						action="?/uploadAvatar"
						enctype="multipart/form-data"
						use:enhance={() => {
							avatarLoading = true;
							return async ({ update }) => {
								await update();
								avatarLoading = false;
							};
						}}
					>
						<input
							bind:this={avatarFileInput}
							type="file"
							name="avatar"
							accept="image/jpeg,image/png,image/webp"
							class="file-input-hidden"
							onchange={onFileChange}
						/>
						<button type="button" class="btn-pick-file" onclick={() => avatarFileInput?.click()}>
							Cambiar foto
						</button>
						{#if avatarFile}
							<Button type="submit" loading={avatarLoading}>
								{avatarLoading ? 'Subiendo…' : 'Guardar foto'}
							</Button>
						{/if}
					</form>
					<p class="avatar-hint">JPEG, PNG o WebP · se optimiza automáticamente</p>
					{#if avatarError}
						<p class="feedback error">{avatarError}</p>
					{/if}
					{#if avatarSuccess}
						<p class="feedback success">Foto actualizada correctamente.</p>
					{/if}
				</div>
			</div>

			<!-- Profile fields -->
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
					<Input label="Correo electrónico" name="email" value={data.user?.email ?? ''} disabled />
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

	/* Avatar */
	.avatar-section {
		display: flex;
		align-items: center;
		gap: 1.25rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 145, 42, 0.1);
	}

	.avatar-display {
		flex-shrink: 0;
	}

	.avatar-img {
		width: 64px;
		height: 64px;
		border-radius: 50%;
		object-fit: cover;
		border: 2px solid var(--border-strong);
	}

	.avatar-initials {
		width: 64px;
		height: 64px;
		border-radius: 50%;
		background: var(--surface-3);
		border: 2px solid var(--border-strong);
		color: var(--amber);
		font-family: var(--font-mono);
		font-size: 1.35rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.avatar-controls {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.avatar-controls form {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		flex-wrap: wrap;
	}

	.file-input-hidden {
		display: none;
	}

	.btn-pick-file {
		padding: 0.45rem 1rem;
		border-radius: 6px;
		border: 1px solid rgba(212, 145, 42, 0.4);
		background: rgba(212, 145, 42, 0.08);
		color: var(--amber);
		font-size: 0.825rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.btn-pick-file:hover {
		background: rgba(212, 145, 42, 0.15);
		border-color: rgba(212, 145, 42, 0.65);
	}

	.avatar-hint {
		margin: 0;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.4);
	}

	/* Form */
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
