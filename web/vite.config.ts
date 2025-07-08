import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		fs: {
			allow: ['static']
		},
		proxy: {
			'/api': {
				target: 'http://localhost:8090',
				changeOrigin: true,
				secure: false,
				ws: true,
				rewrite: (path) => path.replace(/^\/api/, '/api')
			}
		}
	},
	resolve: {
		alias: {
			$lib: path.resolve('./src/lib'),
			$components: path.resolve('./src/lib/components')
		}
	},
	optimizeDeps: {
		exclude: ['@internationalized/date']
	},
	build: {
		target: 'esnext'
	}
});
