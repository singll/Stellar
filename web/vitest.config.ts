/// <reference types="vitest" />
import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { fileURLToPath } from 'url';

export default defineConfig({
	plugins: [svelte()],
	resolve: {
		alias: {
			$lib: fileURLToPath(new URL('./src/lib', import.meta.url))
		}
	},
	test: {
		include: ['src/**/*.{test,spec}.{js,ts,jsx,tsx}'],
		globals: true,
		environment: 'jsdom',
		setupFiles: ['./src/setupTests.ts'],
		deps: {
			inline: ['@sveltejs/kit']
		}
	}
});
