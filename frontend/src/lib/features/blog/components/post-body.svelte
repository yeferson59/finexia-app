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
	/*
	 * El texto corrido va en Fraunces y en tinta plena, no en el gris de la
	 * interfaz: es lo que se viene a leer. A 19px y 1,7 de interlínea la línea
	 * se queda en unos 65 caracteres. Los encabezados, las tablas y el código
	 * vuelven a la Archivo de la interfaz: son para orientarse, no para leer.
	 */
	.post-body {
		max-width: 680px;
		font-family: var(--blog-serif);
		font-optical-sizing: auto;
		color: var(--lp-ink);
	}

	.post-body > :global(:first-child) {
		margin-top: 0;
	}

	/* ── Texto corrido ────────────────────────────────────────────────────── */

	.post-body :global(p) {
		margin: 0 0 24px;
		font-size: 19px;
		font-weight: 380;
		line-height: 1.7;
		text-wrap: pretty;
		hanging-punctuation: first;
	}

	.post-body :global(strong) {
		font-weight: 600;
	}

	/* Subrayado ámbar: la marca en el único sitio donde el texto la admite sin
	   perder contraste, porque la letra sigue en tinta. */
	.post-body :global(a) {
		color: var(--lp-ink);
		text-decoration: underline;
		text-decoration-color: var(--lp-jubilacion);
		text-decoration-thickness: 2px;
		text-underline-offset: 4px;
	}

	.post-body :global(a:hover) {
		text-decoration-color: var(--lp-ink);
	}

	/* ── Encabezados ──────────────────────────────────────────────────────── */

	/*
	 * `posts.ts` le pone un `id` a cada uno, así que se puede enlazar a una
	 * sección concreta; el `scroll-margin` deja el ancla por debajo de la
	 * cabecera fija al saltar desde el índice del margen.
	 */
	.post-body :global(h2),
	.post-body :global(h3) {
		scroll-margin-top: calc(var(--blog-header, 72px) + 28px);
		font-family: var(--lp-font);
		color: var(--lp-ink);
		text-wrap: balance;
	}

	.post-body :global(h2) {
		margin: 64px 0 18px;
		font-size: 30px;
		font-stretch: 116%;
		font-weight: 630;
		line-height: 1.12;
		letter-spacing: -0.02em;
	}

	.post-body :global(h3) {
		margin: 40px 0 12px;
		font-size: 21px;
		font-stretch: 108%;
		font-weight: 620;
		line-height: 1.3;
	}

	/* ── Listas ───────────────────────────────────────────────────────────── */

	.post-body :global(ul),
	.post-body :global(ol) {
		margin: 0 0 26px;
		padding-left: 26px;
	}

	.post-body :global(ul) {
		list-style: disc;
	}

	.post-body :global(ol) {
		list-style: decimal;
	}

	.post-body :global(li) {
		margin-bottom: 10px;
		padding-left: 4px;
		font-size: 19px;
		font-weight: 380;
		line-height: 1.6;
	}

	.post-body :global(li::marker) {
		font-family: var(--lp-font);
		font-size: 0.85em;
		font-variant-numeric: tabular-nums;
		color: var(--lp-ink-2);
	}

	.post-body :global(li > p) {
		margin-bottom: 10px;
	}

	/* ── Citas, filetes y figuras ─────────────────────────────────────────── */

	.post-body :global(blockquote) {
		margin: 40px 0;
		padding: 2px 0 2px 24px;
		border-left: 3px solid var(--lp-jubilacion);
	}

	.post-body :global(blockquote p) {
		margin-bottom: 10px;
		font-size: 23px;
		font-style: italic;
		font-weight: 350;
		line-height: 1.5;
	}

	.post-body :global(blockquote p:last-child) {
		margin-bottom: 0;
	}

	.post-body :global(hr) {
		margin: 56px 0;
		border: 0;
		border-top: 1px solid var(--lp-rule);
	}

	.post-body :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 36px 0;
		border-radius: 6px;
	}

	/* ── Código ───────────────────────────────────────────────────────────── */

	.post-body :global(code) {
		padding: 2px 6px;
		border-radius: 4px;
		background: var(--lp-paper-2);
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 0.8em;
		color: var(--lp-ink);
	}

	.post-body :global(pre) {
		margin: 32px 0;
		padding: 18px 20px;
		overflow-x: auto;
		border: 1px solid var(--lp-rule);
		border-radius: 6px;
		background: var(--lp-paper-2);
	}

	.post-body :global(pre code) {
		padding: 0;
		background: none;
		font-size: 14px;
		line-height: 1.6;
	}

	/* ── Tablas ───────────────────────────────────────────────────────────── */

	/*
	 * Las tablas del blog son las de un extracto —portafolio, horizonte,
	 * riesgo—, así que se leen en la letra de la interfaz, con cifras
	 * tabulares y un filete en tinta bajo la cabecera. En pantallas estrechas
	 * se desplazan dentro de su caja en vez de romper la página.
	 */
	.post-body :global(table) {
		display: block;
		width: 100%;
		max-width: 100%;
		margin: 36px 0;
		overflow-x: auto;
		border-collapse: collapse;
		font-family: var(--lp-font);
		font-size: 15px;
		font-variant-numeric: tabular-nums;
		line-height: 1.45;
	}

	.post-body :global(th),
	.post-body :global(td) {
		padding: 11px 16px 11px 0;
		border-bottom: 1px solid var(--lp-rule);
		text-align: left;
		vertical-align: top;
	}

	.post-body :global(th) {
		border-bottom-color: var(--lp-ink);
		font-weight: 600;
		white-space: nowrap;
		color: var(--lp-ink);
	}

	.post-body :global(td) {
		color: var(--lp-ink-2);
	}

	.post-body :global(td:first-child) {
		font-weight: 600;
		color: var(--lp-ink);
	}

	@media (max-width: 640px) {
		.post-body :global(p),
		.post-body :global(li) {
			font-size: 18px;
		}
		.post-body :global(h2) {
			margin-top: 48px;
			font-size: 25px;
		}
		.post-body :global(blockquote p) {
			font-size: 20px;
		}
	}
</style>
