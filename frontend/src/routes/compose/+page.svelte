<script lang="ts">
	import { composeStore } from '$lib/stores/docker';
	import { dockerClient } from '$lib/api/client';
	import { handleApiError } from '$lib/utils/error-handlers';

	$: projects = $composeStore.data;
	$: isLoading = $composeStore.loading;

	async function handleProjectAction(projectName: string, action: 'up' | 'down') {
		try {
			if (action === 'up') {
				await dockerClient.composeUp(projectName);
			} else {
				// Add down method to DockerClient if not already present
				await dockerClient.composeDown(projectName);
			}
			await composeStore.refresh(() => dockerClient.getComposeProjects());
		} catch (error) {
			const errorDetails = await handleApiError(error);
			console.error('Project action failed:', errorDetails);
		}
	}

	function getStatusColor(status: string): string {
		switch (status.toLowerCase()) {
			case 'running':
				return 'bg-green-100 text-green-800';
			case 'partial':
				return 'bg-yellow-100 text-yellow-800';
			case 'stopped':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">Docker Compose Projects</h1>
		<button
			class="rounded-lg bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 disabled:opacity-50"
			disabled={isLoading}
			on:click={() => composeStore.refresh(() => dockerClient.getComposeProjects())}
		>
			{#if isLoading}
				<span>Refreshing...</span>
			{:else}
				<span>Refresh</span>
			{/if}
		</button>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
		{#each projects as project}
			<div class="rounded-lg border bg-white p-4 shadow dark:border-gray-700 dark:bg-gray-800">
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-semibold">{project.name}</h3>
					<span
						class={`rounded-full px-2.5 py-0.5 text-xs font-medium ${getStatusColor(project.status)}`}
					>
						{project.status}
					</span>
				</div>

				<div class="mt-4 space-y-2">
					<div class="text-sm text-gray-600 dark:text-gray-400">
						<span class="font-medium">Services:</span>
						{project.services.join(', ')}
					</div>
				</div>

				<div class="mt-4 flex space-x-2">
					{#if project.status !== 'running'}
						<button
							class="flex-1 rounded bg-green-500 px-3 py-1.5 text-white hover:bg-green-600"
							on:click={() => handleProjectAction(project.name, 'up')}
						>
							Start
						</button>
					{:else}
						<button
							class="flex-1 rounded bg-red-500 px-3 py-1.5 text-white hover:bg-red-600"
							on:click={() => handleProjectAction(project.name, 'down')}
						>
							Stop
						</button>
					{/if}
					<button
						class="rounded bg-blue-500 px-3 py-1.5 text-white hover:bg-blue-600"
						on:click={() => (window.location.href = `/compose/${project.name}`)}
					>
						Details
					</button>
				</div>
			</div>
		{/each}
	</div>

	{#if projects.length === 0}
		<div
			class="rounded-lg border border-gray-200 bg-white p-8 text-center dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="text-gray-600 dark:text-gray-400">No Docker Compose projects found</p>
			<p class="mt-2 text-sm text-gray-500 dark:text-gray-500">
				Add a docker-compose.yml file to the compose directory to get started
			</p>
		</div>
	{/if}
</div>
