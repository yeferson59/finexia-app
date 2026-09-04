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
	 */
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/shared/css';

	type Size = 'sm' | 'md' | 'lg';

	interface Props {
		open: boolean;
		/** Encabezado del diálogo; nombra también el `<dialog>` para el lector. */
		title: string;
		/** Línea de apoyo bajo el título. */
		description?: string;
		/** Ancho máximo: `sm` para confirmar, `lg` para formularios de varias columnas. */
		size?: Size;
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
		title,
		description = '',
		size = 'md',
		onClose,
		class: className = '',
		children
	}: Props = $props();

	let dialog = $state<HTMLDialogElement | null>(null);

	// Único por instancia: dos modales montados a la vez compartirían el `id`
	// del título y `aria-labelledby` apuntaría al del otro.
	const titleId = $props.id();

	// `showModal()` es imperativo y vive fuera del modelo de Svelte, así que
	// sincronizarlo con `open` es justo para lo que está `$effect`.
	$effect(() => {
		if (!dialog) return;
		if (open && !dialog.open) dialog.showModal();
		else if (!open && dialog.open) dialog.close();
	});

	// El `<dialog>` nativo no bloquea el scroll de la página que queda detrás.
	$effect(() => {
		if (!open) return;
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
	aria-labelledby={titleId}
	onclick={onDialogClick}
	onclose={() => open && onClose()}
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
			<div>
				<h2 class="modal-title" id={titleId}>{title}</h2>
				{#if description}
					<p class="modal-description">{description}</p>
				{/if}
			</div>
			<button type="button" class="modal-close" onclick={onClose} aria-label="Cerrar">
				<svg
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.75"
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
		/* El preflight de Tailwind pone `margin: 0` a todo, incluido `dialog`, y
		   con eso se lleva por delante el `margin: auto` con el que el navegador
		   centra un diálogo modal: sin esta línea salía pegado arriba a la
		   izquierda. */
		margin: auto;
		width: min(var(--modal-width), calc(100vw - 2rem));
		max-height: min(85vh, calc(100dvh - 2rem));
		padding: 0;
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		background: #101114;
		color: var(--text);
		overflow: hidden;
		box-shadow: 0 24px 64px rgba(0, 0, 0, 0.55);
	}

	/* Sólo abierto: un `display` suelto pisa el `display: none` que el navegador
	   le da a un `<dialog>` cerrado, y el formulario se quedaba pintado en medio
	   de la página con el modal «cerrado». */
	.modal[open] {
		display: flex;
		flex-direction: column;
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
		background: rgba(6, 7, 9, 0.72);
	}

	@media (prefers-reduced-motion: no-preference) {
		.modal[open] {
			animation: modal-in 0.16s ease-out;
		}
	}

	@keyframes modal-in {
		from {
			opacity: 0;
			transform: translateY(6px);
		}
	}

	.modal-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1.5rem;
		padding: 1.25rem 1.35rem 1rem;
		border-bottom: 1px solid var(--border);
	}

	.modal-title {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	.modal-description {
		margin: 0.35rem 0 0;
		max-width: 52ch;
		font-size: 0.85rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.modal-close {
		flex-shrink: 0;
		display: flex;
		padding: 0.3rem;
		border: none;
		border-radius: 6px;
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

	.modal-close:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.modal-body {
		padding: 1.35rem;
		overflow-y: auto;
	}
</style>
