<script lang="ts">
	/** Envoltorio de pruebas: `Modal` recibe su contenido por snippet. */
	import Modal from './modal.svelte';

	let {
		open = false,
		title = 'Título',
		description = '',
		tone = 'default',
		onClose = () => {}
	}: {
		open?: boolean;
		title?: string;
		description?: string;
		tone?: 'default' | 'danger';
		onClose?: () => void;
	} = $props();
</script>

<Modal {open} {title} {description} {tone} {onClose}>
	<p>contenido del modal</p>
</Modal>
