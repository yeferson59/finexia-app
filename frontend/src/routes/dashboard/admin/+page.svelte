<script lang="ts">
	/*
	 * La portada de administración abre diciendo qué hay que hacer, y por eso el
	 * `<h1>` es esa frase y no el nombre de la pantalla: dónde estás ya lo dicen
	 * el menú y la cabecera del panel, y repetirlo aquí gastaba lo único que se
	 * lee entero antes de decidir a dónde ir.
	 *
	 * Sustituye a cuatro tarjetas de cifras y tres de atajos, que se pintaban
	 * iguales estuviera el sistema al día o hubiera doce precios de hace un mes.
	 */
	import { resolve } from '$app/paths';
	import { Worklist, buildWorklist, describeDesk } from '$lib/features/admin';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const tasks = $derived(buildWorklist(data.desk));
</script>

<svelte:head>
	<title>Administración — FINEXIA</title>
	<meta name="description" content="Quién entra a Finexia y qué dicen los datos compartidos" />
</svelte:head>

<h1 class="state">{describeDesk(tasks)}</h1>

{#if tasks.length > 0}
	<Worklist {tasks} />
{/if}

<nav class="record" aria-label="Lo que hay guardado">
	<a class="entry" href={resolve('/dashboard/admin/users')}>
		<span class="label">Cuentas</span>
		<span class="figure">{data.totalUsers}</span>
	</a>
	<a class="entry" href={resolve('/dashboard/admin/assets')}>
		<span class="label">Activos del catálogo</span>
		<span class="figure">{data.totalAssets}</span>
	</a>
	<a class="entry" href={resolve('/dashboard/admin/exchange-rates')}>
		<span class="label">Tasas compartidas</span>
		<span class="figure">{data.totalRates}</span>
	</a>
</nav>

<style>
	.state {
		max-width: 26ch;
		margin: 0 0 3rem;
		font-family: var(--font-display);
		font-size: clamp(1.7rem, 3.4vw, 2.4rem);
		font-weight: 300;
		line-height: 1.18;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	/*
	 * Lo que hay guardado, que no es lo mismo que lo que hay pendiente: las
	 * cifras van en el color del texto y las de la lista de tareas en el ámbar
	 * apagado, al mismo cuerpo. Son además las tres únicas puertas de esta
	 * pantalla, así que hacen de navegación sin necesitar tres tarjetas con
	 * icono.
	 */
	/* Sin filete propio: cuando hay tareas ya lo puso la última de la lista, y
	   cuando no las hay el titular no necesita que lo subrayen. */
	.record {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
		gap: 2rem;
	}

	.entry {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		text-decoration: none;
	}

	.label {
		font-size: 0.83rem;
		color: var(--text-muted);
	}

	.entry:hover .label {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 4px;
	}

	.figure {
		font-family: var(--font-mono);
		font-size: 1.5rem;
		font-weight: 400;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}
</style>
