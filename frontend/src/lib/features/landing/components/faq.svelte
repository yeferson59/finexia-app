<script lang="ts">
	interface Faq {
		q: string;
		a: string;
	}

	interface Props {
		faqs: Faq[];
	}

	let { faqs }: Props = $props();

	/*
	 * La primera abierta de entrada: se ve el patrón de pregunta y respuesta sin
	 * tener que adivinar que la fila es pulsable.
	 */
	let openFaqIndex = $state<number | null>(0);

	function toggleFaq(index: number) {
		openFaqIndex = openFaqIndex === index ? null : index;
	}
</script>

<section class="wrap block" id="faq">
	<!-- De columna centrada de 720px a raíl más lista: la cabecera deja de
	     empujar las preguntas media pantalla hacia abajo. -->
	<div class="sec-rail">
		<div class="faq-head reveal">
			<div class="eyebrow">Preguntas frecuentes</div>
			<h2 class="sec-title sec-title-sm">Lo que necesitas saber</h2>
			<div class="faq-count">{faqs.length} preguntas</div>
		</div>

		<div class="faq">
			{#each faqs as faq, i (faq.q)}
				<div class="faq-item reveal" class:open={openFaqIndex === i}>
					<h3>
						<button
							id="faq-q-{i}"
							class="faq-q"
							type="button"
							aria-expanded={openFaqIndex === i}
							aria-controls="faq-a-{i}"
							onclick={() => toggleFaq(i)}
						>
							{faq.q}<span class="plus" aria-hidden="true"></span>
						</button>
					</h3>
					<!--
						La respuesta se abre con grid-template-rows 0fr → 1fr en vez de con
						un max-height fijo: así el alto lo decide el texto y una respuesta
						larga no queda recortada en pantallas estrechas.
					-->
					<div id="faq-a-{i}" class="faq-a" role="region" aria-labelledby="faq-q-{i}">
						<div class="faq-a-inner"><p>{faq.a}</p></div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>

<style>
	.faq-count {
		display: inline-flex;
		align-items: center;
		margin-top: 24px;
		padding: 6px 12px;
		border: 1px solid var(--border);
		border-radius: 999px;
		background: var(--surface);
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.faq-item {
		border-bottom: 1px solid var(--border);
	}
	.faq-item:last-child {
		border-bottom: none;
	}
	.faq-item h3 {
		margin: 0;
		font-weight: inherit;
		font-size: inherit;
	}
	.faq-q {
		width: 100%;
		background: none;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		padding: 22px 0;
		text-align: left;
		font-family: var(--font-display);
		font-weight: 300;
		font-size: 18px;
		color: var(--text);
		transition: color 0.2s;
	}
	.faq-q:hover {
		color: var(--amber-light);
	}
	.faq-item.open .faq-q {
		color: var(--amber-light);
	}
	.plus {
		flex-shrink: 0;
		width: 22px;
		height: 22px;
		position: relative;
		transition: transform 0.3s ease;
	}
	.plus::before,
	.plus::after {
		content: '';
		position: absolute;
		background: var(--amber);
		border-radius: 1px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
	}
	.plus::before {
		width: 12px;
		height: 1.5px;
	}
	.plus::after {
		width: 1.5px;
		height: 12px;
		transition: opacity 0.3s ease;
	}
	.faq-item.open .plus {
		transform: rotate(90deg);
	}
	.faq-item.open .plus::after {
		opacity: 0;
	}
	.faq-a {
		display: grid;
		grid-template-rows: 0fr;
		transition: grid-template-rows 0.35s ease;
	}
	.faq-item.open .faq-a {
		grid-template-rows: 1fr;
	}
	.faq-a-inner {
		overflow: hidden;
	}
	.faq-a p {
		max-width: 68ch;
		font-size: 15px;
		color: var(--text-muted);
		line-height: 1.68;
		font-weight: 300;
		padding: 0 40px 24px 0;
		text-wrap: pretty;
	}

	@media (max-width: 640px) {
		.faq-q {
			font-size: 17px;
		}
		.faq-a p {
			padding-right: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.faq-a,
		.plus,
		.plus::after {
			transition: none;
		}
	}
</style>
