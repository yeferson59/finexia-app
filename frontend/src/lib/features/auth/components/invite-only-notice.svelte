<script lang="ts">
	import { resolve } from '$app/paths';

	let {
		slideDirection = 'right',
		onSwitchToLogin
	}: {
		slideDirection?: 'left' | 'right';
		onSwitchToLogin: () => void;
	} = $props();
</script>

<div
	class="form-content invite-only"
	class:slide-left={slideDirection === 'left'}
	id="register-form"
>
	<p class="invite-only-title">Registro por invitación</p>
	<p class="invite-only-copy">
		FINEXIA está en fase beta y el acceso es solo por invitación. Únete a la lista de espera y te
		avisaremos en cuanto tengas un cupo.
	</p>

	<a href="{resolve('/')}#waitlist" class="invite-only-cta"> Unirme a la lista de espera </a>

	<div class="form-switch">
		¿Ya tienes cuenta?
		<button type="button" onclick={onSwitchToLogin} class="switch-link"> Inicia sesión </button>
	</div>
</div>

<style>
	.invite-only {
		text-align: center;
	}

	.invite-only-title {
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.invite-only-copy {
		font-size: 0.9rem;
		line-height: 1.6;
		color: var(--text-secondary);
		margin: 0;
	}

	.invite-only-cta {
		display: block;
		width: 100%;
		padding: 1rem 2rem;
		border-radius: 8px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 600;
		font-size: 0.95rem;
		text-align: center;
		text-decoration: none;
		letter-spacing: 0.5px;
		box-shadow: 0 4px 16px rgba(212, 145, 42, 0.2);
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.invite-only-cta:hover {
		background: var(--amber-light);
		transform: translateY(-2px);
		box-shadow: 0 6px 24px rgba(212, 145, 42, 0.35);
	}
</style>
