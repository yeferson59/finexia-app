<script lang="ts">
	const steps = [
		{
			num: '01',
			title: 'Registra tus plataformas',
			body: 'Añade manualmente cada broker, exchange o banco donde tienes activos y anota qué tienes en cada uno. Sin conectar cuentas.'
		},
		{
			num: '02',
			title: 'Crea tus portafolios',
			body: 'Agrupa los activos según los portafolios que tienes en mente, combinando lo que está en distintas plataformas.'
		},
		{
			num: '03',
			title: 'Visualiza el conjunto',
			body: 'Ve la distribución y el peso de cada portafolio en una sola vista, y decide con tu patrimonio completo a la vista.'
		}
	];
</script>

<section class="block wrap" id="como-funciona">
	<div class="sec-head reveal">
		<div class="eyebrow">Cómo funciona</div>
		<h2 class="sec-title">De plataformas dispersas<br />a un solo mapa</h2>
	</div>
	<div class="steps">
		{#each steps as step (step.num)}
			<div class="step reveal">
				<div class="step-num">{step.num}</div>
				<div class="step-body">
					<h3>{step.title}</h3>
					<p>{step.body}</p>
				</div>
			</div>
		{/each}
	</div>
</section>

<style>
	.steps {
		display: flex;
		flex-direction: column;
		max-width: 660px;
		margin: 0 auto;
	}
	.step {
		display: flex;
		gap: 32px;
		padding: 32px 0;
		border-bottom: 1px solid var(--border);
	}
	.step:last-child {
		border-bottom: none;
	}
	.step-num {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--amber);
		letter-spacing: 0.06em;
		flex-shrink: 0;
		padding-top: 5px;
		min-width: 26px;
	}
	.step-body h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 21px;
		letter-spacing: -0.01em;
		margin-bottom: 10px;
	}
	.step-body p {
		font-size: 14.5px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
	}
</style>
