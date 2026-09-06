<script lang="ts">
	/**
	 * Un bloque de la página: el carril de `ui/rail-section` con la variante que
	 * usa esta página.
	 *
	 * Sustituyó a la tarjeta que envolvía cada canal. La página eran dos tarjetas
	 * iguales una al lado de la otra, así que el canal que todavía no existe
	 * ocupaba la mitad de la pantalla para decir que no existe. En un carril cada
	 * bloque ocupa lo que tiene que contar.
	 *
	 * Es el mismo carril que configuración, y a propósito: son vecinas en el menú
	 * y se hojean igual. Lo único propio es que los controles llegan a 38rem: los
	 * interruptores llevan su explicación al lado, no debajo.
	 */
	import type { Snippet } from 'svelte';
	import RailSection from '$lib/ui/rail-section.svelte';

	interface Props {
		title: string;
		/** Qué es el canal, en una frase; va bajo el título. */
		description: string;
		children: Snippet;
	}

	let { title, description, children }: Props = $props();
</script>

<RailSection {title} {description} contentMax="38rem">
	{@render children()}
</RailSection>
