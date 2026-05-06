<script lang="ts">
</script>

<header class="sticky top-0 z-50 border-b border-[rgba(212,175,55,0.1)] bg-[#0f1419]/95 backdrop-blur-md">
	<nav class="mx-auto flex w-full max-w-7xl items-center justify-between px-6 py-4">
		<div class="flex items-center gap-3">
			<div class="logo-mark" aria-hidden="true"></div>
			<span class="font-display text-xl font-bold tracking-tight text-[#d4af37]">FINEXIA</span>
		</div>
		<div class="hidden items-center gap-10 text-sm font-500 tracking-tight md:flex">
			<a href="#caracteristicas" class="nav-link">Características</a>
			<a href="#como-funciona" class="nav-link">Cómo funciona</a>
			<a href="#caracteristicas-principales" class="nav-link">Features</a>
			<a href="#testimonios" class="nav-link">Testimonios</a>
		</div>
		<div class="flex items-center gap-3">
			<button class="cta-secondary" aria-label="Ver demo" type="button">Ver demo</button>
			<a href="/auth" class="cta-primary" aria-label="Acceder a la plataforma">
				Acceder
			</a>
		</div>
	</nav>
</header>

<style>
	.logo-mark {
		height: 32px;
		width: 32px;
		border-radius: 6px;
		background: linear-gradient(135deg, #d4af37, #2ecc71);
		border: 1px solid #d4af37;
		box-shadow: 0 0 12px rgba(212, 175, 55, 0.3);
	}

	.nav-link {
		position: relative;
		color: #e0e0e0;
		text-decoration: none;
		transition: color 0.25s ease;
	}

	.nav-link::after {
		content: '';
		position: absolute;
		bottom: -4px;
		left: 0;
		width: 0;
		height: 2px;
		background: #d4af37;
		transition: width 0.25s ease;
	}

	.nav-link:hover {
		color: #d4af37;
	}

	.nav-link:hover::after {
		width: 100%;
	}

	.cta-primary {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.75rem 1.75rem;
		background: linear-gradient(135deg, #d4af37 0%, #e8c547 100%);
		color: #0f1419;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.95rem;
		cursor: pointer;
		text-decoration: none;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		box-shadow: 0 4px 16px rgba(212, 175, 55, 0.2);
		letter-spacing: 0.5px;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.cta-primary:hover {
		background: linear-gradient(135deg, #e8c547 0%, #d4af37 100%);
		transform: translateY(-2px);
		box-shadow: 0 6px 24px rgba(212, 175, 55, 0.35);
	}

	.cta-primary:active {
		transform: translateY(0);
	}

	:global(.cta-secondary) {
		padding: 0.75rem 1.5rem;
		background: transparent;
		color: #e0e0e0;
		border: 1.5px solid #d4af37;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.5px;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	:global(.cta-secondary:hover) {
		background: rgba(212, 175, 55, 0.1);
		color: #d4af37;
	}
</style>
