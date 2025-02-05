<script lang="ts">
	import { containersStore } from '$lib/stores/docker';
	import type { Container } from '$lib/types/docker';

	let searchQuery = '';
	let view: 'grid' | 'table' = 'table';

	$: containers = $containersStore.data;
	$: filteredContainers = containers.filter(
		(container) =>
			container.Names[0].toLowerCase().includes(searchQuery.toLowerCase()) ||
			container.Image.toLowerCase().includes(searchQuery.toLowerCase())
	);

	function getStatusColor(state: string): string {
		switch (state.toLowerCase()) {
			case 'running':
				return 'bg-green-400';
			case 'paused':
				return 'bg-yellow-400';
			case 'exited':
				return 'bg-red-400';
			default:
				return 'bg-gray-400';
		}
	}
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<input
			type="search"
			bind:value={searchQuery}
			placeholder="Search containers..."
			class="rounded-lg border border-gray-300 px-4 py-2 dark:border-gray-600 dark:bg-gray-700"
		/>
		<div class="flex items-center space-x-2">
			<button
				class:bg-blue-500={view === 'table'}
				class:text-white={view === 'table'}
				class="rounded px-3 py-1"
				on:click={() => (view = 'table')}
			>
				Table
			</button>
			<button
				class:bg-blue-500={view === 'grid'}
				class:text-white={view === 'grid'}
				class="rounded px-3 py-1"
				on:click={() => (view = 'grid')}
			>
				Grid
			</button>
		</div>
	</div>

	{#if view === 'table'}
		<div class="overflow-x-auto rounded-lg border dark:border-gray-700">
			<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
				<thead class="bg-gray-50 dark:bg-gray-800">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">Status</th>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">Name</th>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">Image</th>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">Ports</th>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider">Actions</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-900">
					{#each filteredContainers as container}
						<tr>
							<td class="whitespace-nowrap px-6 py-4">
								<div class="flex items-center">
									<div
										class={`h-2.5 w-2.5 rounded-full ${getStatusColor(container.State)} mr-2`}
									></div>
									{container.State}
								</div>
							</td>
							<td class="whitespace-nowrap px-6 py-4">{container.Names[0].replace('/', '')}</td>
							<td class="whitespace-nowrap px-6 py-4">{container.Image}</td>
							<td class="whitespace-nowrap px-6 py-4">
								{#each container.Ports as port}
									<div>{port.PublicPort}:{port.PrivatePort}/{port.Type}</div>
								{/each}
							</td>
							<td class="whitespace-nowrap px-6 py-4">
								<div class="flex space-x-2">
									<button class="rounded bg-blue-500 px-2 py-1 text-white hover:bg-blue-600">
										Details
									</button>
									{#if container.State === 'running'}
										<button class="rounded bg-red-500 px-2 py-1 text-white hover:bg-red-600">
											Stop
										</button>
									{:else}
										<button class="rounded bg-green-500 px-2 py-1 text-white hover:bg-green-600">
											Start
										</button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredContainers as container}
				<div class="rounded-lg border p-4 dark:border-gray-700">
					<div class="flex items-center justify-between">
						<div class="flex items-center">
							<div class={`h-2.5 w-2.5 rounded-full ${getStatusColor(container.State)} mr-2`}></div>
							<h3 class="font-medium">{container.Names[0].replace('/', '')}</h3>
						</div>
						<div class="flex space-x-2">
							<button class="rounded bg-blue-500 px-2 py-1 text-white hover:bg-blue-600">
								Details
							</button>
							{#if container.State === 'running'}
								<button class="rounded bg-red-500 px-2 py-1 text-white hover:bg-red-600">
									Stop
								</button>
							{:else}
								<button class="rounded bg-green-500 px-2 py-1 text-white hover:bg-green-600">
									Start
								</button>
							{/if}
						</div>
					</div>
					<div class="mt-2 text-sm text-gray-600 dark:text-gray-400">
						<div>{container.Image}</div>
						{#each container.Ports as port}
							<div>{port.PublicPort}:{port.PrivatePort}/{port.Type}</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
