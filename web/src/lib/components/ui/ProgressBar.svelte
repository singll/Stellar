<!--
进度条组件
显示任务或操作的进度
-->
<script lang="ts">
	interface Props {
		value: number; // 0-100
		max?: number;
		color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'indigo' | 'gray';
		size?: 'sm' | 'md' | 'lg';
		showLabel?: boolean;
		label?: string;
		striped?: boolean;
		animated?: boolean;
	}

	let {
		value,
		max = 100,
		color = 'blue',
		size = 'md',
		showLabel = false,
		label,
		striped = false,
		animated = false
	}: Props = $props();

	// 计算百分比
	let percentage = $derived(Math.min(Math.max((value / max) * 100, 0), 100));

	// 获取颜色类名
	function getColorClass(color: string) {
		const colorMap: Record<string, string> = {
			blue: 'bg-blue-500',
			green: 'bg-green-500',
			yellow: 'bg-yellow-500',
			red: 'bg-red-500',
			purple: 'bg-purple-500',
			indigo: 'bg-indigo-500',
			gray: 'bg-gray-500'
		};
		return colorMap[color] || colorMap.blue;
	}

	// 获取尺寸类名
	function getSizeClass(size: string) {
		const sizeMap: Record<string, string> = {
			sm: 'h-1',
			md: 'h-2',
			lg: 'h-3'
		};
		return sizeMap[size] || sizeMap.md;
	}

	let colorClass = $derived(getColorClass(color));
	let sizeClass = $derived(getSizeClass(size));
</script>

<div class="w-full">
	{#if showLabel || label}
		<div class="flex justify-between items-center mb-1">
			<span class="text-sm font-medium text-gray-700 dark:text-gray-300">
				{label || '进度'}
			</span>
			<span class="text-sm text-gray-600 dark:text-gray-400">
				{Math.round(percentage)}%
			</span>
		</div>
	{/if}

	<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full {sizeClass}">
		<div
			class="{colorClass} {sizeClass} rounded-full transition-all duration-300 ease-out {striped
				? 'bg-stripes'
				: ''} {animated ? 'animate-pulse' : ''}"
			style="width: {percentage}%"
		></div>
	</div>
</div>

<style>
	.bg-stripes {
		background-image: linear-gradient(
			45deg,
			rgba(255, 255, 255, 0.2) 25%,
			transparent 25%,
			transparent 50%,
			rgba(255, 255, 255, 0.2) 50%,
			rgba(255, 255, 255, 0.2) 75%,
			transparent 75%,
			transparent
		);
		background-size: 1rem 1rem;
	}
</style>
