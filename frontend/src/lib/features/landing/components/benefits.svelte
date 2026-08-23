<script lang="ts">
	/*
	 * La numeración 01–04 que llevaba cada tarjeta solo repetía lo que ya decía
	 * el icono y le robaba ancho al texto, así que desaparece. La cabecera se
	 * va a un raíl lateral: la sección anterior abre con un eje horizontal y
	 * esta con uno vertical, de modo que dos cabeceras seguidas nunca son
	 * iguales.
	 */
	const benefits = [
		{
			icon: 'pencil',
			title: 'Registro manual, total control',
			body: 'Anota tus brokers, exchanges, bancos o plataformas y los activos que tienes en cada uno. Tú decides qué incluir y cómo nombrarlo.'
		},
		{
			icon: 'layers',
			title: 'Portafolios a tu medida',
			body: 'Agrupa activos en los portafolios que tienes en mente, aunque vivan en plataformas distintas. Tú trazas la estructura.'
		},
		{
			icon: 'chart',
			title: 'Ganancia real, no solo saldo',
			body: 'Cada portafolio compara su valor de mercado con el capital que invertiste, así que ves lo que has ganado, no solo cuánto tienes.'
		},
		{
			icon: 'pie',
			title: 'Visión de un vistazo',
			body: 'Distribución y peso de cada portafolio claros al instante, sin hojas de cálculo ni cuentas manuales.'
		}
	];
</script>

<section class="wrap block" id="beneficios">
	<div class="sec-rail">
		<div class="rail-head reveal">
			<div class="eyebrow">Por qué Finexia</div>
			<h2 class="sec-title sec-title-sm">Tu patrimonio organizado a tu manera</h2>
			<p class="rail-desc">
				Cuatro cosas que cambian cuando dejas de repartir tu patrimonio entre pestañas y hojas de
				cálculo.
			</p>
		</div>

		<div class="benefits-list">
			{#each benefits as benefit (benefit.title)}
				<div class="bitem reveal">
					<span class="bicon" aria-hidden="true">
						<svg
							width="17"
							height="17"
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
					<h3>{benefit.title}</h3>
					<p>{benefit.body}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<style>
	.rail-desc {
		margin-top: 18px;
		font-size: 15px;
		color: var(--text-muted);
		font-weight: 300;
		line-height: 1.7;
		text-wrap: pretty;
	}

	/* Filete interior de 1px pintado con el `gap` sobre el color del borde,
	   igual que la rejilla de seguridad. */
	.benefits-list {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1px;
		background: var(--border);
		border: 1px solid var(--border);
		border-radius: 12px;
		overflow: hidden;
	}

	.bitem {
		padding: 32px 30px;
		background: var(--bg);
		transition: background 0.2s;
	}

	.bitem:hover {
		background: rgba(255, 255, 255, 0.022);
	}

	.bicon {
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 9px;
		border: 1px solid rgba(212, 145, 42, 0.2);
		background: rgba(212, 145, 42, 0.07);
		color: var(--amber-light);
	}

	.bitem h3 {
		margin-top: 18px;
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 19px;
		letter-spacing: -0.01em;
	}

	.bitem p {
		margin-top: 10px;
		font-size: 14px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
		text-wrap: pretty;
	}

	@media (max-width: 700px) {
		.benefits-list {
			grid-template-columns: minmax(0, 1fr);
		}
		.bitem {
			padding: 28px 22px;
		}
	}
</style>
