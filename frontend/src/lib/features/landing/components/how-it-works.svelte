<script lang="ts">
	/*
	 * Tres pasos. Aquí la numeración sí dice algo: es una secuencia, y el orden
	 * es el que sigue quien empieza en el panel.
	 *
	 * La duda que hay que resolver es el trabajo de apuntarlo todo a mano, así
	 * que el segundo paso nombra la importación del extracto.
	 */
	const steps = [
		{
			title: 'Anota tus plataformas',
			body: 'Da de alta cada broker, exchange o banco donde tienes dinero. Basta con el nombre: no hay nada que conectar.'
		},
		{
			title: 'Registra lo que tienes',
			body: 'Añade tus posiciones y movimientos uno a uno, o sube el extracto de tu plataforma en CSV o Excel y revisa la vista previa antes de confirmar.'
		},
		{
			title: 'Agrúpalo en portafolios',
			body: 'Crea los portafolios que tengas en mente, con su meta y su perfil de riesgo, y mira cada uno contra el capital que pusiste.'
		}
	];
</script>

<section class="lp-section steps-section" id="como-funciona" aria-labelledby="como-funciona-title">
	<div class="lp-wrap">
		<div class="head">
			<h2 class="lp-h2" id="como-funciona-title">Cómo funciona</h2>
			<p class="lp-lead">
				Todo lo escribes tú. Por eso puedes agruparlo como quieras, y por eso nadie más necesita tus
				claves.
			</p>
		</div>

		<ol class="steps">
			{#each steps as step, i (step.title)}
				<li class="step">
					<span class="num" aria-hidden="true">{i + 1}</span>
					<h3>{step.title}</h3>
					<p>{step.body}</p>
				</li>
			{/each}
		</ol>
	</div>
</section>

<style>
	.steps-section {
		background: var(--lp-paper-2);
	}

	.head {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
		gap: 20px 64px;
		align-items: end;
		margin-bottom: 64px;
	}

	.steps {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 40px;
		margin: 0;
		padding: 0;
		list-style: none;
		counter-reset: none;
	}

	.step {
		padding-top: 20px;
		border-top: 2px solid var(--lp-ink);
	}

	.num {
		display: block;
		font-size: 72px;
		font-stretch: 125%;
		font-weight: 300;
		line-height: 0.9;
		letter-spacing: -0.04em;
		font-variant-numeric: lining-nums;
	}

	.step h3 {
		margin: 28px 0 0;
		font-size: var(--lp-fs-h3);
		font-stretch: 110%;
		font-weight: 620;
		line-height: 1.2;
		letter-spacing: -0.01em;
	}

	.step p {
		margin: 10px 0 0;
		max-width: 38ch;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	@media (max-width: 900px) {
		.head {
			grid-template-columns: minmax(0, 1fr);
			margin-bottom: 44px;
		}
		.steps {
			grid-template-columns: minmax(0, 1fr);
			gap: 36px;
		}
		.step {
			display: grid;
			grid-template-columns: 64px minmax(0, 1fr);
			column-gap: 16px;
		}
		.num {
			grid-row: 1 / span 2;
			font-size: 56px;
		}
		.step h3 {
			margin-top: 4px;
		}
	}
</style>
