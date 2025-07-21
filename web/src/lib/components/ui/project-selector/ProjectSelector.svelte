<!-- 项目选择器组件 -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import {
		Popover,
		PopoverContent,
		PopoverTrigger
	} from '$lib/components/ui/popover';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ProjectAPI } from '$lib/api/projects';
	import type { Project } from '$lib/types/project';

	// Props
	interface Props {
		selectedProjectId?: string;
		selectedProjectName?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
		onProjectSelect?: (project: Project | null) => void;
	}

	let {
		selectedProjectId = $bindable(),
		selectedProjectName = $bindable(),
		placeholder = '选择项目...',
		required = false,
		disabled = false,
		onProjectSelect
	}: Props = $props();

	// State
	let projects: Project[] = $state([]);
	let searchQuery = $state('');
	let loading = $state(false);
	let open = $state(false);
	let searchTimeout: NodeJS.Timeout;

	// 搜索项目
	async function searchProjects(query: string = '') {
		try {
			loading = true;
			const response = await ProjectAPI.getProjects({ 
				search: query, 
				limit: 20 
			});
			projects = response.data;
		} catch (error) {
			console.error('搜索项目失败:', error);
			projects = [];
		} finally {
			loading = false;
		}
	}

	// 处理搜索输入
	function handleSearchInput() {
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}
		
		searchTimeout = setTimeout(() => {
			searchProjects(searchQuery);
		}, 300);
	}

	// 选择项目
	function selectProject(project: Project) {
		selectedProjectId = project.id;
		selectedProjectName = project.name;
		open = false;
		searchQuery = '';
		
		if (onProjectSelect) {
			onProjectSelect(project);
		}
	}

	// 清除选择
	function clearSelection() {
		selectedProjectId = '';
		selectedProjectName = '';
		
		if (onProjectSelect) {
			onProjectSelect(null);
		}
	}

	// 组件挂载时加载项目
	onMount(() => {
		searchProjects();
	});

	// 监听搜索输入变化
	$effect(() => {
		if (searchQuery !== undefined) {
			handleSearchInput();
		}
	});
</script>

<div class="space-y-2">
	<Label class="text-sm font-medium">
		选择项目
		{#if required}<span class="text-red-500">*</span>{/if}
	</Label>

	<Popover bind:open>
		<PopoverTrigger>
			{#snippet children({ builder })}
				<Button
					builders={[builder]}
					variant="outline"
					role="combobox"
					aria-expanded={open}
					class="w-full justify-between"
					{disabled}
				>
					<span class="truncate">
						{selectedProjectName || placeholder}
					</span>
					<Icon name="chevrons-up-down" class="ml-2 h-4 w-4 shrink-0 opacity-50" />
				</Button>
			{/snippet}
		</PopoverTrigger>
		
		<PopoverContent class="w-full p-0">
			<div class="p-2">
				<div class="relative">
					<Icon name="search" class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
					<Input
						bind:value={searchQuery}
						placeholder="搜索项目..."
						class="pl-8"
					/>
				</div>
			</div>
			
			<div class="max-h-60 overflow-auto">
				{#if loading}
					<div class="p-4 text-center text-sm text-muted-foreground">
						搜索中...
					</div>
				{:else if projects.length === 0}
					<div class="p-4 text-center text-sm text-muted-foreground">
						{searchQuery ? '没有找到相关项目' : '暂无项目'}
					</div>
				{:else}
					<div class="space-y-1 p-1">
						{#each projects as project}
							<Button
								variant="ghost"
								class="w-full justify-start h-auto p-2"
								onclick={() => selectProject(project)}
							>
								<div class="flex flex-col items-start text-left">
									<span class="font-medium">{project.name}</span>
									<span class="text-xs text-muted-foreground truncate">
										ID: {project.id}
									</span>
									{#if project.description}
										<span class="text-xs text-muted-foreground truncate">
											{project.description}
										</span>
									{/if}
								</div>
							</Button>
						{/each}
					</div>
				{/if}
			</div>
			
			{#if selectedProjectId}
				<div class="border-t p-2">
					<Button
						variant="ghost" 
						class="w-full text-destructive"
						onclick={clearSelection}
					>
						<Icon name="x" class="mr-2 h-4 w-4" />
						清除选择
					</Button>
				</div>
			{/if}
		</PopoverContent>
	</Popover>
	
	{#if selectedProjectId}
		<p class="text-xs text-muted-foreground">
			已选择: {selectedProjectName} (ID: {selectedProjectId})
		</p>
	{/if}
</div>