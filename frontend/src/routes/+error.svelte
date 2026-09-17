<script lang="ts">
	/*
	 * La página de error, con el armazón de las pantallas públicas.
	 *
	 * Era una gráfica que subía y se desplomaba en rojo, y el texto no aparecía
	 * hasta pasados dos segundos, cuando terminaba la animación: quien llegaba
	 * con un lector de pantalla o con prisa no leía nada. Ahora lo primero que se
	 * ve es qué ha pasado, y a la derecha el mensaje concreto, la salida y los
	 * datos para contarlo si hace falta. Qué dice cada código lo decide
	 * `$lib/shared/http-error`.
	 */
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import PublicShell from '$lib/ui/public-shell.svelte';
	import { CONTACT_EMAIL } from '$lib/seo';
	import { errorActions, errorCopy } from '$lib/shared/http-error';

	const copy = $derived(errorCopy(page.status, page.error?.message));
	const actions = $derived(errorActions(page.status, page.url.pathname));
</script>

<svelte:head>
	<title>{copy.title} — Finexia</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<PublicShell>
	{#snippet heading()}
		<h1 class="shell-title">{copy.title}</h1>
		<p class="shell-note">{copy.hint}</p>
	{/snippet}

	{#if copy.detail}
		<p class="shell-copy detail">{copy.detail}</p>
	{/if}

	<div class="actions">
		{#each actions as action, i (action.label)}
			{#if 'reload' in action}
				<button
					type="button"
					class="lp-btn block"
					class:quiet={i > 0}
					onclick={() => location.reload()}
				>
					{action.label}
				</button>
			{:else}
				<a href={resolve(action.route)} class="lp-btn block" class:quiet={i > 0}>
					{action.label}
				</a>
			{/if}
		{/each}
	</div>

	{#if copy.contact}
		<p class="shell-after">
			¿Sigue pasando? Escríbenos a
			<a class="lp-link" href="mailto:{CONTACT_EMAIL}">{CONTACT_EMAIL}</a>
			con la dirección y el código de abajo.
		</p>
	{/if}

	<!--
		La dirección en monoespaciada a propósito: es lo que hay que revisar
		letra a letra para ver si el enlace venía mal escrito.
	-->
	<dl class="facts">
		<div>
			<dt>Dirección</dt>
			<dd><code>{page.url.pathname}</code></dd>
		</div>
		<div>
			<dt>Código</dt>
			<dd>{page.status}</dd>
		</div>
	</dl>
</PublicShell>

<style>
	/* El margen va en lo que sigue: el del mensaje lo fija el armazón. */
	.detail + .actions {
		margin-top: 28px;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.facts {
		margin: 40px 0 0;
		border-top: 1px solid var(--lp-rule);
		font-size: var(--lp-fs-sm);
	}

	.facts div {
		display: grid;
		grid-template-columns: 96px minmax(0, 1fr);
		gap: 16px;
		padding: 10px 0;
		border-bottom: 1px solid var(--lp-rule);
	}

	.facts dt {
		color: var(--lp-ink-2);
	}

	.facts dd {
		margin: 0;
		min-width: 0;
		overflow-wrap: anywhere;
	}

	.facts code {
		font-family: var(--font-mono);
		font-size: 13px;
	}
</style>
