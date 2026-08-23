<script lang="ts">
	/*
	 * Tres pasos en horizontal, unidos por un hilo.
	 *
	 * Antes era otra pila vertical justo después de la de beneficios, y las dos
	 * contaban casi lo mismo con la misma forma. Puestos en fila se lee lo que
	 * son: una secuencia, frente al catálogo de la sección anterior.
	 */
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

<section class="band" id="como-funciona">
	<div class="wrap block">
		<div class="sec-head reveal">
			<div class="eyebrow">Cómo funciona</div>
			<h2 class="sec-title">De plataformas dispersas<br />a un solo mapa</h2>
		</div>

		<div class="steps">
			<div class="thread" aria-hidden="true"></div>
			{#each steps as step (step.num)}
				<div class="step reveal">
					<span class="step-num">{step.num}</span>
					<h3>{step.title}</h3>
					<p>{step.body}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<style>
	.steps {
		position: relative;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 48px;
	}

	.thread {
		position: absolute;
		top: 15px;
		left: 12%;
		right: 12%;
		height: 1px;
		background: linear-gradient(
			90deg,
			transparent,
			rgba(212, 145, 42, 0.35),
			rgba(212, 145, 42, 0.35),
			transparent
		);
	}

	.step {
		position: relative;
	}

	.step-num {
		display: grid;
		place-items: center;
		width: 30px;
		height: 30px;
		border-radius: 999px;
		border: 1px solid rgba(212, 145, 42, 0.28);
		/* Opaco a propósito: tapa el hilo que pasa por detrás. */
		background: #0c0d0e;
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--amber);
	}

	.step h3 {
		margin-top: 22px;
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 21px;
		letter-spacing: -0.01em;
	}

	.step p {
		margin-top: 10px;
		font-size: 14.5px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
		text-wrap: pretty;
	}

	@media (max-width: 860px) {
		.steps {
			grid-template-columns: minmax(0, 1fr);
			gap: 36px;
			max-width: 660px;
			margin: 0 auto;
		}
		/* En columna el hilo pasa a vertical, a la altura de los círculos. */
		.thread {
			top: 30px;
			bottom: 30px;
			left: 15px;
			right: auto;
			width: 1px;
			height: auto;
			background: linear-gradient(
				180deg,
				rgba(212, 145, 42, 0.35),
				rgba(212, 145, 42, 0.35),
				transparent
			);
		}
	}
</style>
