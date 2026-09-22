<script lang="ts">
	/*
	 * Lo que ve quien vuelve de la pasarela de Bold. El estado lo confirma el
	 * servidor preguntándole a Bold (`supportResult`); si Bold no contesta, un
	 * pago nunca se da por aprobado y se queda en «en proceso».
	 */
	import { formatCurrency } from '$lib/shared/format/money';
	import type { SupportResult } from '../support';

	interface Props {
		result: SupportResult;
		contactEmail: string;
	}

	let { result, contactEmail }: Props = $props();

	const amount = $derived(result.total ? ` de ${formatCurrency(result.total, 'COP')}` : '');
</script>

<div class="result {result.outcome}" role="status">
	{#if result.outcome === 'approved'}
		<p class="head">Recibimos tu aporte{amount}. Gracias.</p>
		<p>Bold te envía el comprobante al correo que usaste para pagar.</p>
	{:else if result.outcome === 'pending'}
		<p class="head">Tu pago está en proceso.</p>
		<p>
			Los pagos por PSE pueden tardar unos minutos en confirmarse. Bold te avisa por correo cuando
			termine; no hace falta que pagues otra vez.
		</p>
	{:else}
		<p class="head">El pago no se completó.</p>
		<p>
			Puedes intentarlo de nuevo abajo con otro medio de pago. Si ves un cobro en tu banco, escribe
			a
			<a class="lp-link" href="mailto:{contactEmail}">{contactEmail}</a>.
		</p>
	{/if}
</div>

<style>
	.result {
		max-width: 64ch;
		padding: 20px 24px;
		border-left: 4px solid var(--lp-ink-2);
		background: var(--lp-paper-2);
	}

	.result.approved {
		border-left-color: var(--lp-cripto);
	}

	.result.pending {
		border-left-color: var(--lp-jubilacion);
	}

	.result.failed {
		border-left-color: var(--lp-error);
	}

	p {
		margin: 0;
		color: var(--lp-ink-2);
	}

	.head {
		margin-bottom: 4px;
		font-size: var(--lp-fs-lead);
		font-weight: 600;
		color: var(--lp-ink);
	}
</style>
