<script lang="ts">
	/*
	 * El pie, en columnas: cada bloque se lee por lo que es, e "Iniciar sesión"
	 * tiene un sitio fijo también al final de la página. Va en el verde del
	 * cierre, que en la portada queda justo encima; en las páginas legales es el
	 * único bloque verde, y cierra igual.
	 */
	import { resolve } from '$app/paths';
	import Brand from '$lib/ui/brand.svelte';

	const currentYear = new Date().getFullYear();
	const home = resolve('/');
</script>

<footer class="lp-forest">
	<div class="lp-wrap">
		<div class="cols">
			<div class="brand-col">
				<Brand iconSize={26} nameSize={20} />
				<p>Tu patrimonio en un solo mapa. Registro manual, sin conectar cuentas.</p>
			</div>

			<nav class="col" aria-label="Producto">
				<h2>Producto</h2>
				<a href="{home}#producto">El producto</a>
				<a href="{home}#como-funciona">Cómo funciona</a>
				<a href="{home}#seguridad">Seguridad</a>
				<a href="{home}#faq">Preguntas</a>
				<a href={resolve('/blog')}>Blog</a>
				<a href={resolve('/apoyar')}>Apoyar el proyecto</a>
			</nav>

			<nav class="col" aria-label="Legal">
				<h2>Legal</h2>
				<a href={resolve('/privacidad')}>Política de privacidad</a>
				<a href={resolve('/terminos')}>Términos y condiciones</a>
				<a href={resolve('/cookies')}>Cookies</a>
			</nav>

			<nav class="col" aria-label="Cuenta">
				<h2>Cuenta</h2>
				<a href={resolve('/auth')}>Iniciar sesión</a>
				<a href="{home}#waitlist">Unirme a la lista</a>
			</nav>
		</div>

		<p class="copy">© {currentYear} Finexia. Todos los derechos reservados.</p>
	</div>
</footer>

<style>
	footer {
		padding-top: 56px;
	}

	.cols {
		display: grid;
		grid-template-columns: minmax(0, 1.6fr) repeat(3, minmax(0, 1fr));
		gap: 40px;
		padding-top: 40px;
		border-top: 1px solid var(--lp-forest-rule);
	}

	.brand-col p {
		max-width: 34ch;
		margin: 16px 0 0;
		font-size: var(--lp-fs-sm);
		color: var(--lp-forest-ink-2);
	}

	.col {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 10px;
	}

	.col h2 {
		margin: 0 0 4px;
		font-size: var(--lp-fs-sm);
		font-weight: 600;
		color: var(--lp-forest-ink);
	}

	.col a {
		font-size: var(--lp-fs-sm);
		color: var(--lp-forest-ink-2);
	}

	.col a:hover {
		color: var(--lp-forest-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.copy {
		margin: 56px 0 0;
		padding-block: 20px 32px;
		border-top: 1px solid var(--lp-forest-rule);
		font-size: 13px;
		color: var(--lp-forest-ink-2);
	}

	@media (max-width: 860px) {
		.cols {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.brand-col {
			grid-column: 1 / -1;
		}
	}
</style>
