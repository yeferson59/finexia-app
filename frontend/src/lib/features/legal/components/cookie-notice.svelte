<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';

	// Lightweight cookie/privacy notice required for Colombian compliance
	// (Ley 1581 de 2012). We only use strictly-necessary storage today, so this
	// is an informative acknowledgement rather than a full consent manager.
	const STORAGE_KEY = 'finexia:cookie-notice';

	let visible = $state(false);

	onMount(() => {
		try {
			if (localStorage.getItem(STORAGE_KEY) !== 'dismissed') visible = true;
		} catch {
			visible = true;
		}
	});

	function dismiss() {
		visible = false;
		try {
			localStorage.setItem(STORAGE_KEY, 'dismissed');
		} catch {
			// Storage unavailable (e.g. private mode) — closing for the session is enough.
		}
	}
</script>

{#if visible}
	<div class="cookie-notice" role="region" aria-label="Aviso de cookies">
		<p>
			Solo usamos las cookies necesarias para que el servicio funcione. Más detalles en el
			<a href={resolve('/cookies')}>aviso de cookies</a> y la
			<a href={resolve('/privacidad')}>política de privacidad</a>.
		</p>
		<button type="button" onclick={dismiss}>Entendido</button>
	</div>
{/if}

<style>
	/*
	 * Sale en la primera visita, que casi siempre es la portada o el acceso, en
	 * papel; pero vive en el layout raíz y también puede aparecer sobre el panel
	 * oscuro. En tinta sobre papel claro se lee como aviso, y sobre el panel
	 * sigue destacando por el filete. Por eso los colores van escritos aquí y no
	 * con los tokens de ninguna de las dos superficies: fuera de `.lp` no existen
	 * los del papel.
	 *
	 * La altura es la del aviso anterior a propósito: está fijo al pie y, más
	 * alto, tapaba los botones de los formularios del panel hasta cerrarlo.
	 */
	.cookie-notice {
		position: fixed;
		z-index: 200;
		left: 16px;
		right: 16px;
		bottom: 16px;
		margin: 0 auto;
		max-width: 760px;
		display: flex;
		align-items: center;
		gap: 18px;
		padding: 12px 12px 12px 18px;
		border: 1px solid rgba(238, 240, 236, 0.18);
		border-radius: 10px;
		background: #10231e;
		box-shadow: 0 20px 40px -24px rgba(16, 35, 30, 0.6);
		font-family: 'Archivo', system-ui, sans-serif;
	}
	.cookie-notice p {
		flex: 1;
		margin: 0;
		font-size: 13px;
		line-height: 1.5;
		color: #c9d4cf;
	}
	.cookie-notice a {
		color: #eef0ec;
		text-decoration: underline;
		text-underline-offset: 3px;
	}
	.cookie-notice a:hover {
		text-decoration-thickness: 2px;
	}
	.cookie-notice a:focus-visible,
	.cookie-notice button:focus-visible {
		outline: 2px solid #eef0ec;
		outline-offset: 2px;
	}
	.cookie-notice button {
		flex-shrink: 0;
		min-height: 36px;
		padding: 0 16px;
		border: none;
		border-radius: 6px;
		background: #eef0ec;
		color: #10231e;
		font-family: inherit;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}
	.cookie-notice button:hover {
		background: #ffffff;
	}
	@media (max-width: 560px) {
		.cookie-notice {
			flex-direction: column;
			align-items: stretch;
			gap: 12px;
			padding: 14px;
		}
		.cookie-notice button {
			min-height: 44px;
		}
	}
</style>
