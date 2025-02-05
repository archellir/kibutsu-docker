<script lang="ts">
	import { dockerClient } from '$lib/api/client';
	import type { Container } from '$lib/types/docker';

	export let container: Container;

	let logs: string[] = [];
	let isStreaming = false;
	let logStream: AsyncGenerator<string>;

	async function startLogStream() {
		isStreaming = true;
		logStream = dockerClient.streamLogs(container.Id);

		try {
			for await (const log of logStream) {
				logs = [...logs, log];
			}
		} catch (error) {
			console.error('Log streaming error:', error);
		} finally {
			isStreaming = false;
		}
	}

	function stopLogStream() {
		isStreaming = false;
		logStream?.return?.(undefined);
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<h2 class="text-2xl font-bold">{container.Names[0].replace('/', '')}</h2>
		<div class="flex space-x-2">
			{#if container.State === 'running'}
				<button
					class="rounded bg-red-500 px-3 py-1.5 text-white hover:bg-red-600"
					on:click={() => dockerClient.stopContainer(container.Id)}
				>
					Stop
				</button>
			{:else}
				<button
					class="rounded bg-green-500 px-3 py-1.5 text-white hover:bg-green-600"
					on:click={() => dockerClient.startContainer(container.Id)}
				>
					Start
				</button>
			{/if}
		</div>
	</div>

	<!-- Container Info -->
	<div class="grid grid-cols-2 gap-4 rounded-lg border p-4 dark:border-gray-700">
		<div>
			<span class="text-sm text-gray-500 dark:text-gray-400">ID</span>
			<p class="font-mono">{container.Id.substring(0, 12)}</p>
		</div>
		<div>
			<span class="text-sm text-gray-500 dark:text-gray-400">Image</span>
			<p>{container.Image}</p>
		</div>
		<div>
			<span class="text-sm text-gray-500 dark:text-gray-400">Created</span>
			<p>{new Date(container.Created * 1000).toLocaleString()}</p>
		</div>
		<div>
			<span class="text-sm text-gray-500 dark:text-gray-400">Status</span>
			<p>{container.Status}</p>
		</div>
	</div>

	<!-- Logs -->
	<div class="space-y-2">
		<div class="flex items-center justify-between">
			<h3 class="text-lg font-semibold">Logs</h3>
			<button
				class="rounded bg-blue-500 px-3 py-1.5 text-white hover:bg-blue-600"
				on:click={isStreaming ? stopLogStream : startLogStream}
			>
				{isStreaming ? 'Stop' : 'Start'} Streaming
			</button>
		</div>
		<div
			class="h-96 overflow-auto rounded-lg border bg-gray-900 p-4 font-mono text-sm text-gray-100 dark:border-gray-700"
		>
			{#each logs as log}
				<div>{log}</div>
			{/each}
		</div>
	</div>
</div>
