<script lang="ts">
	/*
	 * Barra superior del panel.
	 *
	 * Ahora dice en qué sección está el usuario: antes solo repetía la marca —que
	 * ya estaba en el menú— y tres botones, así que ocupaba 72 px sin aportar un
	 * dato. El nombre sale de la misma lista que pinta el menú (`nav.ts`), de
	 * modo que los dos no pueden discrepar.
	 *
	 * Se fueron dos cosas que engañaban: el botón de tres puntos «Configuración»,
	 * que no tenía manejador y no hacía nada al pulsarlo, y el punto rojo de
	 * notificaciones, que estaba encendido siempre —hubiera o no algo que leer—
	 * porque nadie le pasaba un contador.
	 */
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { page } from '$app/state';
	import Icon from './icon.svelte';
	import { sectionTitle } from '../nav';

	interface Props {
		sidebarOpen?: boolean;
		data: { user: { name: string; email: string; image?: string } };
	}

	let { sidebarOpen = $bindable(false), data }: Props = $props();

	const section = $derived(sectionTitle(page.url.pathname));
	const initial = $derived(data.user.name.trim().charAt(0).toUpperCase());
</script>

<header class="header">
	<button
		class="menu"
		onclick={() => (sidebarOpen = !sidebarOpen)}
		aria-label="Abrir el menú de secciones"
		aria-expanded={sidebarOpen}
		aria-controls="dashboard-sidebar"
	>
		<Icon name="menu" size={20} width={2} />
	</button>

	<p class="section">{section}</p>

	<div class="actions">
		<button
			class="action"
			onclick={() => privacy.toggle()}
			aria-pressed={privacy.hidden}
			title={privacy.hidden ? 'Mostrar los importes' : 'Ocultar los importes'}
		>
			<Icon name={privacy.hidden ? 'eye-off' : 'eye'} size={18} />
			<span class="sr-only">{privacy.hidden ? 'Mostrar los importes' : 'Ocultar los importes'}</span
			>
		</button>

		<a class="action" href={resolve('/dashboard/notifications')} aria-label="Notificaciones">
			<Icon name="bell" size={18} />
		</a>

		<a class="user" href={resolve('/dashboard/settings')}>
			{#if data.user.image && data.user.image !== 'avatar.png'}
				<img src={data.user.image} alt="" class="avatar" />
			{:else}
				<span class="avatar" aria-hidden="true">{initial}</span>
			{/if}
			<span class="who">
				<span class="name">{data.user.name}</span>
				<span class="mail">{data.user.email}</span>
			</span>
		</a>
	</div>
</header>

<style>
	.header {
		position: fixed;
		top: 0;
		right: 0;
		left: var(--rail);
		z-index: 20;
		display: flex;
		align-items: center;
		gap: 1rem;
		height: var(--header-h);
		padding: 0 1.75rem;
		background: color-mix(in srgb, var(--bg) 88%, transparent);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		border-bottom: 1px solid var(--border);
	}

	.menu {
		display: none;
		padding: 0.4rem;
		margin-left: -0.4rem;
		border: none;
		background: none;
		color: var(--text-muted);
		cursor: pointer;
	}

	.section {
		margin: 0;
		font-size: 0.95rem;
		font-weight: 500;
		color: var(--text);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		margin-left: auto;
	}

	.action {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border: none;
		border-radius: 8px;
		background: none;
		color: var(--text-muted);
		cursor: pointer;
		text-decoration: none;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.action:hover {
		background: var(--panel);
		color: var(--text);
	}

	.user {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		margin-left: 0.5rem;
		padding: 0.3rem 0.5rem 0.3rem 0.3rem;
		border-radius: 999px;
		text-decoration: none;
		transition: background 0.15s ease;
	}

	.user:hover {
		background: var(--panel);
	}

	.avatar {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		flex-shrink: 0;
		border-radius: 50%;
		background: var(--panel-2);
		object-fit: cover;
		color: var(--text);
		font-size: 0.8rem;
		font-weight: 600;
	}

	.who {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.name {
		font-size: 0.8rem;
		color: var(--text);
	}

	.mail {
		font-size: 0.7rem;
		color: var(--text-dim);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	@media (max-width: 1024px) {
		.header {
			left: 0;
			padding: 0 1rem;
		}

		.menu {
			display: flex;
		}

		.who {
			display: none;
		}
	}
</style>
