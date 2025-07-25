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
	let allProjects: Project[] = $state([]);
	let filteredProjects: Project[] = $state([]);
	let selectedProject: Project | null = $state(null);
	let searchQuery = $state('');
	let loading = $state(false);
	let open = $state(false);

	// 加载所有项目
	async function loadAllProjects() {
		try {
			loading = true;
			// 获取全量项目数据用于前端搜索
			allProjects = await ProjectAPI.searchProjects('', 100); // 增加限制数量
			filteredProjects = allProjects;
			console.log('📦 项目选择器加载项目列表:', allProjects.length, '个项目');
		} catch (error) {
			console.error('❌ 项目选择器加载项目列表失败:', error);
			allProjects = [];
			filteredProjects = [];
		} finally {
			loading = false;
		}
	}
	
	// 前端搜索项目
	function filterProjects(query: string) {
		console.log('🔍 过滤项目, 搜索词:', query, '全部项目数:', allProjects.length);
		
		if (!query.trim()) {
			filteredProjects = allProjects;
			console.log('📋 显示全部项目:', filteredProjects.length, '个');
			return;
		}
		
		const searchTerm = query.toLowerCase();
		filteredProjects = allProjects.filter(project => {
			const matches = (
				project.name?.toLowerCase().includes(searchTerm) ||
				project.id?.toLowerCase().includes(searchTerm) ||
				project.tag?.toLowerCase().includes(searchTerm) ||
				project.description?.toLowerCase().includes(searchTerm)
			);
			return matches;
		});
		
		console.log('🎯 搜索结果:', filteredProjects.length, '个项目匹配');
	}


	// 选择项目
	function selectProject(project: Project) {
		selectedProject = project;
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
		selectedProject = null;
		selectedProjectId = '';
		selectedProjectName = '';
		
		if (onProjectSelect) {
			onProjectSelect(null);
		}
	}

	// 组件挂载时加载项目
	onMount(() => {
		loadAllProjects();
	});

	// 监听搜索输入变化
	$effect(() => {
		filterProjects(searchQuery);
	});
	
	// 监听allProjects变化，确保在没有搜索词时显示所有项目
	$effect(() => {
		if (allProjects.length > 0 && !searchQuery) {
			filteredProjects = allProjects;
			console.log('🔄 项目数据更新，显示全部项目:', filteredProjects.length, '个');
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
						placeholder="搜索项目名称、ID或标签..."
						class="pl-8"
					/>
				</div>
			</div>
			
			<div class="max-h-60 overflow-auto">
				{#if loading}
					<div class="p-4 text-center text-sm text-muted-foreground">
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mx-auto mb-2"></div>
						加载项目列表...
					</div>
				{:else if filteredProjects.length === 0}
					<div class="p-4 text-center text-sm text-muted-foreground">
						{#if searchQuery}
							没有找到包含 "{searchQuery}" 的项目
						{:else if allProjects.length === 0}
							暂无项目，请先创建项目
						{:else}
							所有项目 ({allProjects.length} 个)
						{/if}
					</div>
				{:else}
					<div class="space-y-1 p-1">
						{#each filteredProjects as project}
							<Button
								variant="ghost"
								class="w-full justify-start h-auto p-2 hover:bg-blue-50"
								onclick={() => selectProject(project)}
							>
								<div class="flex flex-col items-start text-left w-full">
									<div class="flex items-center justify-between w-full">
										<span class="font-medium text-sm text-gray-900">{project.name || '未命名项目'}</span>
										{#if project.tag}
											<span class="px-2 py-0.5 text-xs bg-blue-100 text-blue-800 rounded-md flex-shrink-0 ml-2">{project.tag}</span>
										{/if}
									</div>
									<span class="text-xs text-gray-500 truncate mt-0.5">
										ID: {project.id}
									</span>
									{#if project.description}
										<span class="text-xs text-gray-500 truncate mt-0.5 max-w-full">
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
	
	{#if selectedProjectId && selectedProject}
		<div class="p-2 bg-blue-50 rounded-md border">
			<div class="flex items-center justify-between">
				<div class="flex-1">
					<p class="text-sm font-medium text-blue-900">{selectedProject.name}</p>
					<p class="text-xs text-blue-600 mt-0.5">ID: {selectedProjectId}</p>
					{#if selectedProject.description}
						<p class="text-xs text-blue-600 mt-0.5 truncate">{selectedProject.description}</p>
					{/if}
				</div>
				{#if selectedProject.tag}
					<span class="px-2 py-1 text-xs bg-blue-200 text-blue-800 rounded-md ml-2">{selectedProject.tag}</span>
				{/if}
			</div>
		</div>
	{/if}
</div>