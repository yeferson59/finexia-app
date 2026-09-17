<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import AuthField from './auth-field.svelte';
	import PasswordInput from './password-input.svelte';
	import type { AuthActionResult } from '../types';
	import { parseErrors } from '../utils';

	let {
		form,
		onSwitchToLogin
	}: {
		form: AuthActionResult;
		onSwitchToLogin: () => void;
	} = $props();

	let registerName = $state('');
	let registerEmail = $state('');
	let registerPassword = $state('');
	let registerConfirmPassword = $state('');
	let agreeTerms = $state(false);
	let isSubmitting = $state(false);

	const errors = $derived(form?.type === 'register' ? parseErrors(form.errors) : {});
</script>

<form
	method="POST"
	action="?/register"
	class="lp-fields"
	id="register-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<AuthField
		label="Nombre"
		id="register-name"
		name="name"
		autocomplete="name"
		placeholder="Laura Méndez"
		bind:value={registerName}
		error={errors['name']}
	/>

	<AuthField
		label="Correo electrónico"
		id="register-email"
		name="email"
		type="email"
		autocomplete="email"
		placeholder="tu@correo.com"
		bind:value={registerEmail}
		error={errors['email']}
	/>

	<PasswordInput
		label="Contraseña"
		id="register-password"
		name="password"
		autocomplete="new-password"
		hint="Entre 8 y 20 caracteres."
		bind:value={registerPassword}
		error={errors['password']}
	/>

	<PasswordInput
		label="Repite la contraseña"
		id="register-confirm"
		name="confirmPassword"
		autocomplete="new-password"
		bind:value={registerConfirmPassword}
		error={errors['confirmPassword']}
	/>

	<div class="consent">
		<input
			type="checkbox"
			id="terms"
			name="terms"
			class="lp-check"
			bind:checked={agreeTerms}
			aria-invalid={errors['terms'] ? 'true' : undefined}
			aria-describedby={errors['terms'] ? 'terms-error' : undefined}
		/>
		<label for="terms">
			Autorizo el tratamiento de mis datos según la
			<a href={resolve('/privacidad')} target="_blank" rel="noopener" class="lp-link"
				>política de privacidad</a
			>
			y acepto los
			<a href={resolve('/terminos')} target="_blank" rel="noopener" class="lp-link"
				>términos y condiciones</a
			>.
		</label>
	</div>
	{#if errors['terms']}
		<p class="lp-field-error terms-error" id="terms-error">{errors['terms']}</p>
	{/if}

	{#if errors['server']}
		<div class="lp-alert error" role="alert">
			<p>{errors['server']}</p>
			{#if form?.type === 'register' && form.duplicateEmail}
				<p class="shortcuts">
					<button type="button" onclick={onSwitchToLogin} class="lp-link">
						Iniciar sesión con este correo
					</button>
					<a href={resolve('/auth/forgot-password')} class="lp-link">¿Olvidaste tu contraseña?</a>
				</p>
			{/if}
		</div>
	{/if}

	<button type="submit" class="lp-btn block" disabled={isSubmitting}>
		{isSubmitting ? 'Creando tu cuenta…' : 'Crear cuenta'}
	</button>
</form>

<style>
	.consent {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		font-size: 15px;
		line-height: 1.5;
		color: var(--lp-ink-2);
	}

	.consent label {
		cursor: pointer;
	}

	.terms-error {
		margin-top: -14px;
	}

	.lp-alert p {
		margin: 0;
	}

	.lp-alert .shortcuts {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 20px;
		margin-top: 8px;
	}

	.shortcuts .lp-link {
		color: var(--lp-ink);
	}
</style>
