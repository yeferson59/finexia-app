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
	import PublicShell from '$lib/ui/public-shell.svelte';
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
	 * "leer tus portafolios" sí, y es lo que se está autorizando.
	 */
	const SCOPE_LABELS: Record<string, string> = {
		'mcp:read':
			'Leer tus portafolios, posiciones, movimientos y datos de mercado. Solo lectura: no puede crear, modificar ni borrar nada.'
	};
</script>

<svelte:head>
	<title>Autorizar aplicación — Finexia</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<PublicShell>
	{#snippet heading()}
		{#if consent.logoUri}
			<img class="logo" src={consent.logoUri} alt="" referrerpolicy="no-referrer" />
		{/if}
		<h1 class="shell-title">{consent.clientName} quiere acceder a tu cuenta</h1>
		<p class="shell-note">
			Autoriza solo si acabas de pedir esta conexión desde {consent.clientName}.
		</p>
	{/snippet}

	<section class="scopes" aria-labelledby="scopes-title">
		<h2 id="scopes-title">Si lo autorizas, podrá:</h2>
		<ul>
			{#each consent.scopes as scope (scope)}
				<li>{SCOPE_LABELS[scope] ?? scope}</li>
			{/each}
		</ul>
	</section>

	<!--
		El host de retorno en monoespaciada a propósito: es lo que hay que leer
		letra a letra, y ahí una «l» y una «I» no pueden parecerse.
	-->
	<dl class="details">
		<div>
			<dt>Volverá a</dt>
			<dd><code>{redirectHost}</code></dd>
		</div>
		{#if consent.clientUri}
			<div>
				<dt>Sitio de la aplicación</dt>
				<dd><code>{consent.clientUri}</code></dd>
			</div>
		{/if}
	</dl>

	<p class="revoke">Podrás retirarle el acceso cuando quieras desde Configuración.</p>

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
			<button type="submit" class="lp-btn quiet block" disabled={submitting}>Cancelar</button>
		</form>
		<form method="POST" action="?/decide" onsubmit={() => (submitting = true)}>
			<input type="hidden" name="request" value={consent.requestId} />
			<input type="hidden" name="decision" value="approve" />
			<button type="submit" class="lp-btn block" disabled={submitting}>Autorizar</button>
		</form>
	</div>
</PublicShell>

<style>
	.logo {
		display: block;
		width: 56px;
		height: 56px;
		margin-bottom: 24px;
		border-radius: 12px;
		object-fit: cover;
	}

	h2 {
		margin: 0 0 12px;
		font-size: var(--lp-fs-lead);
		font-weight: 600;
	}

	.scopes ul {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin: 0;
		padding-left: 20px;
		list-style: disc;
	}

	.scopes li {
		font-size: var(--lp-fs-lead);
		line-height: 1.5;
	}

	.details {
		margin: 32px 0 0;
		border-top: 1px solid var(--lp-rule);
	}

	.details div {
		display: grid;
		grid-template-columns: minmax(0, 160px) minmax(0, 1fr);
		gap: 4px 16px;
		padding: 12px 0;
		border-bottom: 1px solid var(--lp-rule);
		font-size: 15px;
	}

	.details dt {
		color: var(--lp-ink-2);
	}

	.details dd {
		margin: 0;
		min-width: 0;
		overflow-wrap: anywhere;
	}

	.details code {
		font-family: var(--font-mono);
		font-size: 14px;
	}

	.revoke {
		margin: 20px 0 0;
		font-size: 15px;
		color: var(--lp-ink-2);
	}

	.actions {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
		margin-top: 32px;
	}

	@media (max-width: 420px) {
		.details div {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
