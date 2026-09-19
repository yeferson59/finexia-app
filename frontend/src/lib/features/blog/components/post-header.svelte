<script lang="ts">
	/*
	 * La cabecera del artículo: el titular y el resumen.
	 *
	 * El titular manda —la misma voz ancha del índice, un escalón por debajo—
	 * y el resumen va en la serif del cuerpo, que es donde empieza la lectura.
	 * La fecha, el autor y las etiquetas están en el margen (`post-aside`), a la
	 * altura del texto, que es donde se miran solo si hacen falta.
	 */
	import type { PostMeta } from '../blog';

	interface Props {
		post: PostMeta;
	}

	let { post }: Props = $props();
</script>

<header class="post-header">
	<h1>{post.title}</h1>
	<p class="lead">{post.description}</p>
</header>

<style>
	h1 {
		max-width: 18ch;
		margin: 0;
		font-size: clamp(38px, 5.4vw, 72px);
		font-stretch: 122%;
		font-weight: 650;
		line-height: 0.98;
		letter-spacing: -0.032em;
		text-wrap: balance;
		overflow-wrap: anywhere;
	}

	.lead {
		max-width: 40ch;
		margin: 28px 0 0;
		font-family: var(--blog-serif);
		font-size: clamp(20px, 2.2vw, 24px);
		font-weight: 350;
		line-height: 1.45;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	@media (max-width: 640px) {
		.lead {
			margin-top: 18px;
		}
	}
</style>
