<!-- 统一数据列表组件 -->
<script lang="ts">
	import Icon from '$lib/components/ui/Icon.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';

	interface Column {
		key: string;
		title: string;
		sortable?: boolean;
		render?: (value: any, row: any) => any;
		width?: string;
	}

	interface Props {
		title: string;
		description?: string;
		data: any[];
		columns: Column[];
		loading?: boolean;
		searchable?: boolean;
		searchPlaceholder?: string;
		searchValue?: string;
		onSearch?: (query: string) => void;
		onRowClick?: (row: any) => void;
		emptyStateTitle?: string;
		emptyStateDescription?: string;
		emptyStateAction?: {
			text: string;
			icon?: string;
			onClick: () => void;
		};
		actions?: Array<{
			text: string;
			icon?: string;
			variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
			onClick: () => void;
		}>;
	}

	let {
		title,
		description,
		data,
		columns,
		loading = false,
		searchable = true,
		searchPlaceholder = '搜索...',
		searchValue = '',
		onSearch,
		onRowClick,
		emptyStateTitle = '暂无数据',
		emptyStateDescription = '目前没有任何数据',
		emptyStateAction,
		actions = []
	}: Props = $props();

	let searchQuery = $state(searchValue);
	let filteredData = $state<any[]>([]);

	// 同步外部搜索值
	$effect(() => {
		if (searchValue !== undefined) {
			searchQuery = searchValue;
		}
	});

	// 搜索过滤 - 如果有外部搜索控制，则直接显示data；否则进行本地过滤
	$effect(() => {
		if (onSearch) {
			// 外部控制搜索，直接显示传入的数据
			filteredData = data;
		} else {
			// 本地搜索过滤
			if (searchQuery.trim()) {
				filteredData = data.filter(row => 
					columns.some(col => {
						const value = row[col.key];
						return value && String(value).toLowerCase().includes(searchQuery.toLowerCase());
					})
				);
			} else {
				filteredData = data;
			}
		}
	});

	// 处理行点击
	function handleRowClick(row: any) {
		if (onRowClick) {
			onRowClick(row);
		}
	}
</script>

<Card>
	<CardHeader>
		<div class="flex items-center justify-between">
			<div>
				<CardTitle>{title}</CardTitle>
				{#if description}
					<CardDescription>{description}</CardDescription>
				{/if}
			</div>
			
			{#if actions.length > 0}
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
	</CardHeader>

	<CardContent>
		<!-- 搜索栏 -->
		{#if searchable}
			<div class="flex justify-between items-center mb-6">
				<div class="flex gap-4 items-center">
					<div class="relative">
						<Icon name="search" class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
						<Input 
							type="search" 
							placeholder={searchPlaceholder}
							class="pl-8 w-80" 
							bind:value={searchQuery}
							oninput={(e) => {
								if (onSearch) {
									onSearch((e.target as HTMLInputElement).value);
								}
							}}
						/>
					</div>
					{#if searchQuery.trim()}
						<Badge variant="secondary" class="bg-blue-100 text-blue-800">
							找到 {filteredData.length} 个结果
						</Badge>
					{/if}
				</div>
			</div>
		{/if}

		<!-- 数据表格 -->
		<div class="rounded-md border">
			<Table>
				<TableHeader>
					<TableRow>
						{#each columns as column}
							<TableHead style={column.width ? `width: ${column.width}` : ''}>
								{column.title}
							</TableHead>
						{/each}
					</TableRow>
				</TableHeader>
				<TableBody>
					{#if loading}
						<TableRow>
							<td colspan={columns.length} class="text-center py-8">
								<div class="flex flex-col items-center gap-2">
									<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
									<p class="text-gray-500">加载中...</p>
								</div>
							</td>
						</TableRow>
					{:else if filteredData.length === 0}
						<TableRow>
							<td colspan={columns.length} class="text-center py-12">
								<div class="flex flex-col items-center gap-4">
									<Icon name="inbox" class="h-12 w-12 text-gray-300" />
									<div>
										<h3 class="text-lg font-medium text-gray-900 mb-2">
											{searchQuery.trim() ? '没有找到匹配的结果' : emptyStateTitle}
										</h3>
										<p class="text-gray-500">
											{searchQuery.trim() ? `没有找到包含 "${searchQuery}" 的数据` : emptyStateDescription}
										</p>
									</div>
									{#if emptyStateAction && !searchQuery.trim()}
										<Button onclick={emptyStateAction.onClick} class="mt-2">
											{#if emptyStateAction.icon}
												<Icon name={emptyStateAction.icon} class="h-4 w-4 mr-2" />
											{/if}
											{emptyStateAction.text}
										</Button>
									{/if}
								</div>
							</td>
						</TableRow>
					{:else}
						{#each filteredData as row, index}
							<TableRow 
								class={onRowClick ? "cursor-pointer hover:bg-muted/50 transition-colors" : ""}
								onclick={() => handleRowClick(row)}
								role={onRowClick ? "button" : undefined}
								tabindex={onRowClick ? 0 : undefined}
							>
								{#each columns as column}
									<TableCell>
										{#if column.render}
											{@html column.render(row[column.key], row)}
										{:else}
											{row[column.key] || '-'}
										{/if}
									</TableCell>
								{/each}
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</div>
	</CardContent>
</Card>