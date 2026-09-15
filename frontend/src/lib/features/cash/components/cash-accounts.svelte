<script lang="ts">
	/**
	 * Las cuentas de efectivo, agrupadas por plataforma.
	 *
	 * Se leen como en la app del banco: la entidad y, debajo, lo que guarda en
	 * cada moneda. Con una fila por cuenta, el nombre de la plataforma se repetía
	 * en cada moneda y dos filas «Banco Demo» parecían un duplicado.
	 *
	 * El importe grande es el de su propia moneda, el que coincide con el
	 * extracto; su equivalente en la moneda de la pantalla va debajo. Que cada
	 * portafolio lleve por dentro su propio saldo solo se enseña cuando reparte
	 * algo: con un único portafolio no se nombra, y una cuenta repartida lista
	 * cuánto pone cada uno.
	 *
	 * «Registrar» abre el formulario con esa cuenta ya elegida —casi siempre se
	 * anota sobre una que ya existe— y, si está repartida, con el portafolio de
	 * su saldo mayor, que el formulario deja cambiar.
	 *
	 * La tasa de la cuenta va debajo de su historia, como un botón callado: se
	 * consulta más de lo que se cambia. En verde mientras rinde; sin tasa invita
	 * a anotarla.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import { groupCashPlatforms, type CashAccount, type CashBalance } from '../cash';
	import { cashAccountRate, describeCashAccountRate, type CashRate } from '../rates';

	interface Props {
		balances: CashBalance[];
		/** Todas las versiones de las tasas; cada cuenta toma las suyas. */
		rates: CashRate[];
		/** Si se nombra el portafolio de cada cuenta. Con uno solo, sobra. */
		showPortfolio: boolean;
		onRecord: (balance: CashBalance) => void;
		/** Abre la rentabilidad de una cuenta: anotar su tasa o cambiarla. */
		onRate: (account: CashAccount) => void;
	}

	let { balances, rates, showPortfolio, onRecord, onRate }: Props = $props();

	const today = todayLocalDateString();

	const platforms = $derived(groupCashPlatforms(balances));

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	/* Una cuenta vaciada sigue en la lista —es donde cae el próximo depósito— y
	   dice desde cuándo está así, que es lo que la distingue de una sin estrenar. */
	function history(account: CashAccount): string {
		if (!account.lastMovementDate) return 'Sin movimientos todavía';

		const when = formatCalendarDate(account.lastMovementDate, {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});

		return account.balance === 0 ? `Vacía desde el ${when}` : `Último movimiento el ${when}`;
	}
</script>

<div class="platforms">
	{#each platforms as platform (platform.sourceId)}
		{@const name = platform.sourceName || 'Sin plataforma'}
		<section class="platform" aria-labelledby="cash-platform-{platform.sourceId}">
			<header class="platform-head">
				<h3 id="cash-platform-{platform.sourceId}">{name}</h3>
				<!-- Solo cuando suma varias monedas y todas se pudieron convertir: con
				     una sola repetiría la fila, y sin tasa sería un total a medias. -->
				{#if platform.accounts.length > 1 && !platform.partial}
					<p class="platform-total">
						<span class="figure">{money(platform.value, platform.displayCurrency)}</span>
						entre sus {platform.accounts.length} cuentas
					</p>
				{/if}
			</header>

			<ul class="accounts">
				{#each platform.accounts as account (account.key)}
					{@const rate = cashAccountRate(rates, account.sourceId, account.currency, today)}
					{@const rateLine = describeCashAccountRate(rate)}
					{@const earned = account.balances.reduce(
						(sum, b) => sum + (parseFloat(b.interestThisMonth) || 0),
						0
					)}
					{@const pending = account.balances.reduce(
						(sum, b) => sum + (parseFloat(b.pendingInterest) || 0),
						0
					)}
					<li class="account" class:emptied={account.balance === 0}>
						<span class="code">{account.currency}</span>

						<div class="where">
							{#if account.balances.length > 1}
								<p class="line">Suma en {account.balances.length} portafolios</p>
								<ul class="portions">
									{#each account.balances as balance (balance.entryId)}
										<li>
											<span class="portion-name">{balance.portfolioName}</span>
											<span class="portion-amount">
												{money(parseFloat(balance.balance) || 0, balance.currency)}
											</span>
										</li>
									{/each}
								</ul>
							{:else if showPortfolio}
								<p class="line">Suma en {account.balances[0].portfolioName}</p>
							{/if}
							<p class="line quiet">{history(account)}</p>
							<button
								type="button"
								class="rate"
								class:earning={rate.current !== null}
								onclick={() => onRate(account)}
							>
								<span class="sr-only">Rentabilidad de {name}, {account.currency}: </span>
								{rateLine ?? 'Agregar tasa'}
							</button>
							{#if earned > 0}
								<p class="line quiet" style:color="var(--green)">
									+{money(earned, account.currency)} en intereses este mes
								</p>
							{/if}
							<!-- Con abono mensual lo calculado espera al cierre del mes: está
							     ganado, pero todavía no está en el saldo. -->
							{#if pending > 0}
								<p class="line quiet">
									+{money(pending, account.currency)} calculados, se abonan al cerrar el mes
								</p>
							{/if}
						</div>

						<p class="figures">
							<span class="native">{money(account.balance, account.currency)}</span>
							<!-- Vacía no hay nada que convertir: «≈ $0,00» solo añade ruido. -->
							{#if account.currency !== account.displayCurrency && account.balance !== 0}
								{#if account.fxConverted}
									<span class="converted">≈ {money(account.value, account.displayCurrency)}</span>
								{:else}
									<span class="converted">sin tasa a {account.displayCurrency}</span>
								{/if}
							{/if}
						</p>

						<button type="button" class="record" onclick={() => onRecord(account.balances[0])}>
							Registrar<span class="sr-only"> un movimiento en {name}, {account.currency}</span>
						</button>
					</li>
				{/each}
			</ul>
		</section>
	{/each}
</div>

<style>
	.platform + .platform {
		border-top: 1px solid var(--border-strong);
	}

	.platform-head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.25rem 1.5rem;
		padding: 1.15rem 1.5rem 0.2rem;
	}

	h3 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 500;
		color: var(--text);
	}

	.platform-total {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.platform-total .figure {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.accounts {
		margin: 0;
		padding: 0 0 0.35rem;
		list-style: none;
	}

	/*
	 * Cada cuenta es su propia rejilla, pero las dos últimas columnas tienen
	 * ancho fijo: así los importes, alineados a la derecha, caen en la misma
	 * vertical en todas las plataformas y se comparan de un vistazo.
	 */
	.account {
		display: grid;
		grid-template-columns: 3.25rem minmax(0, 1fr) auto 7rem;
		grid-template-areas: 'code where figures record';
		align-items: start;
		column-gap: 1.25rem;
		margin: 0 1.5rem;
		padding: 0.9rem 0;
	}

	.account + .account {
		border-top: 1px solid var(--border);
	}

	.code {
		grid-area: code;
		padding-top: 0.2rem;
		font-family: var(--font-mono);
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		color: var(--text-muted);
	}

	.where {
		grid-area: where;
		min-width: 0;
		padding-top: 0.1rem;
	}

	.line {
		margin: 0;
		font-size: 0.84rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.quiet {
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.portions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.2rem 1.25rem;
		margin: 0.2rem 0 0.35rem;
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

	.figures {
		grid-area: figures;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.2rem;
		margin: 0;
	}

	.native {
		font-family: var(--font-mono);
		font-size: 1.05rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text);
	}

	.converted {
		font-family: var(--font-mono);
		font-size: 0.76rem;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--text-dim);
	}

	/* Una acción por cuenta, callada: cinco botones ámbar seguidos competían con
	   el principal de la página y con las propias cifras. */
	.record {
		grid-area: record;
		justify-self: end;
		margin-top: -0.1rem;
		padding: 0.35rem 0.85rem;
		border: 1px solid var(--border-strong);
		border-radius: 999px;
		background: transparent;
		font: inherit;
		font-size: 0.78rem;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			color 0.2s ease;
	}

	.record:hover {
		border-color: var(--amber);
		color: var(--amber-light);
	}

	/*
	 * La tasa, más callada todavía que «Registrar»: discontinua mientras la
	 * cuenta no rinde, y en el verde de los intereses mientras rinde.
	 */
	.rate {
		display: inline-flex;
		align-items: center;
		margin-top: 0.4rem;
		padding: 0.15rem 0.65rem;
		border: 1px dashed var(--border-strong);
		border-radius: 999px;
		background: transparent;
		font: inherit;
		font-size: 0.76rem;
		font-variant-numeric: tabular-nums;
		text-align: left;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			color 0.2s ease;
	}

	.rate:hover {
		border-color: var(--amber);
		color: var(--amber-light);
	}

	.rate:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.rate.earning {
		border-style: solid;
		border-color: rgba(34, 201, 126, 0.3);
		color: var(--green);
	}

	.rate.earning:hover {
		border-color: var(--green);
		color: var(--green);
	}

	.emptied .code,
	.emptied .native {
		color: var(--text-dim);
	}

	@media (max-width: 640px) {
		.platform-head {
			padding: 1rem 1rem 0.1rem;
		}

		.account {
			grid-template-columns: 2.75rem minmax(0, 1fr) auto;
			grid-template-areas:
				'code figures record'
				'. where where';
			row-gap: 0.4rem;
			column-gap: 0.75rem;
			margin: 0 1rem;
		}

		.figures {
			align-items: flex-start;
		}
	}
</style>
