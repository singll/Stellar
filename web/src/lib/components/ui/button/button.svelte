<script lang="ts" module>
	import type { HTMLButtonAttributes, HTMLAnchorAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	export type ButtonVariant =
		| 'default'
		| 'destructive'
		| 'outline'
		| 'secondary'
		| 'ghost'
		| 'link';
	export type ButtonSize = 'default' | 'sm' | 'lg' | 'icon';

	export type ButtonProps = {
		variant?: ButtonVariant;
		size?: ButtonSize;
		children?: Snippet;
		href?: string;
		class?: string;
		builders?: Array<{ [key: string]: any }>;
		[key: string]: any;
	};

	// 简化的按钮样式变体
	export function getButtonClasses(
		variant: ButtonVariant = 'default',
		size: ButtonSize = 'default'
	): string {
		const baseClasses =
			'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium outline-none transition-all disabled:pointer-events-none disabled:opacity-50 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2';

		const variantClasses = {
			default: 'bg-blue-600 text-white hover:bg-blue-700 shadow-sm',
			destructive: 'bg-red-600 text-white hover:bg-red-700 shadow-sm',
			outline: 'border border-gray-300 bg-white hover:bg-gray-50 text-gray-900',
			secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 shadow-sm',
			ghost: 'hover:bg-gray-100 text-gray-900',
			link: 'text-blue-600 underline-offset-4 hover:underline'
		};

		const sizeClasses = {
			default: 'h-9 px-4 py-2',
			sm: 'h-8 px-3 text-sm',
			lg: 'h-10 px-6',
			icon: 'h-9 w-9'
		};

		return `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]}`;
	}
</script>

<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		variant = 'default',
		size = 'default',
		class: className,
		children,
		href,
		builders = [],
		...restProps
	}: ButtonProps = $props();

	// Merge builder props
	let builderProps = {};
	builders.forEach(builder => {
		builderProps = { ...builderProps, ...builder };
	});

	const baseProps = {
		class: cn(getButtonClasses(variant, size), className),
		...builderProps,
		...restProps
	};
</script>

{#if href}
	<a {href} {...baseProps}>
		{#if children}
			{@render children()}
		{/if}
	</a>
{:else}
	<button {...baseProps}>
		{#if children}
			{@render children()}
		{/if}
	</button>
{/if}
