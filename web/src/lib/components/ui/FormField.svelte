<!--
表单字段组件
-->
<script lang="ts">
	import { cn } from '$lib/utils';
	import Label from './label/label.svelte';

	interface Props {
		label?: string;
		error?: string;
		required?: boolean;
		description?: string;
		class?: string;
		children?: any;
	}

	let {
		label,
		error,
		required = false,
		description,
		class: className = '',
		children
	}: Props = $props();
</script>

<div class={cn('space-y-2', className)}>
	{#if label}
		<Label
			class={cn(
				'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
				required && 'after:content-["*"] after:ml-0.5 after:text-red-500'
			)}
		>
			{label}
		</Label>
	{/if}

	{@render children?.()}

	{#if description}
		<p class="text-sm text-muted-foreground">{description}</p>
	{/if}

	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}
</div>
