<!-- 统一页面布局组件 -->
<script lang="ts">
	import Icon from '$lib/components/ui/Icon.svelte';
	import { Button } from '$lib/components/ui/button';
	import { goto } from '$app/navigation';

	interface Props {
		title: string;
		description?: string;
		icon?: string;
		backUrl?: string;
		backText?: string;
		actions?: Array<{
			text: string;
			icon?: string;
			variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
			onClick: () => void;
		}>;
		showStats?: boolean;
		centered?: boolean;
		children: import('svelte').Snippet;
		stats?: import('svelte').Snippet;
	}

	let {
		title,
		description,
		icon = 'layers',
		backUrl,
		backText = '返回',
		actions = [],
		showStats = false,
		centered = false,
		children,
		stats
	}: Props = $props();
</script>

<div class="min-h-screen bg-gray-50">
	<div class="container mx-auto px-4 py-6 max-w-7xl">
		<!-- 页面头部 -->
		<div class="mb-8">
			<!-- 导航栏 -->
			{#if backUrl}
				<div class="flex items-center gap-4 mb-6">
					<Button 
						variant="ghost" 
						onclick={() => goto(backUrl)} 
						class="flex items-center gap-2 hover:bg-gray-100 text-gray-600"
					>
						<Icon name="chevron-left" class="h-4 w-4" />
						{backText}
					</Button>
				</div>
			{/if}

			<!-- 页面标题 -->
			<div class={centered ? 'text-center' : 'flex items-center justify-between'}>
				<div class={centered ? 'mx-auto' : ''}>
					<div class="flex items-center gap-4 mb-4">
						<div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-100 to-blue-200 rounded-xl">
							<Icon name={icon} class="h-8 w-8 text-blue-600" />
						</div>
						<div>
							<h1 class="text-3xl font-bold text-gray-900">{title}</h1>
							{#if description}
								<p class="text-gray-600 text-lg mt-1">{description}</p>
							{/if}
						</div>
					</div>
				</div>

				<!-- 操作按钮 -->
				{#if actions.length > 0 && !centered}
					<div class="flex items-center gap-3">
						{#each actions as action}
							<Button 
								variant={action.variant || 'default'} 
								onclick={action.onClick}
								class="flex items-center gap-2"
							>
								{#if action.icon}
									<Icon name={action.icon} class="h-4 w-4" />
								{/if}
								{action.text}
							</Button>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- 统计信息 -->
		{#if showStats && stats}
			<div class="mb-8">
				{@render stats()}
			</div>
		{/if}

		<!-- 页面内容 -->
		<div>
			{@render children()}
		</div>

		<!-- 居中布局的操作按钮 -->
		{#if actions.length > 0 && centered}
			<div class="flex justify-center items-center gap-3 mt-8">
				{#each actions as action}
					<Button 
						variant={action.variant || 'default'} 
						onclick={action.onClick}
						class="flex items-center gap-2"
					>
						{#if action.icon}
							<Icon name={action.icon} class="h-4 w-4" />
						{/if}
						{action.text}
					</Button>
				{/each}
			</div>
		{/if}
	</div>
</div>