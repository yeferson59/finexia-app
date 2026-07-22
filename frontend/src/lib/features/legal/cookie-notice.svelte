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
			Usamos cookies estrictamente necesarias para el funcionamiento del servicio. Consulta nuestro
			<a href={resolve('/cookies')}>Aviso de Cookies</a> y la
			<a href={resolve('/privacidad')}>Política de Privacidad</a>.
		</p>
		<button type="button" onclick={dismiss}>Entendido</button>
	</div>
{/if}

<style>
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
		padding: 14px 18px;
		border-radius: 12px;
		background: rgba(16, 17, 19, 0.96);
		border: 1px solid var(--border-strong);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.45);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
	}
	.cookie-notice p {
		flex: 1;
		margin: 0;
		font-size: 13px;
		line-height: 1.5;
		color: var(--text-muted);
	}
	.cookie-notice a {
		color: var(--amber-light);
		text-decoration: underline;
		text-underline-offset: 2px;
	}
	.cookie-notice a:hover {
		color: var(--amber);
	}
	.cookie-notice button {
		flex-shrink: 0;
		padding: 9px 18px;
		border-radius: 8px;
		border: 1px solid var(--border-strong);
		background: var(--amber);
		color: #0d0800;
		font-family: var(--font-body);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition:
			background 0.2s,
			transform 0.1s;
	}
	.cookie-notice button:hover {
		background: var(--amber-light);
	}
	.cookie-notice button:active {
		transform: translateY(1px);
	}
	@media (max-width: 560px) {
		.cookie-notice {
			flex-direction: column;
			align-items: stretch;
			text-align: center;
			gap: 12px;
		}
		.cookie-notice button {
			width: 100%;
		}
	}
</style>
