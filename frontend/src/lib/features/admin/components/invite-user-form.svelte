<script lang="ts">
	/**
	 * Alta de un usuario por invitación.
	 *
	 * No se crean cuentas con contraseña desde aquí: el backend manda un enlace
	 * de un solo uso y la persona elige la suya.
	 */
	import { enhance } from '$app/forms';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import { actionData, actionError } from '$lib/shared/form';
	import { INVITE_ROLES } from '../admin';

	interface Props {
		/** `form` de la página, para el error o el acuse de la invitación. */
		form: Record<string, unknown> | null;
		/**
		 * Se llama cuando la invitación sale. La página cierra el panel desde
		 * aquí y no desde el `form` común, que también cambia con el resto de
		 * actions de la pantalla (banear, eliminar, revocar…).
		 */
		onSuccess?: () => void;
		/** Cierra el modal sin enviar. */
		onCancel?: () => void;
	}

	let { form, onSuccess, onCancel }: Props = $props();

	// Estado local del formulario: la página lo desmonta cuando la invitación
	// sale, así que vuelve vacío la próxima vez que se abre.
	let inviteName = $state('');
	let inviteEmail = $state('');
	let inviteRole = $state('customer');
	let inviting = $state(false);

	// La pantalla tiene seis actions y un solo `form`: sin filtrar por acción,
	// un borrado fallido pintaba «No se pudo eliminar el usuario» aquí dentro.
	const error = $derived(actionError(form, 'inviteUser'));
	const invited = $derived(actionData<string>(form, 'inviteUser', 'invited'));
</script>

<form
	method="POST"
	action="?/inviteUser"
	use:enhance={() => {
		inviting = true;
		return async ({ result, update }) => {
			inviting = false;
			await update();
			if (result.type === 'success') onSuccess?.();
		};
	}}
>
	<div class="form-row">
		<Input label="Nombre (opcional)" name="name" bind:value={inviteName} />
		<Input label="Correo electrónico" name="email" type="email" bind:value={inviteEmail} required />
		<div class="field">
			<span class="field-label">Rol</span>
			<select class="select" name="role" bind:value={inviteRole}>
				{#each INVITE_ROLES as role (role.value)}
					<option value={role.value}>{role.label}</option>
				{/each}
			</select>
		</div>
	</div>
	{#if error}
		<p class="form-error">{error}</p>
	{/if}
	{#if invited}
		<p class="form-success">Invitación enviada a {invited}.</p>
	{/if}
	<div class="form-actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={inviting}>Enviar invitación</Button>
	</div>
</form>

<style>
	/* Dentro del modal no hay ancho para tres columnas: nombre y correo
	   comparten línea y el rol baja a la suya. */
	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-bottom: 1rem;
		align-items: end;
	}

	.field {
		grid-column: 1 / -1;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.select {
		width: 100%;
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
		gap: 0.75rem;
		justify-content: flex-end;
	}

	@media (max-width: 640px) {
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
