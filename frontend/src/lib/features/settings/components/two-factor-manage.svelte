<script lang="ts">
	/**
	 * Gestión de una 2FA ya activa: regenerar códigos de recuperación o
	 * desactivarla. Interno de `two-factor-section`.
	 *
	 * Ambas operaciones piden contraseña + código vigente, así que ninguna se
	 * puede ejecutar desde una sesión secuestrada sin el segundo factor.
	 */
	import { enhance } from '$app/forms';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import {
		actionError,
		actionSucceeded,
		issuedRecoveryCodes,
		type SettingsForm
	} from '../settings';

	interface Props {
		form: SettingsForm;
	}

	let { form }: Props = $props();

	let twoFaDisablePassword = $state('');
	let twoFaDisableCode = $state('');
	let twoFaDisableLoading = $state(false);
	let twoFaShowDisable = $state(false);
	let twoFaRegenPassword = $state('');
	let twoFaRegenCode = $state('');
	let twoFaRegenLoading = $state(false);
	let twoFaShowRegen = $state(false);

	const twoFaDisableError = $derived(actionError(form, 'disable2fa'));
	const twoFaDisableSuccess = $derived(actionSucceeded(form, 'disable2fa'));
	const twoFaRegenError = $derived(actionError(form, 'regenerate2faCodes'));

	// Never leave credentials sitting in the forms after they succeed.
	$effect(() => {
		if (issuedRecoveryCodes(form).length > 0) {
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
</script>

<div class="twofa-actions">
	<button type="button" class="twofa-toggle" onclick={() => (twoFaShowRegen = !twoFaShowRegen)}>
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
			<Button type="submit" variant="secondary" size="sm" loading={twoFaRegenLoading}>
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
			Para desactivar la verificación en dos pasos confirma tu contraseña y un código vigente del
			autenticador (o un código de recuperación).
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
			<Button type="submit" variant="secondary" size="sm" loading={twoFaDisableLoading}
				>Desactivar 2FA</Button
			>
		</div>
	</form>
{/if}

<style>
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
</style>
