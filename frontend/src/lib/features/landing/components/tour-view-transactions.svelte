<script lang="ts">
	/* Maqueta de `/dashboard/transactions`: historial + asistente de importación. */
	import { TOUR_TRANSACTIONS } from '../product-tour';

	const steps = ['Archivo', 'Columnas', 'Vista previa', 'Confirmar'];

	function toneOf(type: string): string {
		if (type === 'Compra') return 'buy';
		if (type === 'Venta') return 'sell';
		return 'income';
	}
</script>

<div class="transactions">
	<div class="mk-card import">
		<div class="import-top">
			<div>
				<div class="mk-eyebrow">Importación</div>
				<div class="mk-title">Sube un CSV de tu plataforma</div>
			</div>
			<span class="mk-pill">movimientos.csv · 128 filas</span>
		</div>
		<ol class="steps">
			{#each steps as step, i (step)}
				<li class:done={i < 2} class:on={i === 2}>
					<span class="dot">{i + 1}</span>
					{step}
				</li>
			{/each}
		</ol>
	</div>

	<div class="mk-card">
		<div class="hist-top">
			<div>
				<div class="mk-eyebrow">Movimientos</div>
				<div class="mk-title">Historial</div>
			</div>
			<span class="mk-pill">Compra · Venta · Dividendo · +5</span>
		</div>

		<div class="mk-table">
			<div class="mk-row mk-head">
				<span>Tipo</span>
				<span>Activo</span>
				<span>Plataforma</span>
				<span>Fecha</span>
				<span class="mk-num">Importe</span>
			</div>
			{#each TOUR_TRANSACTIONS as t (t.type + t.asset + t.date)}
				<div class="mk-row">
					<span class="type {toneOf(t.type)}">{t.type}</span>
					<span class="sym">{t.asset}</span>
					<span>{t.platform}</span>
					<span class="date">{t.date}</span>
					<span class="mk-num" class:mk-up={t.amount.startsWith('+')}>{t.amount}</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.transactions {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.import-top,
	.hist-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		padding-bottom: 12px;
		border-bottom: 1px solid var(--border);
	}

	.hist-top {
		margin-bottom: 4px;
	}

	.steps {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 18px;
		margin: 14px 0 0;
		padding: 0;
		list-style: none;
	}

	.steps li {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 11px;
		color: var(--text-dim);
	}

	.steps .dot {
		display: grid;
		place-items: center;
		width: 17px;
		height: 17px;
		border-radius: 50%;
		border: 1px solid var(--border-strong);
		font-family: var(--font-mono);
		font-size: 9px;
	}

	.steps li.done {
		color: var(--text-muted);
	}

	.steps li.done .dot {
		border-color: rgba(34, 201, 126, 0.35);
		background: rgba(34, 201, 126, 0.12);
		color: var(--green);
	}

	.steps li.on {
		color: var(--amber-light);
	}

	.steps li.on .dot {
		border-color: rgba(212, 145, 42, 0.4);
		background: rgba(212, 145, 42, 0.14);
		color: var(--amber-light);
	}

	.mk-table :global(.mk-row) {
		grid-template-columns:
			minmax(0, 0.9fr) minmax(0, 0.7fr) minmax(0, 1fr) minmax(0, 0.6fr)
			minmax(0, 1fr);
	}

	.type {
		justify-self: start;
		padding: 2px 7px;
		border: 1px solid var(--border);
		border-radius: 4px;
		font-size: 9.5px;
		font-weight: 600;
		white-space: nowrap;
	}

	.type.buy {
		color: var(--amber-light);
		border-color: rgba(212, 145, 42, 0.25);
		background: rgba(212, 145, 42, 0.1);
	}

	.type.sell {
		color: #6b8cef;
		border-color: rgba(107, 140, 239, 0.25);
		background: rgba(107, 140, 239, 0.1);
	}

	.type.income {
		color: var(--green);
		border-color: rgba(34, 201, 126, 0.25);
		background: rgba(34, 201, 126, 0.1);
	}

	.sym {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text);
	}

	.date {
		font-family: var(--font-mono);
		font-size: 10px;
		color: var(--text-dim);
	}

	@media (max-width: 700px) {
		.mk-table :global(.mk-row) {
			grid-template-columns: minmax(0, 0.9fr) minmax(0, 0.6fr) minmax(0, 1fr);
		}

		.mk-table :global(.mk-row > :nth-child(3)),
		.mk-table :global(.mk-row > :nth-child(4)) {
			display: none;
		}
	}
</style>
