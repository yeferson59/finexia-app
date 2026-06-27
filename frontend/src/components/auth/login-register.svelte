<script lang="ts">
	import { enhance } from '$app/forms';
	import Button from '$components/ui/button.svelte';
	import Input from '$components/ui/input.svelte';
	import Checkbox from '$components/ui/checkbox.svelte';
	type FormResult = {
		type: 'login' | 'register';
		errors: Record<string, string> | Array<{ path: PropertyKey[]; message: string }>;
	} | null;

	let { form }: { form: FormResult } = $props();

	function parseErrors(errors: unknown): Record<string, string> {
		if (!errors) return {};
		if (Array.isArray(errors)) {
			const result: Record<string, string> = {};
			for (const issue of errors) {
				const key = issue?.path?.[0] ? String(issue.path[0]) : 'server';
				if (!result[key]) result[key] = issue.message ?? 'Campo inválido';
			}
			return result;
		}
		if (typeof errors === 'object') return errors as Record<string, string>;
		return {};
	}

	// Form state
	let isLoginMode = $state(true);
	let slideDirection = $state<'left' | 'right'>('right');
	let showPassword = $state(false);
	let showConfirmPassword = $state(false);
	let isSubmitting = $state(false);

	// Login form
	let loginEmail = $state('');
	let loginPassword = $state('');
	let loginErrors: Record<string, string> = $state({});

	// Register form
	let registerEmail = $state('');
	let registerPassword = $state('');
	let registerConfirmPassword = $state('');
	let registerName = $state('');
	let agreeTerms = $state(false);
	let registerErrors: Record<string, string> = $state({});

	// Sync server-side form errors into local state
	$effect(() => {
		if (!form) return;
		if (form.type === 'login') {
			loginErrors = parseErrors(form.errors);
		} else if (form.type === 'register') {
			registerErrors = parseErrors(form.errors);
		}
	});

	const switchToLogin = () => {
		if (isLoginMode) return;
		slideDirection = 'left';
		isLoginMode = true;
		loginErrors = {};
		registerErrors = {};
	};

	const switchToRegister = () => {
		if (!isLoginMode) return;
		slideDirection = 'right';
		isLoginMode = false;
		loginErrors = {};
		registerErrors = {};
	};
</script>

<main class="auth-container">
	<!-- Left: Brand Panel (desktop only) -->
	<aside class="brand-panel">
		<div class="brand-content">
			<div class="brand-logo">
				<div class="logo-mark" aria-hidden="true">
					<svg width="28" height="28" viewBox="0 0 30 30" fill="none">
						<path
							d="M7 22L12.5 14.5L16.5 18.5L23 9"
							stroke="#0c0a06"
							stroke-width="2.6"
							stroke-linecap="round"
							stroke-linejoin="round"
						/>
					</svg>
				</div>
				<span class="brand-name">FINEXIA</span>
			</div>

			<p class="brand-tagline">Tu patrimonio,<br />bajo control.</p>

			<div class="brand-chart" aria-hidden="true">
				<svg viewBox="0 0 320 160" class="chart-svg" fill="none" xmlns="http://www.w3.org/2000/svg">
					<!-- Subtle grid lines -->
					<line x1="0" y1="40" x2="320" y2="40" stroke="rgba(255,255,255,0.04)" stroke-width="1" />
					<line x1="0" y1="80" x2="320" y2="80" stroke="rgba(255,255,255,0.04)" stroke-width="1" />
					<line
						x1="0"
						y1="120"
						x2="320"
						y2="120"
						stroke="rgba(255,255,255,0.04)"
						stroke-width="1"
					/>
					<!-- Vertical tick marks -->
					<line x1="80" y1="0" x2="80" y2="160" stroke="rgba(255,255,255,0.025)" stroke-width="1" />
					<line
						x1="160"
						y1="0"
						x2="160"
						y2="160"
						stroke="rgba(255,255,255,0.025)"
						stroke-width="1"
					/>
					<line
						x1="240"
						y1="0"
						x2="240"
						y2="160"
						stroke="rgba(255,255,255,0.025)"
						stroke-width="1"
					/>

					<!-- Area fill under the line -->
					<path
						class="chart-area"
						d="M0,140 C30,130 55,118 80,105 C110,90 130,88 155,72 C180,56 205,48 230,35 C255,22 280,14 320,8 L320,160 L0,160 Z"
						fill="url(#area-grad)"
					/>

					<!-- Portfolio trend line -->
					<path
						class="chart-line"
						d="M0,140 C30,130 55,118 80,105 C110,90 130,88 155,72 C180,56 205,48 230,35 C255,22 280,14 320,8"
						stroke="url(#chart-grad)"
						stroke-width="2.5"
						stroke-linecap="round"
					/>

					<!-- End dot -->
					<circle class="chart-dot" cx="320" cy="8" r="4.5" fill="var(--amber)" />
					<circle
						class="chart-dot-ring"
						cx="320"
						cy="8"
						r="8"
						fill="none"
						stroke="var(--amber)"
						stroke-width="1"
					/>

					<defs>
						<linearGradient id="chart-grad" x1="0" y1="0" x2="1" y2="0">
							<stop offset="0%" stop-color="rgba(212,145,42,0.2)" />
							<stop offset="100%" stop-color="rgba(212,145,42,0.9)" />
						</linearGradient>
						<linearGradient id="area-grad" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="rgba(212,145,42,0.08)" />
							<stop offset="100%" stop-color="rgba(212,145,42,0)" />
						</linearGradient>
					</defs>
				</svg>
			</div>

			<p class="brand-footnote">Visualiza, analiza y optimiza<br />tu riqueza en tiempo real.</p>
		</div>
	</aside>

	<!-- Right: Auth Panel -->
	<div class="auth-panel">
		<div class="auth-card">
			<!-- Header: mobile only -->
			<header class="auth-header">
				<div class="logo-container">
					<div class="logo-mark" aria-hidden="true">
						<svg width="32" height="32" viewBox="0 0 30 30" fill="none">
							<path
								d="M7 22L12.5 14.5L16.5 18.5L23 9"
								stroke="#0c0a06"
								stroke-width="2.6"
								stroke-linecap="round"
								stroke-linejoin="round"
							/>
						</svg>
					</div>
				</div>
				<div class="header-text">
					<h1 class="auth-title">FINEXIA</h1>
					<p class="auth-subtitle">Tu patrimonio, bajo control</p>
				</div>
			</header>

			<!-- Mode Toggle -->
			<div class="mode-toggle" role="tablist">
				<span class="toggle-slide" class:register={!isLoginMode}></span>
				<button
					role="tab"
					aria-selected={isLoginMode}
					class:active={isLoginMode}
					onclick={switchToLogin}
					aria-controls="login-form"
				>
					Iniciar sesión
				</button>
				<button
					role="tab"
					aria-selected={!isLoginMode}
					class:active={!isLoginMode}
					onclick={switchToRegister}
					aria-controls="register-form"
				>
					Crear cuenta
				</button>
			</div>

			<!-- Forms Section -->
			<section class="forms-container">
				{#if isLoginMode}
					<form
						method="POST"
						action="?/login"
						class="form-content"
						class:slide-left={slideDirection === 'left'}
						id="login-form"
						use:enhance={() => {
							isSubmitting = true;
							loginErrors = {};
							return async ({ update }) => {
								await update({ reset: false });
								isSubmitting = false;
							};
						}}
					>
						<Input
							label="Email"
							id="login-email"
							name="email"
							type="email"
							placeholder="tu@email.com"
							bind:value={loginEmail}
							error={loginErrors['email']}
							required
						/>

						<div class="password-wrapper">
							<Input
								label="Contraseña"
								id="login-password"
								name="password"
								type={showPassword ? 'text' : 'password'}
								placeholder="Ingresa tu contraseña"
								bind:value={loginPassword}
								error={loginErrors['password']}
								required
							/>
							<button
								type="button"
								class="password-toggle"
								onclick={() => (showPassword = !showPassword)}
								aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
							>
								{#if showPassword}
									<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
										<circle cx="12" cy="12" r="3"></circle>
									</svg>
								{:else}
									<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path
											d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1 4.24 4.24"
										></path>
										<line x1="1" y1="1" x2="23" y2="23"></line>
									</svg>
								{/if}
							</button>
						</div>

						<div class="form-footer">
							<a href="#forgot" class="forgot-link">¿Olvidaste tu contraseña?</a>
						</div>

						{#if loginErrors['server']}
							<p class="error-server" role="alert">{loginErrors['server']}</p>
						{/if}

						<Button variant="primary" size="lg" loading={isSubmitting} fullWidth>
							{isSubmitting ? 'Iniciando sesión...' : 'Iniciar sesión'}
						</Button>

						<div class="form-switch">
							¿No tienes cuenta?
							<button type="button" onclick={switchToRegister} class="switch-link">
								Crear una
							</button>
						</div>
					</form>
				{:else}
					<form
						method="POST"
						action="?/register"
						class="form-content"
						class:slide-left={slideDirection === 'left'}
						id="register-form"
						use:enhance={() => {
							isSubmitting = true;
							registerErrors = {};
							return async ({ update }) => {
								await update({ reset: false });
								isSubmitting = false;
							};
						}}
					>
						<Input
							label="Nombre completo"
							id="register-name"
							name="name"
							type="text"
							placeholder="Juan Pérez"
							bind:value={registerName}
							error={registerErrors['name']}
							required
						/>

						<Input
							label="Email"
							id="register-email"
							name="email"
							type="email"
							placeholder="tu@email.com"
							bind:value={registerEmail}
							error={registerErrors['email']}
							required
						/>

						<div class="password-wrapper">
							<Input
								label="Contraseña"
								id="register-password"
								name="password"
								type={showPassword ? 'text' : 'password'}
								placeholder="Crea una contraseña segura"
								bind:value={registerPassword}
								error={registerErrors['password']}
								required
							/>
							<button
								type="button"
								class="password-toggle"
								onclick={() => (showPassword = !showPassword)}
								aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
							>
								{#if showPassword}
									<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
										<circle cx="12" cy="12" r="3"></circle>
									</svg>
								{:else}
									<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path
											d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1 4.24 4.24"
										></path>
										<line x1="1" y1="1" x2="23" y2="23"></line>
									</svg>
								{/if}
							</button>
						</div>

						<div class="password-wrapper">
							<Input
								label="Confirmar contraseña"
								id="register-confirm"
								name="confirmPassword"
								type={showConfirmPassword ? 'text' : 'password'}
								placeholder="Repite tu contraseña"
								bind:value={registerConfirmPassword}
								error={registerErrors['confirmPassword']}
								required
							/>
							<button
								type="button"
								class="password-toggle"
								onclick={() => (showConfirmPassword = !showConfirmPassword)}
								aria-label={showConfirmPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
							>
								{#if showConfirmPassword}
									<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
										<circle cx="12" cy="12" r="3"></circle>
									</svg>
								{:else}
									<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path
											d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1 4.24 4.24"
										></path>
										<line x1="1" y1="1" x2="23" y2="23"></line>
									</svg>
								{/if}
							</button>
						</div>

						<Checkbox
							id="terms"
							name="terms"
							label="Acepto los términos y condiciones"
							bind:checked={agreeTerms}
						/>
						{#if registerErrors['terms']}
							<span class="error-message">{registerErrors['terms']}</span>
						{/if}

						{#if registerErrors['server']}
							<p class="error-server" role="alert">{registerErrors['server']}</p>
						{/if}

						<Button variant="primary" size="lg" loading={isSubmitting} fullWidth>
							{isSubmitting ? 'Creando cuenta...' : 'Crear cuenta'}
						</Button>

						<div class="form-switch">
							¿Ya tienes cuenta?
							<button type="button" onclick={switchToLogin} class="switch-link">
								Inicia sesión
							</button>
						</div>
					</form>
				{/if}
			</section>

			<!-- Social Divider -->
			<div class="divider">
				<span>o continúa con</span>
			</div>

			<!-- Social Auth -->
			<footer class="social-auth">
				<button type="button" class="social-button" aria-label="Iniciar sesión con Google">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
						<path
							d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
						/>
						<path
							d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
						/>
						<path
							d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
						/>
						<path
							d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
						/>
					</svg>
					Google
				</button>
				<button type="button" class="social-button" aria-label="Iniciar sesión con GitHub">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
						<path
							d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
						/>
					</svg>
					GitHub
				</button>
			</footer>
		</div>
	</div>
</main>

<style>
	/* ── Layout ──────────────────────────────────────────────── */
	main.auth-container {
		--gold-primary: var(--amber);
		--gold-light: var(--amber-light);
		--text-primary: var(--text);
		--text-secondary: rgba(236, 234, 229, 0.6);
		--error-color: var(--red);

		display: grid;
		grid-template-columns: 1fr 1fr;
		min-height: 100dvh;
	}

	/* ── Brand Panel ─────────────────────────────────────────── */
	.brand-panel {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(3rem, 6vw, 5rem) clamp(2.5rem, 5vw, 4.5rem);
		background: #07080a;
		border-right: 1px solid var(--border);
		position: relative;
		overflow: hidden;
	}

	/* Ambient glow top-right */
	.brand-panel::before {
		content: '';
		position: absolute;
		top: -20%;
		right: -10%;
		width: 420px;
		height: 420px;
		border-radius: 50%;
		background: radial-gradient(circle, rgba(212, 145, 42, 0.06) 0%, transparent 65%);
		pointer-events: none;
	}

	.brand-content {
		position: relative;
		z-index: 1;
		width: 100%;
		max-width: 420px;
	}

	.brand-logo {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 2.5rem;
	}

	.brand-name {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 3rem);
		font-weight: 600;
		letter-spacing: 0.12em;
		color: var(--text);
		line-height: 1;
	}

	.brand-tagline {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.8vw, 2.25rem);
		font-weight: 300;
		color: var(--text-muted);
		line-height: 1.4;
		margin: 0;
		letter-spacing: 0.01em;
	}

	.brand-chart {
		margin-top: 3rem;
		margin-bottom: 2rem;
	}

	.chart-svg {
		width: 100%;
		overflow: visible;
	}

	.chart-area {
		opacity: 0;
		animation: fade-area 0.6s 2s ease forwards;
	}

	.chart-line {
		stroke-dasharray: 700;
		stroke-dashoffset: 700;
		animation: draw-chart 1.8s 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	.chart-dot {
		opacity: 0;
		animation: fade-dot 0.4s 2.1s ease forwards;
	}

	.chart-dot-ring {
		opacity: 0;
		animation: pulse-ring 1.5s 2.3s ease-out infinite;
	}

	@keyframes draw-chart {
		to {
			stroke-dashoffset: 0;
		}
	}
	@keyframes fade-area {
		to {
			opacity: 1;
		}
	}
	@keyframes fade-dot {
		to {
			opacity: 1;
		}
	}
	@keyframes pulse-ring {
		0% {
			opacity: 0.6;
			transform: scale(1);
			transform-origin: 320px 8px;
		}
		100% {
			opacity: 0;
			transform: scale(2.2);
			transform-origin: 320px 8px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.chart-line {
			animation: none;
			stroke-dashoffset: 0;
		}
		.chart-area,
		.chart-dot {
			animation: none;
			opacity: 1;
		}
		.chart-dot-ring {
			animation: none;
			opacity: 0;
		}
	}

	.brand-footnote {
		font-size: 0.8rem;
		color: var(--text-dim);
		line-height: 1.6;
		letter-spacing: 0.3px;
		margin: 0;
	}

	/* ── Auth Panel ──────────────────────────────────────────── */
	.auth-panel {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(1.5rem, 4vw, 3rem);
		background: linear-gradient(160deg, #08090a 0%, #0a0b0d 100%);
		position: relative;
		overflow-y: auto;
	}

	.auth-panel::before {
		content: '';
		position: absolute;
		bottom: -20%;
		left: -10%;
		width: 400px;
		height: 400px;
		border-radius: 50%;
		background: radial-gradient(circle, rgba(34, 201, 126, 0.03) 0%, transparent 65%);
		pointer-events: none;
	}

	.auth-card {
		width: 100%;
		max-width: 440px;
		position: relative;
		z-index: 10;
	}

	/* ── Card Header (mobile only) ───────────────────────────── */
	.auth-header {
		display: none; /* shown on mobile */
		flex-direction: column;
		align-items: center;
		gap: 1.25rem;
		margin-bottom: 2.5rem;
		text-align: center;
	}

	.logo-container {
		position: relative;
	}

	.logo-mark {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 52px;
		height: 52px;
		border-radius: 13px;
		background: var(--amber);
		border: 1px solid rgba(232, 165, 53, 0.4);
		box-shadow:
			0 0 25px rgba(212, 145, 42, 0.25),
			inset 0 1px 2px rgba(255, 255, 255, 0.2);
		flex-shrink: 0;
	}

	/* Brand panel logo is smaller */
	.brand-logo .logo-mark {
		width: 44px;
		height: 44px;
		border-radius: 11px;
		animation: logo-float 3s ease-in-out infinite;
	}

	@keyframes logo-float {
		0%,
		100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-4px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.brand-logo .logo-mark {
			animation: none;
		}
	}

	.header-text {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.auth-title {
		font-size: clamp(1.5rem, 5vw, 2rem);
		font-weight: 600;
		letter-spacing: 0.1em;
		color: var(--text);
		font-family: var(--font-display);
		margin: 0;
	}

	.auth-subtitle {
		font-size: 0.9rem;
		color: var(--text-secondary);
		letter-spacing: 0.8px;
		font-weight: 400;
		margin: 0;
	}

	/* ── Mode Toggle ─────────────────────────────────────────── */
	.mode-toggle {
		position: relative;
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0;
		margin-bottom: 2.5rem;
		background: rgba(0, 0, 0, 0.3);
		border-radius: 12px;
		padding: 4px;
		border: 1px solid var(--border);
	}

	.toggle-slide {
		position: absolute;
		left: 4px;
		top: 4px;
		width: calc(50% - 4px);
		height: calc(100% - 8px);
		background: rgba(212, 145, 42, 0.11);
		border: 1px solid rgba(212, 145, 42, 0.28);
		border-radius: 9px;
		box-shadow: 0 2px 10px rgba(0, 0, 0, 0.25);
		pointer-events: none;
		transition: transform 0.35s cubic-bezier(0.34, 1.3, 0.64, 1);
	}

	.toggle-slide.register {
		transform: translateX(100%);
	}

	.mode-toggle button {
		position: relative;
		z-index: 1;
		padding: 0.875rem 1.25rem;
		background: transparent;
		border: none;
		color: var(--text-secondary);
		font-size: 0.9rem;
		font-weight: 600;
		font-family: var(--font-body);
		border-radius: 9px;
		cursor: pointer;
		transition:
			color 0.25s ease,
			opacity 0.25s ease;
		letter-spacing: 0.4px;
		text-transform: uppercase;
	}

	.mode-toggle button.active {
		color: var(--gold-primary);
	}

	.mode-toggle button:hover:not(.active) {
		color: rgba(236, 234, 229, 0.75);
	}

	/* ── Forms ───────────────────────────────────────────────── */
	.forms-container {
		margin-bottom: 0.5rem;
	}

	.form-content {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		animation: slide-in-right 0.32s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	.form-content.slide-left {
		animation: slide-in-left 0.32s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	@keyframes slide-in-right {
		from {
			opacity: 0;
			transform: translateX(14px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes slide-in-left {
		from {
			opacity: 0;
			transform: translateX(-14px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	.password-wrapper {
		position: relative;
		width: 100%;
	}

	.password-toggle {
		position: absolute;
		right: 0.75rem;
		top: 50%;
		margin-top: 1.125rem;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		transition: all 0.25s ease;
		padding: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		z-index: 10;
	}

	.password-toggle:hover {
		color: var(--gold-primary);
		transform: translateY(calc(-50% + 1.125rem)) scale(1.1);
	}

	.password-toggle svg {
		stroke-width: 2;
	}

	.form-footer {
		display: flex;
		justify-content: flex-end;
		margin-top: -0.75rem;
	}

	.forgot-link {
		font-size: 0.8rem;
		color: var(--gold-primary);
		text-decoration: none;
		transition: color 0.25s ease;
		font-weight: 500;
		letter-spacing: 0.3px;
	}

	.forgot-link:hover {
		color: var(--gold-light);
		text-decoration: underline;
	}

	.form-switch {
		text-align: center;
		font-size: 0.9rem;
		color: var(--text-secondary);
		letter-spacing: 0.3px;
		margin-top: 0.5rem;
	}

	.switch-link {
		background: none;
		border: none;
		color: var(--gold-primary);
		cursor: pointer;
		font-weight: 700;
		font-family: var(--font-body);
		transition: color 0.25s ease;
		padding: 0;
		letter-spacing: 0.4px;
		font-size: 0.9rem;
	}

	.switch-link:hover {
		color: var(--gold-light);
		text-decoration: underline;
	}

	.error-message {
		font-size: 0.75rem;
		color: var(--error-color);
		margin-top: -1rem;
		letter-spacing: 0.2px;
		font-weight: 500;
	}

	.error-server {
		font-size: 0.85rem;
		color: var(--error-color);
		background: rgba(var(--red-rgb, 220, 53, 69), 0.08);
		border: 1px solid rgba(var(--red-rgb, 220, 53, 69), 0.25);
		border-radius: 8px;
		padding: 0.75rem 1rem;
		margin: 0;
		text-align: center;
		font-weight: 500;
		letter-spacing: 0.2px;
	}

	/* ── Divider + Social ────────────────────────────────────── */
	.divider {
		display: flex;
		align-items: center;
		gap: 1.25rem;
		margin: 2rem 0;
		color: var(--text-secondary);
		font-size: 0.8rem;
		letter-spacing: 0.3px;
		text-transform: uppercase;
		font-weight: 500;
	}

	.divider::before,
	.divider::after {
		content: '';
		flex: 1;
		height: 1px;
		background: var(--border);
	}

	.social-auth {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.social-button {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 0.95rem 1.25rem;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 10px;
		color: var(--text-primary);
		font-size: 0.85rem;
		font-weight: 600;
		font-family: var(--font-body);
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		letter-spacing: 0.3px;
		text-transform: uppercase;
		position: relative;
		overflow: hidden;
	}

	.social-button::before {
		content: '';
		position: absolute;
		top: 0;
		left: -100%;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, transparent, var(--border-strong), transparent);
		transition: left 0.5s ease;
	}

	.social-button:hover {
		background: rgba(212, 145, 42, 0.1);
		border-color: rgba(212, 145, 42, 0.25);
		color: var(--gold-primary);
		transform: translateY(-2px);
	}

	.social-button:hover::before {
		left: 100%;
	}

	.social-button svg {
		width: 18px;
		height: 18px;
		transition: transform 0.25s ease;
		flex-shrink: 0;
	}

	.social-button:hover svg {
		transform: scale(1.15);
	}

	/* ── Responsive ──────────────────────────────────────────── */

	/* Desktop: hide card header, brand panel shows branding */
	@media (min-width: 769px) {
		.auth-header {
			display: none;
		}
		.auth-card {
			max-width: 440px;
		}
	}

	/* Tablet / mobile: single column */
	@media (max-width: 768px) {
		main.auth-container {
			grid-template-columns: 1fr;
			background: linear-gradient(135deg, #0d0800 0%, #08090a 50%, #0d0800 100%);
		}

		.brand-panel {
			display: none;
		}

		.auth-panel {
			min-height: 100dvh;
			padding: clamp(1rem, 3vw, 2rem);
			align-items: center;
			background: transparent;
		}

		.auth-panel::before {
			display: none;
		}

		.auth-header {
			display: flex;
		}

		.auth-card {
			background: rgba(255, 255, 255, 0.03);
			backdrop-filter: blur(16px) saturate(180%);
			border: 1px solid var(--border);
			border-radius: 20px;
			padding: clamp(2rem, 5vw, 3rem);
			box-shadow:
				0 25px 65px rgba(0, 0, 0, 0.35),
				inset 0 1px 0 rgba(255, 255, 255, 0.08);
		}
	}

	@media (max-width: 480px) {
		.auth-panel {
			padding: 1rem;
		}

		.auth-card {
			padding: 1.75rem 1.25rem;
			border-radius: 16px;
		}

		.auth-header {
			margin-bottom: 2rem;
			gap: 1rem;
		}

		.mode-toggle {
			margin-bottom: 2rem;
		}

		.form-content {
			gap: 1.25rem;
		}

		.social-auth {
			grid-template-columns: 1fr;
		}

		.divider {
			margin: 1.5rem 0;
		}
	}

	@media (max-width: 360px) {
		.auth-card {
			padding: 1.5rem 1rem;
		}

		.form-content {
			gap: 1rem;
		}
	}
</style>
