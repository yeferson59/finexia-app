<script lang="ts">
	import '../auth-forms.css';
	import LoginForm from './login-form.svelte';
	import TwoFactorChallenge from './two-factor-challenge.svelte';
	import RegisterForm from './register-form.svelte';
	import InviteOnlyNotice from './invite-only-notice.svelte';
	import AuthBrandPanel from './auth-brand-panel.svelte';
	import AuthSocialButtons from './auth-social-buttons.svelte';
	import type { AuthActionResult } from '../types';

	// Resolved server-side from the `selfRegistration` feature flag so the
	// component never needs `$env/dynamic/public` itself — that module is only
	// populated once a real SvelteKit page has hydrated, which breaks
	// isolated component tests. Defaults closed: Finexia is invite-only during
	// the beta, so an omitted prop should fail toward "registration closed".
	let {
		form,
		selfRegistrationEnabled = false
	}: { form: AuthActionResult; selfRegistrationEnabled?: boolean } = $props();

	// Form state
	let isLoginMode = $state(true);
	let slideDirection = $state<'left' | 'right'>('right');

	// Two-factor step: the server validated the password and handed back a
	// short-lived token; the session only exists after a valid TOTP code.
	const twoFactorRequired = $derived(
		form?.type === 'login' && form.twoFactorRequired === true && !!form.twoFactorToken
	);
	const twoFactorToken = $derived(
		form?.type === 'login' && form.twoFactorToken ? form.twoFactorToken : ''
	);

	const switchToLogin = () => {
		if (isLoginMode) return;
		slideDirection = 'left';
		isLoginMode = true;
	};

	const switchToRegister = () => {
		if (!isLoginMode) return;
		slideDirection = 'right';
		isLoginMode = false;
	};
</script>

<main class="auth-container">
	<!-- Left: Brand Panel (desktop only) -->
	<AuthBrandPanel />

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
					{#if !selfRegistrationEnabled}
						<span class="beta-badge">Invitación</span>
					{/if}
				</button>
			</div>

			<!-- Forms Section -->
			<section class="forms-container">
				{#if isLoginMode && twoFactorRequired}
					<TwoFactorChallenge {form} token={twoFactorToken} />
				{:else if isLoginMode}
					<LoginForm {form} {slideDirection} onSwitchToRegister={switchToRegister} />
				{:else if !selfRegistrationEnabled}
					<InviteOnlyNotice {slideDirection} onSwitchToLogin={switchToLogin} />
				{:else}
					<RegisterForm {form} {slideDirection} onSwitchToLogin={switchToLogin} />
				{/if}
			</section>

			<!-- Social Divider -->
			<AuthSocialButtons />
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

	/* Ambient glow top-right */

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

	/* ── Invite-only badge (beta) ────────────────────────────── */

	.beta-badge {
		display: inline-block;
		margin-left: 0.4rem;
		padding: 0.1rem 0.45rem;
		border-radius: 999px;
		background: rgba(212, 145, 42, 0.15);
		border: 1px solid rgba(212, 145, 42, 0.35);
		color: var(--gold-primary);
		font-size: 0.6rem;
		font-weight: 700;
		letter-spacing: 0.4px;
		text-transform: uppercase;
		vertical-align: middle;
	}

	/* ── Divider + Social ────────────────────────────────────── */

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
	}

	@media (max-width: 360px) {
		.auth-card {
			padding: 1.5rem 1rem;
		}
	}
</style>
