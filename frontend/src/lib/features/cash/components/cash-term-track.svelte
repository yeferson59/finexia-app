<script lang="ts">
	/**
	 * El plazo de un depósito como una línea: el día en que se abrió a la
	 * izquierda, el del vencimiento a la derecha y el tramo recorrido hasta hoy.
	 *
	 * Es la pregunta que se le hace a un CDT —cuánto le falta— y como fecha suelta
	 * obligaba a restar. El tramo recorrido va en el verde de lo que rinde: son
	 * los días que ya ganó.
	 */
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { cashDepositTerm } from '../yield';

	interface Props {
		openedOn: string;
		maturesOn: string;
		today: string;
		/** Qué se lee debajo de la línea: los días que faltan o lo que ganará. */
		caption?: string;
	}

	let { openedOn, maturesOn, today, caption }: Props = $props();

	const term = $derived(cashDepositTerm(openedOn, maturesOn, today));

	const short = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'short' }).replace('.', '');

	const long = (iso: string) =>
		formatCalendarDate(iso.slice(0, 10), { day: 'numeric', month: 'long', year: 'numeric' });

	const status = $derived.by(() => {
		if (!term) return '';
		if (term.elapsed === 0 && today < openedOn.slice(0, 10)) return `Empieza el ${long(openedOn)}`;
		if (term.left === 0) {
			return today === maturesOn.slice(0, 10) ? 'Vence hoy' : `Venció el ${long(maturesOn)}`;
		}
		return term.left === 1 ? 'Falta 1 día' : `Faltan ${term.left} días de ${term.total}`;
	});
</script>

{#if term}
	<div class="term" style:--progress={term.progress}>
		<span class="edge">{short(openedOn)}</span>
		<span
			class="track"
			role="meter"
			aria-valuemin={0}
			aria-valuemax={term.total}
			aria-valuenow={term.elapsed}
			aria-valuetext="{status}: del {long(openedOn)} al {long(maturesOn)}"
			aria-label="Plazo del depósito"
		>
			<span class="run"></span>
			<span class="now"></span>
		</span>
		<span class="edge">{short(maturesOn)}</span>
		<span class="status">
			{status}{#if caption}<span class="caption">{caption}</span>{/if}
		</span>
	</div>
{/if}

<style>
	.term {
		display: grid;
		grid-template-columns: auto minmax(4rem, 1fr) auto;
		grid-template-areas:
			'from track to'
			'status status status';
		align-items: center;
		column-gap: 0.6rem;
		row-gap: 0.25rem;
		max-width: 26rem;
	}

	.edge {
		font-size: 0.72rem;
		white-space: nowrap;
		color: var(--text-dim);
	}

	.edge:first-child {
		grid-area: from;
	}

	.edge:nth-child(3) {
		grid-area: to;
	}

	.track {
		grid-area: track;
		position: relative;
		height: 2px;
		border-radius: 2px;
		background: var(--border-strong);
	}

	.run {
		position: absolute;
		inset: 0 auto 0 0;
		width: calc(var(--progress) * 100%);
		border-radius: 2px;
		background: var(--green);
	}

	/* Hoy: un punto con el anillo del fondo, para que se lea encima de la línea. */
	.now {
		position: absolute;
		top: 50%;
		left: calc(var(--progress) * 100%);
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--green);
		box-shadow: 0 0 0 2px var(--ring, var(--bg));
		transform: translate(-50%, -50%);
	}

	.status {
		grid-area: status;
		display: flex;
		flex-wrap: wrap;
		gap: 0 0.75rem;
		font-size: 0.76rem;
		color: var(--text-muted);
	}

	.caption {
		color: var(--text-dim);
	}
</style>
