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
	import Icon from './icon.svelte';

	interface Props {
		/** Estado de las claves del usuario: se avisa cuando no hay ninguna usable. */
		hasUsableKey: boolean;
		/** Alguna clave existe pero el proveedor la rechazó o agotó su cuota. */
		hasBrokenKey?: boolean;
	}

	let { hasUsableKey, hasBrokenKey = false }: Props = $props();
</script>

{#if !hasUsableKey}
	<div class="notice" role="status">
		<span class="icon"><Icon name="alert" size={17} /></span>
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
				{hasBrokenKey ? 'Revisar mi clave' : 'Configurar mi clave'}
			</a>
		</div>
	</div>
{/if}

<style>
	/*
	 * Un aviso, no una tarjeta. Estaba montado sobre `Card variant="elevated"`
	 * —con su sombra de 60 px— y pintaba con tokens que este proyecto no define
	 * (`--color-warning`, `--color-text-primary`), así que se quedaba siempre en
	 * los literales del respaldo y no seguía a la paleta.
	 */
	.notice {
		display: flex;
		gap: 0.85rem;
		align-items: flex-start;
		margin-bottom: 2rem;
		padding: 0.9rem 1.1rem;
		border: 1px solid rgba(212, 145, 42, 0.28);
		border-radius: 10px;
		background: rgba(212, 145, 42, 0.06);
	}

	.icon {
		flex-shrink: 0;
		margin-top: 0.1rem;
		color: var(--amber-light);
	}

	.body {
		min-width: 0;
	}

	.title {
		margin: 0 0 0.2rem;
		font-size: 0.88rem;
		font-weight: 500;
		color: var(--text);
	}

	.text {
		max-width: 68ch;
		margin: 0 0 0.5rem;
		font-size: 0.82rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.link {
		font-size: 0.82rem;
		color: var(--amber-light);
		text-decoration: underline;
		text-underline-offset: 3px;
	}
</style>
