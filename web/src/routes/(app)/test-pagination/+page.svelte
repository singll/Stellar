<!--
分页组件演示页面
展示分页组件的各种功能和配置选项
-->
<script lang="ts">
	import { Pagination } from '$lib/components/ui/pagination';
	import { Button } from '$lib/components/ui/button';
	import PageLayout from '$lib/components/ui/page-layout/PageLayout.svelte';

	// 模拟数据
	let mockData = Array.from({ length: 235 }, (_, i) => ({
		id: i + 1,
		name: `项目 ${i + 1}`,
		description: `这是第 ${i + 1} 个测试项目的描述信息`,
		created_at: new Date(Date.now() - Math.random() * 10000000000).toISOString()
	}));

	// 分页状态
	let currentPage = $state(1);
	let pageSize = $state(20);
	let loading = $state(false);

	// 计算分页信息
	$: totalItems = mockData.length;
	$: totalPages = Math.ceil(totalItems / pageSize);
	$: startIndex = (currentPage - 1) * pageSize;
	$: endIndex = Math.min(startIndex + pageSize, totalItems);
	$: currentData = mockData.slice(startIndex, endIndex);

	// 分页处理函数
	const handlePageChange = async (newPage: number) => {
		loading = true;
		// 模拟异步加载
		await new Promise(resolve => setTimeout(resolve, 300));
		currentPage = newPage;
		loading = false;
	};

	const handlePageSizeChange = async (newPageSize: number) => {
		loading = true;
		// 模拟异步加载
		await new Promise(resolve => setTimeout(resolve, 300));
		pageSize = newPageSize;
		currentPage = 1; // 重置到第一页
		loading = false;
	};

	// 重置数据
	const resetData = () => {
		currentPage = 1;
		pageSize = 20;
	};

	// 添加更多数据
	const addMoreData = () => {
		const newItems = Array.from({ length: 50 }, (_, i) => ({
			id: mockData.length + i + 1,
			name: `新项目 ${mockData.length + i + 1}`,
			description: `这是新添加的第 ${mockData.length + i + 1} 个项目`,
			created_at: new Date().toISOString()
		}));
		mockData = [...mockData, ...newItems];
	};
</script>

<svelte:head>
	<title>分页组件演示 - Stellar</title>
</svelte:head>

<PageLayout
	title="分页组件演示"
	description="展示通用分页组件的各种功能和配置选项"
	icon="layout"
>
	<!-- 控制面板 -->
	<div class="bg-white rounded-lg border p-6 mb-6">
		<h3 class="text-lg font-medium mb-4">控制面板</h3>
		<div class="flex flex-wrap gap-4">
			<Button variant="outline" onclick={resetData}>
				重置分页
			</Button>
			<Button variant="outline" onclick={addMoreData}>
				添加50条数据
			</Button>
			<div class="text-sm text-gray-600 flex items-center">
				当前数据总数: <span class="font-semibold ml-1">{totalItems}</span> 条
			</div>
		</div>
	</div>

	<!-- 数据展示区域 -->
	<div class="bg-white rounded-lg border overflow-hidden mb-6">
		<div class="p-6 border-b">
			<h3 class="text-lg font-medium">数据列表</h3>
			<p class="text-sm text-gray-600 mt-1">
				显示第 {startIndex + 1} - {endIndex} 条，共 {totalItems} 条数据
			</p>
		</div>

		<!-- 数据表格 -->
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							ID
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							项目名称
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							描述
						</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							创建时间
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#if loading}
						{#each Array(pageSize) as _, i}
							<tr class="animate-pulse">
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="h-4 bg-gray-200 rounded w-12"></div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="h-4 bg-gray-200 rounded w-24"></div>
								</td>
								<td class="px-6 py-4">
									<div class="h-4 bg-gray-200 rounded w-48"></div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="h-4 bg-gray-200 rounded w-20"></div>
								</td>
							</tr>
						{/each}
					{:else}
						{#each currentData as item}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
									{item.id}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
									{item.name}
								</td>
								<td class="px-6 py-4 text-sm text-gray-500">
									{item.description}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{new Date(item.created_at).toLocaleDateString('zh-CN')}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>

	<!-- 分页组件演示 -->
	<div class="space-y-8">
		<!-- 完整功能的分页组件 -->
		<div class="bg-white rounded-lg border p-6">
			<h3 class="text-lg font-medium mb-4">完整功能分页组件</h3>
			<p class="text-sm text-gray-600 mb-6">
				包含页面信息、页面大小选择器、首末页按钮、快速跳转等所有功能
			</p>
			<Pagination
				{currentPage}
				{totalPages}
				{totalItems}
				{pageSize}
				pageSizeOptions={[10, 20, 50, 100]}
				showPageSizeSelector={true}
				showPageInfo={true}
				showFirstLast={true}
				maxVisiblePages={7}
				disabled={loading}
				onPageChange={handlePageChange}
				onPageSizeChange={handlePageSizeChange}
			/>
		</div>

		<!-- 简化版分页组件 -->
		<div class="bg-white rounded-lg border p-6">
			<h3 class="text-lg font-medium mb-4">简化版分页组件</h3>
			<p class="text-sm text-gray-600 mb-6">
				只显示基本的页码导航功能，不显示页面信息和页面大小选择器
			</p>
			<Pagination
				{currentPage}
				{totalPages}
				{totalItems}
				{pageSize}
				showPageSizeSelector={false}
				showPageInfo={false}
				showFirstLast={false}
				maxVisiblePages={5}
				disabled={loading}
				onPageChange={handlePageChange}
				onPageSizeChange={handlePageSizeChange}
			/>
		</div>

		<!-- 紧凑版分页组件 -->
		<div class="bg-white rounded-lg border p-6">
			<h3 class="text-lg font-medium mb-4">紧凑版分页组件</h3>
			<p class="text-sm text-gray-600 mb-6">
				适用于移动端或空间有限的场景，只显示3个可见页码
			</p>
			<Pagination
				{currentPage}
				{totalPages}
				{totalItems}
				{pageSize}
				showPageSizeSelector={false}
				showPageInfo={false}
				showFirstLast={true}
				maxVisiblePages={3}
				disabled={loading}
				onPageChange={handlePageChange}
				onPageSizeChange={handlePageSizeChange}
				class="justify-center"
			/>
		</div>

		<!-- 禁用状态演示 -->
		<div class="bg-white rounded-lg border p-6">
			<h3 class="text-lg font-medium mb-4">禁用状态演示</h3>
			<p class="text-sm text-gray-600 mb-6">
				分页组件在加载状态下的禁用效果
			</p>
			<Pagination
				{currentPage}
				{totalPages}
				{totalItems}
				{pageSize}
				pageSizeOptions={[10, 20, 50, 100]}
				showPageSizeSelector={true}
				showPageInfo={true}
				showFirstLast={true}
				maxVisiblePages={7}
				disabled={true}
				onPageChange={handlePageChange}
				onPageSizeChange={handlePageSizeChange}
			/>
		</div>
	</div>

	<!-- 组件说明 -->
	<div class="bg-blue-50 rounded-lg border border-blue-200 p-6 mt-8">
		<h3 class="text-lg font-medium text-blue-900 mb-4">组件特性说明</h3>
		<div class="space-y-3 text-sm text-blue-800">
			<div class="flex items-start gap-2">
				<span class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
				<span><strong>智能页码显示：</strong>自动计算可见页码范围，支持省略号显示</span>
			</div>
			<div class="flex items-start gap-2">
				<span class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
				<span><strong>页面大小选择：</strong>支持动态切换每页显示数量</span>
			</div>
			<div class="flex items-start gap-2">
				<span class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
				<span><strong>快速跳转：</strong>输入页码直接跳转到指定页面</span>
			</div>
			<div class="flex items-start gap-2">
				<span class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
				<span><strong>首末页按钮：</strong>快速跳转到第一页和最后一页</span>
			</div>
			<div class="flex items-start gap-2">
				<span class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
				<span><strong>禁用状态：</strong>支持在加载时禁用所有交互</span>
			</div>
			<div class="flex items-start gap-2">
				<span class="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></span>
				<span><strong>自适应布局：</strong>根据屏幕大小和配置自动调整显示</span>
			</div>
		</div>
	</div>
</PageLayout>