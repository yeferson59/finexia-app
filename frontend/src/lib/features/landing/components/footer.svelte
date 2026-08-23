<script lang="ts">
	/*
	 * Antes era una sola fila flex con la marca, seis enlaces de sección, tres
	 * legales y el copyright: al partirse quedaban en cualquier orden y sin
	 * jerarquía. En columnas cada bloque se lee por lo que es, y "Iniciar
	 * sesión" tiene un sitio fijo también al final de la página.
	 */
	import { resolve } from '$app/paths';
	import Brand from './brand.svelte';

	const currentYear = new Date().getFullYear();
	const home = resolve('/');
</script>

<footer>
	<div class="wrap">
		<div class="foot-cols">
			<div class="foot-brand">
				<Brand iconSize={26} nameSize={18} />
				<p>Tu patrimonio en un solo mapa. Registro manual, sin conectar cuentas.</p>
			</div>

			<nav class="foot-col" aria-label="Producto">
				<div class="foot-label">Producto</div>
				<a href="{home}#producto">El producto</a>
				<a href="{home}#beneficios">Beneficios</a>
				<a href="{home}#como-funciona">Cómo funciona</a>
				<a href="{home}#seguridad">Seguridad</a>
			</nav>

			<nav class="foot-col" aria-label="Legal">
				<div class="foot-label">Legal</div>
				<a href={resolve('/privacidad')}>Política de Privacidad</a>
				<a href={resolve('/terminos')}>Términos y Condiciones</a>
				<a href={resolve('/cookies')}>Cookies</a>
			</nav>

			<nav class="foot-col" aria-label="Cuenta">
				<div class="foot-label">Cuenta</div>
				<a href={resolve('/auth')}>Iniciar sesión</a>
				<a href="{home}#waitlist">Lista de espera</a>
				<a href="{home}#faq">Preguntas</a>
			</nav>
		</div>

		<div class="foot-bottom">
			<div class="foot-copy">© {currentYear} Finexia. Todos los derechos reservados.</div>
			<div class="foot-launch">Lanzamiento · 1 oct 2026</div>
		</div>
	</div>
</footer>

<style>
	footer {
		border-top: 1px solid var(--border);
		margin-top: 40px;
		padding-top: 56px;
	}

	.foot-cols {
		display: grid;
		grid-template-columns: 1.6fr 1fr 1fr 1fr;
		gap: 48px;
		align-items: start;
	}

	.foot-brand p {
		margin-top: 16px;
		max-width: 34ch;
		font-size: 13.5px;
		font-weight: 300;
		line-height: 1.65;
		color: var(--text-dim);
		text-wrap: pretty;
	}

	.foot-col {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.foot-label {
		padding-bottom: 2px;
		font-family: var(--font-mono);
		font-size: 9.5px;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.foot-col a {
		font-size: 13.5px;
		color: var(--text-muted);
		transition: color 0.2s;
	}

	.foot-col a:hover {
		color: var(--text);
	}

	.foot-bottom {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		margin-top: 48px;
		padding: 22px 0 40px;
		border-top: 1px solid var(--border);
	}

	.foot-copy {
		font-size: 13px;
		color: var(--text-dim);
	}

	.foot-launch {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	@media (max-width: 860px) {
		.foot-cols {
			grid-template-columns: repeat(2, minmax(0, 1fr));
			gap: 40px 32px;
		}
		.foot-brand {
			grid-column: 1 / -1;
		}
	}

	@media (max-width: 480px) {
		footer {
			padding-top: 44px;
		}
		.foot-bottom {
			flex-direction: column;
			align-items: flex-start;
			gap: 12px;
		}
	}
</style>
