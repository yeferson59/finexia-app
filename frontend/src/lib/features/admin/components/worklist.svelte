<script lang="ts">
	/**
	 * Lo que hay abierto en el escritorio, y solo eso.
	 *
	 * Cada fila lleva a la pantalla donde se arregla, así que la portada de
	 * administración no necesita además una rejilla de atajos: los tres únicos
	 * sitios a los que se va desde aquí son estos, y aparecen cuando tienen algo
	 * que ver. Si no hay nada pendiente no se pinta nada; lo dice el titular.
	 */
	import { resolve } from '$app/paths';
	import type { AdminTask, AdminTaskKey } from '../desk';

	interface Props {
		tasks: AdminTask[];
	}

	let { tasks }: Props = $props();

	const HREFS: Record<AdminTaskKey, string> = {
		waitlist: resolve('/dashboard/admin/users'),
		invitations: resolve('/dashboard/admin/users'),
		prices: resolve('/dashboard/admin/assets'),
		rates: resolve('/dashboard/admin/exchange-rates')
	};
</script>

<ul class="worklist">
	{#each tasks as task (task.key)}
		<li>
			<!-- Las rutas ya están resueltas arriba; la regla no puede seguir el
			     rastro hasta la tabla. -->
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a class="task" href={HREFS[task.key]}>
				<span class="count">{task.count}</span>
				<span class="text">
					<span class="title">{task.title}</span>
					<span class="detail">{task.detail}</span>
				</span>
			</a>
		</li>
	{/each}
</ul>

<style>
	.worklist {
		margin: 0 0 2.5rem;
		padding: 0;
		list-style: none;
		border-top: 1px solid var(--border-strong);
	}

	.task {
		display: grid;
		grid-template-columns: 3.5rem minmax(0, 1fr);
		align-items: baseline;
		gap: 1.5rem;
		padding: 1.15rem 0.5rem 1.15rem 0;
		border-bottom: 1px solid var(--border);
		text-decoration: none;
		transition: background 0.15s ease;
	}

	.task:hover {
		background: rgba(255, 255, 255, 0.02);
	}

	/*
	 * La cifra en el ámbar apagado del área: es la medida de algo que lleva
	 * esperando, no un dato del sistema. Tabular para que las columnas de dos y
	 * de tres dígitos empiecen en el mismo sitio.
	 */
	.count {
		font-family: var(--font-mono);
		font-size: 1.5rem;
		font-weight: 400;
		font-variant-numeric: tabular-nums;
		text-align: right;
		color: var(--stale);
	}

	.text {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		min-width: 0;
	}

	.title {
		font-size: 1rem;
		font-weight: 500;
		color: var(--text);
	}

	.task:hover .title {
		text-decoration: underline;
		text-underline-offset: 4px;
	}

	.detail {
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.task {
			transition: none;
		}
	}

	@media (max-width: 560px) {
		.task {
			grid-template-columns: 2.75rem minmax(0, 1fr);
			gap: 1rem;
		}

		.count {
			font-size: 1.25rem;
		}
	}
</style>
