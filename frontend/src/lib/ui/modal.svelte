<script lang="ts">
	/**
	 * Diálogo modal del dashboard.
	 *
	 * Sobre el `<dialog>` nativo y `showModal()`, que es lo que trae de serie lo
	 * que las tres copias anteriores —`div` fijos con un backdrop propio— no
	 * tenían: el foco atrapado dentro mientras está abierto, Escape para cerrar,
	 * el fondo inerte para el lector de pantalla y el foco devuelto a quien lo
	 * abrió al cerrarse. El backdrop era además un `div role="button"
	 * tabindex="0"`, o sea un botón anunciado que sólo servía de fondo.
	 *
	 * El estado vive en el padre (`open` + `onClose`): quien abre el modal es
	 * quien sabe cuándo se ha terminado con él.
	 *
	 * Tres zonas: la cabecera, el cuerpo que desplaza y el pie con los botones.
	 * El pie no lo pinta el modal sino el contenido —el botón que envía tiene que
	 * vivir dentro de su `<form>`—, así que el modal aporta la clase
	 * `modal-actions` y la fija al borde de abajo. Sin eso, en un formulario largo
	 * o en un móvil «Guardar» quedaba por debajo del borde, y cada cuerpo traía su
	 * propia fila de botones con su propio tamaño.
	 */
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/shared/css';

	type Size = 'sm' | 'md' | 'lg';
	type Tone = 'default' | 'danger';

	interface Props {
		open: boolean;
		/**
		 * Cerrado a la vista pero con el contenido montado. Un envío optimista
		 * cierra el diálogo al pulsar y, si el servidor lo rechaza, lo vuelve a
		 * abrir con lo que se había escrito: desmontarlo lo habría vaciado.
		 */
		hidden?: boolean;
		/** Encabezado del diálogo; nombra también el `<dialog>` para el lector. */
		title: string;
		/** Línea de apoyo bajo el título; la lee también el lector al abrir. */
		description?: string;
		/** Ancho máximo: `sm` para confirmar, `lg` para formularios de varias columnas. */
		size?: Size;
		/**
		 * `danger` cuando lo que confirma el diálogo no se puede deshacer: el
		 * filete de arriba sale en rojo en vez de ámbar, y avisa antes de leer.
		 */
		tone?: Tone;
		/**
		 * Se llama al cerrar por cualquier vía: Escape, la X, el fondo o el
		 * propio contenido. El padre baja `open` desde aquí.
		 */
		onClose: () => void;
		class?: string;
		children: Snippet;
	}

	let {
		open,
		hidden = false,
		title,
		description = '',
		size = 'md',
		tone = 'default',
		onClose,
		class: className = '',
		children
	}: Props = $props();

	let dialog = $state<HTMLDialogElement | null>(null);

	// Único por instancia: dos modales montados a la vez compartirían el `id`
	// del título y `aria-labelledby` apuntaría al del otro.
	const titleId = $props.id();
	const descriptionId = `${titleId}-description`;

	// `showModal()` es imperativo y vive fuera del modelo de Svelte, así que
	// sincronizarlo con `open` es justo para lo que está `$effect`.
	const shown = $derived(open && !hidden);

	$effect(() => {
		if (!dialog) return;
		if (shown && !dialog.open) dialog.showModal();
		else if (!shown && dialog.open) dialog.close();
	});

	// El `<dialog>` nativo no bloquea el scroll de la página que queda detrás.
	$effect(() => {
		if (!shown) return;
		const previous = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		return () => {
			document.body.style.overflow = previous;
		};
	});

	/**
	 * Un clic en el `::backdrop` tiene como destino el propio `<dialog>`: si el
	 * destino es cualquier otra cosa, el clic fue dentro del contenido.
	 */
	function onDialogClick(event: MouseEvent) {
		if (event.target === dialog) onClose();
	}
</script>

<dialog
	bind:this={dialog}
	class={cn('modal', `modal-${size}`, className)}
	data-tone={tone}
	aria-labelledby={titleId}
	aria-describedby={description ? descriptionId : undefined}
	onclick={onDialogClick}
	onclose={() => shown && onClose()}
>
	<!--
		El contenido se monta sólo mientras está abierto. Un `<dialog>` cerrado
		sigue en el DOM, así que dejarlo dentro montaba los ocho formularios de la
		pantalla en cada carga, sus efectos corrían de fondo y sus textos ocultos
		competían con los de la página en cualquier búsqueda. Además así el
		formulario vuelve en blanco cada vez que se abre.
	-->
	{#if open}
		<header class="modal-header">
			<div class="modal-heading">
				<h2 class="modal-title" id={titleId}>{title}</h2>
				{#if description}
					<p class="modal-description" id={descriptionId}>{description}</p>
				{/if}
			</div>
			<button type="button" class="modal-close" onclick={onClose} aria-label="Cerrar">
				<svg
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
					stroke-linecap="round"
					aria-hidden="true"
				>
					<path d="M18 6L6 18M6 6l12 12" />
				</svg>
			</button>
		</header>

		<div class="modal-body">
			{@render children()}
		</div>
	{/if}
</dialog>

<style>
	.modal {
		--modal-pad: 1.75rem;
		--modal-surface: #0e0f11;
		--modal-accent: var(--amber);

		/* El preflight de Tailwind pone `margin: 0` a todo, incluido `dialog`, y
		   con eso se lleva por delante el `margin: auto` con el que el navegador
		   centra un diálogo modal: sin esta línea salía pegado arriba a la
		   izquierda. */
		margin: auto;
		width: min(var(--modal-width), calc(100vw - 2rem));
		max-height: min(88vh, calc(100dvh - 2rem));
		padding: 0;
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		background: var(--modal-surface);
		color: var(--text);
		overflow: hidden;
		box-shadow: 0 32px 80px -16px rgba(0, 0, 0, 0.75);
		/* La cabecera escucha el desplazamiento del cuerpo, que es su hermano y
		   no su padre: el nombre de la línea de tiempo tiene que subir hasta aquí. */
		timeline-scope: --modal-scroll;
	}

	.modal[data-tone='danger'] {
		--modal-accent: var(--red);
	}

	/* Sólo abierto: un `display` suelto pisa el `display: none` que el navegador
	   le da a un `<dialog>` cerrado, y el formulario se quedaba pintado en medio
	   de la página con el modal «cerrado». */
	.modal[open] {
		display: flex;
		flex-direction: column;
	}

	/*
	 * El filete del borde de arriba. Es el mismo trazo con el que el panel marca
	 * sus avisos —prosa con una línea del color de lo que dice—, puesto en
	 * horizontal: ámbar para lo que se guarda, rojo para lo que se pierde.
	 */
	.modal::before {
		content: '';
		position: absolute;
		inset: 0 0 auto;
		height: 2px;
		background: var(--modal-accent);
		transform-origin: left center;
		pointer-events: none;
	}

	.modal-sm {
		--modal-width: 26rem;
	}

	.modal-md {
		--modal-width: 34rem;
	}

	.modal-lg {
		--modal-width: 46rem;
	}

	.modal::backdrop {
		background: rgba(4, 5, 6, 0.8);
	}

	/* Un solo movimiento: el diálogo llega y el filete se traza sobre él. */
	@media (prefers-reduced-motion: no-preference) {
		.modal[open] {
			animation: modal-in 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
		}

		.modal[open]::before {
			animation: rule-draw 0.5s cubic-bezier(0.65, 0, 0.35, 1) 0.08s both;
		}

		.modal[open]::backdrop {
			animation: backdrop-in 0.2s ease-out;
		}
	}

	@keyframes modal-in {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}

	@keyframes rule-draw {
		from {
			transform: scaleX(0);
		}
	}

	@keyframes backdrop-in {
		from {
			opacity: 0;
		}
	}

	.modal-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1.25rem;
		padding: var(--modal-pad) var(--modal-pad) 1.25rem;
		border-bottom: 1px solid transparent;
	}

	/*
	 * Sin filete bajo la cabecera mientras el contenido está arriba del todo: el
	 * diálogo se lee como una sola hoja. En cuanto el cuerpo desplaza, el filete
	 * aparece para marcar dónde se corta lo que queda debajo. Donde el navegador
	 * no sabe atar una animación al desplazamiento, simplemente no aparece.
	 */
	@supports (animation-timeline: scroll()) {
		.modal-header {
			animation: header-rule linear both;
			animation-timeline: --modal-scroll;
			animation-range: 0 1.5rem;
		}
	}

	@keyframes header-rule {
		to {
			border-bottom-color: var(--border);
		}
	}

	.modal-heading {
		min-width: 0;
	}

	/* La voz de los títulos de página —«AAPL», «Mis activos»— a escala de diálogo. */
	.modal-title {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.6rem;
		font-weight: 300;
		line-height: 1.15;
		letter-spacing: -0.02em;
		color: var(--text);
		overflow-wrap: anywhere;
		text-wrap: balance;
	}

	.modal-description {
		max-width: 56ch;
		margin: 0.5rem 0 0;
		font-size: 0.87rem;
		line-height: 1.55;
		color: var(--text-muted);
		text-wrap: pretty;
	}

	/* 36px de diana aunque el trazo mida 18: la esquina es fácil de fallar. */
	.modal-close {
		flex-shrink: 0;
		display: grid;
		place-items: center;
		width: 2.25rem;
		height: 2.25rem;
		margin: -0.35rem -0.6rem 0 0;
		padding: 0;
		border: none;
		border-radius: 8px;
		background: none;
		color: var(--text-dim);
		cursor: pointer;
		transition:
			color 0.2s ease,
			background-color 0.2s ease;
	}

	.modal-close:hover {
		background: var(--surface-2);
		color: var(--text);
	}

	.modal-body {
		padding: 0 var(--modal-pad) var(--modal-pad);
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-width: thin;
		scrollbar-color: var(--border-strong) transparent;
		scroll-timeline: --modal-scroll block;
	}

	/*
	 * El pie cierra el cuerpo cuando la fila de botones es lo último que hay:
	 * sin el relleno de abajo, la fila fija toca el borde igual desplazando que
	 * en reposo. La fila puede venir anidada —el panel de venta la mete dentro de
	 * su `<form>`, que va dentro de su propio contenedor—; lo que cuenta es que
	 * cierre el último bloque. Si debajo queda algo más —el resultado de un
	 * import—, el cuerpo conserva su margen.
	 */
	.modal-body:global(:has(> .modal-actions:last-child)),
	.modal-body:global(:has(> :last-child .modal-actions:last-child)) {
		padding-bottom: 0;
	}

	.modal-body :global(.modal-actions) {
		position: sticky;
		bottom: 0;
		z-index: 1;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: 0.75rem;
		margin: 0.75rem calc(-1 * var(--modal-pad)) 0;
		padding: 1rem var(--modal-pad) 1.25rem;
		border-top: 1px solid var(--border);
		background: var(--modal-surface);
	}

	/*
	 * Los botones del pie a una sola medida. `ui/button` en tamaño normal es el
	 * de la portada —más alto, con halo y un salto al pasar por encima—, y aquí
	 * convivía con botones escritos a mano de la mitad de alto.
	 */
	.modal-body :global(.modal-actions button) {
		padding: 0.7rem 1.3rem;
		font-size: 0.9rem;
		letter-spacing: 0.01em;
		/* Si no caben, que baje el botón entero a otra línea y no media etiqueta. */
		white-space: nowrap;
		box-shadow: none;
	}

	.modal-body :global(.modal-actions button:hover:not(:disabled)) {
		transform: none;
		box-shadow: none;
	}

	/* En el móvil el diálogo es una hoja que sube desde abajo, donde está el pulgar. */
	@media (max-width: 640px) {
		.modal {
			--modal-pad: 1.25rem;
			width: 100%;
			max-width: 100%;
			max-height: calc(100dvh - 1rem);
			margin: auto 0 0;
			border-width: 1px 0 0;
			border-radius: 18px 18px 0 0;
		}

		.modal-title {
			font-size: 1.4rem;
		}

		.modal-body :global(.modal-actions) {
			padding-bottom: max(1.25rem, env(safe-area-inset-bottom));
		}

		/* Cada botón parte de lo que mide su etiqueta y el sobrante se reparte:
		   a partes iguales, «Registrar transacción» se partía en dos líneas. */
		.modal-body :global(.modal-actions > *) {
			flex: 1 1 auto;
		}
	}

	@media (max-width: 640px) and (prefers-reduced-motion: no-preference) {
		.modal[open] {
			animation: sheet-in 0.28s cubic-bezier(0.2, 0.8, 0.2, 1);
		}
	}

	@keyframes sheet-in {
		from {
			transform: translateY(100%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.modal-close {
			transition: none;
		}
	}
</style>
