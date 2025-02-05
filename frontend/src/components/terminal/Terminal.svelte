<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Terminal as Xterm } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import { WebglAddon } from '@xterm/addon-webgl';
	import { theme } from '$lib/stores/theme';

	// Props
	export let websocketUrl = 'ws://localhost:3000/terminal';
	export let fontSize = 14;

	// Component state
	let terminal: Xterm;
	let fitAddon: FitAddon;
	let websocket: WebSocket;
	let terminalElement: HTMLDivElement;
	let commandHistory: string[] = [];
	let historyIndex = -1;
	let currentCommand = '';

	// Terminal themes
	const themes = {
		light: {
			background: '#ffffff',
			foreground: '#2e3440',
			cursor: '#434c5e'
		},
		dark: {
			background: '#2e3440',
			foreground: '#d8dee9',
			cursor: '#88c0d0'
		}
	};

	onMount(() => {
		initializeTerminal();
		connectWebSocket();
		setupResizeHandler();
	});

	onDestroy(() => {
		websocket?.close();
		terminal?.dispose();
		window.removeEventListener('resize', handleResize);
	});

	function initializeTerminal() {
		terminal = new Xterm({
			fontSize,
			fontFamily: 'Menlo, Monaco, "Courier New", monospace',
			theme: themes[$theme],
			cursorBlink: true,
			scrollback: 1000
		});

		fitAddon = new FitAddon();
		terminal.loadAddon(fitAddon);
		terminal.loadAddon(new WebLinksAddon());
		terminal.loadAddon(new WebglAddon());

		terminal.onData(handleTerminalData);
		terminal.onKey(handleTerminalKey);

		terminal.open(terminalElement);
		fitAddon.fit();
	}

	function connectWebSocket() {
		websocket = new WebSocket(websocketUrl);

		websocket.onopen = () => {
			terminal.writeln('Connected to terminal server');
		};

		websocket.onmessage = (event) => {
			terminal.write(event.data);
		};

		websocket.onclose = () => {
			terminal.writeln('\r\nConnection closed. Reconnecting in 3s...');
			setTimeout(connectWebSocket, 3000);
		};
	}

	function handleTerminalData(data: string) {
		if (websocket?.readyState === WebSocket.OPEN) {
			websocket.send(data);
		}
	}

	function handleTerminalKey(event: { key: string; domEvent: KeyboardEvent }) {
		const { domEvent } = event;

		if (domEvent.key === 'ArrowUp') {
			if (historyIndex < commandHistory.length - 1) {
				historyIndex++;
				currentCommand = commandHistory[historyIndex];
				terminal.write('\x1b[2K\r$ ' + currentCommand);
			}
		} else if (domEvent.key === 'ArrowDown') {
			if (historyIndex > 0) {
				historyIndex--;
				currentCommand = commandHistory[historyIndex];
				terminal.write('\x1b[2K\r$ ' + currentCommand);
			}
		}
	}

	function handleResize() {
		fitAddon?.fit();
	}

	function setupResizeHandler() {
		window.addEventListener('resize', handleResize);
		handleResize();
	}

	// Toolbar actions
	function clearScreen() {
		terminal.clear();
	}

	function resetConnection() {
		websocket?.close();
		connectWebSocket();
	}

	function changeFontSize(delta: number) {
		fontSize = Math.max(8, Math.min(24, fontSize + delta));
		terminal.options.fontSize = fontSize;
		fitAddon.fit();
	}

	$: if (terminal && $theme) {
		terminal.options.theme = themes[$theme];
	}
</script>

<div class="flex h-full flex-col rounded-lg border dark:border-gray-700">
	<div class="flex items-center justify-between border-b p-2 dark:border-gray-700">
		<div class="flex items-center space-x-2">
			<button
				class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
				on:click={clearScreen}
				title="Clear Screen"
			>
				🗑️
			</button>
			<button
				class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
				on:click={resetConnection}
				title="Reset Connection"
			>
				🔄
			</button>
		</div>
		<div class="flex items-center space-x-2">
			<button
				class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
				on:click={() => changeFontSize(-1)}
				title="Decrease Font Size"
			>
				A-
			</button>
			<span class="text-sm">{fontSize}px</span>
			<button
				class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
				on:click={() => changeFontSize(1)}
				title="Increase Font Size"
			>
				A+
			</button>
		</div>
	</div>

	<div bind:this={terminalElement} class="flex-1 overflow-hidden"></div>
</div>

<style>
	:global(.xterm) {
		height: 100%;
		padding: 8px;
	}
</style>
