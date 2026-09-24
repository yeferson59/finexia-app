<script lang="ts">
	/**
	 * Las cuentas de efectivo, una hoja por plataforma.
	 *
	 * Se leen como en la app del banco: la entidad y, debajo, lo que guarda en
	 * cada moneda, con sus bolsillos y depósitos colgando de la cuenta a la que
	 * pertenecen. Cada fila es un `cash-drawer`, y todas comparten columnas, así
	 * que los importes caen en la misma vertical de una plataforma a otra.
	 *
	 * Lo que se le hace a una cuenta entera —abrirle un bolsillo, un depósito,
	 * mover dinero entre sus cajones— va al pie de sus cajones, una sola vez, y
	 * no repetido en cada fila.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { CashPocket } from '$lib/api/types';
	import type { CashAccount, CashBalance } from '../cash';
	import {
		emptyCashPockets,
		groupCashPlatforms,
		openCashPockets,
		pocketAsCashAccount
	} from '../pockets';
	import type { CashRate } from '../rates';
	import CashDrawer from './cash-drawer.svelte';
	import CashRecalcNotice from './cash-recalc-notice.svelte';
	import { cashRecalc } from '../recalc.svelte';

	interface Props {
		balances: CashBalance[];
		/** Todas las versiones de las tasas; cada cuenta toma las suyas. */
		rates: CashRate[];
		/** Todos los bolsillos: los que aún no guardan nada no llegan con los saldos. */
		pockets: CashPocket[];
		/** Si se nombra el portafolio de cada cuenta. Con uno solo, sobra. */
		showPortfolio: boolean;
		today: string;
		onRecord: (balance: CashBalance) => void;
		/** Abre la rentabilidad de una cuenta o de un bolsillo. */
		onRate: (account: CashAccount) => void;
		/** Abre el formulario de un bolsillo: uno nuevo en la cuenta, o el que ya hay. */
		onPocket: (account: CashAccount, pocket: CashPocket | null) => void;
		/**
		 * Si hay otra plataforma a la que trasladar el dinero. Sin ella un
		 * traslado solo puede ir a otro cajón de la misma cuenta, así que una
		 * cuenta sin bolsillos no tendría a dónde mandarlo.
		 */
		canTransfer: boolean;
		/** Mueve dinero a otro cajón de la cuenta, o a otra plataforma. */
		onMove: (account: CashAccount) => void;
		/** Abre un depósito a tasa fija en la cuenta, o mira el que ya está abierto. */
		onDeposit: (account: CashAccount, pocket: CashPocket | null) => void;
	}

	let {
		balances,
		rates,
		pockets,
		showPortfolio,
		today,
		canTransfer,
		onRecord,
		onRate,
		onPocket,
		onMove,
		onDeposit
	}: Props = $props();

	const platforms = $derived(groupCashPlatforms(balances));

	let list = $state<HTMLElement>();

	/*
	 * La fila recalculada destella y se apaga sola. Con la Web Animations API,
	 * desde aquí: las filas no cargan un estado que solo dura un momento, y el
	 * destello vuelve al fondo que tenga cada una.
	 */
	/* Salir de la página cierra el aviso: al volver ya no habla de lo último. */
	$effect(() => () => cashRecalc.dismiss());

	$effect(() => {
		const highlight = cashRecalc.highlight;
		if (!highlight || !list) return;

		list.querySelector(`[data-account="${CSS.escape(highlight)}"]`)?.animate(
			[
				{ backgroundColor: 'rgba(34, 201, 126, 0.1)', offset: 0 },
				{ backgroundColor: 'rgba(34, 201, 126, 0.1)', offset: 0.35 }
			],
			{ duration: 2400, easing: 'ease-out' }
		);
	});

	const pocketOf = (account: CashAccount) => pockets.find((p) => p.id === account.pocketId) ?? null;

	/* Los cajones de una cuenta, en el orden en que cuelgan: los que guardan algo
	   y luego los que todavía están vacíos. */
	const drawersOf = (account: CashAccount) => [
		...openCashPockets(account, pockets),
		...emptyCashPockets(account, pockets).map((blank) => pocketAsCashAccount(account, blank))
	];

	/* Cuántos cajones suma la plataforma: sus cuentas y lo que cuelga de ellas. */
	const drawerCount = (accounts: CashAccount[]) =>
		accounts.reduce((sum, account) => sum + 1 + drawersOf(account).length, 0);

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));
</script>

<!-- Arriba de las cuentas, que es donde cambiaron los números. -->
<CashRecalcNotice feedback={cashRecalc} />

<div class="platforms" bind:this={list}>
	{#each platforms as platform (platform.sourceId)}
		{@const name = platform.sourceName || 'Sin plataforma'}
		{@const count = drawerCount(platform.accounts)}
		<section class="platform" aria-labelledby="cash-platform-{platform.sourceId}">
			<header class="platform-head">
				<h3 id="cash-platform-{platform.sourceId}">{name}</h3>
				<!-- Solo cuando suma más de un cajón y todos se pudieron convertir: con
				     uno repetiría la fila, y sin tasa sería un total a medias. -->
				{#if count > 1 && !platform.partial}
					<p class="platform-total">
						<span class="figure">{money(platform.value, platform.displayCurrency)}</span>
						entre {count} cajones
					</p>
				{/if}
			</header>

			<ul class="groups">
				{#each platform.accounts as account (account.key)}
					{@const drawers = drawersOf(account)}
					<li class="group">
						<ul class="drawers">
							<CashDrawer
								{account}
								pocket={null}
								{rates}
								platformName={name}
								{showPortfolio}
								{today}
								rail={drawers.length > 0}
								onRecord={() => onRecord(account.balances[0])}
								onRate={() => onRate(account)}
								onPocket={() => onPocket(account, null)}
								onDeposit={() => onDeposit(account, null)}
							/>
							{#each drawers as drawer, index (drawer.key)}
								{@const pocket = pocketOf(drawer)}
								<CashDrawer
									account={drawer}
									{pocket}
									{rates}
									platformName={name}
									{showPortfolio}
									{today}
									nested
									last={index === drawers.length - 1}
									onRecord={() => onRecord(drawer.balances[0])}
									onRate={() => onRate(drawer)}
									onPocket={() => onPocket(drawer, pocket)}
									onDeposit={() => onDeposit(drawer, pocket)}
								/>
							{/each}
						</ul>

						<div class="group-actions">
							<button type="button" class="tool" onclick={() => onPocket(account, null)}>
								<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M8 3v10M3 8h10" /></svg>
								Agregar bolsillo<span class="sr-only"> a {name}, {account.currency}</span>
							</button>
							<button type="button" class="tool" onclick={() => onDeposit(account, null)}>
								<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M8 3v10M3 8h10" /></svg>
								Abrir depósito a plazo<span class="sr-only"> en {name}, {account.currency}</span>
							</button>
							{#if drawers.length > 0 || canTransfer}
								<button type="button" class="tool" onclick={() => onMove(account)}>
									<svg viewBox="0 0 16 16" aria-hidden="true">
										<path d="M3 5.5h9.5M10 3l2.5 2.5L10 8M13 10.5H3.5M6 8l-2.5 2.5L6 13" />
									</svg>
									Mover dinero<span class="sr-only"> de {name}, {account.currency}</span>
								</button>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
		</section>
	{/each}
</div>

<style>
	.platforms {
		display: grid;
		gap: 1rem;
	}

	/* Una hoja por entidad: la superficie de lo que es un objeto suelto en el
	   panel, sin sombra, con su nombre arriba como el membrete de un extracto. */
	.platform {
		border: 1px solid var(--border);
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.022);
	}

	.platform-head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.25rem 1.5rem;
		padding: 1.1rem 1.5rem 1rem 1.25rem;
		border-bottom: 1px solid var(--border);
	}

	h3 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-weight: 400;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	.platform-total {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.platform-total .figure {
		margin-right: 0.25rem;
		font-family: var(--font-mono);
		font-size: 0.88rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.groups,
	.drawers {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.group + .group {
		border-top: 1px solid var(--border);
	}

	/* Las acciones de la cuenta entera, al pie de sus cajones y en la columna de
	   los nombres: se hacen de vez en cuando, así que no llevan borde. */
	.group-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem 0.5rem;
		padding: 0 1.5rem 0.85rem calc(1.25rem + 3.5rem + 1.25rem - 0.55rem);
	}

	.tool {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.3rem 0.55rem;
		border: none;
		border-radius: 6px;
		background: transparent;
		font: inherit;
		font-size: 0.78rem;
		color: var(--text-dim);
		cursor: pointer;
		transition:
			background-color 0.15s ease,
			color 0.15s ease;
	}

	.tool:hover {
		background: var(--surface-2);
		color: var(--text);
	}

	.tool svg {
		width: 0.8rem;
		height: 0.8rem;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.5;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	@media (max-width: 980px) {
		.platform-head {
			padding: 1rem 1rem 0.9rem;
		}

		.group-actions {
			padding: 0 1rem 0.85rem calc(0.75rem + 2.75rem + 0.85rem - 0.55rem);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.tool {
			transition: none;
		}
	}
</style>
