<script lang="ts">
	/* Índice de la guía, con la misma numeración que el documento. */
	import CardHeader from '$lib/ui/card-header.svelte';
	import { manual } from '../manual-meta';
</script>

<section class="contents">
	<CardHeader eyebrow="Índice" title="Qué encontrarás en la guía" />

	<ol class="list">
		{#each manual.sections as section (section.number)}
			<li>
				<span class="num">{String(section.number).padStart(2, '0')}</span>
				<span class="title">{section.title}</span>
			</li>
		{/each}
	</ol>
</section>

<style>
	.contents {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 1.75rem;
		backdrop-filter: blur(10px);
	}

	.list {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0 2.5rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.list li {
		display: flex;
		align-items: baseline;
		gap: 0.85rem;
		padding: 0.6rem 0;
		border-bottom: 1px solid var(--border);
		font-size: 0.87rem;
		color: var(--text-muted);
	}

	.num {
		font-family: var(--font-mono);
		font-size: 0.68rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		color: var(--amber);
		flex-shrink: 0;
	}

	.title {
		min-width: 0;
		overflow-wrap: anywhere;
	}

	@media (max-width: 768px) {
		.contents {
			padding: 1.35rem;
		}

		.list {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
