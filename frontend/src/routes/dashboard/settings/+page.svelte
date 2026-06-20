<script lang="ts">
	let emailAlerts = $state(true);
	let weeklySummary = $state(true);
	let darkGoldTheme = $state(true);
</script>

<svelte:head>
	<title>Configuración - FINEXIA</title>
	<meta name="description" content="Preferencias de cuenta y notificaciones de FINEXIA" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Configuración</h1>
	<p class="page-subtitle">Ajusta preferencias de experiencia, alertas y seguridad de cuenta.</p>
</header>

<section class="panel settings-grid">
	<article class="setting-block">
		<h2>Notificaciones</h2>
		<label class="toggle-row">
			<span>Alertas por correo</span>
			<input type="checkbox" bind:checked={emailAlerts} />
		</label>
		<label class="toggle-row">
			<span>Resumen semanal</span>
			<input type="checkbox" bind:checked={weeklySummary} />
		</label>
	</article>

	<article class="setting-block">
		<h2>Apariencia</h2>
		<label class="toggle-row">
			<span>Tema oscuro FINEXIA</span>
			<input type="checkbox" bind:checked={darkGoldTheme} />
		</label>
		<p class="hint">El tema actual mantiene el estilo premium oscuro con acentos dorados.</p>
	</article>

	<article class="setting-block">
		<h2>Seguridad</h2>
		<button class="action">Cambiar contraseña</button>
		<button class="action secondary">Gestionar dispositivos</button>
	</article>
</section>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 300;
		color: var(--text);
		font-family: var(--font-display);
		letter-spacing: -0.02em;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.62);
	}

	.panel {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.settings-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		padding: 1.2rem;
	}

	.setting-block {
		padding: 1rem;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
	}

	.setting-block h2 {
		margin: 0 0 1rem;
		font-size: 1rem;
		color: var(--text);
	}

	.toggle-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.85rem;
		font-size: 0.86rem;
		color: rgba(236, 234, 229, 0.7);
	}

	input[type='checkbox'] {
		width: 18px;
		height: 18px;
		accent-color: var(--amber);
	}

	.hint {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.action {
		width: 100%;
		margin-bottom: 0.55rem;
		border: none;
		border-radius: 8px;
		padding: 0.75rem;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		cursor: pointer;
	}

	.action.secondary {
		background: rgba(212, 145, 42, 0.12);
		color: var(--text);
		border: 1px solid rgba(212, 145, 42, 0.25);
	}

	@media (max-width: 1024px) {
		.settings-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}
	}
</style>
