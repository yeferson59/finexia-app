<script lang="ts">
	/*
	 * Lo que alguien quiere saber antes de mandar dinero a un proyecto: a qué
	 * se destina y qué recibe a cambio. La respuesta honesta a lo segundo es
	 * «nada distinto», y se dice.
	 */
	interface Props {
		contactEmail: string;
	}

	let { contactEmail }: Props = $props();

	const notes = [
		{
			q: 'A qué se destina',
			a: 'Al servidor y la base de datos donde viven tus portafolios, al dominio y al correo, y al tiempo para seguir construyendo lo que falta.'
		},
		{
			q: 'Qué cambia en tu cuenta',
			a: 'Nada. Finexia funciona igual para quien apoya y para quien no: no hay funciones reservadas ni insignias.'
		},
		{
			q: 'Cuánto aportar',
			a: 'Los montos de arriba son atajos: con «Otro monto» escribes el que quieras desde $1.000. Cada aporte es un pago único.'
		},
		{
			q: 'Factura o certificado',
			a: 'Es un aporte voluntario a una persona natural: no genera factura ni certificado de donación, y no es deducible de impuestos.'
		}
	];
</script>

<section class="notes" aria-labelledby="notes-title">
	<h2 class="title" id="notes-title">Antes de aportar</h2>
	<dl>
		{#each notes as note (note.q)}
			<div>
				<dt>{note.q}</dt>
				<dd>{note.a}</dd>
			</div>
		{/each}
	</dl>
	<p class="contact">
		Si pagaste algo por error o tienes una pregunta, escribe a
		<a class="lp-link" href="mailto:{contactEmail}">{contactEmail}</a>.
	</p>
</section>

<style>
	.title {
		margin: 0 0 24px;
		font-size: var(--lp-fs-h3);
		font-stretch: 110%;
		font-weight: 620;
		letter-spacing: -0.01em;
	}

	dl {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 28px 56px;
		margin: 0;
	}

	dt {
		font-weight: 600;
	}

	dd {
		max-width: 46ch;
		margin: 6px 0 0;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.contact {
		margin: 40px 0 0;
		font-size: var(--lp-fs-sm);
		color: var(--lp-ink-2);
	}

	@media (max-width: 760px) {
		dl {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
