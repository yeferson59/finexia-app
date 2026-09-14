<script lang="ts">
	/**
	 * Las cuentas de efectivo: una por plataforma y moneda.
	 *
	 * Se agrupa por cuenta porque es la cifra que el usuario reconoce, la de su
	 * extracto. Que cada portafolio lleve por dentro su propio saldo es cómo el
	 * dinero suma en su valor, y solo se enseña cuando reparte algo: con un único
	 * portafolio no se nombra, y una cuenta repartida lista cuánto pone cada uno.
	 *
	 * El botón de cada fila abre el formulario con esa cuenta ya elegida —casi
	 * siempre se anota sobre una que ya existe— y, si está repartida, con el
	 * portafolio de su saldo mayor, que el formulario deja cambiar.
	 */
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { groupCashAccounts, type CashAccount, type CashBalance } from '../cash';

	interface Props {
		balances: CashBalance[];
		/** Si se nombra el portafolio de cada cuenta. Con uno solo, sobra. */
		showPortfolio: boolean;
		onRecord: (balance: CashBalance) => void;
	}

	let { balances, showPortfolio, onRecord }: Props = $props();

	const accounts = $derived(groupCashAccounts(balances));

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	const displayCurrency = $derived(balances[0]?.displayCurrency ?? '');

	/* Si todo está en la moneda de la pantalla, la columna repetiría el saldo. */
	const showValue = $derived(balances.some((b) => b.currency !== b.displayCurrency));

	function lastMovement(account: CashAccount): string {
		if (!account.lastMovementDate) return 'sin movimientos';
		return `último movimiento el ${formatCalendarDate(account.lastMovementDate, {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		})}`;
	}

	function meta(account: CashAccount): string {
		const last = lastMovement(account);

		if (account.balances.length > 1) {
			return `Repartida en ${account.balances.length} portafolios · ${last}`;
		}
		if (showPortfolio) {
			return `${account.balances[0].portfolioName} · ${last}`;
		}

		return last.charAt(0).toUpperCase() + last.slice(1);
	}
</script>

<DataTable caption="Cuentas de efectivo por plataforma y moneda">
	<thead>
		<tr>
			<th scope="col">Cuenta</th>
			<th scope="col" class="col-currency">Moneda</th>
			<th scope="col" class="num">Saldo</th>
			{#if showValue}
				<th scope="col" class="num col-value">En {displayCurrency}</th>
			{/if}
			<th scope="col" class="col-actions" aria-label="Acciones"></th>
		</tr>
	</thead>
	<tbody>
		{#each accounts as account (account.key)}
			<tr class="row" class:emptied={account.balance === 0}>
				<td class="cell-account">
					<span class="source">{account.sourceName || 'Sin plataforma'}</span>
					<span class="meta">{meta(account)}</span>
					{#if account.balances.length > 1}
						<ul class="portions" aria-label="Reparto por portafolio">
							{#each account.balances as balance (balance.entryId)}
								<li>
									<span class="portion-name">{balance.portfolioName}</span>
									<span class="portion-amount">
										{money(parseFloat(balance.balance) || 0, balance.currency)}
									</span>
								</li>
							{/each}
						</ul>
					{/if}
				</td>
				<td class="col-currency"><span class="code">{account.currency}</span></td>
				<td class="num figure">{money(account.balance, account.currency)}</td>
				{#if showValue}
					<td class="num figure col-value muted">
						{#if account.fxConverted}
							{money(account.value, account.displayCurrency)}
						{:else}
							<span title="No hay tasa de cambio guardada para convertir este saldo">—</span>
						{/if}
					</td>
				{/if}
				<td class="col-actions">
					<Button
						type="button"
						variant="secondary"
						size="sm"
						onclick={() => onRecord(account.balances[0])}
					>
						Movimiento
					</Button>
				</td>
			</tr>
		{/each}
	</tbody>
</DataTable>

<style>
	.row :global(td) {
		padding-top: 0.85rem;
		padding-bottom: 0.85rem;
	}

	.cell-account {
		width: 100%;
		min-width: 12rem;
	}

	.source {
		display: block;
		font-family: var(--font-display);
		font-size: 1.02rem;
		font-weight: 500;
		color: var(--text);
	}

	.meta {
		display: block;
		margin-top: 0.2rem;
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.portions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem 1.25rem;
		margin: 0.45rem 0 0;
		padding: 0;
		list-style: none;
		font-size: 0.8rem;
	}

	.portions li {
		display: flex;
		align-items: baseline;
		gap: 0.45rem;
	}

	.portion-name {
		color: var(--text-muted);
	}

	.portion-amount {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.code {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		letter-spacing: 0.06em;
		color: var(--amber);
	}

	.figure {
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.muted {
		color: var(--text-muted);
	}

	.col-actions {
		text-align: right;
		white-space: nowrap;
	}

	/* Una cuenta vaciada sigue en la lista —es donde cae el próximo depósito—,
	   pero sin competir con las que tienen dinero. */
	.emptied .source,
	.emptied .figure {
		color: var(--text-dim);
	}

	@media (max-width: 720px) {
		.col-currency,
		.col-value {
			display: none;
		}
	}
</style>
