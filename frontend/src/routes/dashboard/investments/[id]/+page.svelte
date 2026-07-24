<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { InvestmentDetail, findInvestmentProduct } from '$lib/features/investments';

	const id = $derived(page.params.id ?? '');
	const investment = $derived(findInvestmentProduct(id));

	function handleBack() {
		goto(resolve('/dashboard/investments'));
	}
</script>

<svelte:head>
	<title>{investment?.name} - FINEXIA</title>
	<meta name="description" content={investment?.description} />
</svelte:head>

<InvestmentDetail {id} onBack={handleBack} />
