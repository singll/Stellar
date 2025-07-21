<!-- 统计信息网格组件 -->
<script lang="ts">
	import Icon from '$lib/components/ui/Icon.svelte';

	interface StatItem {
		title: string;
		value: number | string;
		icon: string;
		color: 'blue' | 'green' | 'purple' | 'orange' | 'red' | 'gray' | 'teal' | 'indigo';
		trend?: {
			value: number;
			isPositive: boolean;
		};
	}

	interface Props {
		stats: StatItem[];
		columns?: 2 | 3 | 4;
	}

	let { stats, columns = 4 }: Props = $props();

	function getColorClasses(color: string) {
		const colors = {
			blue: 'from-blue-50 to-blue-100 text-blue-600 text-blue-900',
			green: 'from-green-50 to-green-100 text-green-600 text-green-900',
			purple: 'from-purple-50 to-purple-100 text-purple-600 text-purple-900',
			orange: 'from-orange-50 to-orange-100 text-orange-600 text-orange-900',
			red: 'from-red-50 to-red-100 text-red-600 text-red-900',
			gray: 'from-gray-50 to-gray-100 text-gray-600 text-gray-900',
			teal: 'from-teal-50 to-teal-100 text-teal-600 text-teal-900',
			indigo: 'from-indigo-50 to-indigo-100 text-indigo-600 text-indigo-900'
		};
		return colors[color] || colors.blue;
	}

	function getGridCols(cols: number) {
		const gridCols = {
			2: 'grid-cols-1 md:grid-cols-2',
			3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
			4: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4'
		};
		return gridCols[cols] || gridCols[4];
	}
</script>

<div class="grid gap-6 {getGridCols(columns)}">
	{#each stats as stat}
		{@const colorClasses = getColorClasses(stat.color).split(' ')}
		<div class="bg-gradient-to-br {colorClasses[0]} {colorClasses[1]} p-6 rounded-xl border border-gray-100 hover:shadow-md transition-all duration-200">
			<div class="flex items-center justify-between mb-4">
				<div class="flex items-center gap-3">
					<div class="p-2 rounded-lg bg-white/70">
						<Icon name={stat.icon} class="h-5 w-5 {colorClasses[2]}" />
					</div>
					<span class="font-medium {colorClasses[3]} text-sm">{stat.title}</span>
				</div>
				
				{#if stat.trend}
					<div class="flex items-center gap-1 text-xs">
						<Icon 
							name={stat.trend.isPositive ? "trending-up" : "trending-down"} 
							class="h-3 w-3 {stat.trend.isPositive ? 'text-green-600' : 'text-red-600'}"
						/>
						<span class="{stat.trend.isPositive ? 'text-green-600' : 'text-red-600'}">
							{Math.abs(stat.trend.value)}%
						</span>
					</div>
				{/if}
			</div>
			
			<p class="text-3xl font-bold {colorClasses[3]}">{stat.value}</p>
		</div>
	{/each}
</div>