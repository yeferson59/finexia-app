<script lang="ts">
	/*
	 * Seguridad. Cada garantía corresponde a algo que la aplicación hace hoy
	 * —2FA, cierre de sesiones, ocultar importes, exportación en XLSX—, no a una
	 * promesa a futuro: la portada no debe prometer lo que el panel no tiene.
	 *
	 * La primera, no pedir credenciales, es la que sostiene a las demás, así que
	 * deja de ser una tarjeta más de seis y pasa a ser el titular de la sección.
	 */
	const guarantees = [
		{
			title: 'Verificación en dos pasos',
			body: 'Activa el segundo factor con tu app de autenticación y guarda los códigos de recuperación por si pierdes el teléfono.'
		},
		{
			title: 'Tus sesiones, bajo control',
			body: 'Consulta desde qué dispositivos hay una sesión abierta y ciérralas una a una o todas a la vez.'
		},
		{
			/* El nombre que lleva el botón del panel. */
			title: 'Ocultar los importes',
			body: 'Un clic y todas las cifras quedan tapadas. Puedes abrir el panel en un café o en una reunión sin enseñarlas.'
		},
		{
			title: 'Tus datos salen contigo',
			body: 'Descarga tu resumen, tus movimientos y tu informe de riesgo en XLSX cuando quieras, sin pedir permiso.'
		},
		{
			title: 'Tú eres la única fuente',
			body: 'Todo lo que aparece en el panel lo has escrito tú. Nada se añade, se estima ni se comparte por detrás.'
		}
	];
</script>

<section class="lp-section lp-forest" id="seguridad" aria-labelledby="seguridad-title">
	<div class="lp-wrap grid">
		<div class="head">
			<h2 class="lp-h2" id="seguridad-title">No te pedimos la contraseña de ninguna cuenta.</h2>
			<p class="lp-lead">
				Finexia no se conecta a tu broker, exchange ni banco. No guardamos claves de terceros, así
				que no hay ninguna que se pueda filtrar. Sobre eso está construido lo demás.
			</p>
		</div>

		<dl class="list">
			{#each guarantees as item (item.title)}
				<div class="item">
					<dt>{item.title}</dt>
					<dd>{item.body}</dd>
				</div>
			{/each}
		</dl>
	</div>
</section>

<style>
	.grid {
		display: grid;
		grid-template-columns: minmax(0, 5fr) minmax(0, 6fr);
		gap: 48px 96px;
		align-items: start;
	}

	.head {
		position: sticky;
		top: 112px;
	}

	.head .lp-h2 {
		font-size: clamp(34px, 4.6vw, 60px);
	}

	.head .lp-lead {
		margin-top: 28px;
	}

	.list {
		margin: 0;
		border-top: 1px solid var(--lp-forest-rule);
	}

	.item {
		display: grid;
		grid-template-columns: minmax(0, 2fr) minmax(0, 3fr);
		gap: 8px 32px;
		padding: 24px 0;
		border-bottom: 1px solid var(--lp-forest-rule);
	}

	dt {
		font-size: var(--lp-fs-lead);
		font-stretch: 108%;
		font-weight: 600;
		line-height: 1.3;
	}

	dd {
		margin: 0;
		color: var(--lp-forest-ink-2);
		text-wrap: pretty;
	}

	@media (max-width: 1000px) {
		.grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.head {
			position: static;
		}
	}

	@media (max-width: 640px) {
		.item {
			grid-template-columns: minmax(0, 1fr);
			padding: 20px 0;
		}
	}
</style>
