<script lang="ts">
	/**
	 * "¿Le das acceso a esta aplicación?" — la única pantalla del flujo OAuth
	 * que ve una persona.
	 *
	 * `clientName`, `clientUri` y `logoUri` los eligió quien registró el
	 * cliente, y el registro es abierto: son texto de un desconocido. Por eso el
	 * nombre se pinta como texto (Svelte lo escapa), el logo como imagen con
	 * `referrerpolicy` cerrado, y la URI de retorno se muestra sin ser un
	 * enlace — es dato a verificar, no un sitio al que invitar a ir.
	 */
	import Button from '$lib/ui/button.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	/**
	 * Sin `use:enhance`, y es deliberado: la acción responde con un 303 hacia el
	 * cliente, que está en otro origen. `enhance` resuelve los redirects con
	 * `goto()`, y `goto()` se niega a salir del sitio — el flujo moriría aquí
	 * con el usuario mirando una pantalla que no hace nada. Un submit normal
	 * deja que lo siga el navegador, que es quien sabe hacerlo.
	 *
	 * Lo único que se pierde así es el estado de carga entre el clic y la
	 * navegación, y para eso basta con bloquear el doble envío.
	 */
	let submitting = $state(false);

	const consent = $derived(data.consent);

	/** El host al que volverá el código, que es lo que de verdad hay que mirar. */
	const redirectHost = $derived.by(() => {
		try {
			return new URL(consent.redirectUri).host;
		} catch {
			return consent.redirectUri;
		}
	});

	/**
	 * Los ámbitos, en lo que significan. Un `mcp:read` no le dice nada a nadie;
	 * "leer tus carteras" sí, y es lo que se está autorizando.
	 */
	const SCOPE_LABELS: Record<string, string> = {
		'mcp:read':
			'Leer tus carteras, posiciones, transacciones y datos de mercado. Solo lectura: no puede crear, modificar ni borrar nada.'
	};
</script>

<svelte:head>
	<title>Autorizar aplicación - FINEXIA</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<main class="consent">
	<article class="card">
		<header class="head">
			{#if consent.logoUri}
				<img class="logo" src={consent.logoUri} alt="" referrerpolicy="no-referrer" />
			{/if}
			<h1>
				<strong>{consent.clientName}</strong> quiere acceder a tu cuenta de Finexia
			</h1>
		</header>

		<section class="scopes" aria-label="Permisos solicitados">
			<h2>Le estarías dando permiso para:</h2>
			<ul>
				{#each consent.scopes as scope (scope)}
					<li>{SCOPE_LABELS[scope] ?? scope}</li>
				{/each}
			</ul>
		</section>

		<dl class="details">
			<dt>Volverá a</dt>
			<dd><code>{redirectHost}</code></dd>
			{#if consent.clientUri}
				<dt>Sitio de la aplicación</dt>
				<dd><code>{consent.clientUri}</code></dd>
			{/if}
		</dl>

		<p class="warning">
			Autoriza solo si acabas de pedir esta conexión desde <strong>{consent.clientName}</strong>.
			Podrás retirarle el acceso cuando quieras desde Ajustes.
		</p>

		<!--
			Un formulario por decisión, en vez de dos botones dentro de uno. Con
			un solo formulario la respuesta viajaría en el `value` del botón
			pulsado, y basta con que ese detalle se pierda —un `Button` que no
			propague `name`, un envío por Enter— para que "cancelar" mande un
			"autorizar". Aquí cada botón solo puede enviar lo que tiene al lado.

			El id va en el cuerpo y no en la URL porque `action="?/decide"`
			reemplaza la query entera: el `?request=…` con el que se llegó a
			esta página no sobrevive al POST.
		-->
		<div class="actions">
			<form method="POST" action="?/decide" onsubmit={() => (submitting = true)}>
				<input type="hidden" name="request" value={consent.requestId} />
				<input type="hidden" name="decision" value="deny" />
				<Button type="submit" variant="secondary" disabled={submitting}>Cancelar</Button>
			</form>
			<form method="POST" action="?/decide" onsubmit={() => (submitting = true)}>
				<input type="hidden" name="request" value={consent.requestId} />
				<input type="hidden" name="decision" value="approve" />
				<Button type="submit" variant="primary" disabled={submitting}>Autorizar</Button>
			</form>
		</div>
	</article>
</main>

<style>
	.consent {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 100dvh;
		padding: 1.5rem;
	}

	.card {
		width: 100%;
		max-width: 32rem;
		padding: 2rem;
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		background: var(--surface);
	}

	.head {
		display: flex;
		align-items: center;
		gap: 0.875rem;
		margin-bottom: 1.75rem;
	}

	.logo {
		width: 2.75rem;
		height: 2.75rem;
		flex-shrink: 0;
		border-radius: 0.5rem;
		object-fit: cover;
	}

	h1 {
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-weight: 500;
		line-height: 1.35;
	}

	h1 strong {
		font-weight: 600;
		color: var(--amber);
	}

	h2 {
		margin-bottom: 0.625rem;
		font-size: 0.75rem;
		font-weight: 500;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	.scopes ul {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding-left: 1.1rem;
		list-style: disc;
	}

	.scopes li {
		font-size: 0.9375rem;
		line-height: 1.55;
	}

	.details {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.4rem 1rem;
		margin: 1.5rem 0;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
		font-size: 0.8125rem;
	}

	.details dt {
		color: var(--text-muted);
	}

	.details dd {
		min-width: 0;
		overflow-wrap: anywhere;
	}

	.details code {
		font-family: var(--font-mono);
		font-size: 0.8125rem;
	}

	.warning {
		padding: 0.75rem 0.875rem;
		border: 1px solid rgba(212, 145, 42, 0.25);
		border-radius: 0.5rem;
		background: rgba(212, 145, 42, 0.06);
		font-size: 0.8125rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 1.5rem;
	}
</style>
