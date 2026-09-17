<script lang="ts">
	/*
	 * `/auth`: entrar o crear cuenta.
	 *
	 * Las dos opciones son el titular de la página, una debajo de la otra y del
	 * mismo tamaño: la elegida en tinta y la otra apagada. Es un `tablist`
	 * vertical, así que se recorre con las flechas. Antes eran un conmutador en
	 * versalitas encima del formulario, y además una línea «¿No tienes cuenta?»
	 * debajo que hacía lo mismo.
	 */
	import { untrack } from 'svelte';
	import PublicShell from '$lib/ui/public-shell.svelte';
	import LoginForm from './login-form.svelte';
	import TwoFactorChallenge from './two-factor-challenge.svelte';
	import RegisterForm from './register-form.svelte';
	import InviteOnlyNotice from './invite-only-notice.svelte';
	import type { AuthActionResult } from '../types';

	// Resolved server-side from the `selfRegistration` feature flag so the
	// component never needs `$env/dynamic/public` itself — that module is only
	// populated once a real SvelteKit page has hydrated, which breaks
	// isolated component tests. Defaults closed: Finexia is invite-only during
	// the beta, so an omitted prop should fail toward "registration closed".
	let {
		form,
		selfRegistrationEnabled = false,
		notice
	}: {
		form: AuthActionResult;
		selfRegistrationEnabled?: boolean;
		/** Lo que viene de otra pantalla: cuenta creada, correo verificado… */
		notice?: string;
	} = $props();

	type Mode = 'login' | 'register';

	/* Si la acción que vuelve es la de registro, la pestaña tiene que seguir en
	   registro para enseñar sus errores. Solo cuenta al montar: después manda
	   lo que elija la persona, aunque llegue otra respuesta. */
	let mode = $state<Mode>(untrack(() => (form?.type === 'register' ? 'register' : 'login')));

	// Two-factor step: the server validated the password and handed back a
	// short-lived token; the session only exists after a valid TOTP code.
	const twoFactorRequired = $derived(
		form?.type === 'login' && form.twoFactorRequired === true && !!form.twoFactorToken
	);
	const twoFactorToken = $derived(
		form?.type === 'login' && form.twoFactorToken ? form.twoFactorToken : ''
	);

	const modes: { id: Mode; label: string }[] = [
		{ id: 'login', label: 'Iniciar sesión' },
		{ id: 'register', label: 'Crear cuenta' }
	];

	function onTabKey(event: KeyboardEvent) {
		const keys = ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'];
		if (!keys.includes(event.key)) return;
		event.preventDefault();
		const next: Mode =
			event.key === 'Home'
				? 'login'
				: event.key === 'End'
					? 'register'
					: mode === 'login'
						? 'register'
						: 'login';
		mode = next;
		document.getElementById(`auth-tab-${next}`)?.focus();
	}
</script>

<PublicShell>
	{#snippet heading()}
		{#if twoFactorRequired}
			<h1 class="shell-title">Verificación en dos pasos</h1>
			<p class="shell-note">
				Tu contraseña es correcta. Falta el código que te da tu app de autenticación.
			</p>
		{:else}
			<h1 class="lp-sr-only">Accede a Finexia</h1>
			<div
				class="modes"
				role="tablist"
				aria-label="Iniciar sesión o crear cuenta"
				aria-orientation="vertical"
			>
				{#each modes as option (option.id)}
					<button
						id="auth-tab-{option.id}"
						class="mode"
						type="button"
						role="tab"
						aria-selected={mode === option.id}
						aria-controls="auth-panel"
						tabindex={mode === option.id ? 0 : -1}
						onclick={() => (mode = option.id)}
						onkeydown={onTabKey}
					>
						{option.label}
						{#if option.id === 'register' && !selfRegistrationEnabled}
							<span class="badge">con invitación</span>
						{/if}
					</button>
				{/each}
			</div>
			<p class="shell-note">
				Finexia no se conecta a tu broker, exchange ni banco: la única contraseña que te pedimos es
				la de tu cuenta aquí.
			</p>
		{/if}
	{/snippet}

	<div id="auth-panel" role="tabpanel" aria-labelledby="auth-tab-{mode}">
		{#if notice}
			<p class="lp-alert success notice" role="status">{notice}</p>
		{/if}

		{#key mode + String(twoFactorRequired)}
			<div class="panel">
				{#if twoFactorRequired}
					<TwoFactorChallenge {form} token={twoFactorToken} />
				{:else if mode === 'login'}
					<LoginForm {form} />
				{:else if !selfRegistrationEnabled}
					<InviteOnlyNotice />
				{:else}
					<RegisterForm {form} onSwitchToLogin={() => (mode = 'login')} />
				{/if}
			</div>
		{/key}
	</div>
</PublicShell>

<style>
	.modes {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
	}

	.mode {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 0;
		border: none;
		background: none;
		color: var(--lp-ink-3);
		font-family: var(--lp-font);
		font-size: clamp(40px, 5.4vw, 72px);
		font-stretch: 118%;
		font-weight: 640;
		line-height: 1.02;
		letter-spacing: -0.03em;
		text-align: left;
		cursor: pointer;
		transition: color 0.18s ease;
	}

	.mode:hover {
		color: var(--lp-ink-2);
	}

	.mode[aria-selected='true'] {
		color: var(--lp-ink);
		cursor: default;
	}

	/* Cuelga del renglón a la altura de la x, en la letra del texto. */
	.badge {
		flex-shrink: 0;
		padding: 3px 10px;
		border: 1px solid currentColor;
		border-radius: 999px;
		font-size: 13px;
		font-stretch: 100%;
		font-weight: 500;
		letter-spacing: 0;
		line-height: 1.4;
	}

	.notice {
		margin-bottom: 28px;
	}

	.panel {
		animation: panel-in 0.22s ease both;
	}

	@keyframes panel-in {
		from {
			opacity: 0;
		}
	}

	@media (max-width: 860px) {
		.mode {
			font-size: clamp(34px, 9vw, 52px);
			gap: 12px;
		}
		.badge {
			font-size: 12px;
			padding: 2px 8px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.mode {
			transition: none;
		}
		.panel {
			animation: none;
		}
	}
</style>
