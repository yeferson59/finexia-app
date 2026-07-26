<script lang="ts">
	/**
	 * Aviso de que las posiciones se están valorando a coste.
	 *
	 * Antes siempre había un precio de mercado porque lo traía la clave del
	 * operador. Con BYO-key, quien no configura la suya ve sus posiciones
	 * valoradas a su precio de compra — y eso hay que decirlo, en vez de
	 * presentar un número que parece de mercado y no lo es.
	 */
	import { resolve } from '$app/paths';
	import Card from '$lib/ui/card.svelte';

	interface Props {
		/** Estado de las claves del usuario: se avisa cuando no hay ninguna usable. */
		hasUsableKey: boolean;
		/** Alguna clave existe pero el proveedor la rechazó o agotó su cuota. */
		hasBrokenKey?: boolean;
	}

	let { hasUsableKey, hasBrokenKey = false }: Props = $props();
</script>

{#if !hasUsableKey}
	<Card variant="elevated" padding="none">
		<div class="notice">
			<span class="icon" aria-hidden="true">
				<svg
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="8" x2="12" y2="12"></line>
					<line x1="12" y1="16" x2="12.01" y2="16"></line>
				</svg>
			</span>
			<div class="body">
				<p class="title">
					{#if hasBrokenKey}
						Tu clave de datos de mercado no está funcionando
					{:else}
						Estás viendo tus posiciones valoradas a precio de compra
					{/if}
				</p>
				<p class="text">
					{#if hasBrokenKey}
						El proveedor rechazó tu clave o agotó su cuota, así que no se han podido actualizar los
						precios. Mientras tanto, las posiciones se valoran a su precio de compra.
					{:else}
						Finexia usa tu propia clave de proveedor para consultar precios, de modo que la cuota y
						los datos son tuyos. Sin clave no hay valoración de mercado.
					{/if}
				</p>
				<a class="link" href="{resolve('/dashboard/settings')}#datos-de-mercado">
					{hasBrokenKey ? 'Revisar mi clave' : 'Configurar mi clave'} →
				</a>
			</div>
		</div>
	</Card>
{/if}

<style>
	.notice {
		display: flex;
		gap: 0.85rem;
		align-items: flex-start;
		padding: 1rem 1.25rem;
		border-left: 3px solid var(--color-warning, #e0a800);
		border-radius: 0.5rem;
	}

	.icon {
		flex-shrink: 0;
		margin-top: 0.1rem;
		color: var(--color-warning, #e0a800);
	}

	.body {
		min-width: 0;
	}

	.title {
		margin: 0 0 0.25rem;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--color-text-primary, #f5f5f5);
	}

	.text {
		margin: 0 0 0.5rem;
		font-size: 0.83rem;
		line-height: 1.5;
		color: var(--color-text-muted, #9a9a9a);
	}

	.link {
		font-size: 0.83rem;
		font-weight: 500;
		color: var(--color-accent, #d4af37);
		text-decoration: none;
	}

	.link:hover,
	.link:focus-visible {
		text-decoration: underline;
	}
</style>
