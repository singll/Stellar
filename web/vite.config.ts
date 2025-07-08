import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: '0.0.0.0', // 前端也监听所有接口
		port: 5173,
		fs: {
			allow: ['static']
		},
		proxy: {
			'/api': {
				target: 'http://0.0.0.0:8090', // 后端服务地址
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
