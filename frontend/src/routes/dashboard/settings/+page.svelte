<script lang="ts">
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import { renderSVG } from 'uqr';
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

	// Two-factor authentication section
	let twoFaPassword = $state('');
	let twoFaSetupLoading = $state(false);
	let twoFaConfirmCode = $state('');
	let twoFaConfirmLoading = $state(false);
	let twoFaDisablePassword = $state('');
	let twoFaDisableCode = $state('');
	let twoFaDisableLoading = $state(false);
	let twoFaShowDisable = $state(false);
	let twoFaRegenPassword = $state('');
	let twoFaRegenCode = $state('');
	let twoFaRegenLoading = $state(false);
	let twoFaShowRegen = $state(false);

	// Sessions section
	let revokingSessionId = $state<string | null>(null);
	let revokeOthersLoading = $state(false);

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
	const twoFaSetupError = $derived(
		form?.action === 'setup2fa' ? ((form as { error?: string })?.error ?? '') : ''
	);
	const twoFaSetupData = $derived(
		form?.action === 'setup2fa' && (form as { success?: boolean })?.success
			? (form as { secret?: string; otpauthUrl?: string })
			: null
	);
	const twoFaEnableError = $derived(
		form?.action === 'enable2fa' ? ((form as { error?: string })?.error ?? '') : ''
	);
	const twoFaRecoveryCodes = $derived(
		(form?.action === 'enable2fa' || form?.action === 'regenerate2faCodes') &&
			(form as { success?: boolean })?.success
			? ((form as { recoveryCodes?: string[] })?.recoveryCodes ?? [])
			: []
	);
	const twoFaDisableError = $derived(
		form?.action === 'disable2fa' ? ((form as { error?: string })?.error ?? '') : ''
	);
	const twoFaDisableSuccess = $derived(
		form?.action === 'disable2fa' && (form as { success?: boolean })?.success
	);
	const twoFaRegenError = $derived(
		form?.action === 'regenerate2faCodes' ? ((form as { error?: string })?.error ?? '') : ''
	);
	// The QR is rendered locally from the otpauth URL; the secret never
	// touches a third-party service.
	const twoFaQrSvg = $derived(
		twoFaSetupData?.otpauthUrl ? renderSVG(twoFaSetupData.otpauthUrl) : ''
	);

	const sessionsError = $derived(
		form?.action === 'revokeSession' || form?.action === 'revokeOtherSessions'
			? ((form as { error?: string })?.error ?? '')
			: ''
	);
	const sessionsSuccess = $derived(
		(form?.action === 'revokeSession' || form?.action === 'revokeOtherSessions') &&
			(form as { success?: boolean })?.success
	);

	const otherSessionsCount = $derived((data.sessions ?? []).filter((s) => !s.current).length);

	function describeDevice(userAgent: string | null): string {
		if (!userAgent) return 'Dispositivo desconocido';
		const ua = userAgent.toLowerCase();

		let browser = 'Navegador desconocido';
		if (ua.includes('edg/')) browser = 'Edge';
		else if (ua.includes('opr/') || ua.includes('opera')) browser = 'Opera';
		else if (ua.includes('chrome')) browser = 'Chrome';
		else if (ua.includes('safari')) browser = 'Safari';
		else if (ua.includes('firefox')) browser = 'Firefox';

		let os = '';
		if (ua.includes('windows')) os = 'Windows';
		else if (ua.includes('android')) os = 'Android';
		else if (ua.includes('iphone') || ua.includes('ipad')) os = 'iOS';
		else if (ua.includes('mac os') || ua.includes('macintosh')) os = 'macOS';
		else if (ua.includes('linux')) os = 'Linux';

		return os ? `${browser} · ${os}` : browser;
	}

	const dateFormatter = new Intl.DateTimeFormat('es', {
		day: '2-digit',
		month: 'short',
		hour: '2-digit',
		minute: '2-digit'
	});

	function formatSessionDate(value: string): string {
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? '—' : dateFormatter.format(date);
	}

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

	// Never leave credentials sitting in the 2FA forms after they succeed.
	$effect(() => {
		if (twoFaSetupData) twoFaPassword = '';
	});
	$effect(() => {
		if (twoFaRecoveryCodes.length > 0) {
			twoFaConfirmCode = '';
			twoFaRegenPassword = '';
			twoFaRegenCode = '';
			twoFaShowRegen = false;
		}
	});
	$effect(() => {
		if (twoFaDisableSuccess) {
			twoFaDisablePassword = '';
			twoFaDisableCode = '';
			twoFaShowDisable = false;
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

	<!-- Two-factor authentication -->
	<Card variant="elevated" padding="none">
		<div class="section">
			<h2 class="section-title">Verificación en dos pasos (2FA)</h2>

			{#if twoFaRecoveryCodes.length > 0}
				<div class="twofa-recovery" role="status">
					<p class="twofa-recovery-title">Guarda tus códigos de recuperación</p>
					<p class="hint">
						Cada código funciona una sola vez y te permitirá entrar si pierdes acceso a tu
						aplicación de autenticación. No volverán a mostrarse.
					</p>
					<ul class="twofa-code-list">
						{#each twoFaRecoveryCodes as code (code)}
							<li class="twofa-code">{code}</li>
						{/each}
					</ul>
				</div>
			{/if}

			{#if twoFaSetupData}
				<!-- Step 2: scan the QR and confirm the first code -->
				<p class="hint twofa-intro">
					Escanea este código QR con tu aplicación de autenticación (Google Authenticator, Authy,
					1Password…) o ingresa la clave manualmente. Luego confirma con el código de 6 dígitos. La
					verificación no quedará activa hasta que confirmes.
				</p>
				<div class="twofa-setup">
					<div class="twofa-qr" aria-label="Código QR para la aplicación de autenticación">
						<!-- eslint-disable-next-line svelte/no-at-html-tags -- SVG generated locally by uqr from the otpauth URL -->
						{@html twoFaQrSvg}
					</div>
					<div class="twofa-setup-info">
						<p class="twofa-secret-label">Clave para ingreso manual</p>
						<code class="twofa-secret">{twoFaSetupData.secret}</code>
						<form
							method="POST"
							action="?/enable2fa"
							use:enhance={() => {
								twoFaConfirmLoading = true;
								return async ({ update }) => {
									await update({ reset: false });
									twoFaConfirmLoading = false;
								};
							}}
						>
							<div class="form-fields">
								<Input
									label="Código de verificación"
									name="code"
									placeholder="123456"
									autocomplete="one-time-code"
									bind:value={twoFaConfirmCode}
									required
								/>
							</div>
							{#if twoFaEnableError}
								<p class="feedback error">{twoFaEnableError}</p>
							{/if}
							<div class="form-actions">
								<Button type="submit" loading={twoFaConfirmLoading}>Confirmar y activar</Button>
							</div>
						</form>
					</div>
				</div>
			{:else if data.twoFactor?.enabled}
				<div class="twofa-status">
					<span class="twofa-badge enabled">Activada</span>
					<p class="hint">
						Tu cuenta pide un código del autenticador en cada inicio de sesión. Te quedan
						{data.twoFactor.recoveryCodesLeft} códigos de recuperación sin usar.
					</p>
				</div>

				{#if twoFaDisableSuccess}
					<p class="feedback success">La verificación en dos pasos fue desactivada.</p>
				{/if}

				<div class="twofa-actions">
					<button
						type="button"
						class="twofa-toggle"
						onclick={() => (twoFaShowRegen = !twoFaShowRegen)}
					>
						Regenerar códigos de recuperación
					</button>
					<button
						type="button"
						class="twofa-toggle danger"
						onclick={() => (twoFaShowDisable = !twoFaShowDisable)}
					>
						Desactivar 2FA
					</button>
				</div>

				{#if twoFaShowRegen}
					<form
						method="POST"
						action="?/regenerate2faCodes"
						class="twofa-subform"
						use:enhance={() => {
							twoFaRegenLoading = true;
							return async ({ update }) => {
								await update({ reset: false });
								twoFaRegenLoading = false;
							};
						}}
					>
						<div class="form-fields">
							<Input
								label="Contraseña actual"
								type="password"
								name="password"
								bind:value={twoFaRegenPassword}
								required
							/>
							<Input
								label="Código del autenticador o de recuperación"
								name="code"
								placeholder="123456"
								autocomplete="one-time-code"
								bind:value={twoFaRegenCode}
								required
							/>
						</div>
						{#if twoFaRegenError}
							<p class="feedback error">{twoFaRegenError}</p>
						{/if}
						<div class="form-actions">
							<Button type="submit" variant="secondary" loading={twoFaRegenLoading}>
								Regenerar códigos
							</Button>
						</div>
					</form>
				{/if}

				{#if twoFaShowDisable}
					<form
						method="POST"
						action="?/disable2fa"
						class="twofa-subform"
						use:enhance={() => {
							twoFaDisableLoading = true;
							return async ({ update }) => {
								await update({ reset: false });
								twoFaDisableLoading = false;
							};
						}}
					>
						<p class="hint">
							Para desactivar la verificación en dos pasos confirma tu contraseña y un código
							vigente del autenticador (o un código de recuperación).
						</p>
						<div class="form-fields">
							<Input
								label="Contraseña actual"
								type="password"
								name="password"
								bind:value={twoFaDisablePassword}
								required
							/>
							<Input
								label="Código del autenticador o de recuperación"
								name="code"
								placeholder="123456"
								autocomplete="one-time-code"
								bind:value={twoFaDisableCode}
								required
							/>
						</div>
						{#if twoFaDisableError}
							<p class="feedback error">{twoFaDisableError}</p>
						{/if}
						<div class="form-actions">
							<Button type="submit" variant="secondary" loading={twoFaDisableLoading}>
								Desactivar 2FA
							</Button>
						</div>
					</form>
				{/if}
			{:else}
				<div class="twofa-status">
					<span class="twofa-badge">Desactivada</span>
					<p class="hint">
						Añade una segunda barrera a tu cuenta: además de tu contraseña, se pedirá un código
						temporal de una aplicación de autenticación al iniciar sesión. Es opcional y puedes
						desactivarla cuando quieras.
					</p>
				</div>

				{#if twoFaDisableSuccess}
					<p class="feedback success">La verificación en dos pasos fue desactivada.</p>
				{/if}

				<form
					method="POST"
					action="?/setup2fa"
					use:enhance={() => {
						twoFaSetupLoading = true;
						return async ({ update }) => {
							await update({ reset: false });
							twoFaSetupLoading = false;
						};
					}}
				>
					<div class="form-fields">
						<Input
							label="Contraseña actual"
							type="password"
							name="password"
							bind:value={twoFaPassword}
							required
						/>
					</div>
					{#if twoFaSetupError}
						<p class="feedback error">{twoFaSetupError}</p>
					{/if}
					<div class="form-actions">
						<Button type="submit" loading={twoFaSetupLoading}>Activar 2FA</Button>
					</div>
				</form>
			{/if}
		</div>
	</Card>

	<!-- Active sessions -->
	<Card variant="elevated" padding="none">
		<div class="section">
			<h2 class="section-title">Sesiones activas</h2>
			<p class="hint sessions-intro">
				Estos son los dispositivos con acceso a tu cuenta y a la información de tu patrimonio.
				Cierra cualquier sesión que no reconozcas.
			</p>

			{#if (data.sessions ?? []).length === 0}
				<p class="hint">No se pudieron cargar las sesiones activas.</p>
			{:else}
				<ul class="session-list">
					{#each data.sessions as session (session.id)}
						<li class="session-item">
							<div class="session-info">
								<div class="session-device">
									<span class="session-name">{describeDevice(session.userAgent)}</span>
									{#if session.current}
										<span class="session-badge">Este dispositivo</span>
									{/if}
								</div>
								<p class="session-meta">
									{session.ipAddress ?? 'IP desconocida'} · Última actividad: {formatSessionDate(
										session.lastActiveAt
									)}
								</p>
							</div>
							{#if !session.current}
								<form
									method="POST"
									action="?/revokeSession"
									use:enhance={() => {
										revokingSessionId = session.id;
										return async ({ update }) => {
											await update();
											revokingSessionId = null;
										};
									}}
								>
									<input type="hidden" name="sessionId" value={session.id} />
									<button
										type="submit"
										class="btn-revoke"
										disabled={revokingSessionId === session.id}
									>
										{revokingSessionId === session.id ? 'Cerrando…' : 'Cerrar sesión'}
									</button>
								</form>
							{/if}
						</li>
					{/each}
				</ul>

				{#if sessionsError}
					<p class="feedback error">{sessionsError}</p>
				{/if}
				{#if sessionsSuccess}
					<p class="feedback success">Sesión cerrada correctamente.</p>
				{/if}

				{#if otherSessionsCount > 0}
					<div class="form-actions">
						<form
							method="POST"
							action="?/revokeOtherSessions"
							use:enhance={() => {
								revokeOthersLoading = true;
								return async ({ update }) => {
									await update();
									revokeOthersLoading = false;
								};
							}}
						>
							<Button type="submit" variant="secondary" loading={revokeOthersLoading}>
								{revokeOthersLoading
									? 'Cerrando sesiones…'
									: `Cerrar las demás sesiones (${otherSessionsCount})`}
							</Button>
						</form>
					</div>
				{/if}
			{/if}
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

	/* Two-factor authentication */
	.twofa-status {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-bottom: 1.25rem;
	}

	.twofa-badge {
		width: fit-content;
		font-size: 0.675rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.55);
		background: rgba(236, 234, 229, 0.06);
		border: 1px solid rgba(236, 234, 229, 0.2);
		border-radius: 20px;
		padding: 0.2rem 0.65rem;
	}

	.twofa-badge.enabled {
		color: #4ade80;
		background: rgba(74, 222, 128, 0.08);
		border-color: rgba(74, 222, 128, 0.3);
	}

	.twofa-intro {
		margin-bottom: 1.25rem;
	}

	.twofa-setup {
		display: flex;
		gap: 1.25rem;
		align-items: flex-start;
		flex-wrap: wrap;
	}

	.twofa-qr {
		flex-shrink: 0;
		width: 148px;
		height: 148px;
		padding: 8px;
		background: #fff;
		border-radius: 8px;
	}

	.twofa-qr :global(svg) {
		width: 100%;
		height: 100%;
		display: block;
	}

	.twofa-setup-info {
		flex: 1;
		min-width: 220px;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.twofa-secret-label {
		margin: 0;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.45);
	}

	.twofa-secret {
		font-family: var(--font-mono);
		font-size: 0.8rem;
		color: var(--amber);
		background: rgba(212, 145, 42, 0.08);
		border: 1px solid rgba(212, 145, 42, 0.25);
		border-radius: 6px;
		padding: 0.5rem 0.75rem;
		overflow-wrap: anywhere;
	}

	.twofa-recovery {
		margin-bottom: 1.5rem;
		padding: 1rem;
		border: 1px solid rgba(212, 145, 42, 0.35);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.06);
	}

	.twofa-recovery-title {
		margin: 0 0 0.5rem;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--amber);
	}

	.twofa-code-list {
		list-style: none;
		margin: 0.875rem 0 0;
		padding: 0;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
		gap: 0.5rem;
	}

	.twofa-code {
		font-family: var(--font-mono);
		font-size: 0.8rem;
		color: var(--text);
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(212, 145, 42, 0.15);
		border-radius: 6px;
		padding: 0.4rem 0.6rem;
		text-align: center;
	}

	.twofa-actions {
		display: flex;
		gap: 0.625rem;
		flex-wrap: wrap;
		margin-top: 0.5rem;
	}

	.twofa-toggle {
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

	.twofa-toggle:hover {
		background: rgba(212, 145, 42, 0.15);
		border-color: rgba(212, 145, 42, 0.65);
	}

	.twofa-toggle.danger {
		border-color: rgba(224, 90, 90, 0.35);
		background: rgba(224, 90, 90, 0.06);
		color: var(--red, #e05a5a);
	}

	.twofa-toggle.danger:hover {
		background: rgba(224, 90, 90, 0.14);
		border-color: rgba(224, 90, 90, 0.6);
	}

	.twofa-subform {
		margin-top: 1.25rem;
		padding-top: 1.25rem;
		border-top: 1px solid rgba(212, 145, 42, 0.1);
	}

	.twofa-subform .hint {
		margin-bottom: 1rem;
	}

	/* Sessions */
	.sessions-intro {
		margin-bottom: 1.25rem;
	}

	.session-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.session-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.875rem 1rem;
		border: 1px solid rgba(212, 145, 42, 0.12);
		border-radius: 8px;
		background: var(--surface-2, rgba(255, 255, 255, 0.02));
	}

	.session-info {
		min-width: 0;
	}

	.session-device {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.session-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
	}

	.session-badge {
		font-size: 0.675rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--amber);
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 20px;
		padding: 0.15rem 0.55rem;
	}

	.session-meta {
		margin: 0.3rem 0 0;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.45);
		font-family: var(--font-mono);
		overflow-wrap: anywhere;
	}

	.btn-revoke {
		flex-shrink: 0;
		padding: 0.4rem 0.875rem;
		border-radius: 6px;
		border: 1px solid rgba(224, 90, 90, 0.35);
		background: rgba(224, 90, 90, 0.06);
		color: var(--red, #e05a5a);
		font-size: 0.775rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.btn-revoke:hover:not(:disabled) {
		background: rgba(224, 90, 90, 0.14);
		border-color: rgba(224, 90, 90, 0.6);
	}

	.btn-revoke:disabled {
		opacity: 0.6;
		cursor: default;
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
