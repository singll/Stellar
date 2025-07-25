<script lang="ts">
	import { onMount } from 'svelte';
	import type { Task } from '$lib/types/task';
	import { Button } from '$lib/components/ui/button';
	import Icon from '$lib/components/ui/Icon.svelte';

	interface Props {
		tasks: Task[];
		selectedTaskId?: string;
		onTaskSelect?: (taskId: string) => void;
		readonly?: boolean;
	}

	let {
		tasks,
		selectedTaskId,
		onTaskSelect,
		readonly = false
	}: Props = $props();

	let canvasElement: HTMLCanvasElement;
	let ctx: CanvasRenderingContext2D | null = null;
	let canvasWidth = 800;
	let canvasHeight = 600;

	// 节点位置计算
	interface NodePosition {
		x: number;
		y: number;
		task: Task;
		level: number;
	}

	let nodePositions = $state<NodePosition[]>([]);

	// 构建依赖图
	const buildDependencyGraph = () => {
		const taskMap = new Map(tasks.map(task => [task.id, task]));
		const levels = new Map<string, number>();
		const visited = new Set<string>();

		// 计算每个任务的层级
		const calculateLevel = (taskId: string): number => {
			if (levels.has(taskId)) {
				return levels.get(taskId)!;
			}

			const task = taskMap.get(taskId);
			if (!task || !task.dependencies || task.dependencies.length === 0) {
				levels.set(taskId, 0);
				return 0;
			}

			let maxLevel = -1;
			for (const depId of task.dependencies) {
				if (visited.has(depId)) {
					// 检测循环依赖
					console.warn(`Circular dependency detected: ${taskId} -> ${depId}`);
					continue;
				}
				visited.add(depId);
				maxLevel = Math.max(maxLevel, calculateLevel(depId));
				visited.delete(depId);
			}

			const level = maxLevel + 1;
			levels.set(taskId, level);
			return level;
		};

		// 为所有任务计算层级
		tasks.forEach(task => calculateLevel(task.id));

		// 按层级分组任务
		const levelGroups = new Map<number, Task[]>();
		levels.forEach((level, taskId) => {
			const task = taskMap.get(taskId)!;
			if (!levelGroups.has(level)) {
				levelGroups.set(level, []);
			}
			levelGroups.get(level)!.push(task);
		});

		// 计算节点位置
		const positions: NodePosition[] = [];
		const nodeWidth = 120;
		const nodeHeight = 60;
		const horizontalSpacing = 180;
		const verticalSpacing = 100;
		const maxLevels = Math.max(...levels.values()) + 1;

		levelGroups.forEach((tasksInLevel, level) => {
			const levelY = 50 + level * verticalSpacing;
			const levelWidth = tasksInLevel.length * horizontalSpacing;
			const startX = (canvasWidth - levelWidth) / 2 + horizontalSpacing / 2;

			tasksInLevel.forEach((task, index) => {
				positions.push({
					x: startX + index * horizontalSpacing,
					y: levelY,
					task,
					level
				});
			});
		});

		nodePositions = positions;
		canvasHeight = Math.max(400, maxLevels * verticalSpacing + 100);
	};

	// 绘制依赖图
	const drawGraph = () => {
		if (!ctx) return;

		// 清空画布
		ctx.clearRect(0, 0, canvasWidth, canvasHeight);

		// 绘制连接线
		ctx.strokeStyle = '#64748b';
		ctx.lineWidth = 2;

		nodePositions.forEach(nodePos => {
			const task = nodePos.task;
			if (task.dependencies) {
				task.dependencies.forEach(depId => {
					const depNode = nodePositions.find(n => n.task.id === depId);
					if (depNode) {
						// 绘制箭头连接线
						drawArrow(ctx!, depNode.x, depNode.y + 30, nodePos.x, nodePos.y + 30);
					}
				});
			}
		});

		// 绘制任务节点
		nodePositions.forEach(nodePos => {
			drawTaskNode(ctx!, nodePos);
		});
	};

	// 绘制箭头
	const drawArrow = (ctx: CanvasRenderingContext2D, fromX: number, fromY: number, toX: number, toY: number) => {
		const arrowLength = 10;
		const arrowAngle = Math.PI / 6;

		ctx.beginPath();
		ctx.moveTo(fromX, fromY);
		ctx.lineTo(toX, toY);
		ctx.stroke();

		// 箭头头部
		const angle = Math.atan2(toY - fromY, toX - fromX);
		ctx.beginPath();
		ctx.moveTo(toX, toY);
		ctx.lineTo(
			toX - arrowLength * Math.cos(angle - arrowAngle),
			toY - arrowLength * Math.sin(angle - arrowAngle)
		);
		ctx.moveTo(toX, toY);
		ctx.lineTo(
			toX - arrowLength * Math.cos(angle + arrowAngle),
			toY - arrowLength * Math.sin(angle + arrowAngle)
		);
		ctx.stroke();
	};

	// 绘制任务节点
	const drawTaskNode = (ctx: CanvasRenderingContext2D, nodePos: NodePosition) => {
		const { x, y, task } = nodePos;
		const nodeWidth = 120;
		const nodeHeight = 60;
		const isSelected = task.id === selectedTaskId;

		// 节点背景
		ctx.fillStyle = isSelected ? '#3b82f6' : getStatusColor(task.status);
		ctx.strokeStyle = isSelected ? '#1d4ed8' : '#64748b';
		ctx.lineWidth = isSelected ? 3 : 1;

		ctx.beginPath();
		ctx.roundRect(x - nodeWidth/2, y - nodeHeight/2, nodeWidth, nodeHeight, 8);
		ctx.fill();
		ctx.stroke();

		// 任务名称
		ctx.fillStyle = isSelected ? 'white' : '#1f2937';
		ctx.font = '12px sans-serif';
		ctx.textAlign = 'center';
		ctx.textBaseline = 'middle';
		
		const taskName = task.name.length > 12 ? task.name.substring(0, 12) + '...' : task.name;
		ctx.fillText(taskName, x, y - 8);

		// 任务状态
		ctx.font = '10px sans-serif';
		ctx.fillStyle = isSelected ? '#e5e7eb' : '#6b7280';
		ctx.fillText(task.status.toUpperCase(), x, y + 8);

		// 进度指示器
		if (task.status === 'running' && task.progress !== undefined) {
			const progressWidth = nodeWidth - 20;
			const progressHeight = 4;
			const progressX = x - progressWidth/2;
			const progressY = y + 18;

			// 进度条背景
			ctx.fillStyle = isSelected ? 'rgba(255,255,255,0.3)' : '#e5e7eb';
			ctx.fillRect(progressX, progressY, progressWidth, progressHeight);

			// 进度条
			ctx.fillStyle = isSelected ? 'white' : '#10b981';
			ctx.fillRect(progressX, progressY, progressWidth * (task.progress / 100), progressHeight);
		}
	};

	// 获取状态颜色
	const getStatusColor = (status: string): string => {
		switch (status) {
			case 'pending': return '#f3f4f6';
			case 'queued': return '#fbbf24';
			case 'running': return '#3b82f6';
			case 'completed': return '#10b981';
			case 'failed': return '#ef4444';
			case 'cancelled': return '#6b7280';
			default: return '#f3f4f6';
		}
	};

	// 处理画布点击
	const handleCanvasClick = (event: MouseEvent) => {
		if (!onTaskSelect) return;

		const rect = canvasElement.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const y = event.clientY - rect.top;

		// 检查点击的节点
		const clickedNode = nodePositions.find(nodePos => {
			const dx = x - nodePos.x;
			const dy = y - nodePos.y;
			return Math.abs(dx) <= 60 && Math.abs(dy) <= 30;
		});

		if (clickedNode) {
			onTaskSelect(clickedNode.task.id);
		}
	};

	onMount(() => {
		if (canvasElement) {
			ctx = canvasElement.getContext('2d');
			buildDependencyGraph();
			drawGraph();
		}
	});

	// 响应数据变化
	$effect(() => {
		if (tasks && ctx) {
			buildDependencyGraph();
			drawGraph();
		}
	});

	// 导出图像
	const exportGraph = () => {
		if (!canvasElement) return;
		
		const link = document.createElement('a');
		link.download = 'task-dependency-graph.png';
		link.href = canvasElement.toDataURL();
		link.click();
	};
</script>

<div class="w-full bg-white rounded-lg border shadow-sm">
	<div class="flex items-center justify-between p-4 border-b">
		<div class="flex items-center gap-2">
			<Icon name="git-branch" class="h-5 w-5 text-blue-600" />
			<h3 class="text-lg font-semibold">任务依赖关系图</h3>
		</div>
		<div class="flex items-center gap-2">
			<Button variant="outline" size="sm" onclick={exportGraph}>
				<Icon name="download" class="h-4 w-4 mr-2" />
				导出图像
			</Button>
		</div>
	</div>

	<div class="p-4">
		{#if tasks.length === 0}
			<div class="text-center py-8 text-gray-500">
				<Icon name="git-branch" class="h-12 w-12 mx-auto mb-3 text-gray-300" />
				<p>暂无任务数据</p>
			</div>
		{:else}
			<div class="relative">
				<canvas
					bind:this={canvasElement}
					width={canvasWidth}
					height={canvasHeight}
					class="border rounded-lg bg-gray-50 cursor-pointer max-w-full"
					onclick={handleCanvasClick}
				></canvas>
			</div>

			<!-- 图例 -->
			<div class="mt-4 flex flex-wrap gap-4 text-sm">
				<div class="flex items-center gap-2">
					<div class="w-4 h-4 rounded bg-gray-100 border"></div>
					<span>等待中</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="w-4 h-4 rounded bg-yellow-400"></div>
					<span>队列中</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="w-4 h-4 rounded bg-blue-500"></div>
					<span>执行中</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="w-4 h-4 rounded bg-green-500"></div>
					<span>已完成</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="w-4 h-4 rounded bg-red-500"></div>
					<span>已失败</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="w-4 h-4 rounded bg-gray-500"></div>
					<span>已取消</span>
				</div>
			</div>
		{/if}
	</div>
</div>