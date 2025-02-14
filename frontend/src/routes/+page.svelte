<script lang="ts">
	import { systemStore, containersStore } from '$lib/stores/docker';
	import { formatBytes } from '$lib/utils/format';

	$: systemInfo = $systemStore.data;
	$: containers = $containersStore.data;

	$: containerStats = containers.reduce(
		(acc, container) => {
			const state = container.State.toLowerCase() as 'running' | 'exited' | 'paused';
			if (state in acc) {
				acc[state]++;
			}
			return acc;
		},
		{ running: 0, exited: 0, paused: 0 }
	);

	$: recentContainers = [...containers].sort((a, b) => b.Created - a.Created).slice(0, 5);
</script>

<div class="space-y-6">
	<h1 class="text-2xl font-bold">Dashboard</h1>

	<!-- System Overview -->
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Containers</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.containers || 0}</p>
			<div class="mt-2 text-sm text-gray-600 dark:text-gray-400">
				{containerStats.running} running · {containerStats.exited} stopped · {containerStats.paused}
				paused
			</div>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Images</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.images || 0}</p>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">CPU Usage</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.cpuUsage?.toFixed(1) || 0}%</p>
			<div class="mt-2 text-sm text-gray-600 dark:text-gray-400">
				{systemInfo?.NCPU || 0} CPUs Available
			</div>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">Memory Usage</h3>
			<p class="mt-2 text-2xl font-semibold">{systemInfo?.memoryUsage?.toFixed(1) || 0}%</p>
			<div class="mt-2 text-sm text-gray-600 dark:text-gray-400">
				{formatBytes(systemInfo?.MemTotal || 0)} Total
			</div>
		</div>
	</div>

	<!-- Recent Containers -->
	<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
		<h2 class="mb-4 text-lg font-semibold">Recent Containers</h2>
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
				<thead>
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">Name</th>
						<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">Image</th>
						<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">Status</th>
						<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">Created</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
					{#each recentContainers as container}
						<tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
							<td class="whitespace-nowrap px-4 py-3">{container.Names[0].replace('/', '')}</td>
							<td class="whitespace-nowrap px-4 py-3">{container.Image}</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span
									class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
									class:bg-green-100={container.State === 'running'}
									class:text-green-800={container.State === 'running'}
									class:bg-red-100={container.State === 'exited'}
									class:text-red-800={container.State === 'exited'}
									class:bg-yellow-100={container.State === 'paused'}
									class:text-yellow-800={container.State === 'paused'}
								>
									{container.State}
								</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								{new Date(container.Created * 1000).toLocaleString()}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>
