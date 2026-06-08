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
		href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=Hanken+Grotesk:wght@400;500;600;700&display=swap"
	/>
</svelte:head>

<div class="atmosphere" aria-hidden="true">
	<div class="glow g1"></div>
	<div class="glow g2"></div>
	<div class="grid-overlay"></div>
</div>

<header bind:this={headerEl}>
	<div class="wrap nav">
		<div class="brand">
			<div class="brand-mark"></div>
			<div class="brand-name">FINEXIA</div>
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
		<div class="badge reveal"><span class="dot"></span> Próximamente · Lanzamiento 1 oct 2026</div>
		<h1 class="hero-title reveal">Todo tu patrimonio,<br /><span class="accent">en tu mapa.</span></h1>
		<p class="hero-sub reveal">
			Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú imaginas,
			aunque estén repartidos entre varias plataformas. Sin conectar cuentas, sin dar acceso a
			nadie. Tú controlas la información.
		</p>

		<div class="countdown reveal" aria-label="Cuenta regresiva para el lanzamiento">
			<div class="cd-cell"><div class="cd-num" bind:this={cdDays}>00</div><div class="cd-label">Días</div></div>
			<div class="cd-cell"><div class="cd-num" bind:this={cdHours}>00</div><div class="cd-label">Horas</div></div>
			<div class="cd-cell"><div class="cd-num" bind:this={cdMins}>00</div><div class="cd-label">Minutos</div></div>
			<div class="cd-cell"><div class="cd-num" bind:this={cdSecs}>00</div><div class="cd-label">Segundos</div></div>
		</div>

		<div class="waitlist reveal" id="waitlist">
			{#if waitlistSuccess}
				<div class="wl-success">
					<span class="check-ico" aria-hidden="true">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>
					</span>
					<span>¡Listo! Te avisaremos en cuanto Finexia esté disponible.</span>
				</div>
			{:else}
				<form
					class="wl-form"
					class:error={waitlistError}
					onsubmit={submitWaitlist}
					novalidate
				>
					<input
						type="email"
						bind:value={waitlistEmail}
						placeholder="tu@email.com"
						autocomplete="email"
						required
						aria-label="Correo electrónico"
					/>
					<button type="submit" class="btn-gold">Acceso anticipado</button>
				</form>
				<div class="wl-note">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
					Sin spam. Te escribiremos solo el día del lanzamiento.
				</div>
			{/if}
		</div>
	</section>

	<div class="wrap"><div class="divider"></div></div>

	<!-- BENEFITS -->
	<section class="block wrap" id="beneficios">
		<div class="sec-head reveal">
			<div class="eyebrow">Por qué Finexia</div>
			<h2 class="sec-title">Tu patrimonio organizado a tu manera</h2>
			<p class="sec-desc">
				Tú registras dónde está cada activo y cómo quieres agruparlo. Finexia nunca se conecta a
				tus plataformas: solo organiza lo que tú anotas.
			</p>
		</div>
		<div class="benefits-grid">
			<div class="bcard reveal">
				<div class="ico">
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" rx="1.5" /><rect x="14" y="3" width="7" height="7" rx="1.5" /><rect x="3" y="14" width="7" height="7" rx="1.5" /><rect x="14" y="14" width="7" height="7" rx="1.5" /></svg>
				</div>
				<h3>Registro manual</h3>
				<p>Anota tus brokers, exchanges, bancos o plataformas y los activos que tienes en cada uno. Tú decides qué incluir.</p>
			</div>
			<div class="bcard green reveal">
				<div class="ico">
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="m12 2 9 5-9 5-9-5 9-5Z" /><path d="m3 12 9 5 9-5" /><path d="m3 17 9 5 9-5" /></svg>
				</div>
				<h3>Portafolios a tu medida</h3>
				<p>Agrupa activos en los portafolios que tienes en mente, aunque vivan en plataformas distintas. Tú trazas la estructura.</p>
			</div>
			<div class="bcard reveal">
				<div class="ico">
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" /><path d="m9 12 2 2 4-4" /></svg>
				</div>
				<h3>Cero acceso a tus cuentas</h3>
				<p>Finexia nunca se conecta a tus plataformas ni pide credenciales. La información es tuya y solo tú la editas.</p>
			</div>
			<div class="bcard green reveal">
				<div class="ico">
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18" /><path d="M7 15l4-5 3 3 5-7" /></svg>
				</div>
				<h3>Visión de un vistazo</h3>
				<p>Distribución y peso de cada portafolio claros al instante, sin hojas de cálculo ni cuentas manuales.</p>
			</div>
		</div>
	</section>

	<div class="wrap"><div class="divider"></div></div>

	<!-- HOW IT WORKS -->
	<section class="block wrap" id="como-funciona">
		<div class="sec-head reveal">
			<div class="eyebrow">Cómo funciona</div>
			<h2 class="sec-title">De plataformas dispersas a un solo mapa</h2>
		</div>
		<div class="steps">
			<div class="step reveal">
				<div class="num">1</div>
				<div class="connector"></div>
				<h3>Registra tus plataformas</h3>
				<p>Añade manualmente cada broker, exchange o banco donde tienes activos y anota qué tienes en cada uno. Sin conectar cuentas.</p>
			</div>
			<div class="step reveal">
				<div class="num">2</div>
				<div class="connector"></div>
				<h3>Crea tus portafolios</h3>
				<p>Agrupa los activos según los portafolios que tienes en mente, combinando lo que está en distintas plataformas.</p>
			</div>
			<div class="step reveal">
				<div class="num">3</div>
				<h3>Visualiza el conjunto</h3>
				<p>Ve la distribución y el peso de cada portafolio en una sola vista, y decide con tu patrimonio completo a la vista.</p>
			</div>
		</div>
	</section>

	<div class="wrap"><div class="divider"></div></div>

	<!-- METRICS SHOWCASE -->
	<section class="block wrap">
		<div class="sec-head reveal">
			<div class="eyebrow">Métricas que importan</div>
			<h2 class="sec-title">Mira crecer tu patrimonio con claridad</h2>
			<p class="sec-desc">
				Valor, peso, rendimiento y distribución de cada portafolio, en una vista hecha para entender
				de un vistazo.
			</p>
		</div>

		<div class="metrics-wrap reveal" bind:this={metricsEl}>
			<div class="metrics-top">
				<div class="metrics-headline">
					<div class="eyebrow" style="color: var(--green)">Patrimonio total</div>
					<h2>$248.500 <span class="growth-badge">+12,4%</span></h2>
					<p>Sigue la evolución de tu patrimonio agregado y el peso de cada activo dentro de tus portafolios.</p>
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
						<div class="val">38%<span class="delta gold">Acciones</span></div>
					</div>
				</div>
			</div>

			<div class="chart-zone">
				<svg class="chart-svg" viewBox="0 0 1000 280" preserveAspectRatio="none" aria-hidden="true">
					<defs>
						<linearGradient id="growthFill" x1="0" y1="0" x2="0" y2="1">
							<stop offset="0%" stop-color="rgba(61,220,132,0.28)" />
							<stop offset="100%" stop-color="rgba(61,220,132,0)" />
						</linearGradient>
						<linearGradient id="growthStroke" x1="0" y1="0" x2="1" y2="0">
							<stop offset="0%" stop-color="#3ddc84" />
							<stop offset="100%" stop-color="#e6c43c" />
						</linearGradient>
					</defs>
					<g class="chart-grid">
						<line x1="0" y1="40" x2="1000" y2="40" />
						<line x1="0" y1="110" x2="1000" y2="110" />
						<line x1="0" y1="180" x2="1000" y2="180" />
						<line x1="0" y1="250" x2="1000" y2="250" />
					</g>
					<path class="chart-area" d="M0,230 C90,225 150,210 230,200 C320,188 360,170 450,150 C540,130 590,140 670,110 C760,78 820,72 900,52 C935,42 960,34 982,26 L982,280 L0,280 Z" />
					<path class="chart-line" d="M0,230 C90,225 150,210 230,200 C320,188 360,170 450,150 C540,130 590,140 670,110 C760,78 820,72 900,52 C935,42 960,34 982,26" />
					<g class="chart-dot">
						<circle class="pulse-ring" cx="982" cy="26" r="6" fill="rgba(61,220,132,0.4)" />
						<circle cx="982" cy="26" r="5.5" fill="#3ddc84" stroke="#080b11" stroke-width="2.5" />
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
			<h2>Sé de los primeros en tener<br />el control.</h2>
			<p>Cupos de acceso anticipado limitados. Reserva tu lugar antes del lanzamiento.</p>
			<a href="#waitlist" class="btn-gold">Unirme a la lista de espera</a>
		</div>
	</section>
</main>

<footer>
	<div class="wrap foot">
		<div class="brand">
			<div class="brand-mark"></div>
			<div class="brand-name" style="font-size:18px;">FINEXIA</div>
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
		background: #080b11;
		color: #f4f6f8;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		line-height: 1.55;
		-webkit-font-smoothing: antialiased;
		overflow-x: hidden;
	}
	:global(::selection) {
		background: rgba(230, 196, 60, 0.28);
		color: #fff;
	}
	:global(a) {
		color: inherit;
		text-decoration: none;
	}

	/* CSS VARIABLES */
	:global(:root) {
		--bg: #080b11;
		--surface: rgba(255, 255, 255, 0.025);
		--surface-2: rgba(255, 255, 255, 0.045);
		--border: rgba(255, 255, 255, 0.08);
		--border-strong: rgba(255, 255, 255, 0.14);
		--text: #f4f6f8;
		--text-muted: #8b95a3;
		--text-dim: #5c6675;
		--gold: #e6c43c;
		--gold-deep: #d4ae2c;
		--green: #3ddc84;
		--maxw: 1180px;
		--font-display: 'Space Grotesk', system-ui, sans-serif;
		--font-body: 'Hanken Grotesk', system-ui, sans-serif;
	}

	.wrap {
		max-width: var(--maxw);
		margin: 0 auto;
		padding: 0 32px;
	}

	/* ATMOSPHERE */
	.atmosphere {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
		overflow: hidden;
	}
	.glow {
		position: absolute;
		border-radius: 50%;
		filter: blur(90px);
		opacity: 0.5;
	}
	.glow.g1 {
		width: 620px;
		height: 620px;
		top: -260px;
		left: -160px;
		background: radial-gradient(circle, rgba(230, 196, 60, 0.16), transparent 68%);
	}
	.glow.g2 {
		width: 680px;
		height: 680px;
		top: -200px;
		right: -220px;
		background: radial-gradient(circle, rgba(61, 220, 132, 0.12), transparent 70%);
	}
	.grid-overlay {
		position: absolute;
		inset: 0;
		background-image: linear-gradient(rgba(255, 255, 255, 0.022) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255, 255, 255, 0.022) 1px, transparent 1px);
		background-size: 64px 64px;
		-webkit-mask-image: radial-gradient(ellipse 90% 60% at 50% 0%, #000 30%, transparent 78%);
		mask-image: radial-gradient(ellipse 90% 60% at 50% 0%, #000 30%, transparent 78%);
	}

	main,
	header,
	footer {
		position: relative;
		z-index: 1;
	}

	/* NAV */
	header {
		position: sticky;
		top: 0;
		z-index: 50;
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		background: rgba(8, 11, 17, 0.72);
		border-bottom: 1px solid var(--border);
	}
	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 76px;
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.brand-mark {
		width: 36px;
		height: 36px;
		border-radius: 10px;
		background: linear-gradient(150deg, var(--gold) 0%, var(--green) 130%);
		box-shadow: 0 4px 18px rgba(230, 196, 60, 0.22);
	}
	.brand-name {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 22px;
		letter-spacing: 0.02em;
		color: var(--gold);
	}
	.nav-links {
		display: flex;
		align-items: center;
		gap: 36px;
	}
	.nav-links a {
		font-size: 15px;
		color: var(--text-muted);
		font-weight: 500;
		transition: color 0.2s ease;
	}
	.nav-links a:hover {
		color: var(--text);
	}
	.nav-cta {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 10px;
		border: 1px solid var(--border-strong);
		font-size: 14px;
		font-weight: 600;
		color: var(--text);
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}
	.nav-cta:hover {
		border-color: var(--gold);
		background: rgba(230, 196, 60, 0.06);
	}
	@media (max-width: 860px) {
		.nav-links {
			display: none;
		}
	}

	/* HERO */
	.hero {
		text-align: center;
		padding: 96px 0 88px;
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		padding: 8px 16px 8px 12px;
		border-radius: 100px;
		border: 1px solid rgba(230, 196, 60, 0.3);
		background: rgba(230, 196, 60, 0.05);
		font-size: 12.5px;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--gold);
	}
	.dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--green);
		box-shadow: 0 0 0 0 rgba(61, 220, 132, 0.6);
		animation: pulse 2.4s ease-out infinite;
		flex-shrink: 0;
	}
	@keyframes pulse {
		0% {
			box-shadow: 0 0 0 0 rgba(61, 220, 132, 0.55);
		}
		70% {
			box-shadow: 0 0 0 9px rgba(61, 220, 132, 0);
		}
		100% {
			box-shadow: 0 0 0 0 rgba(61, 220, 132, 0);
		}
	}
	.hero-title {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: clamp(44px, 7.4vw, 92px);
		line-height: 0.98;
		letter-spacing: -0.025em;
		margin: 30px 0 0;
		max-width: 14ch;
		text-wrap: balance;
	}
	.accent {
		background: linear-gradient(105deg, var(--gold) 0%, #f3da6b 50%, var(--gold-deep) 100%);
		-webkit-background-clip: text;
		background-clip: text;
		-webkit-text-fill-color: transparent;
	}
	.hero-sub {
		margin: 28px auto 0;
		max-width: 56ch;
		font-size: clamp(17px, 2vw, 20px);
		color: var(--text-muted);
		text-wrap: pretty;
	}

	/* COUNTDOWN */
	.countdown {
		display: flex;
		gap: 14px;
		margin-top: 48px;
		flex-wrap: wrap;
		justify-content: center;
	}
	.cd-cell {
		min-width: 104px;
		padding: 22px 18px 16px;
		border-radius: 16px;
		background: var(--surface);
		border: 1px solid var(--border);
		backdrop-filter: blur(6px);
	}
	.cd-num {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: clamp(36px, 5vw, 48px);
		line-height: 1;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}
	.cd-label {
		margin-top: 10px;
		font-size: 11.5px;
		font-weight: 600;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	/* WAITLIST */
	.waitlist {
		margin-top: 52px;
		width: 100%;
		max-width: 520px;
	}
	.wl-form {
		display: flex;
		gap: 10px;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 7px;
		transition:
			border-color 0.2s ease,
			box-shadow 0.2s ease;
	}
	.wl-form:focus-within {
		border-color: rgba(230, 196, 60, 0.55);
		box-shadow: 0 0 0 4px rgba(230, 196, 60, 0.08);
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
		font-size: 16px;
		padding: 0 16px;
		min-width: 0;
	}
	.wl-form input::placeholder {
		color: var(--text-dim);
	}
	.btn-gold {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 13px 24px;
		border-radius: 9px;
		border: none;
		cursor: pointer;
		font-family: var(--font-body);
		font-size: 15px;
		font-weight: 700;
		color: #1c1606;
		background: linear-gradient(135deg, var(--gold) 0%, var(--gold-deep) 100%);
		box-shadow: 0 6px 20px rgba(230, 196, 60, 0.22);
		transition:
			transform 0.15s ease,
			box-shadow 0.2s ease;
	}
	.btn-gold:hover {
		transform: translateY(-1px);
		box-shadow: 0 10px 28px rgba(230, 196, 60, 0.3);
	}
	.btn-gold:active {
		transform: translateY(0);
	}
	.wl-note {
		margin-top: 14px;
		font-size: 13.5px;
		color: var(--text-dim);
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}
	.wl-success {
		display: flex;
		align-items: center;
		gap: 12px;
		justify-content: center;
		padding: 18px 22px;
		border-radius: 14px;
		background: rgba(61, 220, 132, 0.07);
		border: 1px solid rgba(61, 220, 132, 0.28);
		color: var(--text);
		font-size: 15.5px;
		font-weight: 500;
	}
	.check-ico {
		flex-shrink: 0;
		width: 26px;
		height: 26px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--green);
		color: #06140c;
	}
	@media (max-width: 560px) {
		.wl-form {
			flex-direction: column;
		}
		.btn-gold {
			justify-content: center;
		}
		.cd-cell {
			min-width: 72px;
			padding: 16px 10px 12px;
		}
	}

	/* SECTION SHELL */
	.block {
		padding: 92px 0;
	}
	.sec-head {
		max-width: 660px;
		margin: 0 auto 56px;
		text-align: center;
	}
	.eyebrow {
		font-size: 12.5px;
		font-weight: 600;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--gold);
	}
	.sec-title {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: clamp(30px, 4.4vw, 46px);
		line-height: 1.08;
		letter-spacing: -0.02em;
		margin-top: 14px;
		text-wrap: balance;
	}
	.sec-desc {
		margin-top: 16px;
		font-size: 17px;
		color: var(--text-muted);
		text-wrap: pretty;
	}
	.divider {
		height: 1px;
		background: linear-gradient(90deg, transparent, var(--border), transparent);
	}

	/* BENEFITS */
	.benefits-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 20px;
	}
	.bcard {
		padding: 28px 26px;
		border-radius: 18px;
		background: var(--surface);
		border: 1px solid var(--border);
		transition:
			border-color 0.25s ease,
			transform 0.25s ease,
			background 0.25s ease;
	}
	.bcard:hover {
		border-color: var(--border-strong);
		transform: translateY(-3px);
		background: var(--surface-2);
	}
	.bcard .ico {
		width: 46px;
		height: 46px;
		border-radius: 12px;
		display: grid;
		place-items: center;
		background: rgba(230, 196, 60, 0.08);
		border: 1px solid rgba(230, 196, 60, 0.18);
		color: var(--gold);
		margin-bottom: 22px;
	}
	.bcard.green .ico {
		background: rgba(61, 220, 132, 0.08);
		border-color: rgba(61, 220, 132, 0.2);
		color: var(--green);
	}
	.bcard h3 {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 19px;
		letter-spacing: -0.01em;
	}
	.bcard p {
		margin-top: 10px;
		font-size: 14.5px;
		color: var(--text-muted);
		line-height: 1.6;
	}
	@media (max-width: 980px) {
		.benefits-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	@media (max-width: 540px) {
		.benefits-grid {
			grid-template-columns: 1fr;
		}
	}

	/* HOW IT WORKS */
	.steps {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 24px;
	}
	.step {
		position: relative;
		padding: 34px 30px 30px;
		border-radius: 20px;
		background: linear-gradient(180deg, var(--surface-2), var(--surface));
		border: 1px solid var(--border);
	}
	.step .num {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 16px;
		width: 40px;
		height: 40px;
		border-radius: 11px;
		display: grid;
		place-items: center;
		color: #1c1606;
		background: linear-gradient(135deg, var(--gold), var(--gold-deep));
		margin-bottom: 22px;
	}
	.step h3 {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 21px;
		letter-spacing: -0.01em;
	}
	.step p {
		margin-top: 12px;
		font-size: 15px;
		color: var(--text-muted);
		line-height: 1.62;
	}
	.connector {
		position: absolute;
		top: 54px;
		right: -16px;
		width: 32px;
		height: 1px;
		background: linear-gradient(90deg, var(--border-strong), transparent);
	}
	.step:last-child .connector {
		display: none;
	}
	@media (max-width: 860px) {
		.steps {
			grid-template-columns: 1fr;
		}
		.connector {
			display: none;
		}
	}

	/* FAQ */
	.faq {
		max-width: 800px;
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
		padding: 26px 4px;
		text-align: left;
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 19px;
		color: var(--text);
	}
	.plus {
		flex-shrink: 0;
		width: 26px;
		height: 26px;
		position: relative;
		transition: transform 0.3s ease;
	}
	.plus::before,
	.plus::after {
		content: '';
		position: absolute;
		background: var(--gold);
		border-radius: 2px;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
	}
	.plus::before {
		width: 14px;
		height: 2px;
	}
	.plus::after {
		width: 2px;
		height: 14px;
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
		font-size: 15.5px;
		color: var(--text-muted);
		line-height: 1.66;
		padding: 0 40px 4px 4px;
	}
	.faq-item.open .faq-a {
		max-height: 280px;
		padding-bottom: 26px;
	}

	/* METRICS SHOWCASE */
	.metrics-wrap {
		position: relative;
		border-radius: 28px;
		overflow: hidden;
		border: 1px solid var(--border-strong);
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.04), rgba(255, 255, 255, 0.012));
		padding: 48px 48px 0;
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
		font-family: var(--font-display);
		font-weight: 700;
		font-size: clamp(28px, 3.6vw, 42px);
		line-height: 1.06;
		letter-spacing: -0.02em;
		margin-top: 12px;
		text-wrap: balance;
	}
	.growth-badge {
		color: var(--green);
		font-size: 0.5em;
		font-weight: 600;
		letter-spacing: 0;
	}
	.metrics-headline p {
		margin-top: 14px;
		font-size: 16px;
		color: var(--text-muted);
	}
	.metric-chips {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}
	.mchip {
		padding: 16px 20px;
		border-radius: 14px;
		min-width: 132px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
		backdrop-filter: blur(6px);
	}
	.mchip .lbl {
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
	}
	.mchip .val {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 25px;
		letter-spacing: -0.02em;
		margin-top: 8px;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}
	.delta {
		font-size: 13px;
		font-weight: 600;
		margin-left: 7px;
	}
	.delta.up {
		color: var(--green);
	}
	.delta.gold {
		color: var(--gold);
	}

	/* CHART */
	.chart-zone {
		position: relative;
		margin-top: 36px;
		height: 280px;
	}
	.chart-svg {
		display: block;
		width: 100%;
		height: 100%;
	}
	.chart-grid line {
		stroke: rgba(255, 255, 255, 0.05);
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
		stroke-width: 2.5;
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
		padding: 10px 14px;
		border-radius: 12px;
		background: rgba(8, 11, 17, 0.82);
		border: 1px solid rgba(61, 220, 132, 0.32);
		backdrop-filter: blur(8px);
		opacity: 0;
		transform: translateY(8px);
		transition:
			opacity 0.5s ease 1.5s,
			transform 0.5s ease 1.5s;
	}
	:global(.metrics-wrap.in) .chart-flag {
		opacity: 1;
		transform: none;
	}
	.chart-flag .t {
		font-size: 11px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-dim);
	}
	.chart-flag .v {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 19px;
		color: var(--green);
		margin-top: 4px;
	}
	@media (max-width: 620px) {
		.metrics-wrap {
			padding: 32px 22px 0;
		}
		.chart-zone {
			height: 200px;
		}
		.mchip {
			min-width: 0;
			flex: 1;
			padding: 13px 14px;
		}
		.mchip .val {
			font-size: 20px;
		}
	}

	/* FINAL CTA */
	.final-cta {
		text-align: center;
		padding: 80px 40px;
		border-radius: 28px;
		position: relative;
		overflow: hidden;
		background: radial-gradient(ellipse 80% 130% at 50% -10%, rgba(230, 196, 60, 0.1), transparent 60%),
			linear-gradient(180deg, rgba(255, 255, 255, 0.04), rgba(255, 255, 255, 0.015));
		border: 1px solid var(--border-strong);
	}
	.final-cta h2 {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: clamp(30px, 5vw, 52px);
		line-height: 1.04;
		letter-spacing: -0.025em;
		text-wrap: balance;
	}
	.final-cta p {
		margin: 18px auto 0;
		max-width: 50ch;
		font-size: 18px;
		color: var(--text-muted);
	}
	.final-cta .btn-gold {
		margin-top: 36px;
		padding: 16px 34px;
		font-size: 16px;
		display: inline-flex;
	}

	/* FOOTER */
	footer {
		border-top: 1px solid var(--border);
		padding: 44px 0 56px;
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
		gap: 28px;
	}
	.foot-links a {
		font-size: 14px;
		color: var(--text-muted);
		transition: color 0.2s;
	}
	.foot-links a:hover {
		color: var(--text);
	}
	.foot-copy {
		font-size: 13.5px;
		color: var(--text-dim);
	}

	/* REVEAL */
	.reveal {
		opacity: 0;
		transform: translateY(22px);
		transition:
			opacity 0.7s ease,
			transform 0.7s ease;
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
		.dot {
			animation: none;
		}
	}
</style>
