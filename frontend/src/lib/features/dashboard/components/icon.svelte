<script lang="ts">
	/** Pinta un icono de `icons.ts`. Decorativo: quien lo usa pone la etiqueta. */
	import { ICONS } from '../icons';

	interface Props {
		name: string;
		size?: number;
		/** Grosor del trazo; 1.75 es el del menú, 2 el de los botones. */
		width?: number;
	}

	let { name, size = 18, width = 1.75 }: Props = $props();

	const paths = $derived(ICONS[name] ?? []);
</script>

<svg
	{...{ width: size, height: size }}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width={width}
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
	focusable="false"
>
	{#each paths as d (d)}
		<path {d} />
	{/each}
</svg>
