<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { dockerClient } from '$lib/api/client';

	export let containerId: string;

	let cpuData: number[] = Array(30).fill(0);
	let memoryData: number[] = Array(30).fill(0);
	let networkRx: number = 0;
	let networkTx: number = 0;

	let interval: NodeJS.Timeout;

	onMount(() => {
		// Update stats every second
		interval = setInterval(async () => {
			try {
				const stats = await dockerClient.getContainerStats(containerId);

				// Update CPU usage
				cpuData = [...cpuData.slice(1), stats.cpuPercentage];

				// Update memory usage
				memoryData = [...memoryData.slice(1), stats.memoryUsage];

				// Update network stats
				networkRx = stats.networkRx;
				networkTx = stats.networkTx;
			} catch (error) {
				console.error('Failed to fetch container stats:', error);
			}
		}, 1000);
	});

	onDestroy(() => {
		clearInterval(interval);
	});

	function getGraphPath(data: number[]): string {
		const width = 500;
		const height = 100;
		const points = data.map(
			(value, index) => `${(index * width) / (data.length - 1)},${height - (value * height) / 100}`
		);
		return `M ${points.join(' L ')}`;
	}
</script>

<div class="space-y-6">
	<div class="grid grid-cols-2 gap-4">
		<!-- CPU Usage Graph -->
		<div class="rounded-lg border p-4 dark:border-gray-700">
			<h3 class="mb-2 text-lg font-semibold">CPU Usage</h3>
			<svg class="h-[100px] w-full">
				<path d={getGraphPath(cpuData)} class="fill-none stroke-blue-500 stroke-2" />
			</svg>
			<div class="mt-2 text-center text-sm">
				Current: {cpuData[cpuData.length - 1].toFixed(1)}%
			</div>
		</div>

		<!-- Memory Usage Graph -->
		<div class="rounded-lg border p-4 dark:border-gray-700">
			<h3 class="mb-2 text-lg font-semibold">Memory Usage</h3>
			<svg class="h-[100px] w-full">
				<path d={getGraphPath(memoryData)} class="fill-none stroke-green-500 stroke-2" />
			</svg>
			<div class="mt-2 text-center text-sm">
				Current: {memoryData[memoryData.length - 1].toFixed(1)}%
			</div>
		</div>
	</div>

	<!-- Network Stats -->
	<div class="rounded-lg border p-4 dark:border-gray-700">
		<h3 class="mb-2 text-lg font-semibold">Network I/O</h3>
		<div class="grid grid-cols-2 gap-4">
			<div>
				<span class="text-sm text-gray-500 dark:text-gray-400">Received</span>
				<p class="text-lg font-medium">{(networkRx / 1024 / 1024).toFixed(2)} MB</p>
			</div>
			<div>
				<span class="text-sm text-gray-500 dark:text-gray-400">Transmitted</span>
				<p class="text-lg font-medium">{(networkTx / 1024 / 1024).toFixed(2)} MB</p>
			</div>
		</div>
	</div>
</div>
