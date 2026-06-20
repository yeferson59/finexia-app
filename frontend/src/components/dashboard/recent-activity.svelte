<script lang="ts">
	const activities = [
		{
			id: 1,
			type: 'purchase',
			title: 'Compra de Acciones',
			description: 'Apple Inc. (AAPL)',
			amount: 5000,
			date: '2024-06-04',
			time: '14:30',
			icon: 'buy'
		},
		{
			id: 2,
			type: 'dividend',
			title: 'Dividendo Recibido',
			description: 'Microsoft Corp (MSFT)',
			amount: 250,
			date: '2024-06-02',
			time: '09:15',
			icon: 'dividend'
		},
		{
			id: 3,
			type: 'deposit',
			title: 'Depósito de Fondos',
			description: 'Transferencia bancaria',
			amount: 10000,
			date: '2024-05-31',
			time: '11:45',
			icon: 'deposit'
		},
		{
			id: 4,
			type: 'sale',
			title: 'Venta de Fondos',
			description: 'Fondo Mutuo XYZ',
			amount: 3500,
			date: '2024-05-28',
			time: '16:20',
			icon: 'sell'
		},
		{
			id: 5,
			type: 'transfer',
			title: 'Transferencia Realizada',
			description: 'A cuenta de ahorros',
			amount: 2000,
			date: '2024-05-25',
			time: '10:00',
			icon: 'transfer'
		}
	];

	function formatDate(dateString: string): string {
		const date = new Date(`${dateString}T00:00:00`);
		const now = new Date();
		const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
		const startOfTarget = new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
		const dayDiff = Math.round((startOfTarget - startOfToday) / 86400000);

		if (dayDiff === 0) return 'Hoy';
		if (dayDiff === -1) return 'Ayer';
		return date.toLocaleDateString('es-CO', { month: 'short', day: 'numeric' });
	}
</script>

<div class="activity-card">
	<div class="card-header">
		<div>
			<p class="card-eyebrow">Movimientos</p>
			<h2 class="card-title">Actividad Reciente</h2>
		</div>
		<a href="/dashboard/transactions" class="view-all">Ver todo →</a>
	</div>

	<div class="activity-list">
		{#each activities as activity (activity.id)}
			<article class={`activity-item activity-${activity.type}`}>
				<div class={`activity-icon ${activity.type}`}>
					{#if activity.icon === 'buy'}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
							<path
								d="M17 3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V7l-4-4zm-5 16c-1.66 0-3-1.34-3-3s1.34-3 3-3 3 1.34 3 3-1.34 3-3 3zm3-10H5V5h10v4z"
							></path>
						</svg>
					{:else if activity.icon === 'dividend'}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
							<path
								d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"
							></path>
						</svg>
					{:else if activity.icon === 'deposit'}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
							<path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"></path>
						</svg>
					{:else if activity.icon === 'sell'}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
							<path
								d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z"
							></path>
						</svg>
					{:else}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
							<path
								d="M16 8v-2c0-.55-.45-1-1-1H9c-.55 0-1 .45-1 1v2H6v2h2v7c0 .55.45 1 1 1h6c.55 0 1-.45 1-1v-7h2V8h-2zm-3 8h-4v-7h4v7z"
							></path>
						</svg>
					{/if}
				</div>

				<div class="activity-info">
					<p class="activity-title">{activity.title}</p>
					<p class="activity-description">{activity.description}</p>
				</div>

				<div class="activity-details">
					<p
						class="activity-amount"
						class:positive={['deposit', 'dividend'].includes(activity.type)}
					>
						{['deposit', 'dividend'].includes(activity.type) ? '+' : '-'}${new Intl.NumberFormat(
							'es-CO'
						).format(activity.amount)}
					</p>
					<time class="activity-date" datetime={`${activity.date}T${activity.time}`}
						>{formatDate(activity.date)}</time
					>
				</div>
			</article>
		{/each}
	</div>

	<div class="card-footer">
		<a href="/dashboard/reports" class="footer-link">Descargar extracto →</a>
	</div>
</div>

<style>
	.activity-card {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1.5rem;
		padding-bottom: 1.25rem;
		border-bottom: 1px solid var(--border);
		gap: 1rem;
	}

	.card-eyebrow {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin: 0 0 0.4rem 0;
	}

	.card-title {
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
		margin: 0;
	}

	.view-all {
		font-size: 0.8rem;
		color: var(--amber);
		text-decoration: none;
		font-weight: 500;
		white-space: nowrap;
		transition: color 0.2s ease;
	}

	.view-all:hover {
		color: var(--amber-light);
	}

	.activity-list {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-bottom: 1.5rem;
		overflow-y: auto;
		max-height: 400px;
	}

	.activity-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border-radius: 8px;
		transition: background 0.25s ease;
	}

	.activity-item:hover {
		background: var(--surface);
	}

	.activity-icon {
		width: 38px;
		height: 38px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		border: 1px solid var(--border);
		color: var(--text-muted);
	}

	.activity-icon.purchase {
		background: rgba(212, 145, 42, 0.12);
		border-color: rgba(212, 145, 42, 0.25);
		color: var(--amber-light);
	}

	.activity-icon.deposit,
	.activity-icon.dividend {
		background: rgba(34, 201, 126, 0.12);
		border-color: rgba(34, 201, 126, 0.25);
		color: var(--green);
	}

	.activity-icon.sale,
	.activity-icon.transfer {
		background: rgba(107, 140, 239, 0.12);
		border-color: rgba(107, 140, 239, 0.25);
		color: #6b8cef;
	}

	.activity-info {
		flex: 1;
		min-width: 0;
	}

	.activity-title {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
		margin: 0 0 0.2rem 0;
	}

	.activity-description {
		font-size: 0.775rem;
		color: var(--text-muted);
		margin: 0;
	}

	.activity-details {
		text-align: right;
	}

	.activity-amount {
		font-family: var(--font-mono);
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text);
		margin: 0 0 0.2rem 0;
		font-variant-numeric: tabular-nums;
	}

	.activity-amount.positive {
		color: var(--green);
	}

	.activity-date {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		color: var(--text-dim);
		margin: 0;
	}

	.card-footer {
		border-top: 1px solid var(--border);
		padding-top: 1.25rem;
	}

	.footer-link {
		font-size: 0.8rem;
		color: var(--amber);
		text-decoration: none;
		font-weight: 500;
		transition: color 0.2s ease;
		display: inline-block;
	}

	.footer-link:hover {
		color: var(--amber-light);
	}

	/* Scrollbar styling */
	.activity-list::-webkit-scrollbar {
		width: 4px;
	}

	.activity-list::-webkit-scrollbar-track {
		background: transparent;
	}

	.activity-list::-webkit-scrollbar-thumb {
		background: var(--border-strong);
		border-radius: 2px;
	}

	.activity-list::-webkit-scrollbar-thumb:hover {
		background: rgba(212, 145, 42, 0.4);
	}

	@media (max-width: 768px) {
		.activity-card {
			padding: 1.5rem;
		}

		.activity-list {
			max-height: 300px;
		}

		.activity-icon {
			width: 36px;
			height: 36px;
		}

		.activity-title {
			font-size: 0.85rem;
		}
	}
</style>
