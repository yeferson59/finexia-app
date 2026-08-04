<script lang="ts">
	/* Maqueta de `/dashboard/portfolios`: lista de portafolios + posiciones. */
	import { TOUR_PORTFOLIOS, TOUR_HOLDINGS } from '../product-tour';
</script>

<div class="portfolios">
	<div class="cards">
		{#each TOUR_PORTFOLIOS as p (p.name)}
			<div class="mk-card pcard">
				<div class="pcard-top">
					<span class="mk-title">{p.name}</span>
					<span class="mk-pill">{p.kind}</span>
				</div>
				<div class="mk-stat-value">{p.value}</div>
				<div class="pcard-delta" class:mk-up={p.up} class:mk-dn={!p.up}>{p.delta}</div>
			</div>
		{/each}
	</div>

	<div class="mk-card">
		<div class="holdings-top">
			<div>
				<div class="mk-eyebrow">Jubilación</div>
				<div class="mk-title">Posiciones</div>
			</div>
			<span class="mk-pill">4 activos</span>
		</div>

		<div class="mk-table">
			<div class="mk-row mk-head">
				<span>Activo</span>
				<span>Peso</span>
				<span class="mk-num">Valor</span>
				<span class="mk-num">Rend.</span>
			</div>
			{#each TOUR_HOLDINGS as h (h.symbol)}
				<div class="mk-row">
					<span class="asset">
						<b>{h.symbol}</b>
						<em>{h.name}</em>
					</span>
					<span class="weight">
						<span class="mk-bar"><span style="width:{h.weight}%"></span></span>
						<i>{h.weight}%</i>
					</span>
					<span class="mk-num">{h.value}</span>
					<span class="mk-num mk-up">{h.delta}</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.portfolios {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.cards {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 10px;
	}

	.pcard-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		margin-bottom: 12px;
	}

	.pcard .mk-title {
		margin-top: 0;
		font-size: 13px;
	}

	.pcard-delta {
		margin-top: 5px;
		font-size: 11px;
	}

	.holdings-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		padding-bottom: 12px;
		margin-bottom: 4px;
		border-bottom: 1px solid var(--border);
	}

	.mk-table :global(.mk-row) {
		grid-template-columns: minmax(0, 1.6fr) minmax(0, 1.3fr) minmax(0, 1fr) minmax(0, 0.7fr);
	}

	.asset {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.asset b {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--text);
	}

	.asset em {
		font-style: normal;
		font-size: 10px;
		color: var(--text-dim);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.weight {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.weight :global(.mk-bar) {
		flex: 1;
	}

	.weight i {
		font-style: normal;
		font-family: var(--font-mono);
		font-size: 9.5px;
		color: var(--text-dim);
		width: 26px;
		text-align: right;
	}

	@media (max-width: 700px) {
		.cards {
			grid-template-columns: minmax(0, 1fr);
		}

		.mk-table :global(.mk-row > :nth-child(2)) {
			display: none;
		}

		.mk-table :global(.mk-row) {
			grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr) minmax(0, 0.7fr);
		}
	}
</style>
