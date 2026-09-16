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
	import type { CashPocket } from '$lib/api/types';
	import { cashAccountRate, describeCashAccountRate, type CashRate } from '../rates';

	interface Props {
		balances: CashBalance[];
		/** Todas las versiones de las tasas; cada cuenta toma las suyas. */
		rates: CashRate[];
		/** Todos los bolsillos: los que aún no guardan nada no llegan con los saldos. */
		pockets: CashPocket[];
		/** Si se nombra el portafolio de cada cuenta. Con uno solo, sobra. */
		showPortfolio: boolean;
		onRecord: (balance: CashBalance) => void;
		/** Abre la rentabilidad de una cuenta o de un bolsillo. */
		onRate: (account: CashAccount) => void;
		/** Abre el formulario de un bolsillo: uno nuevo en la cuenta, o el que ya hay. */
		onPocket: (account: CashAccount, pocket: CashPocket | null) => void;
		/** Mueve dinero entre los cajones de una cuenta. */
		onMove: (account: CashAccount) => void;
	}

	let { balances, rates, pockets, showPortfolio, onRecord, onRate, onPocket, onMove }: Props =
		$props();

	const today = todayLocalDateString();

	const platforms = $derived(groupCashPlatforms(balances));

	/* El bolsillo de una fila, para pasarlo a los formularios que lo piden. */
	const pocketOf = (account: CashAccount) => pockets.find((p) => p.id === account.pocketId) ?? null;

	/*
	 * Los bolsillos de una cuenta que todavía no guardan nada. No llegan con los
	 * saldos —no hay posición que listar—, y aun así tienen que verse: es donde
	 * se les da una tasa y desde donde se mueve el primer dinero.
	 */
	function emptyPockets(account: CashAccount): CashPocket[] {
		const held = new Set(account.pockets.map((p) => p.pocketId));

		return pockets.filter(
			(p) =>
				p.sourceId === account.sourceId &&
				p.currency === account.currency &&
				p.closedOn === null &&
				!held.has(p.id)
		);
	}

	/* Un bolsillo vacío como una cuenta más, para que la fila sea la misma. */
	const asAccount = (account: CashAccount, pocket: CashPocket): CashAccount => ({
		...account,
		key: `${account.sourceId}:${account.currency}:${pocket.id}`,
		pocketId: pocket.id,
		pocketName: pocket.name,
		pocketKind: pocket.kind,
		balance: 0,
		value: 0,
		fxConverted: true,
		lastMovementDate: null,
		balances: [],
		pockets: []
	});

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

<!--
	Cada cajón de una cuenta es la misma fila: la principal y, debajo, sus
	bolsillos. Un bolsillo cuenta dentro de la plataforma —su dinero suma en
	ella— y lo que es suyo es la tasa, así que se anida en lugar de listarse
	aparte.
-->
{#snippet drawer(account: CashAccount, platformName: string)}
	{@const rate = cashAccountRate(
		rates,
		account.sourceId,
		account.currency,
		today,
		account.pocketId
	)}
	{@const rateLine = describeCashAccountRate(rate, (amount) => money(amount, account.currency))}
	{@const earned = account.balances.reduce(
		(sum, b) => sum + (parseFloat(b.interestThisMonth) || 0),
		0
	)}
	{@const pending = account.balances.reduce(
		(sum, b) => sum + (parseFloat(b.pendingInterest) || 0),
		0
	)}
	{@const where = account.pocketName
		? `${platformName}, ${account.pocketName}`
		: `${platformName}, ${account.currency}`}
	<li
		class="account"
		class:emptied={account.balance === 0}
		class:pocket={account.pocketId !== null}
	>
		<span class="code">
			{#if account.pocketId === null}
				{account.currency}
			{:else}
				<span class="sr-only">Bolsillo</span>
			{/if}
		</span>

		<div class="where">
			{#if account.pocketName}
				<p class="line name">{account.pocketName}</p>
			{/if}
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
			{:else if showPortfolio && account.balances.length === 1}
				<p class="line">Suma en {account.balances[0].portfolioName}</p>
			{/if}
			<p class="line quiet">{history(account)}</p>
			<button
				type="button"
				class="rate"
				class:earning={rate.current !== null}
				onclick={() => onRate(account)}
			>
				<span class="sr-only">Rentabilidad de {where}: </span>
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

		<div class="row-actions">
			{#if account.pocketId === null}
				<button type="button" class="record" onclick={() => onRecord(account.balances[0])}>
					Registrar<span class="sr-only"> un movimiento en {where}</span>
				</button>
			{:else}
				<button
					type="button"
					class="record link"
					onclick={() => onPocket(account, pocketOf(account))}
				>
					Editar<span class="sr-only"> el bolsillo {account.pocketName}</span>
				</button>
			{/if}
		</div>
	</li>
{/snippet}

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
					{@render drawer(account, name)}
					{#each account.pockets as held (held.key)}
						{@render drawer(held, name)}
					{/each}
					{#each emptyPockets(account) as empty (empty.id)}
						{@render drawer(asAccount(account, empty), name)}
					{/each}
					<li class="drawer-actions">
						<button type="button" class="link" onclick={() => onPocket(account, null)}>
							Agregar bolsillo<span class="sr-only"> a {name}, {account.currency}</span>
						</button>
						{#if account.pockets.length > 0 || emptyPockets(account).length > 0}
							<button
								type="button"
								class="link"
								disabled={account.balances.length === 0 && account.pockets.length === 0}
								onclick={() => onMove(account)}
							>
								Mover dinero<span class="sr-only">
									entre los saldos de {name}, {account.currency}</span
								>
							</button>
						{/if}
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
	.row-actions {
		grid-area: record;
		justify-self: end;
		display: flex;
		gap: 0.4rem;
	}

	.record {
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

	/*
	 * Un bolsillo se sangra bajo su cuenta y pierde el código de moneda: es la
	 * misma cuenta, en la misma moneda, en otra cajita. La línea de arriba lo
	 * separa menos que a otra cuenta, porque no lo es.
	 */
	.account.pocket {
		margin-left: 3rem;
		padding: 0.7rem 0;
	}

	.account.pocket + .account.pocket,
	.account + .account.pocket {
		border-top-color: var(--border-subtle, var(--border));
	}

	.name {
		font-weight: 500;
		color: var(--text);
	}

	/* Las dos acciones de la cuenta, al pie de sus cajones y muy calladas: no
	   compiten con «Registrar», que es lo que se hace todos los días. */
	.drawer-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		margin: 0 1.5rem;
		padding: 0.25rem 0 0.7rem 3rem;
	}

	.link {
		padding: 0;
		border: 0;
		background: transparent;
		font: inherit;
		font-size: 0.76rem;
		color: var(--text-dim);
		cursor: pointer;
		transition: color 0.2s ease;
	}

	.link:hover:not(:disabled) {
		color: var(--amber-light);
	}

	.link:disabled {
		cursor: default;
		opacity: 0.5;
	}

	.link:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
		border-radius: 4px;
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

		.account.pocket {
			margin-left: 1.9rem;
		}

		.drawer-actions {
			margin: 0 1rem;
			padding-left: 1.9rem;
		}

		.figures {
			align-items: flex-start;
		}
	}
</style>
