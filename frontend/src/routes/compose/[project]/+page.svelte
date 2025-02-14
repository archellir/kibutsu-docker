<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { dockerClient } from '$lib/api/client';
	import { handleApiError } from '$lib/utils/error-handlers';
	import { wsManager } from '$lib/websocket/manager';

	interface ServiceConfig {
		image: string;
		deploy?: {
			replicas: number;
		};
	}

	interface ProjectConfig {
		services: Record<string, ServiceConfig>;
	}

	let projectName: string;
	let projectConfig: ProjectConfig | null = null;
	let loading = true;
	let error: string | null = null;
	// This object will hold the current input value for scaling each service
	let scaleInputs: Record<string, number> = {};

	// Extract the project name from the URL parameters.
	$: projectName = $page.params.project;

	async function loadProject() {
		loading = true;
		error = null;
		try {
			const res = await fetch(`/api/compose/projects/${projectName}`);
			if (!res.ok) {
				throw new Error('Failed to load project details');
			}
			projectConfig = (await res.json()) as ProjectConfig;
			// Initialize the scale inputs: use the deploy replicas if provided or default to 1.
			for (const [service, config] of Object.entries(projectConfig.services)) {
				scaleInputs[service] = config.deploy?.replicas || 1;
			}
		} catch (err: any) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function scaleService(service: string) {
		try {
			const replicas = scaleInputs[service];
			await dockerClient.scaleService(projectName, service, replicas);
			alert(`Scaled service "${service}" to ${replicas} replicas.`);
		} catch (err: any) {
			const details = await handleApiError(err);
			console.error(`Failed to scale service ${service}:`, details);
		}
	}

	function viewLogs() {
		const ws = wsManager.connectToComposeLogs(projectName);
		// For simplicity, log messages to the console.
		ws.onmessage = (event) => {
			console.log('Compose logs:', event.data);
		};
	}

	onMount(() => {
		loadProject();
	});
</script>

{#if loading}
	<p>Loading project details...</p>
{:else if error}
	<p class="text-red-500">Error: {error}</p>
{:else}
	<h1 class="mb-4 text-2xl font-bold">Project: {projectName}</h1>
	<section>
		<h2 class="mb-2 text-xl">Services</h2>
		<ul>
			{#each Object.entries(projectConfig!.services) as [serviceName, service]}
				<li class="mb-2 rounded border p-4">
					<div class="flex items-center justify-between">
						<div>
							<h3 class="font-semibold">{serviceName}</h3>
							<p>Image: {service.image}</p>
							<p>Default Replicas: {service.deploy?.replicas || 1}</p>
						</div>
						<div class="flex items-center space-x-2">
							<input
								type="number"
								min="1"
								class="w-16 border p-1"
								bind:value={scaleInputs[serviceName]}
							/>
							<button
								class="rounded bg-blue-500 px-3 py-1 text-white"
								on:click={() => scaleService(serviceName)}
							>
								Scale
							</button>
							<button class="rounded bg-green-500 px-3 py-1 text-white" on:click={viewLogs}>
								Logs
							</button>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	</section>
{/if}
