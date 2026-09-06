<script lang="ts">
	/*
	 * Lo que rindió la cuenta, en cuánto tiempo y sobre cuánto dinero.
	 *
	 * Ocupa el sitio de seis mosaicos del panel «Estadísticas clave» —periodo
	 * cubierto, capital invertido, valor actual, rentabilidad del periodo,
	 * anualizada y ganancia— que estaban repartidos entre el primer bloque y el
	 * último de la misma tarjeta, así que las dos cifras que hay que leer juntas
	 * caían a treinta centímetros una de otra.
	 *
	 * La rentabilidad y la ganancia sobre coste miden cosas distintas y a menudo
	 * no se parecen. Antes eso vivía en un `title` que en un móvil no se abre;
	 * aquí se dice en la línea de debajo, y solo cuando de verdad se separan.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { formatLongDate, type RecordSummary } from '../reports';

	let { record }: { record: RecordSummary } = $props();

	const money = (amount: number) => privacy.money(formatCurrency(amount, record.currency));

	/*
	 * Cuándo hay que explicar que las dos cifras no coinciden: dos puntos
	 * porcentuales. Por debajo la diferencia es ruido de redondeo y la nota
	 * sobraría; por encima, quien vea un −0,3 % encima de un +14,0 % sin más
	 * necesita saber que ninguna de las dos está mal.
	 */
	const diverge = $derived(
		record.periodReturn !== null && Math.abs(record.periodReturn - record.gainPct) >= 2
	);

	/* Una ganancia que redondea a cero no es una ganancia y no se pinta en verde. */
	const flat = $derived(Math.abs(record.gain) < 0.005);

	/* La frase se compone aquí y no en la plantilla: partida en dos por un
	   `{#if}` en mitad del párrafo, se quedaba sin el espacio del punto. */
	const historyLine = $derived.by(() => {
		const span = `Del ${formatLongDate(record.from)} al ${formatLongDate(record.to)}, ${record.months} ${record.months === 1 ? 'mes' : 'meses'} de historial.`;

		return record.annualized === null
			? span
			: `${span} Anualizado, eso es un ${formatSignedPercent(record.annualized)}.`;
	});
</script>

<section class="record" aria-labelledby="record-return">
	{#if record.periodReturn === null}
		<h2 class="label" id="record-return">Tu cuenta vale</h2>
		<p class="amount">{money(record.value)}</p>
		<p class="span">
			Todavía no hay dos cierres que comparar, así que la rentabilidad tendrá que esperar al de
			mañana.
		</p>
	{:else}
		<h2 class="label" id="record-return">Lo que rindió tu dinero</h2>
		<p class="amount">{formatSignedPercent(record.periodReturn)}</p>
		<p class="span">{historyLine}</p>
	{/if}

	{#if record.cost > 0}
		<p class="money" class:up={!flat && record.gain > 0} class:down={!flat && record.gain < 0}>
			Hoy la cuenta vale {money(record.value)} sobre los {money(record.cost)} que has puesto:
			{#if flat}
				ni ganancia ni pérdida.
			{:else}
				{record.gain > 0 ? '+' : '−'}{money(Math.abs(record.gain))} ({formatSignedPercent(
					record.gainPct
				)}).
			{/if}
		</p>
	{:else}
		<p class="money">Todavía no hay capital invertido que comparar.</p>
	{/if}

	{#if diverge}
		<p class="qualifier">
			Las dos cifras no dicen lo mismo y ninguna está mal: la de arriba mide cómo se comportó el
			dinero mientras estuvo dentro, y la ganancia depende además de cuándo lo aportaste.
		</p>
	{/if}
</section>

<style>
	.record {
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--border);
	}

	/* Nombra la cifra de debajo, así que no es una etiqueta de adorno. En caja
	   normal: la página tenía cuatro antetítulos en versalitas. */
	.label {
		margin: 0 0 0.5rem;
		font-family: var(--font-body);
		font-size: 0.9rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	.amount {
		margin: 0;
		font-family: var(--font-mono);
		font-size: clamp(2rem, 4.5vw, 2.75rem);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.03em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.span,
	.money,
	.qualifier {
		max-width: 64ch;
		margin: 0.85rem 0 0;
		font-size: 0.9rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.money {
		margin-top: 0.35rem;
	}

	.money.up {
		color: var(--green);
	}

	.money.down {
		color: var(--red);
	}

	/* En ámbar porque cambia cómo se leen las dos cifras de arriba, no porque
	   haga falta adornar el párrafo. */
	.qualifier {
		margin-top: 0.9rem;
		padding-left: 0.75rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		font-size: 0.85rem;
	}
</style>
