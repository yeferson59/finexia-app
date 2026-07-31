<script lang="ts">
	/** Apariencia: hoy solo informa de que el tema es único. */
	import SettingsSection from './settings-section.svelte';
</script>

<SettingsSection title="Apariencia">
	<div class="appearance-info">
		<div class="theme-badge">
			<span class="theme-dot"></span>
			<span>Tema oscuro FINEXIA</span>
		</div>
		<p class="hint">
			La aplicación utiliza exclusivamente el tema premium oscuro con acentos dorados. No hay temas
			adicionales disponibles actualmente.
		</p>
	</div>
</SettingsSection>

<style>
	.appearance-info {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.theme-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.875rem;
		border-radius: 20px;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.25);
		font-size: 0.875rem;
		color: var(--amber);
		font-weight: 500;
		width: fit-content;
	}

	.theme-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--amber);
		flex-shrink: 0;
	}
</style>
