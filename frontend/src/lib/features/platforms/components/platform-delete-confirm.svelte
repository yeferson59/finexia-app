<script lang="ts">
	/**
	 * Cuerpo de la confirmación de borrado de una plataforma.
	 *
	 * El diálogo lo pone el `Modal` del detalle, que es quien tiene el estado que
	 * lo abre; aquí sólo va el aviso, el motivo del rechazo y los dos botones.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';

	let { platformName, onCancel }: { platformName: string; onCancel: () => void } = $props();

	let isDeleting = $state(false);

	/**
	 * El motivo por el que el borrado no se hizo.
	 *
	 * El backend se niega a borrar una plataforma que todavía apuntan posiciones
	 * —no las arrastra consigo: son el historial del dueño— y devuelve un 409
	 * diciendo qué hay que quitar antes. Sin esto, la acción del servidor
	 * devolvía ese motivo y nadie lo leía: el modal se quedaba igual, sin
	 * explicar nada, y parecía que el botón no funcionaba.
	 */
	let error = $state<string | null>(null);
</script>

<p class="warning">
	¿Seguro que quieres eliminar <strong>{platformName}</strong>? Esta acción no se puede deshacer.
</p>

{#if error}
	<p class="error" role="alert">{error}</p>
{/if}

<form
	method="POST"
	action="?/delete"
	use:enhance={() => {
		isDeleting = true;
		error = null;
		return async ({ result, update }) => {
			isDeleting = false;

			// Un borrado que sale bien redirige, así que sólo llega aquí con datos
			// cuando fue rechazado. En ese caso el modal se queda abierto con el
			// motivo: cerrarlo obligaría a releer la página para descubrir que no
			// pasó nada.
			const data = result.type === 'success' ? result.data : undefined;
			if (typeof data?.error === 'string') {
				error = data.error;
				return;
			}

			await update();
		};
	}}
>
	<div class="actions">
		<Button type="button" variant="ghost" onclick={onCancel} disabled={isDeleting}>Cancelar</Button>
		<Button type="submit" variant="danger" loading={isDeleting}>Eliminar</Button>
	</div>
</form>

<style>
	.warning {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.9rem;
		line-height: 1.6;
	}

	.warning strong {
		color: var(--text);
		font-weight: 500;
	}

	.error {
		margin: 1rem 0 0;
		padding: 0.6rem 0.85rem;
		border-left: 2px solid var(--red);
		background: rgba(224, 90, 90, 0.08);
		color: var(--red);
		font-size: 0.85rem;
		line-height: 1.5;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		margin-top: 1.5rem;
	}
</style>
