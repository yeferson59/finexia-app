<script lang="ts">
	/**
	 * Verificación en dos pasos.
	 *
	 * Tres estados excluyentes: alta en curso (hay `secret` del servidor), 2FA
	 * activa, o desactivada. `two-factor-setup` y `two-factor-manage` son
	 * internos de esta sección (import relativo).
	 */
	import { enhance } from '$app/forms';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import SettingsSection from './settings-section.svelte';
	import TwoFactorSetup from './two-factor-setup.svelte';
	import TwoFactorManage from './two-factor-manage.svelte';
	import {
		actionData,
		actionError,
		actionSucceeded,
		issuedRecoveryCodes,
		type SettingsForm,
		type TwoFactorStatus
	} from '../settings';

	interface Props {
		twoFactor: TwoFactorStatus | undefined;
		form: SettingsForm;
	}

	let { twoFactor, form }: Props = $props();

	let twoFaPassword = $state('');
	let twoFaSetupLoading = $state(false);

	const twoFaSetupError = $derived(actionError(form, 'setup2fa'));
	const twoFaSetupData = $derived(
		actionSucceeded(form, 'setup2fa')
			? {
					secret: actionData<string>(form, 'setup2fa', 'secret') ?? '',
					otpauthUrl: actionData<string>(form, 'setup2fa', 'otpauthUrl') ?? ''
				}
			: null
	);
	const twoFaRecoveryCodes = $derived(issuedRecoveryCodes(form));
	const twoFaDisableSuccess = $derived(actionSucceeded(form, 'disable2fa'));

	// Never leave credentials sitting in the 2FA forms after they succeed.
	$effect(() => {
		if (twoFaSetupData) twoFaPassword = '';
	});
</script>

<SettingsSection title="Verificación en dos pasos (2FA)">
	{#if twoFaRecoveryCodes.length > 0}
		<div class="twofa-recovery" role="status">
			<p class="twofa-recovery-title">Guarda tus códigos de recuperación</p>
			<p class="hint">
				Cada código funciona una sola vez y te permitirá entrar si pierdes acceso a tu aplicación de
				autenticación. No volverán a mostrarse.
			</p>
			<ul class="twofa-code-list">
				{#each twoFaRecoveryCodes as code (code)}
					<li class="twofa-code">{code}</li>
				{/each}
			</ul>
		</div>
	{/if}

	{#if twoFaSetupData}
		<TwoFactorSetup secret={twoFaSetupData.secret} otpauthUrl={twoFaSetupData.otpauthUrl} {form} />
	{:else if twoFactor?.enabled}
		<div class="twofa-status">
			<span class="twofa-badge enabled">Activada</span>
			<p class="hint">
				Tu cuenta pide un código del autenticador en cada inicio de sesión. Te quedan
				{twoFactor.recoveryCodesLeft} códigos de recuperación sin usar.
			</p>
		</div>

		{#if twoFaDisableSuccess}
			<p class="feedback success">La verificación en dos pasos fue desactivada.</p>
		{/if}

		<TwoFactorManage {form} />
	{:else}
		<div class="twofa-status">
			<span class="twofa-badge">Desactivada</span>
			<p class="hint">
				Añade una segunda barrera a tu cuenta: además de tu contraseña, se pedirá un código temporal
				de una aplicación de autenticación al iniciar sesión. Es opcional y puedes desactivarla
				cuando quieras.
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
</SettingsSection>

<style>
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
</style>
