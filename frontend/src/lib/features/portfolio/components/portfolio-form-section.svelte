<script lang="ts">
	/**
	 * Un bloque de un formulario de portafolio: el carril de `ui/rail-section` en
	 * modo `fields`.
	 *
	 * Sustituyó a las tarjetas apiladas que tenían estos formularios, cada una
	 * con su sombra y su leyenda EN VERSALITAS ÁMBAR. Es el mismo carril que la
	 * página de configuración: siete controles seguidos son una lista larga, y en
	 * tres bloques rotulados se sabe en cuál está lo que se busca.
	 *
	 * El modo `fields` es lo único propio, y lo comparten los seis sitios que lo
	 * usan: apila los campos y les da el aspecto de `ui/input`, que es su vecino
	 * en pantalla.
	 */
	import type { Snippet } from 'svelte';
	import RailSection from '$lib/ui/rail-section.svelte';

	interface Props {
		title: string;
		/** Qué se decide en el bloque, en una o dos frases. */
		description: string;
		children: Snippet;
	}

	let { title, description, children }: Props = $props();
</script>

<RailSection {title} {description} fields>
	{@render children()}
</RailSection>
