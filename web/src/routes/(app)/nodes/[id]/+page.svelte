<!--
节点详情页面
显示节点的详细信息、状态、配置、任务统计等
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { nodeAPI } from '$lib/api/nodes';
	import { NodeStatus, NodeRole } from '$lib/types/node';
	import type {
		Node,
		NodeHealth,
		NodeConfig,
		NodeTaskStats,
		NodeStatusType
	} from '$lib/types/node';
	import { formatDateTime } from '$lib/utils/date';

	// 组件导入
	import Button from '$lib/components/ui/Button.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import StatCard from '$lib/components/ui/StatCard.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';

	// 获取节点ID
	let nodeId = $derived($page.params.id);

	// 状态变量
	let node: Node | null = $state(null);
	let health: NodeHealth | null = $state(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let activeTab = $state('overview');
	let editMode = $state(false);
	let editData = $state<Partial<Node>>({});

	// 标签页选项
	const tabs = [
		{ id: 'overview', label: '概览', icon: '📊' },
		{ id: 'config', label: '配置', icon: '⚙️' },
		{ id: 'tasks', label: '任务', icon: '📋' },
		{ id: 'monitor', label: '监控', icon: '📈' },
		{ id: 'logs', label: '日志', icon: '📄' }
	];

	// 加载节点详情
	async function loadNode() {
		if (!nodeId) return;

		try {
			loading = true;
			error = null;

			const [nodeData, healthData] = await Promise.all([
				nodeAPI.getNode(nodeId),
				nodeAPI.getNodeHealth(nodeId)
			]);

			node = nodeData;
			health = healthData;
			editData = {
				name: node.name,
				role: node.role,
				tags: node.tags,
				config: node.config
			};
		} catch (err) {
			error = err instanceof Error ? err.message : '加载节点详情失败';
		} finally {
			loading = false;
		}
	}

	// 保存节点信息
	async function saveNode() {
		if (!nodeId || !editData) return;

		try {
			await nodeAPI.updateNode(nodeId, editData);
			editMode = false;
			await loadNode();
		} catch (err) {
			error = err instanceof Error ? err.message : '保存节点信息失败';
		}
	}

	// 更新节点状态
	async function updateNodeStatus(status: NodeStatusType) {
		if (!nodeId) return;

		try {
			await nodeAPI.updateNodeStatus(nodeId, { status });
			await loadNode();
		} catch (err) {
			error = err instanceof Error ? err.message : '更新节点状态失败';
		}
	}

	// 删除节点
	async function deleteNode() {
		if (!nodeId || !confirm('确定要删除此节点吗？')) return;

		try {
			await nodeAPI.deleteNode(nodeId);
			goto('/nodes');
		} catch (err) {
			error = err instanceof Error ? err.message : '删除节点失败';
		}
	}

	// 重启节点
	async function restartNode() {
		if (!nodeId || !confirm('确定要重启此节点吗？')) return;

		try {
			await nodeAPI.restartNode(nodeId);
			await loadNode();
		} catch (err) {
			error = err instanceof Error ? err.message : '重启节点失败';
		}
	}

	// 获取状态标签样式
	function getStatusBadgeVariant(status: NodeStatusType) {
		switch (status) {
			case NodeStatus.ONLINE:
				return 'default';
			case NodeStatus.OFFLINE:
				return 'destructive';
			case NodeStatus.DISABLED:
				return 'secondary';
			case NodeStatus.MAINTAIN:
				return 'outline';
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
	function getRoleText(role: string) {
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

	// 格式化正常运行时间
	function formatUptime(seconds: number) {
		const days = Math.floor(seconds / 86400);
		const hours = Math.floor((seconds % 86400) / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);

		if (days > 0) {
			return `${days}天 ${hours}小时 ${minutes}分钟`;
		} else if (hours > 0) {
			return `${hours}小时 ${minutes}分钟`;
		} else {
			return `${minutes}分钟`;
		}
	}

	// 格式化内存大小
	function formatMemory(mb: number) {
		if (mb >= 1024) {
			return `${(mb / 1024).toFixed(1)} GB`;
		}
		return `${mb} MB`;
	}

	// 页面加载
	onMount(() => {
		loadNode();
	});
</script>

<svelte:head>
	<title>节点详情 - {node?.name || '加载中...'} - Stellar</title>
</svelte:head>

<div class="p-6 space-y-6">
	<!-- 加载状态 -->
	{#if loading}
		<div class="flex items-center justify-center p-8">
			<LoadingSpinner size="lg" />
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 rounded-lg p-4">
			<div class="flex items-center space-x-2">
				<span class="text-red-800">❌</span>
				<span class="text-red-800">{error}</span>
				<Button variant="outline" size="sm" onclick={() => (error = null)}>关闭</Button>
			</div>
		</div>
	{:else if node}
		<!-- 页面标题 -->
		<div class="flex items-center justify-between">
			<div class="flex items-center space-x-4">
				<Button variant="outline" onclick={() => goto('/nodes')}>← 返回</Button>
				<div>
					<h1 class="text-2xl font-bold text-gray-900">{node.name}</h1>
					<p class="text-gray-600">{node.ip}:{node.port}</p>
				</div>
				<Badge variant={getStatusBadgeVariant(node.status)}>
					{getStatusText(node.status)}
				</Badge>
			</div>
			<div class="flex items-center space-x-2">
				{#if !editMode}
					<Button variant="outline" onclick={() => (editMode = true)}>编辑</Button>
				{:else}
					<Button variant="outline" onclick={() => (editMode = false)}>取消</Button>
					<Button variant="default" onclick={saveNode}>保存</Button>
				{/if}

				{#if node.status === NodeStatus.ONLINE}
					<Button variant="outline" onclick={() => updateNodeStatus(NodeStatus.MAINTAIN)}>
						维护
					</Button>
					<Button variant="secondary" onclick={restartNode}>重启</Button>
				{:else if node.status === NodeStatus.MAINTAIN}
					<Button variant="outline" onclick={() => updateNodeStatus(NodeStatus.ONLINE)}>
						恢复
					</Button>
				{:else if node.status === NodeStatus.DISABLED}
					<Button variant="outline" onclick={() => updateNodeStatus(NodeStatus.ONLINE)}>
						启用
					</Button>
				{:else}
					<Button variant="outline" onclick={() => updateNodeStatus(NodeStatus.DISABLED)}>
						禁用
					</Button>
				{/if}

				<Button variant="destructive" onclick={deleteNode}>删除</Button>
			</div>
		</div>

		<!-- 基本信息卡片 -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<StatCard
				title="CPU使用率"
				value={`${node.nodeStatus.cpuUsage.toFixed(1)}%`}
				icon="🖥️"
				color="blue"
			/>
			<StatCard
				title="内存使用"
				value={formatMemory(node.nodeStatus.memoryUsage)}
				icon="💾"
				color="blue"
			/>
			<StatCard
				title="运行任务"
				value={node.nodeStatus.runningTasks.toString()}
				icon="⚡"
				color="gray"
			/>
			<StatCard
				title="正常运行"
				value={formatUptime(node.nodeStatus.uptimeSeconds)}
				icon="⏰"
				color="green"
			/>
		</div>

		<!-- 健康状态 -->
		{#if health}
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">健康状态</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<div class="flex items-center space-x-2 mb-2">
							<span class="text-sm font-medium text-gray-700">健康评分</span>
							<Badge
								variant={health.score >= 80
									? 'default'
									: health.score >= 60
										? 'secondary'
										: 'destructive'}
							>
								{health.score}/100
							</Badge>
						</div>
						<ProgressBar
							value={health.score}
							max={100}
							color={health.score >= 80 ? 'green' : health.score >= 60 ? 'yellow' : 'red'}
						/>
					</div>
					<div>
						<div class="text-sm font-medium text-gray-700 mb-2">健康问题</div>
						{#if health.issues.length > 0}
							<ul class="space-y-1">
								{#each health.issues as issue}
									<li class="text-sm text-red-600">• {issue}</li>
								{/each}
							</ul>
						{:else}
							<span class="text-sm text-green-600">无健康问题</span>
						{/if}
					</div>
				</div>
			</div>
		{/if}

		<!-- 标签页 -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<Tabs bind:activeTab {tabs} />

			<div class="p-6">
				{#if activeTab === 'overview'}
					<!-- 概览标签页 -->
					<div class="space-y-6">
						<!-- 基本信息 -->
						<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
							<div>
								<h4 class="text-lg font-semibold text-gray-900 mb-4">基本信息</h4>
								<div class="space-y-3">
									<div class="flex justify-between">
										<span class="text-gray-600">节点名称:</span>
										<span class="font-medium">{node.name}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">角色:</span>
										<span class="font-medium">{getRoleText(node.role)}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">IP地址:</span>
										<span class="font-medium">{node.ip}:{node.port}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">注册时间:</span>
										<span class="font-medium">{formatDateTime(node.registerTime)}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">最后心跳:</span>
										<span class="font-medium">{formatDateTime(node.lastHeartbeatTime)}</span>
									</div>
								</div>
							</div>

							<div>
								<h4 class="text-lg font-semibold text-gray-900 mb-4">任务统计</h4>
								<div class="space-y-3">
									<div class="flex justify-between">
										<span class="text-gray-600">总任务数:</span>
										<span class="font-medium">{node.taskStats.totalTasks}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">成功任务:</span>
										<span class="font-medium text-green-600">{node.taskStats.successTasks}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">失败任务:</span>
										<span class="font-medium text-red-600">{node.taskStats.failedTasks}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">成功率:</span>
										<span class="font-medium">
											{node.taskStats.totalTasks > 0
												? ((node.taskStats.successTasks / node.taskStats.totalTasks) * 100).toFixed(
														1
													)
												: 0}%
										</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">平均执行时间:</span>
										<span class="font-medium">{node.taskStats.avgExecuteTime}秒</span>
									</div>
								</div>
							</div>
						</div>

						<!-- 标签 -->
						<div>
							<h4 class="text-lg font-semibold text-gray-900 mb-4">标签</h4>
							<div class="flex flex-wrap gap-2">
								{#each node.tags as tag}
									<Badge variant="secondary">{tag}</Badge>
								{:else}
									<span class="text-gray-500">暂无标签</span>
								{/each}
							</div>
						</div>
					</div>
				{:else if activeTab === 'config'}
					<!-- 配置标签页 -->
					<div class="space-y-6">
						<div class="flex items-center justify-between mb-6">
							<h4 class="text-lg font-semibold text-gray-900">节点配置</h4>
							{#if !editMode}
								<Button variant="outline" onclick={() => (editMode = true)}>编辑配置</Button>
							{/if}
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
							<div class="space-y-4">
								<div>
									<label for="max-concurrent-tasks" class="block text-sm font-medium text-gray-700 mb-2">
										最大并发任务数
									</label>
									{#if editMode && editData.config}
										<Input
											id="max-concurrent-tasks"
											type="number"
											bind:value={editData.config.maxConcurrentTasks}
											min={1}
											max={100}
										/>
									{:else if editMode}
										<span class="text-sm text-gray-500">配置加载中...</span>
									{:else}
										<span class="text-sm text-gray-900">{node.config.maxConcurrentTasks}</span>
									{/if}
								</div>

								<div>
									<label for="max-memory-usage" class="block text-sm font-medium text-gray-700 mb-2">
										最大内存使用量 (MB)
									</label>
									{#if editMode && editData.config}
										<Input
											id="max-memory-usage"
											type="number"
											bind:value={editData.config.maxMemoryUsage}
											min={512}
											max={32768}
										/>
									{:else if editMode}
										<span class="text-sm text-gray-500">配置加载中...</span>
									{:else}
										<span class="text-sm text-gray-900">{node.config.maxMemoryUsage}</span>
									{/if}
								</div>

								<div>
									<label for="max-cpu-usage" class="block text-sm font-medium text-gray-700 mb-2">
										最大CPU使用率 (%)
									</label>
									{#if editMode && editData.config}
										<Input
											id="max-cpu-usage"
											type="number"
											bind:value={editData.config.maxCpuUsage}
											min={10}
											max={100}
										/>
									{:else if editMode}
										<span class="text-sm text-gray-500">配置加载中...</span>
									{:else}
										<span class="text-sm text-gray-900">{node.config.maxCpuUsage}</span>
									{/if}
								</div>
							</div>

							<div class="space-y-4">
								<div>
									<label for="heartbeat-interval" class="block text-sm font-medium text-gray-700 mb-2">
										心跳间隔 (秒)
									</label>
									{#if editMode && editData.config}
										<Input
											id="heartbeat-interval"
											type="number"
											bind:value={editData.config.heartbeatInterval}
											min={5}
											max={300}
										/>
									{:else if editMode}
										<span class="text-sm text-gray-500">配置加载中...</span>
									{:else}
										<span class="text-sm text-gray-900">{node.config.heartbeatInterval}</span>
									{/if}
								</div>

								<div>
									<label for="task-timeout" class="block text-sm font-medium text-gray-700 mb-2">
										任务超时时间 (秒)
									</label>
									{#if editMode && editData.config}
										<Input
											id="task-timeout"
											type="number"
											bind:value={editData.config.taskTimeout}
											min={60}
											max={3600}
										/>
									{:else if editMode}
										<span class="text-sm text-gray-500">配置加载中...</span>
									{:else}
										<span class="text-sm text-gray-900">{node.config.taskTimeout}</span>
									{/if}
								</div>

								<div>
									<span class="block text-sm font-medium text-gray-700 mb-2"> 日志级别 </span>
									{#if editMode && editData.config}
										<Select
											bind:value={editData.config.logLevel}
											options={[
												{ value: 'debug', label: 'Debug' },
												{ value: 'info', label: 'Info' },
												{ value: 'warn', label: 'Warn' },
												{ value: 'error', label: 'Error' }
											]}
										/>
									{:else if editMode}
										<span class="text-sm text-gray-500">配置加载中...</span>
									{:else}
										<span class="text-sm text-gray-900">{node.config.logLevel}</span>
									{/if}
								</div>
							</div>
						</div>
					</div>
				{:else if activeTab === 'tasks'}
					<!-- 任务标签页 -->
					<div class="space-y-6">
						<h4 class="text-lg font-semibold text-gray-900">任务管理</h4>
						<div class="text-gray-600">此功能正在开发中...</div>
					</div>
				{:else if activeTab === 'monitor'}
					<!-- 监控标签页 -->
					<div class="space-y-6">
						<h4 class="text-lg font-semibold text-gray-900">实时监控</h4>
						<div class="text-gray-600">此功能正在开发中...</div>
					</div>
				{:else if activeTab === 'logs'}
					<!-- 日志标签页 -->
					<div class="space-y-6">
						<h4 class="text-lg font-semibold text-gray-900">节点日志</h4>
						<div class="text-gray-600">此功能正在开发中...</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
