<script lang="ts">
	import { onMount } from 'svelte';

	let openFaqIndex = $state<number | null>(null);
	let waitlistEmail = $state('');
	let waitlistSuccess = $state(false);
	let waitlistError = $state(false);

	let cdDays: HTMLElement;
	let cdHours: HTMLElement;
	let cdMins: HTMLElement;
	let cdSecs: HTMLElement;
	let metricsEl: HTMLElement;
	let headerEl: HTMLElement;

	function pad(n: number) {
		return String(n).padStart(2, '0');
	}

	onMount(() => {
		const target = new Date('2026-10-01T09:00:00').getTime();
		function tick() {
			const diff = Math.max(0, target - Date.now());
			cdDays.textContent = pad(Math.floor(diff / 86400000));
			cdHours.textContent = pad(Math.floor((diff % 86400000) / 3600000));
			cdMins.textContent = pad(Math.floor((diff % 3600000) / 60000));
			cdSecs.textContent = pad(Math.floor((diff % 60000) / 1000));
		}
		tick();
		const countdownInterval = setInterval(tick, 1000);

		const revealIo = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						entry.target.classList.add('in');
						revealIo.unobserve(entry.target);
					}
				});
			},
			{ threshold: 0.12 }
		);
		document.querySelectorAll('.reveal').forEach((el, i) => {
			(el as HTMLElement).style.transitionDelay = Math.min(i * 60, 240) + 'ms';
			revealIo.observe(el);
		});

		const metricsIo = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						metricsEl.classList.add('in');
						metricsIo.unobserve(entry.target);
					}
				});
			},
			{ threshold: 0.15 }
		);
		if (metricsEl) metricsIo.observe(metricsEl);

		const scrollHandler = () => {
			headerEl.style.boxShadow =
				window.scrollY > 10 ? '0 1px 0 rgba(255,255,255,0.04)' : 'none';
		};
		window.addEventListener('scroll', scrollHandler);

		return () => {
			clearInterval(countdownInterval);
			revealIo.disconnect();
			metricsIo.disconnect();
			window.removeEventListener('scroll', scrollHandler);
		};
	});

	function submitWaitlist(e: Event) {
		e.preventDefault();
		const val = waitlistEmail.trim();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val)) {
			waitlistError = true;
			setTimeout(() => (waitlistError = false), 1200);
			return;
		}
		try {
			localStorage.setItem('finexia_waitlist', val);
		} catch {}
		waitlistSuccess = true;
	}

	function toggleFaq(index: number) {
		openFaqIndex = openFaqIndex === index ? null : index;
	}

	const faqs = [
		{
			q: '¿Cuándo estará disponible Finexia?',
			a: 'El lanzamiento oficial está previsto para el 1 de octubre de 2026. Quienes se unan a la lista de espera recibirán acceso anticipado antes que el público general.'
		},
		{
			q: '¿Tiene algún costo unirse ahora?',
			a: 'No. Unirse a la lista de espera es totalmente gratuito y sin compromiso. Solo necesitamos tu correo para avisarte el día del lanzamiento.'
		},
		{
			q: '¿Finexia se conecta a mis brokers o plataformas?',
			a: 'No. Finexia nunca accede a tus plataformas ni te pide credenciales. Tú registras manualmente dónde tienes tus activos, así que la información siempre está bajo tu control.'
		},
		{
			q: '¿Puedo agrupar activos de varias plataformas en un mismo portafolio?',
			a: 'Sí, esa es la idea. Creas los portafolios que tienes en mente y añades a cada uno los activos que quieras, sin importar en qué broker o plataforma estén registrados.'
		},
		{
			q: '¿Habrá conexión automática con plataformas en el futuro?',
			a: 'Es algo que estamos contemplando para más adelante. Por ahora todo el registro es manual para que tú mantengas el control, pero tenemos en mente ofrecer en el futuro una zona centralizada y automatizada que conecte con plataformas. Lo evaluaremos próximamente según las condiciones técnicas y de seguridad que permitan hacerlo bien.'
		}
	];
</script>

<svelte:head>
	<title>Finexia — Próximamente</title>
	<meta
		name="description"
		content="Registra tus activos manualmente y agrúpalos en los portafolios que tú imaginas. Sin conectar cuentas. Lanzamiento 1 oct 2026."
	/>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="" />
	<link
		rel="stylesheet"
		href="https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,300;0,9..144,500;0,9..144,600;1,9..144,300;1,9..144,500&family=JetBrains+Mono:wght@400;600&family=Outfit:wght@300;400;500;600&display=swap"
	/>
</svelte:head>

<div class="scanlines" aria-hidden="true"></div>

<!-- Scrolling market ticker -->
<div class="ticker-banner" aria-hidden="true">
	<div class="ticker-track">
		<span>BTC <em>$67.240</em> <b class="up">+2,1%</b></span>
		<span class="sep">·</span>
		<span>ETH <em>$3.482</em> <b class="dn">−0,4%</b></span>
		<span class="sep">·</span>
		<span>AAPL <em>$189,45</em> <b class="up">+0,8%</b></span>
		<span class="sep">·</span>
		<span>S&P500 <em>5.234</em> <b class="up">+0,6%</b></span>
		<span class="sep">·</span>
		<span>EUR/USD <em>1,0842</em> <b class="up">+0,1%</b></span>
		<span class="sep">·</span>
		<span>ORO <em>$2.387</em> <b class="up">+0,3%</b></span>
		<span class="sep">·</span>
		<span>MSFT <em>$415</em> <b class="up">+1,2%</b></span>
		<span class="sep">·</span>
		<span>NVDA <em>$876</em> <b class="up">+3,8%</b></span>
		<span class="sep">·</span>
		<!-- duplicate for seamless loop -->
		<span>BTC <em>$67.240</em> <b class="up">+2,1%</b></span>
		<span class="sep">·</span>
		<span>ETH <em>$3.482</em> <b class="dn">−0,4%</b></span>
		<span class="sep">·</span>
		<span>AAPL <em>$189,45</em> <b class="up">+0,8%</b></span>
		<span class="sep">·</span>
		<span>S&P500 <em>5.234</em> <b class="up">+0,6%</b></span>
		<span class="sep">·</span>
		<span>EUR/USD <em>1,0842</em> <b class="up">+0,1%</b></span>
		<span class="sep">·</span>
		<span>ORO <em>$2.387</em> <b class="up">+0,3%</b></span>
		<span class="sep">·</span>
		<span>MSFT <em>$415</em> <b class="up">+1,2%</b></span>
		<span class="sep">·</span>
		<span>NVDA <em>$876</em> <b class="up">+3,8%</b></span>
		<span class="sep">·</span>
	</div>
</div>

<header bind:this={headerEl}>
	<div class="wrap nav">
		<div class="brand">
			<svg class="brand-icon" width="30" height="30" viewBox="0 0 30 30" fill="none">
				<rect width="30" height="30" rx="7" fill="var(--amber)" />
				<path
					d="M7 22L12.5 14.5L16.5 18.5L23 9"
					stroke="#0c0a06"
					stroke-width="2.6"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
			<span class="brand-name">FINEXIA</span>
		</div>
		<nav class="nav-links">
			<a href="#beneficios">Beneficios</a>
			<a href="#como-funciona">Cómo funciona</a>
			<a href="#faq">Preguntas</a>
		</nav>
		<a href="#waitlist" class="nav-cta">Unirme a la lista</a>
	</div>
</header>

<main>
	<!-- HERO -->
	<section class="hero wrap">
		<div class="hero-left">
			<h1 class="hero-title reveal">
				Todo tu patrimonio,<br /><em>en tu mapa.</em>
			</h1>
			<p class="hero-sub reveal">
				Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que
				tú imaginas, aunque estén en distintas plataformas. Sin conectar cuentas, sin dar
				acceso a nadie.
			</p>

			<div class="countdown reveal" aria-label="Cuenta regresiva para el lanzamiento">
				<div class="cd-cell">
					<div class="cd-num" bind:this={cdDays}>00</div>
					<div class="cd-label">días</div>
				</div>
				<div class="cd-cell">
					<div class="cd-num" bind:this={cdHours}>00</div>
					<div class="cd-label">hrs</div>
				</div>
				<div class="cd-cell">
					<div class="cd-num" bind:this={cdMins}>00</div>
					<div class="cd-label">min</div>
				</div>
				<div class="cd-cell">
					<div class="cd-num" bind:this={cdSecs}>00</div>
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
								stroke-linejoin="round"
								><path d="M20 6 9 17l-5-5" /></svg
							>
						</span>
						<span>¡Listo! Te avisaremos en cuanto Finexia esté disponible.</span>
					</div>
				{:else}
					<form class="wl-form" class:error={waitlistError} onsubmit={submitWaitlist} novalidate>
						<input
							type="email"
							bind:value={waitlistEmail}
							placeholder="tu@email.com"
							autocomplete="email"
							required
							aria-label="Correo electrónico"
						/>
						<button type="submit" class="btn-amber">Acceso anticipado</button>
					</form>
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

	<div class="wrap"><div class="divider"></div></div>

	<!-- BENEFITS -->
	<section class="block wrap" id="beneficios">
		<div class="sec-head reveal">
			<div class="eyebrow">Por qué Finexia</div>
			<h2 class="sec-title">Tu patrimonio organizado<br />a tu manera</h2>
		</div>
		<div class="benefits-list">
			<div class="bitem reveal">
				<div class="bnum">01</div>
				<div class="bcontent">
					<h3>Registro manual, total control</h3>
					<p>
						Anota tus brokers, exchanges, bancos o plataformas y los activos que tienes en cada
						uno. Tú decides qué incluir y cómo nombrarlo.
					</p>
				</div>
			</div>
			<div class="bitem reveal">
				<div class="bnum">02</div>
				<div class="bcontent">
					<h3>Portafolios a tu medida</h3>
					<p>
						Agrupa activos en los portafolios que tienes en mente, aunque vivan en plataformas
						distintas. Tú trazas la estructura.
					</p>
				</div>
			</div>
			<div class="bitem reveal">
				<div class="bnum">03</div>
				<div class="bcontent">
					<h3>Cero acceso a tus cuentas</h3>
					<p>
						Finexia nunca se conecta a tus plataformas ni pide credenciales. La información es tuya
						y solo tú la editas.
					</p>
				</div>
			</div>
			<div class="bitem reveal">
				<div class="bnum">04</div>
				<div class="bcontent">
					<h3>Visión de un vistazo</h3>
					<p>
						Distribución y peso de cada portafolio claros al instante, sin hojas de cálculo ni
						cuentas manuales.
					</p>
				</div>
			</div>
		</div>
	</section>

	<div class="wrap"><div class="divider"></div></div>

	<!-- HOW IT WORKS -->
	<section class="block wrap" id="como-funciona">
		<div class="sec-head reveal">
			<div class="eyebrow">Cómo funciona</div>
			<h2 class="sec-title">De plataformas dispersas<br />a un solo mapa</h2>
		</div>
		<div class="steps">
			<div class="step reveal">
				<div class="step-num">01</div>
				<div class="step-body">
					<h3>Registra tus plataformas</h3>
					<p>
						Añade manualmente cada broker, exchange o banco donde tienes activos y anota qué tienes
						en cada uno. Sin conectar cuentas.
					</p>
				</div>
			</div>
			<div class="step reveal">
				<div class="step-num">02</div>
				<div class="step-body">
					<h3>Crea tus portafolios</h3>
					<p>
						Agrupa los activos según los portafolios que tienes en mente, combinando lo que está en
						distintas plataformas.
					</p>
				</div>
			</div>
			<div class="step reveal">
				<div class="step-num">03</div>
				<div class="step-body">
					<h3>Visualiza el conjunto</h3>
					<p>
						Ve la distribución y el peso de cada portafolio en una sola vista, y decide con tu
						patrimonio completo a la vista.
					</p>
				</div>
			</div>
		</div>
	</section>

	<div class="wrap"><div class="divider"></div></div>

	<!-- METRICS SHOWCASE -->
	<section class="block wrap">
		<div class="sec-head reveal">
			<div class="eyebrow">Métricas que importan</div>
			<h2 class="sec-title">Mira crecer tu patrimonio<br />con claridad</h2>
			<p class="sec-desc">
				Valor, peso, rendimiento y distribución de cada portafolio, en una vista hecha para entender
				de un vistazo.
			</p>
		</div>

		<div class="metrics-wrap reveal" bind:this={metricsEl}>
			<div class="metrics-top">
				<div class="metrics-headline">
					<div class="eyebrow" style="color: var(--green); text-align:left">Patrimonio total</div>
					<h2>$248.500 <span class="growth-badge">+12,4%</span></h2>
					<p>
						Sigue la evolución de tu patrimonio agregado y el peso de cada activo dentro de tus
						portafolios.
					</p>
				</div>
				<div class="metric-chips">
					<div class="mchip">
						<div class="lbl">Valor</div>
						<div class="val">$248,5K</div>
					</div>
					<div class="mchip">
						<div class="lbl">Rendimiento</div>
						<div class="val">+12,4%<span class="delta up">▲</span></div>
					</div>
					<div class="mchip">
						<div class="lbl">Peso mayor</div>
						<div class="val">38%<span class="delta amb">Acciones</span></div>
					</div>
				</div>
			</div>

			<div class="chart-zone">
				<svg
					class="chart-svg"
					viewBox="0 0 1000 280"
					preserveAspectRatio="none"
					aria-hidden="true"
				>
					<defs>
						<linearGradient id="growthFill" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="rgba(212,145,42,0.2)" />
							<stop offset="100%" stop-color="rgba(212,145,42,0)" />
						</linearGradient>
						<linearGradient id="growthStroke" x1="0" y1="0" x2="1" y2="0">
							<stop offset="0%" stop-color="#22c97e" />
							<stop offset="100%" stop-color="#d4912a" />
						</linearGradient>
					</defs>
					<g class="chart-grid">
						<line x1="0" y1="40" x2="1000" y2="40" />
						<line x1="0" y1="110" x2="1000" y2="110" />
						<line x1="0" y1="180" x2="1000" y2="180" />
						<line x1="0" y1="250" x2="1000" y2="250" />
					</g>
					<path
						class="chart-area"
						d="M0,230 C90,225 150,210 230,200 C320,188 360,170 450,150 C540,130 590,140 670,110 C760,78 820,72 900,52 C935,42 960,34 982,26 L982,280 L0,280 Z"
					/>
					<path
						class="chart-line"
						d="M0,230 C90,225 150,210 230,200 C320,188 360,170 450,150 C540,130 590,140 670,110 C760,78 820,72 900,52 C935,42 960,34 982,26"
					/>
					<g class="chart-dot">
						<circle class="pulse-ring" cx="982" cy="26" r="6" fill="rgba(212,145,42,0.35)" />
						<circle cx="982" cy="26" r="5" fill="var(--amber)" stroke="#08090a" stroke-width="2.5" />
					</g>
				</svg>
				<div class="chart-flag">
					<div class="t">Este año</div>
					<div class="v">+ $27.400</div>
				</div>
			</div>
		</div>
	</section>

	<div class="wrap"><div class="divider"></div></div>

	<!-- FAQ -->
	<section class="block wrap" id="faq">
		<div class="sec-head reveal">
			<div class="eyebrow">Preguntas frecuentes</div>
			<h2 class="sec-title">Lo que necesitas saber</h2>
		</div>
		<div class="faq">
			{#each faqs as faq, i}
				<div class="faq-item reveal" class:open={openFaqIndex === i}>
					<button class="faq-q" onclick={() => toggleFaq(i)}>
						{faq.q}<span class="plus" aria-hidden="true"></span>
					</button>
					<div class="faq-a"><p>{faq.a}</p></div>
				</div>
			{/each}
		</div>
	</section>

	<!-- FINAL CTA -->
	<section class="block wrap">
		<div class="final-cta reveal">
			<div class="cta-eyebrow">Lista de espera</div>
			<h2>Sé de los primeros<br />en tener el control.</h2>
			<p>Cupos de acceso anticipado limitados. Reserva tu lugar antes del lanzamiento.</p>
			<a href="#waitlist" class="btn-amber btn-amber-lg">Unirme a la lista de espera</a>
		</div>
	</section>
</main>

<footer>
	<div class="wrap foot">
		<div class="brand">
			<svg class="brand-icon" width="26" height="26" viewBox="0 0 30 30" fill="none">
				<rect width="30" height="30" rx="7" fill="var(--amber)" />
				<path
					d="M7 22L12.5 14.5L16.5 18.5L23 9"
					stroke="#0c0a06"
					stroke-width="2.6"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
			<span class="brand-name" style="font-size:18px">FINEXIA</span>
		</div>
		<nav class="foot-links">
			<a href="#beneficios">Beneficios</a>
			<a href="#como-funciona">Cómo funciona</a>
			<a href="#faq">Preguntas</a>
			<a href="#waitlist">Lista de espera</a>
		</nav>
		<div class="foot-copy">© 2026 Finexia. Todos los derechos reservados.</div>
	</div>
</footer>

<style>
	:global(*) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}
	:global(html) {
		scroll-behavior: smooth;
	}
	:global(body) {
		background: #08090a;
		color: #eceae5;
		font-family: 'Outfit', system-ui, sans-serif;
		line-height: 1.55;
		-webkit-font-smoothing: antialiased;
		overflow-x: hidden;
	}
	:global(::selection) {
		background: rgba(212, 145, 42, 0.25);
		color: #fff;
	}
	:global(a) {
		color: inherit;
		text-decoration: none;
	}

	:global(:root) {
		--bg: #08090a;
		--surface: rgba(255, 255, 255, 0.028);
		--surface-2: rgba(255, 255, 255, 0.05);
		--border: rgba(255, 255, 255, 0.07);
		--border-strong: rgba(255, 255, 255, 0.12);
		--text: #eceae5;
		--text-muted: #8a8780;
		--text-dim: #555250;
		--amber: #d4912a;
		--amber-light: #e8a535;
		--green: #22c97e;
		--maxw: 1200px;
		--font-display: 'Fraunces', Georgia, serif;
		--font-mono: 'JetBrains Mono', 'Courier New', monospace;
		--font-body: 'Outfit', system-ui, sans-serif;
	}

	.wrap {
		max-width: var(--maxw);
		margin: 0 auto;
		padding: 0 36px;
	}

	/* SUBTLE SCANLINE TEXTURE */
	.scanlines {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
		background: repeating-linear-gradient(
			0deg,
			transparent,
			transparent 3px,
			rgba(255, 255, 255, 0.006) 3px,
			rgba(255, 255, 255, 0.006) 4px
		);
	}

	main,
	header,
	footer {
		position: relative;
		z-index: 1;
	}

	/* TICKER */
	.ticker-banner {
		position: relative;
		z-index: 10;
		background: rgba(255, 255, 255, 0.018);
		border-bottom: 1px solid var(--border);
		overflow: hidden;
		padding: 7px 0;
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.03em;
		color: var(--text-muted);
	}
	.ticker-track {
		display: flex;
		gap: 32px;
		align-items: center;
		white-space: nowrap;
		width: max-content;
		animation: ticker-scroll 40s linear infinite;
	}
	.ticker-track em {
		font-style: normal;
		color: var(--text);
	}
	.ticker-track b {
		font-weight: 600;
	}
	.ticker-track .up {
		color: var(--green);
	}
	.ticker-track .dn {
		color: #e05a5a;
	}
	.ticker-track .sep {
		color: var(--text-dim);
		opacity: 0.5;
	}
	@keyframes ticker-scroll {
		0% {
			transform: translateX(0);
		}
		100% {
			transform: translateX(-50%);
		}
	}

	/* NAV */
	header {
		position: sticky;
		top: 0;
		z-index: 50;
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		background: rgba(8, 9, 10, 0.82);
		border-bottom: 1px solid var(--border);
	}
	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 66px;
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.brand-icon {
		flex-shrink: 0;
	}
	.brand-name {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 20px;
		letter-spacing: 0.1em;
		color: var(--text);
	}
	.nav-links {
		display: flex;
		align-items: center;
		gap: 36px;
	}
	.nav-links a {
		font-size: 14px;
		color: var(--text-muted);
		font-weight: 400;
		transition: color 0.2s;
	}
	.nav-links a:hover {
		color: var(--text);
	}
	.nav-cta {
		display: inline-flex;
		align-items: center;
		padding: 9px 18px;
		border-radius: 6px;
		border: 1px solid var(--border-strong);
		font-size: 13.5px;
		font-weight: 500;
		color: var(--text);
		transition:
			border-color 0.2s,
			background 0.2s;
	}
	.nav-cta:hover {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.06);
	}
	@media (max-width: 860px) {
		.nav-links {
			display: none;
		}
	}

	/* HERO */
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
	.btn-amber {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 12px 22px;
		border-radius: 6px;
		border: none;
		cursor: pointer;
		font-family: var(--font-body);
		font-size: 14px;
		font-weight: 600;
		color: #0d0800;
		background: var(--amber);
		transition:
			background 0.2s,
			transform 0.15s;
	}
	.btn-amber:hover {
		background: var(--amber-light);
		transform: translateY(-1px);
	}
	.btn-amber:active {
		transform: none;
	}
	.btn-amber-lg {
		display: inline-flex;
		padding: 15px 30px;
		font-size: 15.5px;
		border-radius: 8px;
	}
	.wl-note {
		margin-top: 12px;
		font-size: 12px;
		color: var(--text-dim);
		display: flex;
		align-items: center;
		gap: 7px;
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

	/* SECTIONS */
	.block {
		padding: 88px 0;
	}
	.sec-head {
		max-width: 640px;
		margin: 0 auto 56px;
		text-align: center;
	}
	.eyebrow {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 500;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--amber);
	}
	.sec-title {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(28px, 4vw, 46px);
		line-height: 1.1;
		letter-spacing: -0.02em;
		margin-top: 14px;
		text-wrap: balance;
	}
	.sec-desc {
		margin-top: 16px;
		font-size: 16px;
		color: var(--text-muted);
		font-weight: 300;
	}
	.divider {
		height: 1px;
		background: linear-gradient(90deg, transparent, var(--border), transparent);
	}

	/* BENEFITS — editorial numbered grid */
	.benefits-list {
		display: grid;
		grid-template-columns: 1fr 1fr;
		border: 1px solid var(--border);
		border-radius: 12px;
		overflow: hidden;
	}
	.bitem {
		display: flex;
		gap: 22px;
		padding: 36px 30px;
		border-right: 1px solid var(--border);
		border-bottom: 1px solid var(--border);
		transition: background 0.2s;
	}
	.bitem:hover {
		background: var(--surface);
	}
	.bitem:nth-child(2n) {
		border-right: none;
	}
	.bitem:nth-last-child(-n + 2) {
		border-bottom: none;
	}
	.bnum {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--text-dim);
		letter-spacing: 0.04em;
		flex-shrink: 0;
		padding-top: 5px;
	}
	.bcontent h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 20px;
		letter-spacing: -0.01em;
		margin-bottom: 10px;
	}
	.bcontent p {
		font-size: 14px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
	}
	@media (max-width: 700px) {
		.benefits-list {
			grid-template-columns: 1fr;
		}
		.bitem {
			border-right: none;
		}
		.bitem:last-child {
			border-bottom: none;
		}
		.bitem:nth-child(n) {
			border-bottom: 1px solid var(--border);
		}
	}

	/* HOW IT WORKS — timeline */
	.steps {
		display: flex;
		flex-direction: column;
		max-width: 660px;
		margin: 0 auto;
	}
	.step {
		display: flex;
		gap: 32px;
		padding: 32px 0;
		border-bottom: 1px solid var(--border);
	}
	.step:last-child {
		border-bottom: none;
	}
	.step-num {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		color: var(--amber);
		letter-spacing: 0.06em;
		flex-shrink: 0;
		padding-top: 5px;
		min-width: 26px;
	}
	.step-body h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 21px;
		letter-spacing: -0.01em;
		margin-bottom: 10px;
	}
	.step-body p {
		font-size: 14.5px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
	}

	/* METRICS */
	.metrics-wrap {
		position: relative;
		border-radius: 14px;
		overflow: hidden;
		border: 1px solid var(--border-strong);
		background: var(--surface);
		padding: 44px 44px 0;
	}
	.metrics-top {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 28px;
		flex-wrap: wrap;
		position: relative;
		z-index: 2;
	}
	.metrics-headline {
		max-width: 30ch;
	}
	.metrics-headline h2 {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: clamp(24px, 3.2vw, 36px);
		line-height: 1.06;
		letter-spacing: -0.02em;
		margin-top: 10px;
		font-variant-numeric: tabular-nums;
	}
	.growth-badge {
		color: var(--green);
		font-size: 0.52em;
		font-weight: 600;
		letter-spacing: 0;
	}
	.metrics-headline p {
		margin-top: 12px;
		font-size: 14.5px;
		color: var(--text-muted);
		font-weight: 300;
	}
	.metric-chips {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.mchip {
		padding: 14px 18px;
		border-radius: 8px;
		min-width: 118px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
	}
	.mchip .lbl {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 500;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
	}
	.mchip .val {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 21px;
		letter-spacing: -0.02em;
		margin-top: 6px;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}
	.delta {
		font-size: 12px;
		font-weight: 600;
		margin-left: 6px;
	}
	.delta.up {
		color: var(--green);
	}
	.delta.amb {
		color: var(--amber);
	}

	/* CHART */
	.chart-zone {
		position: relative;
		margin-top: 32px;
		height: 260px;
	}
	.chart-svg {
		display: block;
		width: 100%;
		height: 100%;
	}
	.chart-grid line {
		stroke: rgba(255, 255, 255, 0.04);
		stroke-width: 1;
	}
	.chart-area {
		fill: url(#growthFill);
		opacity: 0;
		transition: opacity 1.2s ease 0.5s;
	}
	:global(.metrics-wrap.in) .chart-area {
		opacity: 1;
	}
	.chart-line {
		fill: none;
		stroke: url(#growthStroke);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-dasharray: 1400;
		stroke-dashoffset: 1400;
		transition: stroke-dashoffset 1.8s cubic-bezier(0.4, 0, 0.2, 1) 0.2s;
	}
	:global(.metrics-wrap.in) .chart-line {
		stroke-dashoffset: 0;
	}
	.chart-dot {
		opacity: 0;
		transition: opacity 0.4s ease 1.6s;
	}
	:global(.metrics-wrap.in) .chart-dot {
		opacity: 1;
	}
	.pulse-ring {
		transform-box: fill-box;
		transform-origin: center;
		animation: dotpulse 2.2s ease-out infinite;
	}
	@keyframes dotpulse {
		0% {
			transform: scale(0.6);
			opacity: 0.7;
		}
		70% {
			transform: scale(2.4);
			opacity: 0;
		}
		100% {
			opacity: 0;
		}
	}
	.chart-flag {
		position: absolute;
		top: 8%;
		right: 4%;
		padding: 8px 12px;
		border-radius: 8px;
		background: rgba(8, 9, 10, 0.88);
		border: 1px solid rgba(34, 201, 126, 0.28);
		backdrop-filter: blur(8px);
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.5s ease 1.5s,
			transform 0.5s ease 1.5s;
	}
	:global(.metrics-wrap.in) .chart-flag {
		opacity: 1;
		transform: none;
	}
	.chart-flag .t {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-dim);
	}
	.chart-flag .v {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 17px;
		color: var(--green);
		margin-top: 3px;
	}
	@media (max-width: 620px) {
		.metrics-wrap {
			padding: 28px 20px 0;
		}
		.chart-zone {
			height: 200px;
		}
		.mchip {
			min-width: 0;
			flex: 1;
			padding: 12px 14px;
		}
	}

	/* FAQ */
	.faq {
		max-width: 720px;
		margin: 0 auto;
	}
	.faq-item {
		border-bottom: 1px solid var(--border);
	}
	.faq-q {
		width: 100%;
		background: none;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		padding: 24px 0;
		text-align: left;
		font-family: var(--font-display);
		font-weight: 300;
		font-size: 18px;
		color: var(--text);
	}
	.plus {
		flex-shrink: 0;
		width: 22px;
		height: 22px;
		position: relative;
		transition: transform 0.3s ease;
	}
	.plus::before,
	.plus::after {
		content: '';
		position: absolute;
		background: var(--amber);
		border-radius: 1px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
	}
	.plus::before {
		width: 12px;
		height: 1.5px;
	}
	.plus::after {
		width: 1.5px;
		height: 12px;
		transition: opacity 0.3s ease;
	}
	.faq-item.open .plus {
		transform: rotate(90deg);
	}
	.faq-item.open .plus::after {
		opacity: 0;
	}
	.faq-a {
		max-height: 0;
		overflow: hidden;
		transition:
			max-height 0.35s ease,
			padding 0.35s ease;
	}
	.faq-a p {
		font-size: 15px;
		color: var(--text-muted);
		line-height: 1.68;
		font-weight: 300;
		padding: 0 40px 4px 0;
	}
	.faq-item.open .faq-a {
		max-height: 280px;
		padding-bottom: 24px;
	}

	/* FINAL CTA */
	.final-cta {
		text-align: center;
		padding: 92px 40px;
		border-radius: 14px;
		position: relative;
		overflow: hidden;
		background: var(--surface);
		border: 1px solid var(--border-strong);
	}
	.final-cta::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 65% 110% at 50% -5%,
			rgba(212, 145, 42, 0.07),
			transparent 58%
		);
		pointer-events: none;
	}
	.cta-eyebrow {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 20px;
		position: relative;
	}
	.final-cta h2 {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(30px, 5vw, 58px);
		line-height: 1.06;
		letter-spacing: -0.025em;
		text-wrap: balance;
		position: relative;
	}
	.final-cta p {
		margin: 16px auto 0;
		max-width: 44ch;
		font-size: 16px;
		color: var(--text-muted);
		font-weight: 300;
		position: relative;
	}
	.final-cta .btn-amber-lg {
		margin-top: 32px;
		position: relative;
	}

	/* FOOTER */
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

	/* REVEAL */
	.reveal {
		opacity: 0;
		transform: translateY(16px);
		transition:
			opacity 0.65s ease,
			transform 0.65s ease;
	}
	:global(.reveal.in) {
		opacity: 1;
		transform: none;
	}
	@media (prefers-reduced-motion: reduce) {
		.reveal {
			opacity: 1;
			transform: none;
			transition: none;
		}
		.ticker-track {
			animation: none;
		}
	}
</style>
