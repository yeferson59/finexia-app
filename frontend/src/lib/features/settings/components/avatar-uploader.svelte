<script lang="ts">
	/**
	 * Foto de perfil: previsualización, selector de archivo y subida.
	 *
	 * La imagen se comprime en el navegador antes de enviarla, así que el archivo
	 * que viaja no es el que eligió el usuario: se reemplaza el `FileList` del
	 * input para que el submit mande el comprimido.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import { actionData, actionError, actionSucceeded, type SettingsForm } from '../settings';

	interface Props {
		/** Usuario de sesión, para la imagen guardada y la inicial de respaldo. */
		user: App.Locals['user'];
		form: SettingsForm;
	}

	let { user, form }: Props = $props();

	let avatarPreview = $state<string | null>(null);
	let avatarFile = $state<File | null>(null);
	let avatarLoading = $state(false);
	let avatarFileInput = $state<HTMLInputElement | null>(null);

	const avatarSuccess = $derived(actionSucceeded(form, 'uploadAvatar'));
	const avatarError = $derived(actionError(form, 'uploadAvatar'));

	// Avatar URL: prefer the uploaded URL returned by the server action, then the stored image
	const savedAvatarUrl = $derived(
		avatarSuccess
			? (actionData<string>(form, 'uploadAvatar', 'imageUrl') ?? user?.image ?? '')
			: (user?.image ?? '')
	);

	const displayAvatar = $derived(
		avatarPreview ?? (savedAvatarUrl && savedAvatarUrl !== 'avatar.png' ? savedAvatarUrl : null)
	);

	const userInitial = $derived((user?.name ?? '').trim().charAt(0).toUpperCase());

	$effect(() => {
		if (avatarSuccess) {
			avatarPreview = null;
			avatarFile = null;
		}
	});

	function compressImageForAvatar(file: File): Promise<File> {
		const MAX = 512;
		const QUALITY = 0.85;
		const outputType = file.type === 'image/webp' ? 'image/webp' : 'image/jpeg';

		return new Promise((resolve) => {
			const img = new Image();
			const url = URL.createObjectURL(file);
			img.onload = () => {
				URL.revokeObjectURL(url);
				let { width, height } = img;
				if (width > MAX || height > MAX) {
					const ratio = Math.min(MAX / width, MAX / height);
					width = Math.round(width * ratio);
					height = Math.round(height * ratio);
				}
				const canvas = document.createElement('canvas');
				canvas.width = width;
				canvas.height = height;
				canvas.getContext('2d')!.drawImage(img, 0, 0, width, height);
				canvas.toBlob(
					(blob) => {
						if (!blob) return resolve(file);
						const ext = outputType === 'image/webp' ? '.webp' : '.jpg';
						resolve(new File([blob], 'avatar' + ext, { type: outputType }));
					},
					outputType,
					QUALITY
				);
			};
			img.onerror = () => {
				URL.revokeObjectURL(url);
				resolve(file);
			};
			img.src = url;
		});
	}

	async function onFileChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		if (avatarPreview) URL.revokeObjectURL(avatarPreview);
		const compressed = await compressImageForAvatar(file);
		// Replace input.files so the form submission sends the compressed file
		const dt = new DataTransfer();
		dt.items.add(compressed);
		input.files = dt.files;
		avatarFile = compressed;
		avatarPreview = URL.createObjectURL(compressed);
	}
</script>

<div class="avatar-section">
	<div class="avatar-display">
		{#if displayAvatar}
			<img src={displayAvatar} alt="Avatar de usuario" class="avatar-img" />
		{:else}
			<div class="avatar-initials" aria-hidden="true">{userInitial}</div>
		{/if}
	</div>
	<div class="avatar-controls">
		<form
			method="POST"
			action="?/uploadAvatar"
			enctype="multipart/form-data"
			use:enhance={() => {
				avatarLoading = true;
				return async ({ update }) => {
					await update();
					avatarLoading = false;
				};
			}}
		>
			<input
				bind:this={avatarFileInput}
				type="file"
				name="avatar"
				accept="image/jpeg,image/png,image/webp"
				class="file-input-hidden"
				onchange={onFileChange}
			/>
			<button type="button" class="btn-pick-file" onclick={() => avatarFileInput?.click()}>
				Cambiar foto
			</button>
			{#if avatarFile}
				<Button type="submit" size="sm" loading={avatarLoading}>
					{avatarLoading ? 'Subiendo…' : 'Guardar foto'}
				</Button>
			{/if}
		</form>
		<p class="avatar-hint">JPEG, PNG o WebP. Se encoge sola antes de subirse.</p>
		{#if avatarError}
			<p class="feedback error">{avatarError}</p>
		{/if}
		{#if avatarSuccess}
			<p class="feedback success">Foto actualizada correctamente.</p>
		{/if}
	</div>
</div>

<style>
	/* En vertical: vive en el carril de la sección, no encima del formulario. */
	.avatar-section {
		display: flex;
		align-items: flex-start;
		gap: 0.9rem;
		margin-top: 1.1rem;
	}

	.avatar-display {
		flex-shrink: 0;
	}

	.avatar-img {
		width: 64px;
		height: 64px;
		border-radius: 50%;
		object-fit: cover;
		border: 2px solid var(--border-strong);
	}

	.avatar-initials {
		width: 64px;
		height: 64px;
		border-radius: 50%;
		background: var(--surface-3);
		border: 2px solid var(--border-strong);
		color: var(--amber);
		font-family: var(--font-mono);
		font-size: 1.35rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.avatar-controls {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.avatar-controls form {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		flex-wrap: wrap;
	}

	.file-input-hidden {
		display: none;
	}

	.btn-pick-file {
		padding: 0.4rem 0.85rem;
		border-radius: 7px;
		border: 1px solid var(--border-strong);
		background: transparent;
		color: var(--text);
		font-size: 0.8rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.btn-pick-file:hover {
		border-color: rgba(212, 145, 42, 0.5);
		background: var(--panel);
	}

	.avatar-hint {
		max-width: 22ch;
		margin: 0;
		font-size: 0.72rem;
		line-height: 1.45;
		color: var(--text-dim);
	}
</style>
