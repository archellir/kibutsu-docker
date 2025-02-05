<script lang="ts">
	import { systemStore } from '$lib/stores/docker';

	$: systemInfo = $systemStore.data;

	const navigation = [
		{ name: 'Dashboard', href: '/', icon: '📊' },
		{ name: 'Containers', href: '/containers', icon: '📦' },
		{ name: 'Images', href: '/images', icon: '💿' },
		{ name: 'Compose', href: '/compose', icon: '🔧' },
		{ name: 'Settings', href: '/settings', icon: '⚙️' }
	];
</script>

<aside
	class="flex h-full w-64 flex-col border-r border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800"
>
	<!-- Logo -->
	<div class="p-4">
		<h1 class="text-xl font-bold">Docker UI</h1>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 space-y-1 px-2 py-4">
		{#each navigation as item}
			<a
				href={item.href}
				class="flex items-center rounded-lg px-4 py-2 text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
			>
				<span class="mr-3">{item.icon}</span>
				{item.name}
			</a>
		{/each}
	</nav>

	<!-- System Stats -->
	<div class="border-t border-gray-200 p-4 dark:border-gray-700">
		<div class="space-y-3">
			<div class="flex justify-between">
				<span class="text-gray-600 dark:text-gray-400">Containers</span>
				<span class="font-medium">{systemInfo?.containers || 0}</span>
			</div>
			<div class="flex justify-between">
				<span class="text-gray-600 dark:text-gray-400">Images</span>
				<span class="font-medium">{systemInfo?.images || 0}</span>
			</div>
			<div class="flex justify-between">
				<span class="text-gray-600 dark:text-gray-400">Memory Usage</span>
				<span class="font-medium">{systemInfo?.memoryUsage || 0}%</span>
			</div>
			<div class="flex justify-between">
				<span class="text-gray-600 dark:text-gray-400">CPU Usage</span>
				<span class="font-medium">{systemInfo?.cpuUsage || 0}%</span>
			</div>
		</div>
	</div>
</aside>
