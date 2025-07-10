<!--
任务类型选择器组件
提供任务类型选择功能
-->
<script lang="ts">
	import type { TaskType } from '$lib/types/task';

	interface TaskTypeOption {
		value: TaskType;
		label: string;
		icon: string;
		description?: string;
	}

	interface Props {
		value?: TaskType;
		options: TaskTypeOption[];
		disabled?: boolean;
		class?: string;
	}

	let {
		value = $bindable('subdomain_enum' as TaskType),
		options,
		disabled = false,
		class: className = ''
	}: Props = $props();

	function handleChange(newValue: TaskType) {
		if (disabled) return;
		value = newValue;
	}
</script>

<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 {className}">
	{#each options as option}
		<div
			class="relative p-4 border rounded-lg cursor-pointer transition-colors
				{value === option.value
				? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
				: 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'}
				{disabled ? 'opacity-50 cursor-not-allowed' : ''}
			"
			onclick={() => handleChange(option.value)}
			onkeydown={(e) => {
				if (e.key === 'Enter' || e.key === ' ') {
					e.preventDefault();
					handleChange(option.value);
				}
			}}
			role="radio"
			aria-checked={value === option.value}
			tabindex={disabled ? -1 : 0}
		>
			<div class="flex items-start gap-3">
				<div class="flex-shrink-0">
					<i
						class="{option.icon} text-xl {value === option.value
							? 'text-blue-600 dark:text-blue-400'
							: 'text-gray-400'}"
					></i>
				</div>

				<div class="flex-1">
					<div class="flex items-center gap-2 mb-1">
						<h3 class="font-medium text-gray-900 dark:text-white">
							{option.label}
						</h3>

						{#if value === option.value}
							<i class="fas fa-check text-blue-600 dark:text-blue-400 text-sm"></i>
						{/if}
					</div>

					{#if option.description}
						<p class="text-sm text-gray-600 dark:text-gray-400">
							{option.description}
						</p>
					{/if}
				</div>
			</div>

			<!-- 选中状态的圆形指示器 -->
			<div class="absolute top-2 right-2">
				<div
					class="w-4 h-4 rounded-full border-2 {value === option.value
						? 'border-blue-500 bg-blue-500'
						: 'border-gray-300 dark:border-gray-600'}"
				>
					{#if value === option.value}
						<div class="w-2 h-2 bg-white rounded-full mx-auto mt-0.5"></div>
					{/if}
				</div>
			</div>
		</div>
	{/each}
</div>

<!-- 隐藏的输入框用于表单提交 -->
<input type="hidden" name="taskType" bind:value />
