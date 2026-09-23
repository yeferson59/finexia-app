<script lang="ts">
	interface Faq {
		q: string;
		a: string;
		/** Una salida tras la respuesta, cuando la respuesta sigue en otra página. */
		link?: { href: string; label: string };
	}

	interface Props {
		faqs: Faq[];
	}

	let { faqs }: Props = $props();

	/*
	 * La primera abierta de entrada: se ve el patrón de pregunta y respuesta sin
	 * tener que adivinar que la fila se puede pulsar.
	 */
	let openIndex = $state<number | null>(0);

	function toggle(index: number) {
		openIndex = openIndex === index ? null : index;
	}
</script>

<section class="lp-section" id="faq" aria-labelledby="faq-title">
	<div class="lp-wrap grid">
		<h2 class="lp-h2" id="faq-title">Preguntas frecuentes</h2>

		<div class="list">
			{#each faqs as faq, i (faq.q)}
				<div class="item" class:open={openIndex === i}>
					<h3>
						<button
							id="faq-q-{i}"
							class="q"
							type="button"
							aria-expanded={openIndex === i}
							aria-controls="faq-a-{i}"
							onclick={() => toggle(i)}
						>
							<span>{faq.q}</span>
							<span class="sign" aria-hidden="true"></span>
						</button>
					</h3>
					<!--
						La respuesta se abre con grid-template-rows 0fr → 1fr en vez de con
						un max-height fijo: así el alto lo decide el texto y una respuesta
						larga no queda recortada en pantallas estrechas.
					-->
					<div id="faq-a-{i}" class="a" role="region" aria-labelledby="faq-q-{i}">
						<div class="a-inner">
							<p>{faq.a}</p>
							{#if faq.link}
								<!-- Quien arma la lista ya resolvió la ruta. -->
								<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
								<a class="more" href={faq.link.href}>{faq.link.label}</a>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>

<style>
	.grid {
		display: grid;
		grid-template-columns: minmax(0, 4fr) minmax(0, 7fr);
		gap: 32px 96px;
		align-items: start;
	}

	.grid .lp-h2 {
		max-width: 10ch;
	}

	.list {
		border-top: 2px solid var(--lp-ink);
	}

	.item {
		border-bottom: 1px solid var(--lp-rule);
	}

	.item h3 {
		margin: 0;
		font-size: inherit;
		font-weight: inherit;
	}

	.q {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		width: 100%;
		padding: 22px 0;
		border: none;
		background: none;
		color: var(--lp-ink);
		font-family: var(--lp-font);
		font-size: 19px;
		font-stretch: 104%;
		font-weight: 560;
		line-height: 1.3;
		text-align: left;
		cursor: pointer;
	}

	.q:hover span:first-child {
		text-decoration: underline;
		text-decoration-thickness: 1px;
		text-underline-offset: 4px;
	}

	/* Un más que pierde el trazo vertical al abrirse: menos. */
	.sign {
		position: relative;
		flex-shrink: 0;
		width: 16px;
		height: 16px;
	}

	.sign::before,
	.sign::after {
		content: '';
		position: absolute;
		top: 50%;
		left: 50%;
		background: var(--lp-ink);
		transform: translate(-50%, -50%);
	}

	.sign::before {
		width: 16px;
		height: 2px;
	}

	.sign::after {
		width: 2px;
		height: 16px;
		transition: transform 0.25s ease;
	}

	.item.open .sign::after {
		transform: translate(-50%, -50%) scaleY(0);
	}

	.a {
		display: grid;
		grid-template-rows: 0fr;
		transition: grid-template-rows 0.3s ease;
	}

	.item.open .a {
		grid-template-rows: 1fr;
	}

	.a-inner {
		overflow: hidden;
	}

	.a p {
		max-width: 62ch;
		margin: 0;
		padding: 0 40px 24px 0;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	/* Un solo margen inferior por respuesta: cuando hay enlace, lo pone él. */
	.a p:has(+ .more) {
		padding-bottom: 12px;
	}

	.more {
		display: inline-block;
		margin-bottom: 24px;
		font-weight: 600;
		color: var(--lp-ink);
		text-decoration: underline;
		text-decoration-thickness: 1px;
		text-underline-offset: 4px;
	}

	.more:hover {
		text-decoration-thickness: 2px;
	}

	@media (max-width: 900px) {
		.grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.grid .lp-h2 {
			max-width: none;
		}
	}

	@media (max-width: 640px) {
		.q {
			font-size: 17px;
		}
		.a p {
			padding-right: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.a,
		.sign::after {
			transition: none;
		}
	}
</style>
