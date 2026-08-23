<script lang="ts">
	import { onMount } from 'svelte';
	import { LAUNCH_DATE, countdownBetween, type Countdown } from '../landing';

	interface Props {
		/** Texto que precede a las cifras. */
		label?: string;
	}

	let { label = 'Faltan' }: Props = $props();

	let countdown = $state<Countdown>({ days: '00', hours: '00', mins: '00', secs: '00' });

	onMount(() => {
		const target = new Date(LAUNCH_DATE).getTime();
		const tick = () => (countdown = countdownBetween(target, Date.now()));
		tick();
		const id = setInterval(tick, 1000);
		return () => clearInterval(id);
	});
</script>

<!--
	La cuenta atrás ocupaba cuatro celdas con cifras de hasta 42px justo debajo
	del titular, así que competía con él y empujaba el formulario fuera de la
	primera pantalla. Aquí queda en una línea: sigue dando la urgencia, sin
	disputarle la jerarquía a nada.
-->
<div class="cd" aria-label="Cuenta regresiva para el lanzamiento">
	<span class="cd-label">{label}</span>
	<span class="cd-unit"><b>{countdown.days}</b>d</span>
	<span class="cd-unit"><b>{countdown.hours}</b>h</span>
	<span class="cd-unit"><b>{countdown.mins}</b>m</span>
</div>

<style>
	.cd {
		display: flex;
		align-items: baseline;
		gap: 8px;
		font-family: var(--font-mono);
		font-size: 11.5px;
		letter-spacing: 0.06em;
		color: var(--text-dim);
		white-space: nowrap;
		font-variant-numeric: tabular-nums;
	}

	.cd-label {
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	.cd-unit b {
		font-weight: 600;
		color: var(--text);
	}
</style>
