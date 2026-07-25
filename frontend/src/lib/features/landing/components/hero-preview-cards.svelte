<script lang="ts">
	// Tarjetas decorativas del hero (aria-hidden): maqueta estática de un
	// portafolio. Sin datos reales ni estado.
</script>

<div class="hero-right" aria-hidden="true">
	<div class="hero-card reveal">
		<div class="hc-top">
			<div>
				<div class="hc-eyebrow">Portafolio</div>
				<div class="hc-name">Principal</div>
			</div>
			<div class="hc-status">Activo</div>
		</div>
		<div class="hc-value">$248.500</div>
		<div class="hc-delta">+12,4% este año</div>
		<div class="hc-bars">
			<div class="hc-row">
				<span class="hc-asset">Acciones</span>
				<div class="hc-bar">
					<div class="hc-fill" style="width:38%;--c:var(--amber)"></div>
				</div>
				<span class="hc-pct">38%</span>
			</div>
			<div class="hc-row">
				<span class="hc-asset">Crypto</span>
				<div class="hc-bar">
					<div class="hc-fill" style="width:29%;--c:var(--green)"></div>
				</div>
				<span class="hc-pct">29%</span>
			</div>
			<div class="hc-row">
				<span class="hc-asset">ETFs</span>
				<div class="hc-bar">
					<div class="hc-fill" style="width:23%;--c:#6b8cef"></div>
				</div>
				<span class="hc-pct">23%</span>
			</div>
			<div class="hc-row">
				<span class="hc-asset">Otros</span>
				<div class="hc-bar">
					<div class="hc-fill" style="width:10%;--c:#484542"></div>
				</div>
				<span class="hc-pct">10%</span>
			</div>
		</div>
		<div class="hc-platforms">
			<span>Binance</span><span>Degiro</span><span>Revolut</span><span>+2</span>
		</div>
	</div>

	<div class="hero-card-sm reveal">
		<div class="hcs-label">Portafolio Cripto</div>
		<div class="hcs-row">
			<span class="hcs-sym">BTC</span>
			<span class="hcs-val">$14.200</span>
			<span class="hcs-up">+4,2%</span>
		</div>
		<div class="hcs-row">
			<span class="hcs-sym">ETH</span>
			<span class="hcs-val">$3.480</span>
			<span class="hcs-dn">−1,1%</span>
		</div>
		<div class="hcs-row">
			<span class="hcs-sym">SOL</span>
			<span class="hcs-val">$1.820</span>
			<span class="hcs-up">+2,7%</span>
		</div>
	</div>
</div>

<style>
	.hero-right {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.hero-card {
		width: 100%;
		padding: 26px 24px 20px;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.038);
		border: 1px solid var(--border-strong);
		backdrop-filter: blur(10px);
	}

	.hc-top {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 20px;
	}

	.hc-eyebrow {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 4px;
	}

	.hc-name {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 20px;
		letter-spacing: -0.01em;
	}

	.hc-status {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--green);
		background: rgba(34, 201, 126, 0.09);
		border: 1px solid rgba(34, 201, 126, 0.2);
		padding: 4px 9px;
		border-radius: 4px;
	}

	.hc-value {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 30px;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
	}

	.hc-delta {
		font-size: 13px;
		color: var(--green);
		margin-top: 4px;
		margin-bottom: 20px;
		font-weight: 400;
	}

	.hc-bars {
		display: flex;
		flex-direction: column;
		gap: 11px;
	}

	.hc-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.hc-asset {
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-muted);
		width: 60px;
		flex-shrink: 0;
	}

	.hc-bar {
		flex: 1;
		height: 3px;
		background: rgba(255, 255, 255, 0.07);
		border-radius: 2px;
		overflow: hidden;
	}

	.hc-fill {
		height: 100%;
		border-radius: 2px;
		background: var(--c);
	}

	.hc-pct {
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-dim);
		width: 30px;
		text-align: right;
		flex-shrink: 0;
	}

	.hc-platforms {
		display: flex;
		gap: 6px;
		margin-top: 18px;
		flex-wrap: wrap;
	}

	.hc-platforms span {
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--text-dim);
		background: var(--surface);
		border: 1px solid var(--border);
		padding: 3px 9px;
		border-radius: 4px;
	}

	.hero-card-sm {
		width: 100%;
		padding: 14px 18px;
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid var(--border);
	}

	.hcs-label {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 10px;
	}

	.hcs-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 0;
		border-top: 1px solid var(--border);
		font-family: var(--font-mono);
		font-size: 12.5px;
	}

	.hcs-sym {
		color: var(--text-muted);
		width: 36px;
		flex-shrink: 0;
	}

	.hcs-val {
		color: var(--text);
		font-weight: 600;
		flex: 1;
		font-variant-numeric: tabular-nums;
	}

	.hcs-up {
		color: var(--green);
	}

	.hcs-dn {
		color: #e05a5a;
	}

	@media (max-width: 940px) {
		.hero-right {
			flex-direction: row;
			align-items: flex-start;
		}
		.hero-card {
			flex: 1;
		}
		.hero-card-sm {
			width: 200px;
			flex-shrink: 0;
		}
	}

	@media (max-width: 640px) {
		.hero-right {
			flex-direction: column;
		}
		.hero-card-sm {
			width: 100%;
		}
	}
</style>
