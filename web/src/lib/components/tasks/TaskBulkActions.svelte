<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		DropdownMenu,
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuTrigger,
		DropdownMenuSeparator
	} from '$lib/components/ui/dropdown-menu';
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { DialogContent, DialogHeader, DialogTitle, DialogFooter } from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Select } from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { notifications } from '$lib/stores/notifications';
	import type { Task } from '$lib/types/task';

	interface Props {
		selectedTasks: Task[];
		onTasksUpdate: () => void;
		disabled?: boolean;
	}

	let {
		selectedTasks,
		onTasksUpdate,
		disabled = false
	}: Props = $props();

	// 对话框状态
	let showBulkEditDialog = $state(false);
	let showBulkCreateDialog = $state(false);
	let showConfirmDialog = $state(false);
	let currentAction = $state('');
	let loading = $state(false);

	// 批量编辑表单
	let bulkEditForm = $state({
		priority: '',
		status: '',
		assignedNode: '',
		tags: [] as string[],
		description: ''
	});

	// 批量创建表单
	let bulkCreateForm = $state({
		namePrefix: '',
		count: 5,
		taskType: 'subdomain_enum',
		priority: 'medium',
		description: '',
		projectId: '',
		config: '{}'
	});

	let tagInput = $state('');

	// 可用的操作
	const bulkActions = [
		{
			id: 'start',
			label: '启动任务',
			icon: 'play',
			color: 'text-green-600',
			requiresRunning: false
		},
		{
			id: 'pause',
			label: '暂停任务',
			icon: 'pause',
			color: 'text-yellow-600',
			requiresRunning: true
		},
		{
			id: 'cancel',
			label: '取消任务',
			icon: 'square',
			color: 'text-red-600',
			requiresRunning: true
		},
		{
			id: 'retry',
			label: '重试任务',
			icon: 'refresh-cw',
			color: 'text-blue-600',
			requiresRunning: false
		},
		{
			id: 'delete',
			label: '删除任务',
			icon: 'trash',
			color: 'text-red-600',
			requiresRunning: false,
			dangerous: true
		}
	];

	// 优先级选项
	const priorityOptions = [
		{ value: '', label: '不修改' },
		{ value: 'low', label: '低优先级' },
		{ value: 'medium', label: '中优先级' },
		{ value: 'high', label: '高优先级' },
		{ value: 'urgent', label: '紧急' }
	];

	// 状态选项
	const statusOptions = [
		{ value: '', label: '不修改' },
		{ value: 'pending', label: '等待中' },
		{ value: 'queued', label: '队列中' },
		{ value: 'running', label: '运行中' },
		{ value: 'completed', label: '已完成' },
		{ value: 'failed', label: '失败' },
		{ value: 'cancelled', label: '已取消' }
	];

	// 任务类型选项
	const taskTypeOptions = [
		{ value: 'subdomain_enum', label: '子域名枚举' },
		{ value: 'port_scan', label: '端口扫描' },
		{ value: 'vuln_scan', label: '漏洞扫描' },
		{ value: 'asset_discovery', label: '资产发现' }
	];

	// 标签管理
	const addTag = () => {
		if (tagInput.trim() && !bulkEditForm.tags.includes(tagInput.trim())) {
			bulkEditForm.tags = [...bulkEditForm.tags, tagInput.trim()];
			tagInput = '';
		}
	};

	const removeTag = (tag: string) => {
		bulkEditForm.tags = bulkEditForm.tags.filter(t => t !== tag);
	};

	const handleTagKeydown = (event: KeyboardEvent) => {
		if (event.key === 'Enter') {
			event.preventDefault();
			addTag();
		}
	};

	// 执行批量操作
	const executeBulkAction = async (actionId: string) => {
		if (selectedTasks.length === 0) {
			notifications.add({
				type: 'warning',
				message: '请先选择要操作的任务'
			});
			return;
		}

		const action = bulkActions.find(a => a.id === actionId);
		if (!action) return;

		if (action.dangerous) {
			currentAction = actionId;
			showConfirmDialog = true;
			return;
		}

		await performBulkAction(actionId);
	};

	// 执行批量操作
	const performBulkAction = async (actionId: string) => {
		loading = true;
		try {
			const taskIds = selectedTasks.map(task => task.id);

			switch (actionId) {
				case 'start':
					// await TaskAPI.startTasks(taskIds);
					notifications.add({
						type: 'success',
						message: `成功启动 ${taskIds.length} 个任务`
					});
					break;
				case 'pause':
					// await TaskAPI.pauseTasks(taskIds);
					notifications.add({
						type: 'success',
						message: `成功暂停 ${taskIds.length} 个任务`
					});
					break;
				case 'cancel':
					// await TaskAPI.cancelTasks(taskIds);
					notifications.add({
						type: 'success',
						message: `成功取消 ${taskIds.length} 个任务`
					});
					break;
				case 'retry':
					// await TaskAPI.retryTasks(taskIds);
					notifications.add({
						type: 'success',
						message: `成功重试 ${taskIds.length} 个任务`
					});
					break;
				case 'delete':
					// await TaskAPI.deleteTasks(taskIds);
					notifications.add({
						type: 'success',
						message: `成功删除 ${taskIds.length} 个任务`
					});
					break;
			}

			onTasksUpdate();
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '批量操作失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
			showConfirmDialog = false;
		}
	};

	// 批量编辑
	const handleBulkEdit = async () => {
		loading = true;
		try {
			const taskIds = selectedTasks.map(task => task.id);
			const updateData = Object.fromEntries(
				Object.entries(bulkEditForm).filter(([, value]) => value !== '' && value !== null)
			);

			if (Object.keys(updateData).length === 0) {
				notifications.add({
					type: 'warning',
					message: '请至少修改一个字段'
				});
				return;
			}

			// await TaskAPI.bulkUpdateTasks(taskIds, updateData);
			
			notifications.add({
				type: 'success',
				message: `成功更新 ${taskIds.length} 个任务`
			});

			showBulkEditDialog = false;
			onTasksUpdate();
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '批量编辑失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 批量创建
	const handleBulkCreate = async () => {
		loading = true;
		try {
			const tasks = [];
			for (let i = 1; i <= bulkCreateForm.count; i++) {
				tasks.push({
					name: `${bulkCreateForm.namePrefix}_${i.toString().padStart(3, '0')}`,
					type: bulkCreateForm.taskType,
					priority: bulkCreateForm.priority,
					description: bulkCreateForm.description,
					projectId: bulkCreateForm.projectId,
					config: JSON.parse(bulkCreateForm.config || '{}')
				});
			}

			// await TaskAPI.createTasks(tasks);

			notifications.add({
				type: 'success',
				message: `成功创建 ${tasks.length} 个任务`
			});

			showBulkCreateDialog = false;
			onTasksUpdate();
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '批量创建失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		} finally {
			loading = false;
		}
	};

	// 导出任务
	const exportTasks = async (format: 'csv' | 'json') => {
		try {
			const taskIds = selectedTasks.map(task => task.id);
			// const blob = await TaskAPI.exportTasks(taskIds, format);
			
			// 模拟导出
			const data = selectedTasks.map(task => ({
				id: task.id,
				name: task.name,
				type: task.type,
				status: task.status,
				priority: task.priority,
				created_at: task.created_at,
				updated_at: task.updated_at
			}));

			const blob = new Blob([JSON.stringify(data, null, 2)], { 
				type: format === 'json' ? 'application/json' : 'text/csv' 
			});
			
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `tasks_export_${new Date().toISOString().split('T')[0]}.${format}`;
			a.click();
			URL.revokeObjectURL(url);

			notifications.add({
				type: 'success',
				message: `成功导出 ${taskIds.length} 个任务`
			});
		} catch (error) {
			notifications.add({
				type: 'error',
				message: '导出失败: ' + (error instanceof Error ? error.message : '未知错误')
			});
		}
	};
</script>

<div class="flex items-center gap-2">
	<span class="text-sm text-gray-600">
		已选择 {selectedTasks.length} 个任务
	</span>

	{#if selectedTasks.length > 0}
		<!-- 批量操作菜单 -->
		<DropdownMenu>
			<DropdownMenuTrigger asChild let:builder>
				<Button builders={[builder]} variant="outline" size="sm" {disabled}>
					<Icon name="more-horizontal" class="h-4 w-4 mr-2" />
					批量操作
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent>
				{#each bulkActions as action}
					<DropdownMenuItem 
						onclick={() => executeBulkAction(action.id)}
						class={action.dangerous ? 'text-red-600' : ''}
					>
						<Icon name={action.icon} class="h-4 w-4 mr-2 {action.color}" />
						{action.label}
					</DropdownMenuItem>
				{/each}
				
				<DropdownMenuSeparator />
				
				<DropdownMenuItem onclick={() => { showBulkEditDialog = true; }}>
					<Icon name="edit" class="h-4 w-4 mr-2" />
					批量编辑
				</DropdownMenuItem>
				
				<DropdownMenuSeparator />
				
				<DropdownMenuItem onclick={() => exportTasks('csv')}>
					<Icon name="download" class="h-4 w-4 mr-2" />
					导出CSV
				</DropdownMenuItem>
				<DropdownMenuItem onclick={() => exportTasks('json')}>
					<Icon name="download" class="h-4 w-4 mr-2" />
					导出JSON
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>

		<!-- 批量创建按钮 -->
		<Button 
			variant="outline" 
			size="sm" 
			onclick={() => { showBulkCreateDialog = true; }}
			{disabled}
		>
			<Icon name="copy" class="h-4 w-4 mr-2" />
			批量创建
		</Button>
	{/if}
</div>

<!-- 批量编辑对话框 -->
<DialogPrimitive.Root bind:open={showBulkEditDialog}>
	<DialogContent class="sm:max-w-lg">
		<DialogHeader>
			<DialogTitle>批量编辑任务</DialogTitle>
		</DialogHeader>

		<div class="space-y-4">
			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-2">
					<Label>优先级</Label>
					<Select 
						bind:value={bulkEditForm.priority}
						options={priorityOptions}
						placeholder="选择优先级"
					/>
				</div>
				<div class="space-y-2">
					<Label>状态</Label>
					<Select 
						bind:value={bulkEditForm.status}
						options={statusOptions}
						placeholder="选择状态"
					/>
				</div>
			</div>

			<div class="space-y-2">
				<Label>执行节点</Label>
				<Input
					bind:value={bulkEditForm.assignedNode}
					placeholder="留空表示不修改"
				/>
			</div>

			<div class="space-y-2">
				<Label>描述</Label>
				<Textarea
					bind:value={bulkEditForm.description}
					placeholder="留空表示不修改"
					rows={3}
				/>
			</div>

			<div class="space-y-2">
				<Label>标签</Label>
				<div class="flex flex-wrap gap-1 mb-2">
					{#each bulkEditForm.tags as tag}
						<span class="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 rounded-md text-sm">
							{tag}
							<button 
								onclick={() => removeTag(tag)}
								class="ml-1 hover:text-red-600"
							>
								<Icon name="x" class="h-3 w-3" />
							</button>
						</span>
					{/each}
				</div>
				<div class="flex gap-2">
					<Input
						bind:value={tagInput}
						placeholder="输入标签"
						onkeydown={handleTagKeydown}
						class="flex-1"
					/>
					<Button type="button" variant="outline" onclick={addTag}>
						<Icon name="plus" class="h-4 w-4" />
					</Button>
				</div>
			</div>
		</div>

		<DialogFooter>
			<Button variant="outline" onclick={() => { showBulkEditDialog = false; }}>
				取消
			</Button>
			<Button onclick={handleBulkEdit} {disabled}>
				{#if loading}
					<Icon name="loader-2" class="h-4 w-4 mr-2 animate-spin" />
				{/if}
				确认修改
			</Button>
		</DialogFooter>
	</DialogContent>
</DialogPrimitive.Root>

<!-- 批量创建对话框 -->
<DialogPrimitive.Root bind:open={showBulkCreateDialog}>
	<DialogContent class="sm:max-w-lg">
		<DialogHeader>
			<DialogTitle>批量创建任务</DialogTitle>
		</DialogHeader>

		<div class="space-y-4">
			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-2">
					<Label>名称前缀 <span class="text-red-500">*</span></Label>
					<Input
						bind:value={bulkCreateForm.namePrefix}
						placeholder="例如: scan_task"
						required
					/>
				</div>
				<div class="space-y-2">
					<Label>创建数量</Label>
					<Input
						type="number"
						bind:value={bulkCreateForm.count}
						min={1}
						max={100}
						required
					/>
				</div>
			</div>

			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-2">
					<Label>任务类型</Label>
					<Select 
						bind:value={bulkCreateForm.taskType}
						options={taskTypeOptions}
					/>
				</div>
				<div class="space-y-2">
					<Label>优先级</Label>
					<Select 
						bind:value={bulkCreateForm.priority}
						options={priorityOptions.filter(p => p.value !== '')}
					/>
				</div>
			</div>

			<div class="space-y-2">
				<Label>项目ID</Label>
				<Input
					bind:value={bulkCreateForm.projectId}
					placeholder="关联的项目ID（可选）"
				/>
			</div>

			<div class="space-y-2">
				<Label>任务配置</Label>
				<Textarea
					bind:value={bulkCreateForm.config}
					placeholder="JSON格式的任务配置"
					rows={4}
				/>
			</div>

			<div class="space-y-2">
				<Label>描述</Label>
				<Textarea
					bind:value={bulkCreateForm.description}
					placeholder="任务描述（可选）"
					rows={2}
				/>
			</div>
		</div>

		<DialogFooter>
			<Button variant="outline" onclick={() => { showBulkCreateDialog = false; }}>
				取消
			</Button>
			<Button onclick={handleBulkCreate} {disabled}>
				{#if loading}
					<Icon name="loader-2" class="h-4 w-4 mr-2 animate-spin" />
				{/if}
				创建任务
			</Button>
		</DialogFooter>
	</DialogContent>
</DialogPrimitive.Root>

<!-- 确认对话框 -->
<DialogPrimitive.Root bind:open={showConfirmDialog}>
	<DialogContent class="sm:max-w-md">
		<DialogHeader>
			<DialogTitle class="flex items-center gap-2">
				<Icon name="alert-triangle" class="h-5 w-5 text-red-600" />
				确认操作
			</DialogTitle>
		</DialogHeader>

		<div class="py-4">
			<p class="text-sm text-gray-600">
				确定要对选中的 {selectedTasks.length} 个任务执行此操作吗？此操作不可撤销。
			</p>
		</div>

		<DialogFooter>
			<Button variant="outline" onclick={() => { showConfirmDialog = false; }}>
				取消
			</Button>
			<Button 
				variant="destructive" 
				onclick={() => performBulkAction(currentAction)}
				{disabled}
			>
				{#if loading}
					<Icon name="loader-2" class="h-4 w-4 mr-2 animate-spin" />
				{/if}
				确认执行
			</Button>
		</DialogFooter>
	</DialogContent>
</DialogPrimitive.Root>