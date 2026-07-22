<script lang="ts">
	interface Faq {
		q: string;
		a: string;
	}

	interface Props {
		faqs: Faq[];
	}

	let { faqs }: Props = $props();

	let openFaqIndex = $state<number | null>(null);

	function toggleFaq(index: number) {
		openFaqIndex = openFaqIndex === index ? null : index;
	}
</script>

<section class="block wrap" id="faq">
	<div class="sec-head reveal">
		<div class="eyebrow">Preguntas frecuentes</div>
		<h2 class="sec-title">Lo que necesitas saber</h2>
	</div>
	<div class="faq">
		{#each faqs as faq, i (faq.q)}
			<div class="faq-item reveal" class:open={openFaqIndex === i}>
				<button class="faq-q" onclick={() => toggleFaq(i)}>
					{faq.q}<span class="plus" aria-hidden="true"></span>
				</button>
				<div class="faq-a"><p>{faq.a}</p></div>
			</div>
		{/each}
	</div>
</section>

<style>
	.faq {
		max-width: 720px;
		margin: 0 auto;
	}
	.faq-item {
		border-bottom: 1px solid var(--border);
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
		padding: 24px 0;
		text-align: left;
		font-family: var(--font-display);
		font-weight: 300;
		font-size: 18px;
		color: var(--text);
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
		max-height: 0;
		overflow: hidden;
		transition:
			max-height 0.35s ease,
			padding 0.35s ease;
	}
	.faq-a p {
		font-size: 15px;
		color: var(--text-muted);
		line-height: 1.68;
		font-weight: 300;
		padding: 0 40px 4px 0;
	}
	.faq-item.open .faq-a {
		max-height: 280px;
		padding-bottom: 24px;
	}
</style>
