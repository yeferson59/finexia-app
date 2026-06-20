<script lang="ts">
	import { onMount } from 'svelte';

	let waitlistEmail = $state('');
	let waitlistError = $state(false);
	let waitlistErrorMessage = $state('');
	let waitlistSuccess = $state(false);
	let submitting = $state(false);
	let countdown = $state({ days: '00', hours: '00', mins: '00', secs: '00' });

	function pad(n: number) {
		return String(n).padStart(2, '0');
	}

	async function submitWaitlist(event: SubmitEvent) {
		event.preventDefault();
		const formEl = event.currentTarget as HTMLFormElement;
		waitlistError = false;
		waitlistErrorMessage = '';
		submitting = true;

		try {
			const res = await fetch('/api/waitlist', {
				method: 'POST',
				headers: { accept: 'application/json' },
				body: new FormData(formEl)
			});
			const data = await res.json();

			if (data.success) {
				waitlistSuccess = true;
			} else {
				waitlistError = true;
				waitlistErrorMessage = data.error ?? 'Ocurrió un error. Inténtalo de nuevo.';
			}
		} catch {
			waitlistError = true;
			waitlistErrorMessage = 'Ocurrió un error. Inténtalo de nuevo.';
		} finally {
			submitting = false;
		}
	}

	onMount(() => {
		const target = new Date('2026-10-01T09:00:00').getTime();
		function tick() {
			const diff = Math.max(0, target - Date.now());
			countdown = {
				days: pad(Math.floor(diff / 86400000)),
				hours: pad(Math.floor((diff % 86400000) / 3600000)),
				mins: pad(Math.floor((diff % 3600000) / 60000)),
				secs: pad(Math.floor((diff % 60000) / 1000))
			};
		}
		tick();
		const id = setInterval(tick, 1000);
		return () => clearInterval(id);
	});
</script>

<section class="hero wrap">
	<div class="hero-left">
		<h1 class="hero-title reveal">
			Todo tu patrimonio,<br /><em>en tu mapa.</em>
		</h1>
		<p class="hero-sub reveal">
			Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú imaginas,
			aunque estén en distintas plataformas. Sin conectar cuentas, sin dar acceso a nadie.
		</p>

		<div class="countdown reveal" aria-label="Cuenta regresiva para el lanzamiento">
			<div class="cd-cell">
				<div class="cd-num">{countdown.days}</div>
				<div class="cd-label">días</div>
			</div>
			<div class="cd-cell">
				<div class="cd-num">{countdown.hours}</div>
				<div class="cd-label">hrs</div>
			</div>
			<div class="cd-cell">
				<div class="cd-num">{countdown.mins}</div>
				<div class="cd-label">min</div>
			</div>
			<div class="cd-cell">
				<div class="cd-num">{countdown.secs}</div>
				<div class="cd-label">seg</div>
			</div>
		</div>

		<div class="waitlist reveal" id="waitlist">
			{#if waitlistSuccess}
				<div class="wl-success">
					<span class="check-ico" aria-hidden="true">
						<svg
							width="12"
							height="12"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="3.2"
							stroke-linecap="round"
							stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg
						>
					</span>
					<span>¡Listo! Te avisaremos en cuanto Finexia esté disponible.</span>
				</div>
			{:else}
				<form
					class="wl-form"
					class:error={waitlistError}
					method="POST"
					action="/api/waitlist"
					onsubmit={submitWaitlist}
					novalidate
				>
					<input
						type="email"
						bind:value={waitlistEmail}
						placeholder="tu@email.com"
						autocomplete="email"
						name="email"
						required
						aria-label="Correo electrónico"
					/>
					<button type="submit" class="btn-amber" disabled={submitting}>
						{submitting ? 'Enviando...' : 'Acceso anticipado'}
					</button>
				</form>
				{#if waitlistError && waitlistErrorMessage}
					<p class="wl-error" role="alert">{waitlistErrorMessage}</p>
				{/if}
				<div class="wl-note">
					<svg
						width="13"
						height="13"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						><rect x="3" y="11" width="18" height="11" rx="2" /><path
							d="M7 11V7a5 5 0 0 1 10 0v4"
						/></svg
					>
					Sin spam. Solo te escribimos el día del lanzamiento.
				</div>
			{/if}
		</div>
	</div>

	<div class="hero-right" aria-hidden="true">
		<div class="hero-card reveal">
			<div class="hc-top">
				<div>
					<div class="hc-eyebrow">Portafolio</div>
					<div class="hc-name">Principal</div>
				</div>
				<div class="hc-status">Activo</div>
			</div>
			<div class="hc-value">$248.500</div>
			<div class="hc-delta">+12,4% este año</div>
			<div class="hc-bars">
				<div class="hc-row">
					<span class="hc-asset">Acciones</span>
					<div class="hc-bar">
						<div class="hc-fill" style="width:38%;--c:var(--amber)"></div>
					</div>
					<span class="hc-pct">38%</span>
				</div>
				<div class="hc-row">
					<span class="hc-asset">Crypto</span>
					<div class="hc-bar">
						<div class="hc-fill" style="width:29%;--c:var(--green)"></div>
					</div>
					<span class="hc-pct">29%</span>
				</div>
				<div class="hc-row">
					<span class="hc-asset">ETFs</span>
					<div class="hc-bar">
						<div class="hc-fill" style="width:23%;--c:#6b8cef"></div>
					</div>
					<span class="hc-pct">23%</span>
				</div>
				<div class="hc-row">
					<span class="hc-asset">Otros</span>
					<div class="hc-bar">
						<div class="hc-fill" style="width:10%;--c:#484542"></div>
					</div>
					<span class="hc-pct">10%</span>
				</div>
			</div>
			<div class="hc-platforms">
				<span>Binance</span><span>Degiro</span><span>Revolut</span><span>+2</span>
			</div>
		</div>

		<div class="hero-card-sm reveal">
			<div class="hcs-label">Portafolio Cripto</div>
			<div class="hcs-row">
				<span class="hcs-sym">BTC</span>
				<span class="hcs-val">$14.200</span>
				<span class="hcs-up">+4,2%</span>
			</div>
			<div class="hcs-row">
				<span class="hcs-sym">ETH</span>
				<span class="hcs-val">$3.480</span>
				<span class="hcs-dn">−1,1%</span>
			</div>
			<div class="hcs-row">
				<span class="hcs-sym">SOL</span>
				<span class="hcs-val">$1.820</span>
				<span class="hcs-up">+2,7%</span>
			</div>
		</div>
	</div>
</section>

<style>
	.hero {
		display: grid;
		grid-template-columns: 55fr 45fr;
		gap: 56px;
		align-items: center;
		padding: 84px 0 92px;
	}
	.hero-left {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
	}
	.hero-title {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(42px, 5.6vw, 76px);
		line-height: 1.06;
		letter-spacing: -0.022em;
		margin: 28px 0 0;
	}
	.hero-title em {
		font-style: italic;
		font-weight: 500;
		color: var(--amber-light);
	}
	.hero-sub {
		margin: 22px 0 0;
		max-width: 46ch;
		font-size: clamp(15px, 1.6vw, 17px);
		color: var(--text-muted);
		font-weight: 300;
		line-height: 1.7;
	}

	/* COUNTDOWN */
	.countdown {
		display: flex;
		margin-top: 40px;
		border: 1px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
		background: var(--surface);
	}
	.cd-cell {
		padding: 18px 22px;
		text-align: center;
		border-right: 1px solid var(--border);
		flex: 1;
	}
	.cd-cell:last-child {
		border-right: none;
	}
	.cd-num {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: clamp(28px, 4vw, 42px);
		line-height: 1;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}
	.cd-label {
		margin-top: 8px;
		font-size: 10px;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
		font-family: var(--font-mono);
	}

	/* WAITLIST */
	.waitlist {
		margin-top: 36px;
		width: 100%;
		max-width: 480px;
	}
	.wl-form {
		display: flex;
		gap: 8px;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		padding: 6px;
		transition:
			border-color 0.2s,
			box-shadow 0.2s;
	}
	.wl-form:focus-within {
		border-color: rgba(212, 145, 42, 0.5);
		box-shadow: 0 0 0 3px rgba(212, 145, 42, 0.07);
	}
	.wl-form.error {
		border-color: #e05a5a;
	}
	.wl-form input {
		flex: 1;
		background: transparent;
		border: none;
		outline: none;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 15px;
		padding: 0 14px;
		min-width: 0;
	}
	.wl-form input::placeholder {
		color: var(--text-dim);
	}
	.wl-note {
		margin-top: 12px;
		font-size: 12px;
		color: var(--text-dim);
		display: flex;
		align-items: center;
		gap: 7px;
	}
	.wl-error {
		margin-top: 10px;
		font-size: 12.5px;
		color: #e05a5a;
	}
	.wl-success {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 20px;
		border-radius: 8px;
		background: rgba(34, 201, 126, 0.06);
		border: 1px solid rgba(34, 201, 126, 0.22);
		font-size: 15px;
		font-weight: 300;
	}
	.check-ico {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--green);
		color: #06140c;
	}

	/* HERO RIGHT — MOCK CARDS */
	.hero-right {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.hero-card {
		width: 100%;
		padding: 26px 24px 20px;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.038);
		border: 1px solid var(--border-strong);
		backdrop-filter: blur(10px);
	}
	.hc-top {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 20px;
	}
	.hc-eyebrow {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 4px;
	}
	.hc-name {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 20px;
		letter-spacing: -0.01em;
	}
	.hc-status {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--green);
		background: rgba(34, 201, 126, 0.09);
		border: 1px solid rgba(34, 201, 126, 0.2);
		padding: 4px 9px;
		border-radius: 4px;
	}
	.hc-value {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 30px;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
	}
	.hc-delta {
		font-size: 13px;
		color: var(--green);
		margin-top: 4px;
		margin-bottom: 20px;
		font-weight: 400;
	}
	.hc-bars {
		display: flex;
		flex-direction: column;
		gap: 11px;
	}
	.hc-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.hc-asset {
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-muted);
		width: 60px;
		flex-shrink: 0;
	}
	.hc-bar {
		flex: 1;
		height: 3px;
		background: rgba(255, 255, 255, 0.07);
		border-radius: 2px;
		overflow: hidden;
	}
	.hc-fill {
		height: 100%;
		border-radius: 2px;
		background: var(--c);
	}
	.hc-pct {
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-dim);
		width: 30px;
		text-align: right;
		flex-shrink: 0;
	}
	.hc-platforms {
		display: flex;
		gap: 6px;
		margin-top: 18px;
		flex-wrap: wrap;
	}
	.hc-platforms span {
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--text-dim);
		background: var(--surface);
		border: 1px solid var(--border);
		padding: 3px 9px;
		border-radius: 4px;
	}
	.hero-card-sm {
		width: 100%;
		padding: 14px 18px;
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid var(--border);
	}
	.hcs-label {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 10px;
	}
	.hcs-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 0;
		border-top: 1px solid var(--border);
		font-family: var(--font-mono);
		font-size: 12.5px;
	}
	.hcs-sym {
		color: var(--text-muted);
		width: 36px;
		flex-shrink: 0;
	}
	.hcs-val {
		color: var(--text);
		font-weight: 600;
		flex: 1;
		font-variant-numeric: tabular-nums;
	}
	.hcs-up {
		color: var(--green);
	}
	.hcs-dn {
		color: #e05a5a;
	}

	@media (max-width: 940px) {
		.hero {
			grid-template-columns: 1fr;
			gap: 52px;
		}
		.hero-right {
			flex-direction: row;
			align-items: flex-start;
		}
		.hero-card {
			flex: 1;
		}
		.hero-card-sm {
			width: 200px;
			flex-shrink: 0;
		}
	}
	@media (max-width: 640px) {
		.hero-right {
			flex-direction: column;
		}
		.hero-card-sm {
			width: 100%;
		}
		.wl-form {
			flex-direction: column;
		}
		.countdown {
			flex-wrap: wrap;
		}
	}
</style>
