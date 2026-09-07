<script lang="ts">
	/**
	 * Alta de un usuario por invitación.
	 *
	 * No se crean cuentas con contraseña desde aquí: el backend manda un enlace
	 * de un solo uso y la persona elige la suya.
	 */
	import { enhance } from '$app/forms';
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
	class="rail-fields"
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
	<div class="field">
		<label for="invite-email">Correo</label>
		<input
			id="invite-email"
			name="email"
			type="email"
			bind:value={inviteEmail}
			placeholder="persona@correo.com"
			autocomplete="off"
			required
		/>
	</div>

	<div class="pair">
		<div class="field">
			<label for="invite-name">Nombre <span class="optional">(opcional)</span></label>
			<input id="invite-name" type="text" name="name" bind:value={inviteName} autocomplete="off" />
		</div>
		<div class="field">
			<label for="invite-role">Rol</label>
			<select id="invite-role" name="role" bind:value={inviteRole}>
				{#each INVITE_ROLES as role (role.value)}
					<option value={role.value}>{role.label}</option>
				{/each}
			</select>
		</div>
	</div>

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}
	{#if invited}
		<p class="feedback success">Invitación enviada a {invited}.</p>
	{/if}

	<div class="actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={inviting}>Enviar invitación</Button>
	</div>
</form>

<style>
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}
</style>
