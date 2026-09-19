<script lang="ts">
	/*
	 * El cuerpo del artículo.
	 *
	 * Llega como HTML porque el Markdown se convierte en el build (`posts.ts`),
	 * así que se pinta con `{@html}` y sus estilos se alcanzan con `:global`
	 * dentro de `.post-body`: Svelte los compila con el hash de este componente,
	 * o sea que llegan a los hijos sin escaparse a toda la página. Es el mismo
	 * reparto que usa el layout de las páginas legales.
	 *
	 * Sin sanear a propósito: el HTML sale de un `.md` del repositorio, que pasa
	 * por revisión como cualquier otro archivo. Si algún día el contenido
	 * viniera de fuera —un CMS, un formulario—, esto tendría que sanearse antes
	 * de pintarlo.
	 */
	interface Props {
		html: string;
	}

	let { html }: Props = $props();
</script>

<div class="post-body">
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html html}
</div>

<style>
	.post-body {
		max-width: 720px;
		margin-top: 44px;
	}

	/* ── Texto corrido ────────────────────────────────────────────────────── */

	.post-body :global(p) {
		margin: 0 0 20px;
		font-size: 17px;
		line-height: 1.75;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.post-body :global(strong) {
		font-weight: 620;
		color: var(--lp-ink);
	}

	.post-body :global(a) {
		color: var(--lp-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.post-body :global(a:hover) {
		text-decoration-thickness: 2px;
	}

	/* ── Encabezados ──────────────────────────────────────────────────────── */

	/*
	 * `posts.ts` le pone un `id` a cada uno, así que se puede enlazar a una
	 * sección concreta; el `scroll-margin` evita que el ancla quede pegada al
	 * borde de la ventana al saltar.
	 */
	.post-body :global(h2),
	.post-body :global(h3) {
		scroll-margin-top: 24px;
		color: var(--lp-ink);
	}

	.post-body :global(h2) {
		margin: 48px 0 16px;
		font-size: 28px;
		font-stretch: 110%;
		font-weight: 620;
		line-height: 1.2;
		letter-spacing: -0.015em;
		text-wrap: balance;
	}

	.post-body :global(h3) {
		margin: 34px 0 12px;
		font-size: 20px;
		font-weight: 600;
		line-height: 1.3;
	}

	/* ── Listas ───────────────────────────────────────────────────────────── */

	.post-body :global(ul),
	.post-body :global(ol) {
		margin: 0 0 22px;
		padding-left: 24px;
	}

	.post-body :global(ul) {
		list-style: disc;
	}

	.post-body :global(ol) {
		list-style: decimal;
	}

	.post-body :global(li) {
		margin-bottom: 10px;
		font-size: 17px;
		line-height: 1.7;
		color: var(--lp-ink-2);
	}

	.post-body :global(li > p) {
		margin-bottom: 10px;
	}

	/* ── Citas, filetes y figuras ─────────────────────────────────────────── */

	.post-body :global(blockquote) {
		margin: 32px 0;
		padding: 4px 0 4px 22px;
		border-left: 2px solid var(--lp-ink);
	}

	.post-body :global(blockquote p) {
		margin-bottom: 8px;
		font-size: 19px;
		line-height: 1.6;
		color: var(--lp-ink);
	}

	.post-body :global(blockquote p:last-child) {
		margin-bottom: 0;
	}

	.post-body :global(hr) {
		margin: 44px 0;
		border: 0;
		border-top: 1px solid var(--lp-rule);
	}

	.post-body :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 32px 0;
		border-radius: 10px;
	}

	/* ── Código ───────────────────────────────────────────────────────────── */

	.post-body :global(code) {
		padding: 2px 6px;
		border-radius: 4px;
		background: var(--lp-paper-2);
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 0.88em;
		color: var(--lp-ink);
	}

	.post-body :global(pre) {
		margin: 28px 0;
		padding: 18px 20px;
		overflow-x: auto;
		border: 1px solid var(--lp-rule);
		border-radius: 10px;
		background: var(--lp-paper-2);
	}

	.post-body :global(pre code) {
		padding: 0;
		background: none;
		font-size: 14px;
		line-height: 1.6;
	}

	/* ── Tablas ───────────────────────────────────────────────────────────── */

	.post-body :global(table) {
		width: 100%;
		margin: 28px 0;
		border-collapse: collapse;
		font-size: 15px;
	}

	.post-body :global(th),
	.post-body :global(td) {
		padding: 10px 12px;
		border-bottom: 1px solid var(--lp-rule);
		text-align: left;
		vertical-align: top;
	}

	.post-body :global(th) {
		font-weight: 620;
		color: var(--lp-ink);
	}

	.post-body :global(td) {
		color: var(--lp-ink-2);
	}

	@media (max-width: 640px) {
		.post-body {
			margin-top: 32px;
		}
		.post-body :global(p),
		.post-body :global(li) {
			font-size: 16px;
		}
		.post-body :global(h2) {
			margin-top: 38px;
			font-size: 24px;
		}
	}
</style>
