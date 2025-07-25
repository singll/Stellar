<!--
可搜索项目选择器组件
支持输入框搜索和下拉列表选择功能
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ProjectAPI } from '$lib/api/projects';
	import type { Project } from '$lib/types/project';
	import { cn } from '$lib/utils';

	// Props
	interface Props {
		selectedProjectId?: string;
		selectedProjectName?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
		class?: string;
		onProjectSelect?: (project: Project | null) => void;
	}

	let {
		selectedProjectId = $bindable(),
		selectedProjectName = $bindable(),
		placeholder = '搜索并选择项目...',
		required = false,
		disabled = false,
		class: className = '',
		onProjectSelect
	}: Props = $props();

	// State
	let allProjects: Project[] = $state([]);
	let filteredProjects: Project[] = $state([]);
	let selectedProject: Project | null = $state(null);
	let inputValue = $state(''); // 输入框的值
	let loading = $state(false);
	let isOpen = $state(false);
	let inputRef: HTMLInputElement | null = $state(null);
	let dropdownRef: HTMLDivElement | null = $state(null);

	// 加载所有项目
	async function loadAllProjects() {
		try {
			loading = true;
			allProjects = await ProjectAPI.searchProjects('', 100);
			updateFilteredProjects();
			console.log('📦 [SearchableProjectSelector] 加载项目列表:', allProjects.length, '个项目');
		} catch (error) {
			console.error('❌ [SearchableProjectSelector] 加载项目列表失败:', error);
			allProjects = [];
			filteredProjects = [];
		} finally {
			loading = false;
		}
	}
	
	// 更新过滤的项目列表
	function updateFilteredProjects() {
		if (!inputValue.trim()) {
			filteredProjects = allProjects;
			return;
		}
		
		const searchTerm = inputValue.toLowerCase();
		filteredProjects = allProjects.filter(project => {
			return (
				project.name?.toLowerCase().includes(searchTerm) ||
				project.id?.toLowerCase().includes(searchTerm) ||
				project.tag?.toLowerCase().includes(searchTerm) ||
				project.description?.toLowerCase().includes(searchTerm)
			);
		});
		
		console.log('🎯 [SearchableProjectSelector] 搜索结果:', filteredProjects.length, '个项目匹配');
	}

	// 选择项目
	function selectProject(project: Project) {
		selectedProject = project;
		selectedProjectId = project.id;
		selectedProjectName = project.name;
		inputValue = project.name;
		isOpen = false;
		
		console.log('✅ [SearchableProjectSelector] 项目选择:', {
			id: project.id,
			name: project.name,
			tag: project.tag
		});
		
		if (onProjectSelect) {
			onProjectSelect(project);
		}
	}

	// 清除选择
	function clearSelection() {
		selectedProject = null;
		selectedProjectId = '';
		selectedProjectName = '';
		inputValue = '';
		isOpen = true;
		updateFilteredProjects();
		
		if (onProjectSelect) {
			onProjectSelect(null);
		}
	}

	// 处理输入变化
	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		inputValue = target.value;
		
		// 如果输入的内容与当前选中项目不匹配，清除选择
		if (selectedProject && inputValue !== selectedProject.name) {
			selectedProject = null;
			selectedProjectId = '';
			selectedProjectName = '';
		}
		
		// 更新过滤结果
		updateFilteredProjects();
		
		// 显示下拉列表
		if (!isOpen) {
			isOpen = true;
		}
	}

	// 处理焦点事件
	function handleFocus() {
		if (!disabled) {
			isOpen = true;
			updateFilteredProjects();
		}
	}

	// 处理键盘事件
	function handleKeyDown(event: KeyboardEvent) {
		if (disabled) return;

		switch (event.key) {
			case 'Escape':
				isOpen = false;
				inputRef?.blur();
				break;
			case 'Enter':
				event.preventDefault();
				if (filteredProjects.length === 1) {
					selectProject(filteredProjects[0]);
				}
				break;
			case 'ArrowDown':
				event.preventDefault();
				if (!isOpen) {
					isOpen = true;
				}
				break;
		}
	}

	// 点击外部关闭下拉列表
	function handleClickOutside(event: MouseEvent) {
		if (isOpen && inputRef && dropdownRef) {
			const target = event.target as Element;
			if (!inputRef.contains(target) && !dropdownRef.contains(target)) {
				isOpen = false;
			}
		}
	}

	// 组件挂载时加载项目
	onMount(() => {
		loadAllProjects();
	});

	// 监听外部selectedProjectId变化，只在必要时同步
	$effect(() => {
		if (selectedProjectId && allProjects.length > 0) {
			if (!selectedProject || selectedProject.id !== selectedProjectId) {
				const project = allProjects.find(p => p.id === selectedProjectId);
				if (project) {
					selectedProject = project;
					selectedProjectName = project.name;
					inputValue = project.name;
				}
			}
		} else if (!selectedProjectId && selectedProject) {
			selectedProject = null;
			selectedProjectName = '';
			inputValue = '';
		}
	});
</script>

<svelte:window onclick={handleClickOutside} />

<div class={cn('relative', className)}>
	<div class="space-y-2">
		<Label class="text-sm font-medium">
			{#snippet children()}
				选择项目
				{#if required}<span class="text-red-500">*</span>{/if}
			{/snippet}
		</Label>

		<div class="relative">
			<!-- 搜索输入框 -->
			<div class="relative">
				<Input
					bind:ref={inputRef}
					bind:value={inputValue}
					{placeholder}
					{disabled}
					class="pr-10 {selectedProject ? 'border-blue-500' : ''}"
					oninput={handleInput}
					onfocus={handleFocus}
					onkeydown={handleKeyDown}
				/>
				
				<!-- 右侧图标 -->
				<div class="absolute inset-y-0 right-0 flex items-center pr-3">
					{#if selectedProject}
						<button
							type="button"
							onclick={clearSelection}
							class="text-gray-400 hover:text-gray-600 focus:outline-none"
							aria-label="清除选择"
							{disabled}
						>
							<Icon name="x" class="h-4 w-4" />
						</button>
					{:else}
						<Icon 
							name={isOpen ? "chevron-up" : "chevron-down"} 
							class="h-4 w-4 text-gray-400" 
						/>
					{/if}
				</div>
			</div>

			<!-- 下拉列表 -->
			{#if isOpen}
				<div 
					bind:this={dropdownRef}
					class="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto"
				>
					{#if loading}
						<!-- 加载状态 -->
						<div class="p-4 text-center text-sm text-gray-500">
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mx-auto mb-2"></div>
							加载项目列表...
						</div>
					{:else if filteredProjects.length === 0}
						<!-- 空状态 -->
						<div class="p-4 text-center text-sm text-gray-500">
							{#if inputValue}
								没有找到包含 "{inputValue}" 的项目
							{:else if allProjects.length === 0}
								暂无项目，请先创建项目
							{:else}
								请输入搜索关键词
							{/if}
						</div>
					{:else}
						<!-- 项目列表 -->
						<div class="py-1">
							{#each filteredProjects as project}
								<button
									type="button"
									class="w-full px-4 py-2 text-left hover:bg-blue-50 focus:bg-blue-50 focus:outline-none {selectedProject?.id === project.id ? 'bg-blue-100' : ''}"
									onclick={() => selectProject(project)}
								>
									<div class="flex flex-col">
										<div class="flex items-center justify-between">
											<span class="font-medium text-sm text-gray-900">
												{project.name || '未命名项目'}
											</span>
											{#if project.tag}
												<span class="px-2 py-0.5 text-xs bg-blue-100 text-blue-800 rounded-md flex-shrink-0 ml-2">
													{project.tag}
												</span>
											{/if}
										</div>
										<span class="text-xs text-gray-500 truncate mt-0.5">
											ID: {project.id}
										</span>
										{#if project.description}
											<span class="text-xs text-gray-500 truncate mt-0.5">
												{project.description}
											</span>
										{/if}
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- 已选择项目的显示 -->
		{#if selectedProject}
			<div class="p-2 bg-blue-50 border border-blue-200 rounded-md">
				<div class="flex items-center justify-between">
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium text-blue-900 truncate">
							{selectedProject.name}
						</p>
						<p class="text-xs text-blue-600 mt-0.5">
							ID: {selectedProject.id}
						</p>
						{#if selectedProject.description}
							<p class="text-xs text-blue-600 mt-0.5 truncate">
								{selectedProject.description}
							</p>
						{/if}
					</div>
					{#if selectedProject.tag}
						<span class="px-2 py-1 text-xs bg-blue-200 text-blue-800 rounded-md ml-2 flex-shrink-0">
							{selectedProject.tag}
						</span>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>