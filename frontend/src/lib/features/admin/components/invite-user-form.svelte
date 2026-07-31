<script lang="ts">
	/**
	 * Alta de un usuario por invitación.
	 *
	 * No se crean cuentas con contraseña desde aquí: el backend manda un enlace
	 * de un solo uso y la persona elige la suya.
	 */
	import { enhance } from '$app/forms';
	import Card from '$lib/ui/card.svelte';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import { INVITE_ROLES } from '../admin';

	interface Props {
		/** `form` de la página, para el error o el acuse de la invitación. */
		form: Record<string, unknown> | null;
	}

	let { form }: Props = $props();

	// Estado local del formulario: la página lo desmonta cuando la invitación
	// sale, así que vuelve vacío la próxima vez que se abre.
	let inviteName = $state('');
	let inviteEmail = $state('');
	let inviteRole = $state('customer');
	let inviting = $state(false);
</script>

<div class="create-form-card">
	<Card padding="md">
		<h2 class="form-title">Invitar a un nuevo usuario</h2>
		<p class="form-hint">
			Enviaremos un enlace seguro de un solo uso para que la persona cree su propia contraseña.
		</p>
		<form
			method="POST"
			action="?/inviteUser"
			use:enhance={() => {
				inviting = true;
				return async ({ update }) => {
					inviting = false;
					await update();
				};
			}}
		>
			<div class="form-row">
				<Input label="Nombre (opcional)" name="name" bind:value={inviteName} />
				<Input
					label="Correo electrónico"
					name="email"
					type="email"
					bind:value={inviteEmail}
					required
				/>
				<div class="field">
					<span class="field-label">Rol</span>
					<select class="select" name="role" bind:value={inviteRole}>
						{#each INVITE_ROLES as role (role.value)}
							<option value={role.value}>{role.label}</option>
						{/each}
					</select>
				</div>
			</div>
			{#if form?.error}
				<p class="form-error">{form.error}</p>
			{/if}
			{#if form?.invited}
				<p class="form-success">Invitación enviada a {form.invited}.</p>
			{/if}
			<div class="form-actions">
				<Button type="submit" loading={inviting}>Enviar invitación</Button>
			</div>
		</form>
	</Card>
</div>

<style>
	.create-form-card {
		margin-bottom: 1.5rem;
	}

	.form-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
		margin: 0 0 0.35rem 0;
	}

	.form-hint {
		font-size: 0.82rem;
		color: var(--text-dim);
		margin: 0 0 1.25rem 0;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr auto;
		gap: 1rem;
		margin-bottom: 1rem;
		align-items: end;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.field-label {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text-muted);
	}

	.select {
		appearance: none;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		color: var(--text);
		font-size: 0.875rem;
		padding: 0.6rem 0.75rem;
		min-width: 9rem;
		cursor: pointer;
	}

	.select:focus {
		outline: none;
		border-color: var(--amber);
	}

	.form-error {
		font-size: 0.82rem;
		color: var(--red);
		margin: 0 0 0.75rem 0;
	}

	.form-success {
		font-size: 0.82rem;
		color: var(--green);
		margin: 0 0 0.75rem 0;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
	}

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
