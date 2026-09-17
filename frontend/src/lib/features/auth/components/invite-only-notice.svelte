<script lang="ts">
	/*
	 * Lo que se ve en «Crear cuenta» mientras el registro está cerrado. No es un
	 * formulario deshabilitado: dice por qué no hay formulario y las dos salidas
	 * que sí existen, la lista de espera y el enlace de una invitación.
	 */
	import { resolve } from '$app/paths';
</script>

<div class="invite-only" id="register-form">
	<h2 class="title">Registro por invitación</h2>
	<p class="shell-copy">
		Durante la beta, las cuentas se crean por invitación. Únete a la lista de espera y te escribimos
		cuando tengas acceso.
	</p>

	<a href="{resolve('/')}#waitlist" class="lp-btn cta block">Unirme a la lista de espera</a>

	<p class="shell-after">
		¿Ya te llegó la invitación? Abre el enlace del correo para activar tu cuenta.
	</p>
</div>

<style>
	.title {
		margin: 0 0 12px;
		font-size: var(--lp-fs-h3);
		font-stretch: 110%;
		font-weight: 620;
		line-height: 1.2;
	}

	.cta {
		margin-top: 28px;
	}
</style>
