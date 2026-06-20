<script lang="ts">
	const benefits = [
		{
			num: '01',
			title: 'Registro manual, total control',
			body: 'Anota tus brokers, exchanges, bancos o plataformas y los activos que tienes en cada uno. Tú decides qué incluir y cómo nombrarlo.'
		},
		{
			num: '02',
			title: 'Portafolios a tu medida',
			body: 'Agrupa activos en los portafolios que tienes en mente, aunque vivan en plataformas distintas. Tú trazas la estructura.'
		},
		{
			num: '03',
			title: 'Cero acceso a tus cuentas',
			body: 'Finexia nunca se conecta a tus plataformas ni pide credenciales. La información es tuya y solo tú la editas.'
		},
		{
			num: '04',
			title: 'Visión de un vistazo',
			body: 'Distribución y peso de cada portafolio claros al instante, sin hojas de cálculo ni cuentas manuales.'
		}
	];
</script>

<section class="block wrap" id="beneficios">
	<div class="sec-head reveal">
		<div class="eyebrow">Por qué Finexia</div>
		<h2 class="sec-title">Tu patrimonio organizado<br />a tu manera</h2>
	</div>
	<div class="benefits-list">
		{#each benefits as benefit (benefit.num)}
			<div class="bitem reveal">
				<div class="bnum">{benefit.num}</div>
				<div class="bcontent">
					<h3>{benefit.title}</h3>
					<p>{benefit.body}</p>
				</div>
			</div>
		{/each}
	</div>
</section>

<style>
	.block {
		padding: 88px 0;
	}
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
	.bnum {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--text-dim);
		letter-spacing: 0.04em;
		flex-shrink: 0;
		padding-top: 5px;
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
		.bitem {
			border-right: none;
		}
		.bitem:last-child {
			border-bottom: none;
		}
		.bitem:nth-child(n) {
			border-bottom: 1px solid var(--border);
		}
	}
</style>
