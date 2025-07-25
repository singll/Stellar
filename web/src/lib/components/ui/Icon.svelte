<!--
简单的图标组件
使用 SVG 图标，避免依赖第三方图标库
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';

	interface Props extends HTMLAttributes<SVGElement> {
		name: string;
		size?: number | string;
		class?: string;
		children?: Snippet;
	}

	let { name, size = 24, class: className = '', children, ...restProps }: Props = $props();

	const icons: Record<string, string> = {
		menu: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>',
		x: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 6L6 18M6 6l12 12"/>',
		home: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="9,22 9,12 15,12 15,22"/>',
		folder:
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>',
		layers:
			'<polygon stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="12,2 2,7 12,12 22,7 12,2"/><polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="2,17 12,22 22,17"/><polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="2,12 12,17 22,12"/>',
		activity:
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="22,12 18,12 15,21 9,3 6,12 2,12"/>',
		settings:
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="3"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>',
		server:
			'<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="2" y="3" width="20" height="4" rx="1"/><rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="2" y="9" width="20" height="4" rx="1"/><rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="2" y="15" width="20" height="4" rx="1"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="6" y1="5" x2="6.01" y2="5"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="6" y1="11" x2="6.01" y2="11"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="6" y1="17" x2="6.01" y2="17"/>',
		user: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="7" r="4"/>',
		sun: '<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="5"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 1v2m0 16v2m11-9h-2M4 12H2m15.364-6.364l-1.414 1.414M6.05 6.05L4.636 4.636m12.728 12.728l1.414 1.414M6.05 17.95l-1.414 1.414"/>',
		moon: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>',
		'log-out':
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4m7 14l5-5-5-5m5 5H9"/>',
		'chevron-down':
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="6,9 12,15 18,9"/>',
		'chevron-up':
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="18,15 12,9 6,15"/>',
		'chevron-left':
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="15,18 9,12 15,6"/>',
		'chevron-right':
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="9,18 15,12 9,6"/>',
		plus: '<line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="12" y1="5" x2="12" y2="19"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="5" y1="12" x2="19" y2="12"/>',
		minus:
			'<line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="5" y1="12" x2="19" y2="12"/>',
		edit: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>',
		trash:
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="3,6 5,6 21,6"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
		search:
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="11" cy="11" r="8"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35"/>',
		filter:
			'<polygon stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="22,3 2,3 10,12.46 10,19 14,21 14,12.46 22,3"/>',
		refresh:
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="1,4 1,10 7,10"/><polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="23,20 23,14 17,14"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/>',
		download:
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4m7-10v12m-4-4l4 4 4-4"/>',
		upload:
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4m11-10v12m-4-8l4-4 4 4"/>',
		eye: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="3"/>',
		'eye-off':
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19M14.12 14.12a3 3 0 1 1-4.24-4.24m-1.07 1.07l14.14 14.14"/>',
		play: '<polygon stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="5,3 19,12 5,21 5,3"/>',
		pause:
			'<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="6" y="4" width="4" height="16"/><rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="14" y="4" width="4" height="16"/>',
		stop: '<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="6" y="6" width="12" height="12"/>',
		check:
			'<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="20,6 9,17 4,12"/>',
		'alert-triangle':
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4m0 4h.01"/>',
		info: '<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="10"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="12" y1="16" x2="12" y2="12"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="12" y1="8" x2="12.01" y2="8"/>',
		'more-horizontal':
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="1"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="19" cy="12" r="1"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="5" cy="12" r="1"/>',
		'more-vertical':
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="1"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="5" r="1"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="19" r="1"/>',
		'external-link':
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6m4-3h6v6m-11 5L21 3"/>',
		copy: '<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="9" y="9" width="13" height="13" rx="2" ry="2"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>',
		calendar:
			'<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="3" y="4" width="18" height="18" rx="2" ry="2"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="16" y1="2" x2="16" y2="6"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="8" y1="2" x2="8" y2="6"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="3" y1="10" x2="21" y2="10"/>',
		clock:
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="10"/><polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="12,6 12,12 16,14"/>',
		globe:
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="10"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="2" y1="12" x2="22" y2="12"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>',
		shield:
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>',
		lock: '<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="3" y="11" width="18" height="10" rx="2" ry="2"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="16" r="1"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 11V7a5 5 0 0 1 10 0v4"/>',
		unlock:
			'<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="3" y="11" width="18" height="10" rx="2" ry="2"/><circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="16" r="1"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 11V7a5 5 0 0 1 9.9-1"/>',
		wifi: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12.55a11 11 0 0 1 14.08 0M1.42 9a16 16 0 0 1 21.16 0m-6.7 3.87a5 5 0 0 1 6.26 0M9.53 16.83a3 3 0 0 1 4.94 0"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="12" y1="20" x2="12.01" y2="20"/>',
		database:
			'<ellipse stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="5" rx="9" ry="3"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3m18-7v14c0 1.66-4 3-9 3s-9-1.34-9-3V5"/>',
		'check-circle':
			'<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>',
		'x-circle':
			'<circle stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" cx="12" cy="12" r="10"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="15" y1="9" x2="9" y2="15"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="9" y1="9" x2="15" y2="15"/>',
		list: '<line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="8" y1="6" x2="21" y2="6"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="8" y1="12" x2="21" y2="12"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="8" y1="18" x2="21" y2="18"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="3" y1="6" x2="3.01" y2="6"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="3" y1="12" x2="3.01" y2="12"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="3" y1="18" x2="3.01" y2="18"/>',
		briefcase: '<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="2" y="7" width="20" height="14" rx="2" ry="2"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>',
		inbox: '<polyline stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="22,12 16,12 14,15 10,15 8,12 2,12"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/>',
		link: '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>',
		smartphone: '<rect stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x="5" y="2" width="14" height="20" rx="2" ry="2"/><line stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" x1="12" y1="18" x2="12.01" y2="18"/>'
	};
</script>

<svg width={size} height={size} viewBox="0 0 24 24" fill="none" class={className} {...restProps}>
	{@html icons[name] || icons['info']}
	{#if children}
		{@render children()}
	{/if}
</svg>
