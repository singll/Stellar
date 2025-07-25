<!-- 端口扫描任务详情页面 -->
<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { portScanStore, type PortScanTask, type PortScanResult } from '$lib/stores/portscan';
	import { toastStore } from '$lib/stores/toast';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
	import { formatDateTime, formatDuration } from '$lib/utils/date';

	// 获取任务ID
	const taskId = $page.params.id;

	// 响应式状态
	let task = $state<PortScanTask | null>(null);
	let results = $state<PortScanResult[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let activeTab = $state('results');
	let refreshInterval: NodeJS.Timeout | null = null;

	// Store 订阅
	const currentTask = portScanStore.currentTask;
	const taskResults = portScanStore.taskResults;
	const loading = portScanStore.loading;

	// Tab 选项
	const tabs = [
		{ id: 'results', label: '扫描结果' },
		{ id: 'config', label: '任务配置' },
		{ id: 'logs', label: '执行日志' }
	];

	// 初始化
	onMount(async () => {
		await loadTaskDetail();

		// 如果任务正在运行，设置定时刷新
		if (task?.status === 'running') {
			refreshInterval = setInterval(refreshTask, 5000);
		}
	});

	// 清理
	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});

	// 加载任务详情
	async function loadTaskDetail() {
		isLoading = true;
		try {
			task = await portScanStore.actions.loadTask(taskId);
			await loadTaskResults();
		} catch (error) {
			console.error('加载任务详情失败:', error);
			toastStore.error('加载任务详情失败');
		} finally {
			isLoading = false;
		}
	}

	// 加载任务结果
	async function loadTaskResults() {
		try {
			const response = await portScanStore.actions.loadTaskResults(taskId);
			results = response.results;
		} catch (error) {
			console.error('加载任务结果失败:', error);
		}
	}

	// 刷新任务
	async function refreshTask() {
		if (isLoading || isRefreshing) return;

		isRefreshing = true;
		try {
			task = await portScanStore.actions.loadTask(taskId);

			// 如果任务状态变化，更新刷新策略
			if (task?.status !== 'running' && refreshInterval) {
				clearInterval(refreshInterval);
				refreshInterval = null;
			}

			// 如果是运行中或完成状态，刷新结果
			if (task?.status === 'running' || task?.status === 'completed') {
				await loadTaskResults();
			}
		} catch (error) {
			console.error('刷新任务失败:', error);
		} finally {
			isRefreshing = false;
		}
	}

	// 任务操作
	async function handleTaskAction(action: string) {
		if (!task) return;

		try {
			switch (action) {
				case 'cancel':
					await portScanStore.actions.cancelTask(task.id);
					toastStore.success('任务已取消');
					await loadTaskDetail();
					break;
				case 'retry':
					const newTask = await portScanStore.actions.retryTask(task.id);
					toastStore.success('任务已重新启动');
					await goto(`/portscan/${newTask.id}`);
					break;
				case 'delete':
					if (confirm('确定要删除这个任务吗？')) {
						await portScanStore.actions.deleteTask(task.id);
						toastStore.success('任务已删除');
						await goto('/portscan');
					}
					break;
				case 'export':
					await portScanStore.actions.exportResults(task.id, 'csv');
					toastStore.success('结果导出成功');
					break;
			}
		} catch (error) {
			console.error('任务操作失败:', error);
			toastStore.error('操作失败: ' + (error as Error).message);
		}
	}

	// 获取状态样式
	function getStatusVariant(status: string) {
		switch (status) {
			case 'completed':
				return 'default';
			case 'running':
				return 'secondary';
			case 'failed':
				return 'destructive';
			case 'canceled':
				return 'secondary';
			default:
				return 'default';
		}
	}

	// 获取状态文本
	function getStatusText(status: string) {
		switch (status) {
			case 'pending':
				return '待执行';
			case 'queued':
				return '已排队';
			case 'running':
				return '运行中';
			case 'completed':
				return '已完成';
			case 'failed':
				return '失败';
			case 'canceled':
				return '已取消';
			default:
				return status;
		}
	}

	// 获取端口状态样式
	function getPortStatusVariant(status: string) {
		switch (status) {
			case 'open':
				return 'default';
			case 'closed':
				return 'secondary';
			case 'filtered':
				return 'outline';
			default:
				return 'default';
		}
	}

	// 过滤结果
	let searchTerm = $state('');
	let filterStatus = $state('all');
	let filterService = $state('all');

	let filteredResults = $derived(
		results.filter((result) => {
			// 搜索过滤
			if (searchTerm) {
				const search = searchTerm.toLowerCase();
				if (
					!result.host.toLowerCase().includes(search) &&
					!result.port.toString().includes(search) &&
					!(result.service || '').toLowerCase().includes(search)
				) {
					return false;
				}
			}

			// 状态过滤
			if (filterStatus !== 'all' && result.status !== filterStatus) {
				return false;
			}

			// 服务过滤
			if (filterService !== 'all' && result.service !== filterService) {
				return false;
			}

			return true;
		})
	);

	// 获取可用的服务列表
	let availableServices = $derived([...new Set(results.map((r) => r.service).filter(Boolean))]);

	// 分页
	let currentPage = $state(1);
	let pageSize = $state(50);

	let totalPages = $derived(Math.ceil(filteredResults.length / pageSize));
	let paginatedResults = $derived(
		filteredResults.slice((currentPage - 1) * pageSize, currentPage * pageSize)
	);

	// 统计信息
	let stats = $derived({
		total: results.length,
		open: results.filter((r) => r.status === 'open').length,
		closed: results.filter((r) => r.status === 'closed').length,
		filtered: results.filter((r) => r.status === 'filtered').length
	});
</script>

<svelte:head>
	<title>{task?.name || '任务详情'} - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	{#if isLoading}
		<div class="flex justify-center items-center py-12">
			<LoadingSpinner size="lg" />
		</div>
	{:else if !task}
		<div class="text-center py-12">
			<div class="text-gray-400 text-lg mb-2">任务不存在</div>
			<Button onclick={() => goto('/portscan')}>返回列表</Button>
		</div>
	{:else}
		<!-- 页面标题和操作 -->
		<div class="flex justify-between items-start mb-6">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">{task.name}</h1>
				<div class="flex items-center gap-4 mt-2">
					<Badge variant={getStatusVariant(task.status)}>
						{getStatusText(task.status)}
					</Badge>
					<span class="text-gray-600">目标: {task.config.target}</span>
					<span class="text-gray-600">创建时间: {formatDateTime(task.createdAt)}</span>
				</div>
			</div>

			<div class="flex gap-2">
				<Button variant="outline" onclick={() => goto('/portscan')}>返回列表</Button>

				<Button variant="outline" onclick={refreshTask} disabled={isRefreshing}>
					{#if isRefreshing}
						<LoadingSpinner size="sm" />
					{:else}
						刷新
					{/if}
				</Button>

				{#if task.status === 'running'}
					<Button variant="outline" onclick={() => handleTaskAction('cancel')}>取消任务</Button>
				{/if}

				{#if task.status === 'failed'}
					<Button variant="outline" onclick={() => handleTaskAction('retry')}>重试任务</Button>
				{/if}

				{#if task.status === 'completed'}
					<Button variant="outline" onclick={() => handleTaskAction('export')}>导出结果</Button>
				{/if}

				{#if task.status !== 'running'}
					<Button variant="outline" onclick={() => handleTaskAction('delete')}>删除任务</Button>
				{/if}
			</div>
		</div>

		<!-- 任务进度 -->
		{#if task.status === 'running' && task.progress}
			<div class="bg-white rounded-lg shadow-sm border p-6 mb-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">执行进度</h3>
				<div class="space-y-4">
					<ProgressBar value={task.progress.completed} max={task.progress.total} label="扫描进度" />
					<div class="grid grid-cols-1 md:grid-cols-4 gap-4 text-sm">
						<div>
							<span class="text-gray-600">总端口数:</span>
							<span class="font-medium ml-1">{task.progress.total}</span>
						</div>
						<div>
							<span class="text-gray-600">已完成:</span>
							<span class="font-medium ml-1">{task.progress.completed}</span>
						</div>
						<div>
							<span class="text-gray-600">失败:</span>
							<span class="font-medium ml-1">{task.progress.failed}</span>
						</div>
						<div>
							<span class="text-gray-600">进度:</span>
							<span class="font-medium ml-1">
								{Math.round((task.progress.completed / task.progress.total) * 100)}%
							</span>
						</div>
					</div>
					{#if task.progress.speed}
						<div class="text-sm text-gray-600">
							扫描速度: {task.progress.speed} 端口/秒
							{#if task.progress.estimatedTime}
								| 预计剩余时间: {Math.ceil(task.progress.estimatedTime / 60)} 分钟
							{/if}
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Tab导航 -->
		<Tabs {tabs} bind:activeTab />

		<!-- Tab内容 -->
		{#if activeTab === 'results'}
			<div class="bg-white rounded-lg shadow-sm border">
				<!-- 结果统计 -->
				<div class="p-6 border-b">
					<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
						<div class="text-center">
							<div class="text-2xl font-bold text-gray-900">{stats.total}</div>
							<div class="text-sm text-gray-600">总端口数</div>
						</div>
						<div class="text-center">
							<div class="text-2xl font-bold text-green-600">{stats.open}</div>
							<div class="text-sm text-gray-600">开放端口</div>
						</div>
						<div class="text-center">
							<div class="text-2xl font-bold text-gray-600">{stats.closed}</div>
							<div class="text-sm text-gray-600">关闭端口</div>
						</div>
						<div class="text-center">
							<div class="text-2xl font-bold text-yellow-600">{stats.filtered}</div>
							<div class="text-sm text-gray-600">过滤端口</div>
						</div>
					</div>
				</div>

				<!-- 过滤器 -->
				<div class="p-6 border-b">
					<div class="flex flex-wrap gap-4">
						<input
							type="text"
							bind:value={searchTerm}
							placeholder="搜索主机、端口或服务..."
							class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>

						<select
							bind:value={filterStatus}
							class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						>
							<option value="all">全部状态</option>
							<option value="open">开放</option>
							<option value="closed">关闭</option>
							<option value="filtered">过滤</option>
						</select>

						<select
							bind:value={filterService}
							class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						>
							<option value="all">全部服务</option>
							{#each availableServices as service}
								<option value={service}>{service}</option>
							{/each}
						</select>

						<div class="text-sm text-gray-600 flex items-center">
							显示 {filteredResults.length} / {stats.total} 个结果
						</div>
					</div>
				</div>

				<!-- 结果列表 -->
				{#if paginatedResults.length === 0}
					<div class="text-center py-12">
						<div class="text-gray-400 text-lg mb-2">暂无结果</div>
						<p class="text-gray-500">
							{#if task.status === 'running'}
								任务正在执行中，请稍候...
							{:else}
								未找到匹配的扫描结果
							{/if}
						</p>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="w-full">
							<thead class="bg-gray-50">
								<tr>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">主机</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">端口</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">协议</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">状态</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">服务</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">版本</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">响应时间</th>
									<th class="px-4 py-3 text-left text-sm font-medium text-gray-900">Banner</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-200">
								{#each paginatedResults as result}
									<tr class="hover:bg-gray-50">
										<td class="px-4 py-3 text-sm text-gray-900">{result.host}</td>
										<td class="px-4 py-3 text-sm text-gray-900">{result.port}</td>
										<td class="px-4 py-3 text-sm text-gray-500">{result.protocol.toUpperCase()}</td>
										<td class="px-4 py-3">
											<Badge variant={getPortStatusVariant(result.status)}>
												{result.status}
											</Badge>
										</td>
										<td class="px-4 py-3 text-sm text-gray-900">{result.service || '-'}</td>
										<td class="px-4 py-3 text-sm text-gray-500">{result.version || '-'}</td>
										<td class="px-4 py-3 text-sm text-gray-500">{result.responseTime}ms</td>
										<td class="px-4 py-3 text-sm text-gray-500 max-w-xs truncate">
											{result.banner || '-'}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

					<!-- 分页 -->
					{#if totalPages > 1}
						<div class="p-6 border-t">
							<div class="flex justify-between items-center">
								<div class="text-sm text-gray-600">
									第 {(currentPage - 1) * pageSize + 1} - {Math.min(
										currentPage * pageSize,
										filteredResults.length
									)} 条，共 {filteredResults.length} 条
								</div>
								<div class="flex gap-2">
									<Button
										variant="outline"
										size="sm"
										disabled={currentPage === 1}
										onclick={() => currentPage--}
									>
										上一页
									</Button>
									<span class="px-3 py-1 text-sm text-gray-600">
										{currentPage} / {totalPages}
									</span>
									<Button
										variant="outline"
										size="sm"
										disabled={currentPage === totalPages}
										onclick={() => currentPage++}
									>
										下一页
									</Button>
								</div>
							</div>
						</div>
					{/if}
				{/if}
			</div>
		{:else if activeTab === 'config'}
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">任务配置</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<h4 class="font-medium text-gray-900 mb-2">基本信息</h4>
						<dl class="space-y-2 text-sm">
							<div class="flex justify-between">
								<dt class="text-gray-600">任务名称:</dt>
								<dd class="text-gray-900">{task.name}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">扫描目标:</dt>
								<dd class="text-gray-900">{task.config.target}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">端口配置:</dt>
								<dd class="text-gray-900">{task.config.ports}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">扫描方法:</dt>
								<dd class="text-gray-900">{task.config.scanMethod?.toUpperCase()}</dd>
							</div>
						</dl>
					</div>

					<div>
						<h4 class="font-medium text-gray-900 mb-2">高级配置</h4>
						<dl class="space-y-2 text-sm">
							<div class="flex justify-between">
								<dt class="text-gray-600">最大并发:</dt>
								<dd class="text-gray-900">{task.config.maxWorkers || 50}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">超时时间:</dt>
								<dd class="text-gray-900">{task.config.timeout || 30}秒</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">速率限制:</dt>
								<dd class="text-gray-900">{task.config.rateLimit || 100}/秒</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">服务识别:</dt>
								<dd class="text-gray-900">{task.config.enableService ? '启用' : '禁用'}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">Banner抓取:</dt>
								<dd class="text-gray-900">{task.config.enableBanner ? '启用' : '禁用'}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-600">SSL检测:</dt>
								<dd class="text-gray-900">{task.config.enableSSL ? '启用' : '禁用'}</dd>
							</div>
						</dl>
					</div>
				</div>

				{#if task.description}
					<div class="mt-6">
						<h4 class="font-medium text-gray-900 mb-2">任务描述</h4>
						<p class="text-sm text-gray-600">{task.description}</p>
					</div>
				{/if}
			</div>
		{:else if activeTab === 'logs'}
			<div class="bg-white rounded-lg shadow-sm border p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">执行日志</h3>
				<div class="bg-gray-50 rounded-lg p-4 font-mono text-sm">
					<div class="space-y-1">
						<div class="text-gray-600">[{formatDateTime(task.createdAt)}] 任务创建</div>
						{#if task.startTime}
							<div class="text-blue-600">[{formatDateTime(task.startTime)}] 任务开始执行</div>
						{/if}
						{#if task.status === 'running' && task.progress}
							<div class="text-yellow-600">
								[{formatDateTime(new Date().toISOString())}] 正在扫描 {task.progress
									.completed}/{task.progress.total} 端口
							</div>
						{/if}
						{#if task.endTime}
							<div class="text-green-600">
								[{formatDateTime(task.endTime)}] 任务完成
								{#if task.startTime}
									(耗时: {formatDuration(task.startTime, task.endTime)})
								{/if}
							</div>
						{/if}
						{#if task.error}
							<div class="text-red-600">[错误] {task.error}</div>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>
