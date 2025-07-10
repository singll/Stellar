<!--
按钮组件
提供多种样式和大小的按钮
-->
<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	type ButtonVariant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
	type ButtonSize = 'sm' | 'md' | 'lg';

	interface Props {
		variant?: ButtonVariant;
		size?: ButtonSize;
		disabled?: boolean;
		class?: string;
		href?: string;
		type?: 'button' | 'submit' | 'reset';
		onclick?: (event: MouseEvent) => void;
		children: Snippet;
	}

	let {
		variant = 'default',
		size = 'md',
		disabled = false,
		class: className = '',
		href,
		type = 'button',
		onclick,
		children
	}: Props = $props();

	const variants = {
		default: 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500',
		destructive: 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500',
		outline: 'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 focus:ring-blue-500',
		secondary: 'bg-gray-100 text-gray-900 hover:bg-gray-200 focus:ring-gray-500',
		ghost: 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 focus:ring-gray-500',
		link: 'text-blue-600 underline-offset-4 hover:underline focus:ring-blue-500'
	};

	const sizes = {
		sm: 'px-3 py-1.5 text-sm',
		md: 'px-4 py-2 text-sm',
		lg: 'px-6 py-3 text-base'
	};

	const baseClasses =
		'inline-flex items-center justify-center rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';

	let buttonClasses = $derived(cn(baseClasses, variants[variant], sizes[size], className));

	function handleClick(event: MouseEvent) {
		if (disabled) {
			event.preventDefault();
			return;
		}
		onclick?.(event);
	}
</script>

{#if href}
	<a {href} class={buttonClasses} role="button" onclick={handleClick}>
		{@render children()}
	</a>
{:else}
	<button {type} {disabled} class={buttonClasses} onclick={handleClick}>
		{@render children()}
	</button>
{/if}
