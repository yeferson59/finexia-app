<script lang="ts">
	/**
	 * Alta de 2FA, paso 2: escanear el QR y confirmar el primer código.
	 *
	 * Interno de `two-factor-section`. La verificación no queda activa hasta que
	 * el usuario confirma, así que este paso puede abandonarse sin consecuencias.
	 */
	import { enhance } from '$app/forms';
	import { renderSVG } from 'uqr';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import { actionError, issuedRecoveryCodes, type SettingsForm } from '../settings';

	interface Props {
		secret: string;
		otpauthUrl: string;
		form: SettingsForm;
	}

	let { secret, otpauthUrl, form }: Props = $props();

	let twoFaConfirmCode = $state('');
	let twoFaConfirmLoading = $state(false);

	const twoFaEnableError = $derived(actionError(form, 'enable2fa'));

	// The QR is rendered locally from the otpauth URL; the secret never
	// touches a third-party service.
	const twoFaQrSvg = $derived(otpauthUrl ? renderSVG(otpauthUrl) : '');

	$effect(() => {
		if (issuedRecoveryCodes(form).length > 0) twoFaConfirmCode = '';
	});
</script>

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
		<code class="twofa-secret">{secret}</code>
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

<style>
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
		color: var(--text-dim);
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
</style>
