<script lang="ts">
	import { systemStore } from '$lib/stores/docker';
	import { theme } from '$lib/stores/theme';

	let searchQuery = '';

	$: systemInfo = $systemStore.data;
	$: isLoading = $systemStore.loading;
</script>

<nav class="border-b border-gray-200 bg-white px-4 py-2.5 dark:border-gray-700 dark:bg-gray-800">
	<div class="flex items-center justify-between">
		<!-- System Status -->
		<div class="flex items-center space-x-4">
			<div class="flex items-center space-x-2">
				<div class={`h-2 w-2 rounded-full ${isLoading ? 'bg-yellow-400' : 'bg-green-400'}`}></div>
				<span class="text-sm text-gray-600 dark:text-gray-300">
					{isLoading ? 'Updating...' : 'System Online'}
				</span>
			</div>
			<div class="text-sm text-gray-600 dark:text-gray-300">
				Docker {systemInfo?.version || '---'}
			</div>
		</div>

		<!-- Search Bar -->
		<div class="flex-1 px-8">
			<input
				type="search"
				bind:value={searchQuery}
				placeholder="Search containers and images..."
				class="w-full rounded-lg border border-gray-300 bg-gray-50 px-4 py-2 dark:border-gray-600 dark:bg-gray-700"
			/>
		</div>

		<!-- Actions -->
		<div class="flex items-center space-x-4">
			<button class="rounded-lg bg-blue-500 px-4 py-2 text-white hover:bg-blue-600">
				Quick Actions
			</button>
			<button
				class="rounded-lg bg-gray-200 p-2 dark:bg-gray-700"
				on:click={() => theme.set($theme === 'dark' ? 'light' : 'dark')}
			>
				{$theme === 'dark' ? '🌞' : '🌙'}
			</button>
		</div>
	</div>
</nav>
