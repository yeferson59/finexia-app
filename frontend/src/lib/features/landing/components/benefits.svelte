<script lang="ts">
	const benefits = [
		{
			num: '01',
			icon: 'pencil',
			title: 'Registro manual, total control',
			body: 'Anota tus brokers, exchanges, bancos o plataformas y los activos que tienes en cada uno. Tú decides qué incluir y cómo nombrarlo.'
		},
		{
			num: '02',
			icon: 'layers',
			title: 'Portafolios a tu medida',
			body: 'Agrupa activos en los portafolios que tienes en mente, aunque vivan en plataformas distintas. Tú trazas la estructura.'
		},
		{
			num: '03',
			icon: 'chart',
			title: 'Ganancia real, no solo saldo',
			body: 'Cada portafolio compara su valor de mercado con el capital que invertiste, así que ves lo que has ganado, no solo cuánto tienes.'
		},
		{
			num: '04',
			icon: 'pie',
			title: 'Visión de un vistazo',
			body: 'Distribución y peso de cada portafolio claros al instante, sin hojas de cálculo ni cuentas manuales.'
		}
	];
</script>

<section class="wrap block" id="beneficios">
	<div class="sec-head reveal">
		<div class="eyebrow">Por qué Finexia</div>
		<h2 class="sec-title">Tu patrimonio organizado<br />a tu manera</h2>
	</div>
	<div class="benefits-list">
		{#each benefits as benefit (benefit.num)}
			<div class="bitem reveal">
				<div class="bside">
					<span class="bicon" aria-hidden="true">
						<svg
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.7"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							{#if benefit.icon === 'pencil'}
								<path d="M12 20h9" />
								<path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" />
							{:else if benefit.icon === 'layers'}
								<path d="m12 2 8 4-8 4-8-4 8-4z" />
								<path d="m4 12 8 4 8-4" />
								<path d="m4 17 8 4 8-4" />
							{:else if benefit.icon === 'chart'}
								<path d="M3 3v18h18" />
								<path d="m7 14 4-4 3 3 5-6" />
							{:else}
								<path d="M12 3a9 9 0 1 0 9 9h-9V3z" />
								<path d="M15 3a6 6 0 0 1 6 6h-6V3z" />
							{/if}
						</svg>
					</span>
					<div class="bnum">{benefit.num}</div>
				</div>
				<div class="bcontent">
					<h3>{benefit.title}</h3>
					<p>{benefit.body}</p>
				</div>
			</div>
		{/each}
	</div>
</section>

<style>
	.benefits-list {
		display: grid;
		grid-template-columns: 1fr 1fr;
		border: 1px solid var(--border);
		border-radius: 12px;
		overflow: hidden;
	}
	.bitem {
		display: flex;
		gap: 22px;
		padding: 36px 30px;
		border-right: 1px solid var(--border);
		border-bottom: 1px solid var(--border);
		transition: background 0.2s;
	}
	.bitem:hover {
		background: var(--surface);
	}
	.bitem:nth-child(2n) {
		border-right: none;
	}
	.bitem:nth-last-child(-n + 2) {
		border-bottom: none;
	}
	.bside {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		flex-shrink: 0;
	}
	.bicon {
		display: grid;
		place-items: center;
		width: 32px;
		height: 32px;
		border-radius: 9px;
		border: 1px solid rgba(212, 145, 42, 0.2);
		background: rgba(212, 145, 42, 0.07);
		color: var(--amber-light);
	}
	.bnum {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--text-dim);
		letter-spacing: 0.04em;
	}
	.bcontent h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 20px;
		letter-spacing: -0.01em;
		margin-bottom: 10px;
	}
	.bcontent p {
		font-size: 14px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
	}
	@media (max-width: 700px) {
		.benefits-list {
			grid-template-columns: 1fr;
		}
		/* En una sola columna cada tarjeta vuelve a llevar su separador; el orden
		   importa: `:last-child` va después para poder quitárselo a la última. */
		.bitem,
		.bitem:nth-last-child(-n + 2) {
			border-right: none;
			border-bottom: 1px solid var(--border);
			padding: 28px 22px;
		}
		.bitem:last-child {
			border-bottom: none;
		}
	}
</style>
