<script lang="ts">
	import { systemStore } from '$lib/stores/docker';
	import ContainerStats from '$lib/components/containers/ContainerStats.svelte';

	$: systemInfo = $systemStore.data;
</script>

<div class="space-y-6">
	<h1 class="text-2xl font-bold">Dashboard</h1>

	<!-- System Overview -->
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Containers</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.containers || 0}</p>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Images</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.images || 0}</p>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">CPU Usage</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.cpuUsage?.toFixed(1) || 0}%</p>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Memory Usage</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.memoryUsage?.toFixed(1) || 0}%</p>
		</div>
	</div>
</div>
