<script lang="ts">
	/**
	 * Pegar varios valores de una vez: la tabla de un extracto o de una hoja de
	 * cálculo, una línea por día. Antes de guardar dice cuántas líneas leyó y
	 * cuáles no, porque cada entidad exporta fechas y cifras a su manera.
	 *
	 * Cada día reemplaza el valor que ya tuviera, como al anotarlos uno por uno,
	 * y la gráfica de crecimiento se corrige desde el más antiguo.
	 */
	import { enhance } from '$app/forms';
	import type { SubmitFunction } from '@sveltejs/kit';
	import Button from '$lib/ui/button.svelte';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { parseMarksTable } from '../performance';

	interface Props {
		assetId: string;
		/** Qué es cada cifra: el valor de unidad o el saldo. */
		byBalance: boolean;
	}

	let { assetId, byBalance }: Props = $props();

	let text = $state('');
	let error = $state('');
	let saved = $state(0);
	let submitting = $state(false);

	const table = $derived(parseMarksTable(text));
	const rows = $derived(table.rows);

	const day = (iso: string) =>
		formatCalendarDate(iso, { day: 'numeric', month: 'short', year: 'numeric' });

	const submit: SubmitFunction = () => {
		submitting = true;
		error = '';
		saved = 0;

		return async ({ result, update }) => {
			submitting = false;

			if (result.type === 'success') {
				saved = rows.length;
				text = '';
				await update({ reset: false });
			} else if (result.type === 'failure') {
				error = (result.data?.error as string) ?? 'No pudimos guardar los valores.';
			} else {
				await update();
			}
		};
	};
</script>

<details class="paste">
	<summary>Pegar varios valores del extracto</summary>

	<form method="POST" action="?/saveMarks" class="rail-fields" use:enhance={submit}>
		<input type="hidden" name="id" value={assetId} />
		<input type="hidden" name="tracking" value={byBalance ? 'balance' : 'units'} />
		<input type="hidden" name="marks" value={JSON.stringify(rows)} />

		<div class="field">
			<label for="fund-paste-{assetId}"
				>Una línea por día: la fecha y {byBalance ? 'el saldo' : 'el valor de unidad'}</label
			>
			<textarea
				id="fund-paste-{assetId}"
				rows="6"
				placeholder={byBalance
					? '31/08/2026\t15.110.000\n30/09/2026\t13.050.000'
					: '01/09/2026\t12.345,678901\n30/09/2026\t12.431,22'}
				bind:value={text}></textarea>
			<p class="hint">
				Copia las columnas desde el extracto o una hoja de cálculo. Sirven «30/09/2026» o
				«2026-09-30», y cifras como «12.431,22» o «12,431.22».
			</p>
		</div>

		{#if text.trim()}
			<p class="hint" aria-live="polite">
				{#if rows.length > 0}
					{rows.length}
					{rows.length === 1 ? 'valor leído' : 'valores leídos'}, del {day(rows[0].date)} al {day(
						rows[rows.length - 1].date
					)}.
				{:else}
					No se leyó ningún valor.
				{/if}
				{#if table.errors.length > 0}
					<span class="bad">
						{table.errors.length === 1
							? 'Una línea no se pudo leer'
							: `${table.errors.length} líneas no se pudieron leer`}:
						{table.errors
							.slice(0, 3)
							.map((e) => `línea ${e.line}`)
							.join(', ')}{table.errors.length > 3 ? '…' : ''}.
					</span>
				{/if}
			</p>
		{/if}

		{#if error}
			<p class="feedback error" role="alert">{error}</p>
		{:else if saved > 0}
			<p class="feedback success" role="status">
				{saved === 1 ? 'Se guardó 1 valor' : `Se guardaron ${saved} valores`}.
			</p>
		{/if}

		<div class="act">
			<Button
				type="submit"
				variant="secondary"
				size="sm"
				loading={submitting}
				disabled={rows.length === 0 || table.errors.length > 0}
			>
				{rows.length > 1 ? `Guardar ${rows.length} valores` : 'Guardar valores'}
			</Button>
		</div>
	</form>
</details>

<style>
	.paste {
		margin-top: 1.25rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	.paste summary {
		font-size: 0.87rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	.paste form {
		margin-top: 1rem;
	}

	textarea {
		font-family: var(--font-mono);
		font-size: 0.82rem;
	}

	.bad {
		color: var(--red);
	}

	.feedback {
		margin: 0;
	}

	.act {
		display: flex;
		justify-content: flex-end;
	}
</style>
