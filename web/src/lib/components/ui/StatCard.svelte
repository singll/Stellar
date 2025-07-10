<!--
统计卡片组件
显示统计数据，支持图标、颜色、脉冲效果等
-->
<script lang="ts">
	interface Props {
		title: string;
		value: number | string;
		icon?: string;
		color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'indigo' | 'gray' | 'orange';
		subtitle?: string;
		pulse?: boolean;
		onclick?: () => void;
	}

	let { title, value, icon, color = 'blue', subtitle, pulse = false, onclick }: Props = $props();

	// 获取颜色类名
	function getColorClasses(color: string) {
		const colorMap = {
			blue: {
				bg: 'bg-blue-50 dark:bg-blue-900/20',
				icon: 'text-blue-600 dark:text-blue-400',
				text: 'text-blue-800 dark:text-blue-200'
			},
			green: {
				bg: 'bg-green-50 dark:bg-green-900/20',
				icon: 'text-green-600 dark:text-green-400',
				text: 'text-green-800 dark:text-green-200'
			},
			yellow: {
				bg: 'bg-yellow-50 dark:bg-yellow-900/20',
				icon: 'text-yellow-600 dark:text-yellow-400',
				text: 'text-yellow-800 dark:text-yellow-200'
			},
			red: {
				bg: 'bg-red-50 dark:bg-red-900/20',
				icon: 'text-red-600 dark:text-red-400',
				text: 'text-red-800 dark:text-red-200'
			},
			purple: {
				bg: 'bg-purple-50 dark:bg-purple-900/20',
				icon: 'text-purple-600 dark:text-purple-400',
				text: 'text-purple-800 dark:text-purple-200'
			},
			indigo: {
				bg: 'bg-indigo-50 dark:bg-indigo-900/20',
				icon: 'text-indigo-600 dark:text-indigo-400',
				text: 'text-indigo-800 dark:text-indigo-200'
			},
			gray: {
				bg: 'bg-gray-50 dark:bg-gray-900/20',
				icon: 'text-gray-600 dark:text-gray-400',
				text: 'text-gray-800 dark:text-gray-200'
			},
			orange: {
				bg: 'bg-orange-50 dark:bg-orange-900/20',
				icon: 'text-orange-600 dark:text-orange-400',
				text: 'text-orange-800 dark:text-orange-200'
			}
		};
		return colorMap[color] || colorMap.blue;
	}

	let colorClasses = $derived(getColorClasses(color));
</script>

<div
	class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 {onclick
		? 'cursor-pointer hover:shadow-md transition-shadow'
		: ''}"
	{onclick}
	role={onclick ? 'button' : undefined}
	tabindex={onclick ? 0 : undefined}
>
	<div class="flex items-center justify-between">
		<div class="flex-1">
			<div class="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
				{title}
			</div>
			<div class="text-2xl font-bold text-gray-900 dark:text-white {pulse ? 'animate-pulse' : ''}">
				{value}
			</div>
			{#if subtitle}
				<div class="text-xs text-gray-500 dark:text-gray-500 mt-1">
					{subtitle}
				</div>
			{/if}
		</div>

		{#if icon}
			<div
				class="w-10 h-10 rounded-lg {colorClasses.bg} flex items-center justify-center {pulse
					? 'animate-pulse'
					: ''}"
			>
				<i class="{icon} {colorClasses.icon}"></i>
			</div>
		{/if}
	</div>
</div>
