<script lang="ts">
	import { resolve } from '$app/paths';
	import Brand from './brand.svelte';

	const currentYear = new Date().getFullYear();
	const home = resolve('/');
</script>

<footer>
	<div class="wrap foot">
		<Brand iconSize={26} nameSize={18} />
		<nav class="foot-links">
			<a href="{home}#beneficios">Beneficios</a>
			<a href="{home}#como-funciona">Cómo funciona</a>
			<a href="{home}#faq">Preguntas</a>
			<a href="{home}#waitlist">Lista de espera</a>
		</nav>
		<nav class="foot-legal">
			<a href={resolve('/privacidad')}>Política de Privacidad</a>
			<a href={resolve('/terminos')}>Términos y Condiciones</a>
			<a href={resolve('/cookies')}>Cookies</a>
		</nav>
		<div class="foot-copy">© {currentYear} Finexia. Todos los derechos reservados.</div>
	</div>
</footer>

<style>
	footer {
		border-top: 1px solid var(--border);
		padding: 40px 0 52px;
		margin-top: 40px;
	}
	.foot {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		flex-wrap: wrap;
	}
	.foot-links {
		display: flex;
		gap: 24px;
	}
	.foot-links a,
	.foot-legal a {
		font-size: 13.5px;
		color: var(--text-muted);
		transition: color 0.2s;
	}
	.foot-links a:hover,
	.foot-legal a:hover {
		color: var(--text);
	}
	.foot-legal {
		display: flex;
		gap: 24px;
	}
	.foot-copy {
		font-size: 13px;
		color: var(--text-dim);
	}
	@media (max-width: 640px) {
		.foot {
			flex-direction: column;
			align-items: center;
			text-align: center;
			gap: 20px;
		}
		.foot-links,
		.foot-legal {
			flex-wrap: wrap;
			justify-content: center;
			gap: 16px 20px;
		}
	}
</style>
