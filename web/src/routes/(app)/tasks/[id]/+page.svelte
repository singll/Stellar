<!--
任务详情页面
显示任务的完整信息、执行状态、结果等
-->
<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { taskStore, taskActions } from '$lib/stores/tasks';
	import type { Task, TaskEvent, TaskLog } from '$lib/types/task';

	import TaskHeader from '$lib/components/tasks/TaskHeader.svelte';
	import TaskProgress from '$lib/components/tasks/TaskProgress.svelte';
	import TaskConfig from '$lib/components/tasks/TaskConfig.svelte';
	import TaskResult from '$lib/components/tasks/TaskResult.svelte';
	import TaskEvents from '$lib/components/tasks/TaskEvents.svelte';
	import TaskLogs from '$lib/components/tasks/TaskLogs.svelte';
	import LoadingSpinner from '$lib/components/ui/LoadingSpinner.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	// 任务ID从路由参数获取
	let taskId = $derived($page.params.id);

	// 响应式状态
	let activeTab = $state('overview');
	let task = $state<Task | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let events = $state<TaskEvent[]>([]);
	let logs = $state<TaskLog[]>([]);
	let eventSource = $state<EventSource | null>(null);
	let logSource = $state<EventSource | null>(null);

	// Store 订阅
	let store = $state();
	taskStore.subscribe((value) => {
		store = value;
	});

	// 标签页配置
	const tabs = [
		{ key: 'overview', label: '概览', icon: 'fas fa-info-circle' },
		{ key: 'config', label: '配置', icon: 'fas fa-cog' },
		{ key: 'result', label: '结果', icon: 'fas fa-chart-bar' },
		{ key: 'events', label: '事件', icon: 'fas fa-bell' },
		{ key: 'logs', label: '日志', icon: 'fas fa-file-alt' }
	];

	onMount(async () => {
		await loadTask();

		// 如果任务正在运行，启动实时更新
		if (task?.status === 'running') {
			startRealTimeUpdates();
		}
	});

	onDestroy(() => {
		// 清理SSE连接
		stopRealTimeUpdates();
	});

	// 加载任务详情
	async function loadTask() {
		loading = true;
		error = null;

		try {
			await taskActions.selectTask(taskId);
			task = store?.selectedTask || null;

			if (!task) {
				error = '任务不存在';
				return;
			}

			// 加载任务结果
			if (task.status === 'completed' || task.status === 'failed') {
				await taskActions.loadTaskResult(taskId);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : '加载任务失败';
		} finally {
			loading = false;
		}
	}

	// 启动实时更新
	function startRealTimeUpdates() {
		// 启动事件流
		eventSource = taskActions.getTaskEventStream(
			taskId,
			(event: TaskEvent) => {
				events = [event, ...events].slice(0, 100); // 保留最新100条

				// 如果是任务状态变更事件，重新加载任务
				if (event.type === 'status_changed') {
					loadTask();
				}
			},
			(error: Error) => {
				console.error('事件流错误:', error);
			}
		);

		// 启动日志流
		logSource = taskActions.getTaskLogStream(
			taskId,
			(log: TaskLog) => {
				logs = [log, ...logs].slice(0, 200); // 保留最新200条
			},
			(error: Error) => {
				console.error('日志流错误:', error);
			}
		);
	}

	// 停止实时更新
	function stopRealTimeUpdates() {
		if (eventSource) {
			eventSource.close();
			eventSource = null;
		}

		if (logSource) {
			logSource.close();
			logSource = null;
		}
	}

	// 任务操作
	async function startTask() {
		const success = await taskActions.startTask(taskId);
		if (success) {
			await loadTask();
			startRealTimeUpdates();
		}
	}

	async function cancelTask() {
		const success = await taskActions.cancelTask(taskId);
		if (success) {
			await loadTask();
			stopRealTimeUpdates();
		}
	}

	async function restartTask() {
		const success = await taskActions.restartTask(taskId);
		if (success) {
			await loadTask();
			startRealTimeUpdates();
		}
	}

	async function cloneTask() {
		const success = await taskActions.cloneTask(taskId);
		if (success) {
			// 跳转到新任务
			goto('/tasks');
		}
	}

	async function deleteTask() {
		if (!confirm('确定要删除这个任务吗？此操作不可撤销。')) {
			return;
		}

		const success = await taskActions.deleteTask(taskId);
		if (success) {
			goto('/tasks');
		}
	}

	// 下载任务结果
	async function downloadResult(format: 'json' | 'csv' | 'xml' = 'json') {
		await taskActions.downloadTaskResult(taskId, format);
	}

	// 返回任务列表
	function goBack() {
		goto('/tasks');
	}

	// 获取状态颜色
	function getStatusColor(status: string) {
		switch (status) {
			case 'pending':
				return 'gray';
			case 'queued':
				return 'blue';
			case 'running':
				return 'yellow';
			case 'completed':
				return 'green';
			case 'failed':
				return 'red';
			case 'canceled':
				return 'gray';
			case 'timeout':
				return 'orange';
			default:
				return 'gray';
		}
	}

	// 获取状态文本
	function getStatusText(status: string) {
		switch (status) {
			case 'pending':
				return '等待中';
			case 'queued':
				return '队列中';
			case 'running':
				return '运行中';
			case 'completed':
				return '已完成';
			case 'failed':
				return '失败';
			case 'canceled':
				return '已取消';
			case 'timeout':
				return '超时';
			default:
				return status;
		}
	}
</script>

<svelte:head>
	<title>任务详情 - {task?.name || '加载中'} - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	{#if loading}
		<div class="text-center py-12">
			<LoadingSpinner />
			<p class="text-gray-600 dark:text-gray-400 mt-2">加载任务详情...</p>
		</div>
	{:else if error}
		<div class="text-center py-12">
			<div class="text-red-500 mb-4">
				<i class="fas fa-exclamation-triangle text-3xl"></i>
			</div>
			<p class="text-red-600 dark:text-red-400 mb-4">{error}</p>
			<div class="flex gap-2 justify-center">
				<Button variant="outline" onclick={goBack}>返回列表</Button>
				<Button onclick={loadTask}>重新加载</Button>
			</div>
		</div>
	{:else if task}
		<!-- 任务头部 -->
		<div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
			<div class="flex items-center gap-4">
				<Button variant="ghost" onclick={goBack}>
					<i class="fas fa-arrow-left mr-2"></i>
					返回
				</Button>
				<div>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
						{task.name}
					</h1>
					<div class="flex items-center gap-2 mt-1">
						<Badge color={getStatusColor(task.status)}>
							{getStatusText(task.status)}
						</Badge>
						<span class="text-gray-600 dark:text-gray-400 text-sm">
							ID: {task.id}
						</span>
					</div>
				</div>
			</div>

			<!-- 操作按钮 -->
			<div class="flex gap-2">
				{#if task.status === 'pending' || task.status === 'failed' || task.status === 'canceled'}
					<Button onclick={startTask}>
						<i class="fas fa-play mr-2"></i>
						启动
					</Button>
				{/if}

				{#if task.status === 'running'}
					<Button variant="outline" onclick={cancelTask}>
						<i class="fas fa-stop mr-2"></i>
						取消
					</Button>
				{/if}

				{#if task.status === 'completed' || task.status === 'failed'}
					<Button variant="outline" onclick={restartTask}>
						<i class="fas fa-redo mr-2"></i>
						重启
					</Button>
				{/if}

				<Button variant="outline" onclick={cloneTask}>
					<i class="fas fa-copy mr-2"></i>
					克隆
				</Button>

				<Button variant="destructive" onclick={deleteTask}>
					<i class="fas fa-trash mr-2"></i>
					删除
				</Button>
			</div>
		</div>

		<!-- 任务头部信息 -->
		<TaskHeader {task} />

		<!-- 进度条 -->
		{#if task.status === 'running'}
			<div class="mb-6">
				<TaskProgress {task} />
			</div>
		{/if}

		<!-- 标签页 -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700"
		>
			<Tabs {tabs} bind:active={activeTab} />

			<div class="p-6">
				{#if activeTab === 'overview'}
					<div class="space-y-6">
						<TaskHeader {task} detailed={true} />
					</div>
				{:else if activeTab === 'config'}
					<TaskConfig {task} />
				{:else if activeTab === 'result'}
					<div class="space-y-4">
						{#if task.status === 'completed' || task.status === 'failed'}
							<div class="flex justify-end gap-2">
								<Button variant="outline" onclick={() => downloadResult('json')}>
									<i class="fas fa-download mr-2"></i>
									JSON
								</Button>
								<Button variant="outline" onclick={() => downloadResult('csv')}>
									<i class="fas fa-download mr-2"></i>
									CSV
								</Button>
								<Button variant="outline" onclick={() => downloadResult('xml')}>
									<i class="fas fa-download mr-2"></i>
									XML
								</Button>
							</div>
						{/if}
						<TaskResult {task} result={store?.taskResult} />
					</div>
				{:else if activeTab === 'events'}
					<TaskEvents {task} {events} />
				{:else if activeTab === 'logs'}
					<TaskLogs {task} {logs} />
				{/if}
			</div>
		</div>
	{/if}
</div>
