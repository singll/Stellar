<!--
徽章组件
显示状态或分类信息的小标签
-->
<script lang="ts">
	import { cn } from '$lib/utils';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';

	interface Props {
		variant?: BadgeVariant;
		class?: string;
		children?: any;
	}

	let { variant = 'default', class: className = '', children }: Props = $props();

	const variants = {
		default: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
		secondary: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
		destructive: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
		outline: 'border border-gray-200 text-gray-800 dark:border-gray-700 dark:text-gray-200'
	};

	const baseClasses = 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium';

	let badgeClasses = $derived(cn(baseClasses, variants[variant], className));
</script>

<span class={badgeClasses}>
	{@render children?.()}
</span>
