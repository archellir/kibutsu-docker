import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
	plugins: [sveltekit()],

	test: {
		include: ['src/**/*.{test,spec}.{js,ts}']
	},

	server: {
		proxy: {
			'/api': 'http://localhost:8080',
			'/docker': {
				target: 'ws://localhost:8080',
				ws: true
			}
		}
	}
});
