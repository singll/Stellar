<!--
节点管理主页面
显示节点列表，支持搜索、过滤、批量操作
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { nodeAPI } from '$lib/api/nodes';
	import { NodeStatus, NodeRole } from '$lib/types/node';
	import type {
		Node,
		NodeQueryParams,
		NodeListResponse,
		NodeStats,
		NodeStatusType,
		NodeRoleType
	} from '$lib/types/node';
	import { formatDateTime } from '$lib/utils/date';

	// 组件导入
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import StatCard from '$lib/components/ui/StatCard.svelte';
	import SearchInput from '$lib/components/ui/SearchInput.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	// 状态变量
	let nodes = $state<Node[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let total = $state(0);
	let stats: NodeStats | null = $state(null);
	let selectedNodes = $state<string[]>([]);
	let showBatchActions = $state(false);

	// 查询参数
	let queryParams = $state<NodeQueryParams>({
		page: 1,
		pageSize: 20,
		sortBy: 'registerTime',
		sortDesc: true,
		search: '',
		status: undefined,
		role: undefined,
		onlineOnly: false
	});

	// 过滤器状态
	let searchQuery = $state('');
	let statusFilter = $state<NodeStatusType | ''>('');
	let roleFilter = $state<NodeRoleType | ''>('');
	let onlineOnlyFilter = $state(false);

	// 状态选项
	const statusOptions = [
		{ value: '', label: '全部状态' },
		{ value: NodeStatus.ONLINE, label: '在线' },
		{ value: NodeStatus.OFFLINE, label: '离线' },
		{ value: NodeStatus.DISABLED, label: '禁用' },
		{ value: NodeStatus.MAINTAIN, label: '维护中' },
		{ value: NodeStatus.REGISTING, label: '注册中' }
	];

	// 角色选项
	const roleOptions = [
		{ value: '', label: '全部角色' },
		{ value: NodeRole.MASTER, label: '主节点' },
		{ value: NodeRole.WORKER, label: '工作节点' },
		{ value: NodeRole.SLAVE, label: '从节点' }
	];

	// 排序选项
	const sortOptions = [
		{ value: 'registerTime', label: '注册时间' },
		{ value: 'name', label: '节点名称' },
		{ value: 'status', label: '状态' },
		{ value: 'role', label: '角色' },
		{ value: 'lastHeartbeatTime', label: '最后心跳' }
	];

	// 加载节点列表
	async function loadNodes() {
		try {
			loading = true;
			error = null;

			const params: NodeQueryParams = {
				...queryParams,
				search: searchQuery || undefined,
				status: statusFilter || undefined,
				role: roleFilter || undefined,
				onlineOnly: onlineOnlyFilter
			};

			const response = await nodeAPI.getNodes(params);
			nodes = response.data.items;
			total = response.data.total;
		} catch (err) {
			error = err instanceof Error ? err.message : '加载节点列表失败';
		} finally {
			loading = false;
		}
	}

	// 加载统计信息
	async function loadStats() {
		try {
			stats = await nodeAPI.getNodeStats();
		} catch (err) {
			console.error('加载统计信息失败:', err);
		}
	}

	// 搜索处理
	function handleSearch() {
		queryParams.page = 1;
		loadNodes();
	}

	// 过滤器变化处理
	function handleFilterChange() {
		queryParams.page = 1;
		loadNodes();
	}

	// 分页处理
	function handlePageChange(event: CustomEvent<{ page: number }>) {
		queryParams.page = event.detail.page;
		loadNodes();
	}

	// 排序处理
	function handleSort(field: string) {
		if (queryParams.sortBy === field) {
			queryParams.sortDesc = !queryParams.sortDesc;
		} else {
			queryParams.sortBy = field;
			queryParams.sortDesc = true;
		}
		loadNodes();
	}

	// 节点选择处理
	function handleNodeSelect(nodeId: string, checked: boolean) {
		if (checked) {
			selectedNodes = [...selectedNodes, nodeId];
		} else {
			selectedNodes = selectedNodes.filter((id) => id !== nodeId);
		}
		showBatchActions = selectedNodes.length > 0;
	}

	// 全选处理
	function handleSelectAll(checked: boolean) {
		if (checked) {
			selectedNodes = nodes.map((node) => node.id);
		} else {
			selectedNodes = [];
		}
		showBatchActions = selectedNodes.length > 0;
	}

	// 批量删除
	async function handleBatchDelete() {
		if (!confirm(`确定要删除 ${selectedNodes.length} 个节点吗？`)) return;

		try {
			await nodeAPI.batchDeleteNodes(selectedNodes);
			selectedNodes = [];
			showBatchActions = false;
			loadNodes();
		} catch (err) {
			error = err instanceof Error ? err.message : '批量删除失败';
		}
	}

	// 批量更新状态
	async function handleBatchStatusUpdate(status: NodeStatusType) {
		try {
			await nodeAPI.batchUpdateNodeStatus(selectedNodes, status);
			selectedNodes = [];
			showBatchActions = false;
			loadNodes();
		} catch (err) {
			error = err instanceof Error ? err.message : '批量更新状态失败';
		}
	}

	// 删除节点
	async function handleDeleteNode(nodeId: string) {
		if (!confirm('确定要删除此节点吗？')) return;

		try {
			await nodeAPI.deleteNode(nodeId);
			loadNodes();
		} catch (err) {
			error = err instanceof Error ? err.message : '删除节点失败';
		}
	}

	// 更新节点状态
	async function handleUpdateNodeStatus(nodeId: string, status: NodeStatusType) {
		try {
			await nodeAPI.updateNodeStatus(nodeId, { status });
			loadNodes();
		} catch (err) {
			error = err instanceof Error ? err.message : '更新节点状态失败';
		}
	}

	// 设置维护模式
	async function handleSetMaintenance(nodeId: string, maintenance: boolean) {
		try {
			if (maintenance) {
				await nodeAPI.setNodeMaintenance(nodeId, '手动设置维护模式');
			} else {
				await nodeAPI.cancelNodeMaintenance(nodeId);
			}
			loadNodes();
		} catch (err) {
			error = err instanceof Error ? err.message : '设置维护模式失败';
		}
	}

	// 获取状态标签样式
	function getStatusBadgeVariant(status: NodeStatusType) {
		switch (status) {
			case NodeStatus.ONLINE:
				return 'success';
			case NodeStatus.OFFLINE:
				return 'danger';
			case NodeStatus.DISABLED:
				return 'warning';
			case NodeStatus.MAINTAIN:
				return 'info';
			case NodeStatus.REGISTING:
				return 'secondary';
			default:
				return 'secondary';
		}
	}

	// 获取状态文本
	function getStatusText(status: NodeStatusType) {
		switch (status) {
			case NodeStatus.ONLINE:
				return '在线';
			case NodeStatus.OFFLINE:
				return '离线';
			case NodeStatus.DISABLED:
				return '禁用';
			case NodeStatus.MAINTAIN:
				return '维护中';
			case NodeStatus.REGISTING:
				return '注册中';
			default:
				return '未知';
		}
	}

	// 获取角色文本
	function getRoleText(role: NodeRoleType) {
		switch (role) {
			case NodeRole.MASTER:
				return '主节点';
			case NodeRole.WORKER:
				return '工作节点';
			case NodeRole.SLAVE:
				return '从节点';
			default:
				return '未知';
		}
	}

	// 页面加载
	onMount(() => {
		loadNodes();
		loadStats();
	});

	// 监听查询参数变化
	$effect(() => {
		if (searchQuery !== queryParams.search) {
			handleSearch();
		}
	});

	$effect(() => {
		if (statusFilter !== queryParams.status) {
			handleFilterChange();
		}
	});

	$effect(() => {
		if (roleFilter !== queryParams.role) {
			handleFilterChange();
		}
	});

	$effect(() => {
		if (onlineOnlyFilter !== queryParams.onlineOnly) {
			handleFilterChange();
		}
	});
</script>

<svelte:head>
	<title>节点管理 - Stellar</title>
</svelte:head>

<div class="p-6 space-y-6">
	<!-- 页面标题 -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">节点管理</h1>
			<p class="text-gray-600">管理和监控分布式节点</p>
		</div>
		<div class="flex items-center space-x-4">
			<Button variant="outline" onclick={() => goto('/nodes/register')}>注册节点</Button>
			<Button variant="primary" onclick={() => goto('/nodes/create')}>添加节点</Button>
		</div>
	</div>

	<!-- 统计卡片 -->
	{#if stats}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<StatCard title="总节点数" value={stats.total.toString()} icon="🖥️" variant="primary" />
			<StatCard title="在线节点" value={stats.online.toString()} icon="✅" variant="success" />
			<StatCard title="离线节点" value={stats.offline.toString()} icon="❌" variant="danger" />
			<StatCard
				title="运行任务"
				value={stats.runningTasks.toString()}
				icon="⚡"
				variant="warning"
			/>
		</div>
	{/if}

	<!-- 搜索和过滤 -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
			<div class="md:col-span-2">
				<SearchInput
					bind:value={searchQuery}
					placeholder="搜索节点名称或IP地址..."
					onSearch={handleSearch}
				/>
			</div>
			<Select bind:value={statusFilter} options={statusOptions} placeholder="选择状态" />
			<Select bind:value={roleFilter} options={roleOptions} placeholder="选择角色" />
		</div>

		<div class="flex items-center justify-between">
			<div class="flex items-center space-x-4">
				<label class="flex items-center space-x-2">
					<input
						type="checkbox"
						bind:checked={onlineOnlyFilter}
						class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
					<span class="text-sm text-gray-700">仅显示在线节点</span>
				</label>
			</div>

			<div class="flex items-center space-x-2 text-sm text-gray-600">
				<span>共 {total} 个节点</span>
				{#if selectedNodes.length > 0}
					<span class="text-blue-600">已选择 {selectedNodes.length} 个</span>
				{/if}
			</div>
		</div>
	</div>

	<!-- 批量操作栏 -->
	{#if showBatchActions}
		<div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
			<div class="flex items-center justify-between">
				<span class="text-blue-800">已选择 {selectedNodes.length} 个节点</span>
				<div class="flex items-center space-x-2">
					<Button
						variant="outline"
						size="sm"
						onclick={() => handleBatchStatusUpdate(NodeStatus.ONLINE)}
					>
						启用
					</Button>
					<Button
						variant="outline"
						size="sm"
						onclick={() => handleBatchStatusUpdate(NodeStatus.DISABLED)}
					>
						禁用
					</Button>
					<Button
						variant="outline"
						size="sm"
						onclick={() => handleBatchStatusUpdate(NodeStatus.MAINTAIN)}
					>
						维护
					</Button>
					<Button variant="danger" size="sm" onclick={handleBatchDelete}>删除</Button>
				</div>
			</div>
		</div>
	{/if}

	<!-- 错误提示 -->
	{#if error}
		<div class="bg-red-50 border border-red-200 rounded-lg p-4">
			<div class="flex items-center space-x-2">
				<span class="text-red-800">❌</span>
				<span class="text-red-800">{error}</span>
				<Button variant="outline" size="sm" onclick={() => (error = null)}>关闭</Button>
			</div>
		</div>
	{/if}

	<!-- 节点列表 -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
		{#if loading}
			<div class="flex items-center justify-center p-8">
				<LoadingSpinner size="lg" />
			</div>
		{:else if nodes.length === 0}
			<div class="flex flex-col items-center justify-center p-8">
				<div class="text-gray-400 text-4xl mb-4">🖥️</div>
				<h3 class="text-lg font-medium text-gray-900 mb-2">暂无节点</h3>
				<p class="text-gray-600 mb-4">开始添加您的第一个节点</p>
				<Button variant="primary" onclick={() => goto('/nodes/create')}>添加节点</Button>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left">
								<input
									type="checkbox"
									checked={selectedNodes.length === nodes.length && nodes.length > 0}
									onchange={(e) => handleSelectAll(e.target.checked)}
									class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
								/>
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								<button
									class="flex items-center space-x-1 hover:text-gray-700"
									onclick={() => handleSort('name')}
								>
									<span>节点名称</span>
									{#if queryParams.sortBy === 'name'}
										<span class="text-blue-600">{queryParams.sortDesc ? '↓' : '↑'}</span>
									{/if}
								</button>
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								状态
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								角色
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								IP地址
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								任务统计
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								<button
									class="flex items-center space-x-1 hover:text-gray-700"
									onclick={() => handleSort('lastHeartbeatTime')}
								>
									<span>最后心跳</span>
									{#if queryParams.sortBy === 'lastHeartbeatTime'}
										<span class="text-blue-600">{queryParams.sortDesc ? '↓' : '↑'}</span>
									{/if}
								</button>
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								操作
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each nodes as node (node.id)}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4">
									<input
										type="checkbox"
										checked={selectedNodes.includes(node.id)}
										onchange={(e) => handleNodeSelect(node.id, e.target.checked)}
										class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
									/>
								</td>
								<td class="px-6 py-4">
									<div class="flex items-center space-x-3">
										<div>
											<div class="text-sm font-medium text-gray-900">{node.name}</div>
											<div class="text-sm text-gray-500">{node.ip}:{node.port}</div>
										</div>
									</div>
								</td>
								<td class="px-6 py-4">
									<Badge variant={getStatusBadgeVariant(node.status)}>
										{getStatusText(node.status)}
									</Badge>
								</td>
								<td class="px-6 py-4">
									<span class="text-sm text-gray-900">{getRoleText(node.role)}</span>
								</td>
								<td class="px-6 py-4">
									<span class="text-sm text-gray-900">{node.ip}</span>
								</td>
								<td class="px-6 py-4">
									<div class="text-sm text-gray-900">
										<div>运行: {node.nodeStatus.runningTasks}</div>
										<div>队列: {node.nodeStatus.queuedTasks}</div>
									</div>
								</td>
								<td class="px-6 py-4">
									<span class="text-sm text-gray-900">
										{formatDateTime(node.lastHeartbeatTime)}
									</span>
								</td>
								<td class="px-6 py-4">
									<div class="flex items-center space-x-2">
										<Button variant="outline" size="sm" onclick={() => goto(`/nodes/${node.id}`)}>
											详情
										</Button>
										{#if node.status === NodeStatus.ONLINE}
											<Button
												variant="outline"
												size="sm"
												onclick={() => handleSetMaintenance(node.id, true)}
											>
												维护
											</Button>
										{:else if node.status === NodeStatus.MAINTAIN}
											<Button
												variant="outline"
												size="sm"
												onclick={() => handleSetMaintenance(node.id, false)}
											>
												恢复
											</Button>
										{:else if node.status === NodeStatus.DISABLED}
											<Button
												variant="outline"
												size="sm"
												onclick={() => handleUpdateNodeStatus(node.id, NodeStatus.ONLINE)}
											>
												启用
											</Button>
										{:else}
											<Button
												variant="outline"
												size="sm"
												onclick={() => handleUpdateNodeStatus(node.id, NodeStatus.DISABLED)}
											>
												禁用
											</Button>
										{/if}
										<Button variant="danger" size="sm" onclick={() => handleDeleteNode(node.id)}>
											删除
										</Button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- 分页 -->
	{#if total > queryParams.pageSize}
		<div class="flex justify-center">
			<Pagination
				currentPage={queryParams.page}
				totalPages={Math.ceil(total / queryParams.pageSize)}
				onPageChange={handlePageChange}
			/>
		</div>
	{/if}
</div>
