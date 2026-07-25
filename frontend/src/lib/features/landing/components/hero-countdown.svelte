<script lang="ts">
	import { onMount } from 'svelte';

	let countdown = $state({ days: '00', hours: '00', mins: '00', secs: '00' });

	function pad(n: number) {
		return String(n).padStart(2, '0');
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

<style>
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

	@media (max-width: 640px) {
		.countdown {
			flex-wrap: wrap;
		}
	}

	@media (max-width: 480px) {
		.countdown {
			margin-top: 32px;
		}
		.cd-cell {
			padding: 14px 16px;
		}
	}
</style>
