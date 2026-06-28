<script lang="ts">
	interface Props {
		name: string;
		value?: string;
		required?: boolean;
		class?: string;
	}

	let {
		name,
		value = $bindable(new Date().toISOString().split('T')[0]),
		required: _required = false,
		class: cls = ''
	}: Props = $props();

	const MONTHS = [
		'Enero',
		'Febrero',
		'Marzo',
		'Abril',
		'Mayo',
		'Junio',
		'Julio',
		'Agosto',
		'Septiembre',
		'Octubre',
		'Noviembre',
		'Diciembre'
	];

	const currentYear = new Date().getFullYear();
	const years = Array.from({ length: 80 }, (_, i) => currentYear - i);

	let day = $derived(value ? parseInt(value.split('-')[2]) : new Date().getDate());
	let month = $derived(value ? parseInt(value.split('-')[1]) : new Date().getMonth() + 1);
	let year = $derived(value ? parseInt(value.split('-')[0]) : currentYear);

	function daysInMonth(m: number, y: number) {
		return new Date(y, m, 0).getDate();
	}

	let maxDay = $derived(daysInMonth(month, year));
	let days = $derived(Array.from({ length: maxDay }, (_, i) => i + 1));

	function update(newDay: number, newMonth: number, newYear: number) {
		const safeDay = Math.min(newDay, daysInMonth(newMonth, newYear));
		value = `${newYear}-${String(newMonth).padStart(2, '0')}-${String(safeDay).padStart(2, '0')}`;
	}
</script>

<input type="hidden" {name} {value} />

<div class="date-picker {cls}">
	<select
		class="dp-select"
		aria-label="Día"
		value={day}
		onchange={(e) => update(parseInt((e.target as HTMLSelectElement).value), month, year)}
	>
		{#each days as d (d)}
			<option value={d} selected={d === day}>{String(d).padStart(2, '0')}</option>
		{/each}
	</select>

	<select
		class="dp-select dp-month"
		aria-label="Mes"
		value={month}
		onchange={(e) => update(day, parseInt((e.target as HTMLSelectElement).value), year)}
	>
		{#each MONTHS as m, i (i)}
			<option value={i + 1} selected={i + 1 === month}>{m}</option>
		{/each}
	</select>

	<select
		class="dp-select"
		aria-label="Año"
		value={year}
		onchange={(e) => update(day, month, parseInt((e.target as HTMLSelectElement).value))}
	>
		{#each years as y (y)}
			<option value={y} selected={y === year}>{y}</option>
		{/each}
	</select>
</div>

<style>
	.date-picker {
		display: flex;
		gap: 0.5rem;
	}

	.dp-select {
		padding: 0.6rem 0.5rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		background-color: rgba(255, 255, 255, 0.04);
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='%23d4912a' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.4rem center;
		background-size: 0.85rem;
		appearance: none;
		-webkit-appearance: none;
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		cursor: pointer;
		transition: border-color 0.2s ease;
		padding-right: 1.6rem;
	}

	.dp-select:focus {
		outline: none;
		border-color: var(--amber);
	}

	.dp-month {
		flex: 1;
	}
</style>
