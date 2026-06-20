<script lang="ts">
	import Brand from './brand.svelte';

	const currentYear = new Date().getFullYear();
</script>

<footer>
	<div class="wrap foot">
		<Brand iconSize={26} nameSize={18} />
		<nav class="foot-links">
			<a href="#beneficios">Beneficios</a>
			<a href="#como-funciona">Cómo funciona</a>
			<a href="#faq">Preguntas</a>
			<a href="#waitlist">Lista de espera</a>
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
	.foot-links a {
		font-size: 13.5px;
		color: var(--text-muted);
		transition: color 0.2s;
	}
	.foot-links a:hover {
		color: var(--text);
	}
	.foot-copy {
		font-size: 13px;
		color: var(--text-dim);
	}
</style>
