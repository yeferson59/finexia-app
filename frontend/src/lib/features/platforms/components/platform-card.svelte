<script lang="ts">
	import { privacy } from '$lib/stores/privacy.svelte';

	interface PlatformCardData {
		id: string;
		name: string;
		sourceType: string;
		isActive: boolean;
		investments: number;
		totalValue: string;
	}

	let { platform, onView }: { platform: PlatformCardData; onView: (id: string) => void } = $props();

	function fmtMoney(value: string): string {
		return privacy.money(
			'$' +
				new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(parseFloat(value) || 0)
		);
	}

	function getStatusColor(status: boolean) {
		return status === true ? 'var(--green)' : 'var(--red)';
	}
</script>

<div class="platform-card">
	<div class="card-header">
		<div class="card-title-section">
			<h3 class="platform-name">{platform.name}</h3>
			<span class="platform-type">{platform.sourceType}</span>
		</div>
		<div class="status-badge" style="--status-color: {getStatusColor(platform.isActive)}">
			{platform.isActive ? 'Activo' : 'Inactivo'}
		</div>
	</div>

	<div class="card-stats">
		<div class="stat-item">
			<span class="stat-label">Posiciones</span>
			<span class="stat-value">{platform.investments}</span>
		</div>
		<div class="stat-item">
			<span class="stat-label">Invertido</span>
			<span class="stat-value">{fmtMoney(platform.totalValue)}</span>
		</div>
	</div>

	<div class="card-actions">
		<button
			onclick={() => onView(platform.id)}
			class="action-btn"
			aria-label={`Ver detalles de ${platform.name}`}
		>
			<svg
				width="16"
				height="16"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
				<circle cx="12" cy="12" r="3" />
			</svg>
			Ver detalles
		</button>
	</div>
</div>

<style>
	.platform-card {
		border: 1px solid var(--border-strong);
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
		padding: 1.5rem;
		transition: all 0.3s ease;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.platform-card:hover {
		border-color: rgba(212, 145, 42, 0.3);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 10px 30px var(--border);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
	}

	.card-title-section {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.platform-name {
		margin: 0;
		color: var(--text);
		font-size: 1.1rem;
		font-weight: 700;
	}

	.platform-type {
		color: rgba(212, 145, 42, 0.75);
		font-size: 0.8rem;
		text-transform: uppercase;
		letter-spacing: 0.4px;
		font-weight: 600;
	}

	.status-badge {
		padding: 0.4rem 0.8rem;
		border-radius: 6px;
		background: var(--status-color, var(--amber));
		color: #0d0800;
		font-size: 0.75rem;
		font-weight: 700;
		white-space: nowrap;
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.card-stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		background: var(--border);
		border-radius: 8px;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.stat-label {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		font-weight: 600;
	}

	.stat-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--amber);
		font-size: 1.1rem;
		font-weight: 700;
		font-family: var(--font-mono);
	}

	.card-actions {
		display: flex;
		gap: 0.5rem;
		margin-top: auto;
	}

	.action-btn {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
		padding: 0.65rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		background: transparent;
		color: var(--text);
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		white-space: nowrap;
	}

	.action-btn:hover {
		border-color: rgba(212, 145, 42, 0.5);
		color: var(--amber);
		background: var(--border);
	}

	@media (max-width: 768px) {
		.action-btn {
			font-size: 0.75rem;
			padding: 0.5rem;
		}

		.action-btn svg {
			width: 14px;
			height: 14px;
		}
	}
</style>
