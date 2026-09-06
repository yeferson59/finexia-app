<script lang="ts">
	/**
	 * Una sección de ajustes: el carril de `ui/rail-section` con la variante que
	 * usa esta página.
	 *
	 * Sustituyó en su día a la tarjeta que envolvía cada sección. La página era
	 * una rejilla de ocho tarjetas de dos columnas, así que un párrafo de dos
	 * líneas se estiraba hasta el alto del formulario de al lado y el orden de
	 * lectura iba en zigzag. En un carril, el ojo recorre solo los títulos hasta
	 * encontrar el recado que trae a esta página.
	 *
	 * Lo propio de aquí son dos decisiones, y por eso el componente sigue
	 * existiendo en vez de repetirlas en las siete secciones: el título va en
	 * `<h3>` porque el `<h2>` lo pone el grupo que las envuelve, y el filete va
	 * debajo porque dentro de un grupo las secciones se separan entre sí.
	 */
	import type { Snippet } from 'svelte';
	import RailSection from '$lib/ui/rail-section.svelte';

	interface Props {
		title: string;
		/** Qué hace la sección, en una o dos frases; va bajo el título. */
		description?: string;
		id?: string;
		/** Contenido extra del carril, bajo la descripción. */
		aside?: Snippet;
		children: Snippet;
	}

	let { title, description = '', id = undefined, aside, children }: Props = $props();
</script>

<RailSection {title} {description} {id} {aside} level={3} divider="bottom">
	{@render children()}
</RailSection>
