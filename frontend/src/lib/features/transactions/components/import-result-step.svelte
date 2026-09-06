<script lang="ts">
	/*
	 * Paso tres: qué acabó de pasar.
	 *
	 * Era una palomita ámbar de 2,2rem centrada, un título y, muy por debajo, las
	 * filas omitidas en una caja con scroll. Pero cuando algo se queda fuera, eso
	 * es la noticia: la palomita celebraba una importación a medias y escondía la
	 * parte que el usuario tiene que arreglar en su hoja. Ahora lo que pasó se
	 * dice en una frase y lo que falta va justo detrás, con el número de fila
	 * delante para poder volver al archivo.
	 */
	import Button from '$lib/ui/button.svelte';
	import type { ImportResult } from '../types';

	let {
		result,
		onRestart,
		onViewTransactions
	}: {
		result: ImportResult;
		onRestart: () => void;
		onViewTransactions: () => void;
	} = $props();

	const headline = $derived(
		result.imported === 0
			? 'No entró ninguna fila.'
			: result.skipped === 0
				? `Listo: tus ${result.imported} transacciones están en el portafolio.`
				: `${result.imported} de ${result.totalRows} filas están en tu portafolio.`
	);
</script>

<section class="result">
	<h2>{headline}</h2>

	{#if result.skipped > 0}
		<p class="lede">
			{result.skipped === 1 ? 'Una fila se quedó' : `${result.skipped} filas se quedaron`} fuera. Corrígelas
			en tu hoja y vuelve a subirla: las que ya entraron no se duplican, se omiten por repetidas.
		</p>
	{:else if result.imported > 0}
		<p class="lede">Ya puedes verlas en tu libro de movimientos.</p>
	{/if}

	{#if result.errors.length > 0}
		<ul class="skipped">
			{#each result.errors as error (error.row)}
				<li>
					<span class="line figure">Fila {error.row}</span>
					<span class="reason">{error.message}</span>
				</li>
			{/each}
		</ul>
	{/if}

	<div class="actions">
		<Button type="button" variant="primary" onclick={onViewTransactions}>
			Ver mis transacciones
		</Button>
		<button type="button" class="quiet-action" onclick={onRestart}> Importar otro archivo </button>
	</div>
</section>

<style>
	.result {
		padding-top: 2.25rem;
		border-top: 1px solid var(--border-strong);
	}

	h2 {
		max-width: 24ch;
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(1.6rem, 3.5vw, 2.1rem);
		font-weight: 300;
		line-height: 1.2;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.lede {
		max-width: 58ch;
		margin: 1rem 0 0;
		font-size: 0.9rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	.skipped {
		max-width: 58ch;
		margin: 1.75rem 0 0;
		padding: 0;
		list-style: none;
		max-height: 20rem;
		overflow-y: auto;
	}

	.skipped li {
		display: grid;
		grid-template-columns: minmax(0, 5.5rem) minmax(0, 1fr);
		gap: 0.25rem 1rem;
		padding: 0.7rem 0;
		border-bottom: 1px solid var(--border);
		font-size: 0.83rem;
		line-height: 1.5;
	}

	.skipped li:last-child {
		border-bottom: none;
	}

	.line {
		color: var(--text-dim);
	}

	.reason {
		min-width: 0;
		color: var(--text);
	}

	.actions {
		margin-top: 2.25rem;
	}

	@media (max-width: 560px) {
		.skipped li {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
