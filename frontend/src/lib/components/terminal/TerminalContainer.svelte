<script lang="ts">
	import { notifications } from '$lib/stores/notification';
	import Terminal from './Terminal.svelte';
	import { createEventDispatcher } from 'svelte';

	export let websocketUrl = 'ws://localhost:3000/terminal';
	const dispatch = createEventDispatcher();

	let terminal: Terminal;
	let lastCommand = '';
	let lastError: string | null = null;

	function handleConnect() {
		notifications.success('Terminal connected', {
			title: 'Connection Established',
			autoDismiss: true
		});
		dispatch('connect');
	}

	function handleDisconnect() {
		notifications.warning('Terminal disconnected', {
			title: 'Connection Lost',
			autoDismiss: true
		});
		dispatch('disconnect');
	}

	function handleCommand(event: CustomEvent<{ command: string }>) {
		lastCommand = event.detail.command;
		notifications.info(`Executing: ${lastCommand}`, {
			autoDismiss: true,
			priority: 'low'
		});
		dispatch('command', event.detail);
	}

	function handleCommandError(event: CustomEvent<{ command: string; error: string }>) {
		lastError = event.detail.error;
		notifications.error(event.detail.error, {
			title: `Command failed: ${event.detail.command}`,
			priority: 'high'
		});
		dispatch('commandError', event.detail);
	}

	function handleResize(event: CustomEvent<{ cols: number; rows: number }>) {
		notifications.info(`Terminal resized to ${event.detail.cols}x${event.detail.rows}`, {
			autoDismiss: true,
			priority: 'low'
		});
		dispatch('resize', event.detail);
	}

	function handleExit(event: CustomEvent<{ code: number }>) {
		notifications.info(`Terminal session ended with code ${event.detail.code}`, {
			title: 'Session Ended',
			priority: 'medium'
		});
		dispatch('exit', event.detail);
	}
</script>

<div class="relative flex h-full flex-col">
	{#if lastError}
		<div class="absolute inset-x-0 top-0 z-10 p-2">
			<div
				class="rounded-lg border border-red-400 bg-red-100 p-2 text-sm dark:bg-red-900 dark:bg-opacity-20"
			>
				{lastError}
			</div>
		</div>
	{/if}

	<Terminal
		bind:this={terminal}
		{websocketUrl}
		on:connect={handleConnect}
		on:disconnect={handleDisconnect}
		on:command={handleCommand}
		on:commandError={handleCommandError}
		on:resize={handleResize}
		on:exit={handleExit}
	/>
</div>
