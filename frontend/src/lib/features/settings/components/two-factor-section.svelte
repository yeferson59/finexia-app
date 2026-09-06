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

<SettingsSection
	title="Verificación en dos pasos"
	description="Un código temporal de tu aplicación de autenticación, además de la contraseña, cada vez que inicias sesión. Es opcional y puedes quitarla cuando quieras."
>
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
		<p class="state on">
			Está activada. Te quedan {twoFactor.recoveryCodesLeft}
			{twoFactor.recoveryCodesLeft === 1 ? 'código' : 'códigos'} de recuperación sin usar.
		</p>

		{#if twoFaDisableSuccess}
			<p class="feedback success">La verificación en dos pasos fue desactivada.</p>
		{/if}

		<TwoFactorManage {form} />
	{:else}
		<p class="state">Todavía no la tienes activada. Escribe tu contraseña para empezar.</p>

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
				<Button type="submit" size="sm" loading={twoFaSetupLoading}>Activar 2FA</Button>
			</div>
		</form>
	{/if}
</SettingsSection>

<style>
	/*
	 * El estado en una frase. Eran dos insignias en versalitas —ACTIVADA,
	 * DESACTIVADA— que decían menos que la propia frase y obligaban a leer dos
	 * cosas para entender una.
	 */
	.state {
		max-width: 56ch;
		margin: 0 0 1.1rem;
		font-size: 0.88rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.state.on {
		color: var(--green);
	}

	/* El único bloque de la página que sigue enmarcado: son unos códigos que
	   solo se enseñan una vez y que hay que copiar antes de irse. */
	.twofa-recovery {
		margin-bottom: 1.5rem;
		padding: 1rem 1.1rem;
		border: 1px solid rgba(212, 145, 42, 0.35);
		border-radius: 10px;
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
