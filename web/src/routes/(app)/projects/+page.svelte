<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { ProjectAPI } from '$lib/api/projects';
  import type { Project, ProjectFilters } from '$lib/types/project';
  import { notifications } from '$lib/stores/notifications';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { Search, Plus, MoreHorizontal, Filter, ArrowUpDown } from 'lucide-svelte';
  import { onMount } from 'svelte';

  // 从页面数据获取初始数据
  let { data } = $props();
  
  // 响应式状态
  let projects = $state(data.projects.data);
  let stats = $state(data.stats);
  let loading = $state(false);
  let searchQuery = $state(data.searchParams.search || '');
  let currentPage = $state(data.searchParams.page);
  let totalPages = $state(Math.ceil(data.projects.total / data.searchParams.limit));

  // 筛选器状态
  let filters: ProjectFilters = $state({
    search: searchQuery,
    is_private: undefined,
    scan_status: undefined
  });

  // 搜索功能
  const handleSearch = async () => {
    loading = true;
    try {
      const url = new URL($page.url);
      url.searchParams.set('search', searchQuery);
      url.searchParams.set('page', '1');
      await goto(url.toString());
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '搜索失败: ' + (error instanceof Error ? error.message : '未知错误')
      });
    } finally {
      loading = false;
    }
  };

  // 删除项目
  const handleDeleteProject = async (projectId: string) => {
    if (!confirm('确定要删除这个项目吗？此操作不可逆。')) return;
    
    try {
      await ProjectAPI.deleteProject(projectId);
      projects = projects.filter(p => p.id !== projectId);
      notifications.add({
        type: 'success',
        message: '项目删除成功'
      });
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '删除项目失败: ' + (error instanceof Error ? error.message : '未知错误')
      });
    }
  };

  // 复制项目
  const handleDuplicateProject = async (project: Project) => {
    const newName = prompt('请输入新项目名称:', `${project.name} - 副本`);
    if (!newName) return;

    try {
      const newProject = await ProjectAPI.duplicateProject(project.id, newName);
      projects = [newProject, ...projects];
      notifications.add({
        type: 'success',
        message: '项目复制成功'
      });
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '复制项目失败: ' + (error instanceof Error ? error.message : '未知错误')
      });
    }
  };

  // 分页处理
  const handlePageChange = async (newPage: number) => {
    const url = new URL($page.url);
    url.searchParams.set('page', newPage.toString());
    await goto(url.toString());
  };

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN');
  };

  // 获取扫描状态颜色
  const getScanStatusColor = (status?: string) => {
    switch (status) {
      case 'running': return 'bg-blue-100 text-blue-800';
      case 'completed': return 'bg-green-100 text-green-800';
      case 'failed': return 'bg-red-100 text-red-800';
      case 'paused': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  // 监听搜索输入
  let searchTimeout: ReturnType<typeof setTimeout>;
  $effect(() => {
    // 在 effect 中正确跟踪 searchQuery 的变化
    const currentSearch = searchQuery;
    const initialSearch = data.searchParams.search;
    
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      if (currentSearch !== initialSearch) {
        handleSearch();
      }
    }, 500);
    
    // 清理函数
    return () => {
      clearTimeout(searchTimeout);
    };
  });
</script>

<svelte:head>
  <title>项目管理 - Stellar</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
  <!-- 页面标题和统计 -->
  <div class="mb-6">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">项目管理</h1>
        <p class="text-gray-600 mt-1">管理和监控您的安全扫描项目</p>
      </div>
      
      <Button href="/projects/create" class="flex items-center gap-2">
        <Plus class="h-4 w-4" />
        创建项目
      </Button>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-6">
      <Card>
        <CardHeader class="pb-3">
          <CardDescription>总项目数</CardDescription>
          <CardTitle class="text-2xl">{stats.total_projects}</CardTitle>
        </CardHeader>
      </Card>
      
      <Card>
        <CardHeader class="pb-3">
          <CardDescription>活跃项目</CardDescription>
          <CardTitle class="text-2xl">{stats.active_projects}</CardTitle>
        </CardHeader>
      </Card>
      
      <Card>
        <CardHeader class="pb-3">
          <CardDescription>总资产数</CardDescription>
          <CardTitle class="text-2xl">{stats.total_assets}</CardTitle>
        </CardHeader>
      </Card>
      
      <Card>
        <CardHeader class="pb-3">
          <CardDescription>发现漏洞</CardDescription>
          <CardTitle class="text-2xl">{stats.total_vulnerabilities}</CardTitle>
        </CardHeader>
      </Card>
      
      <Card>
        <CardHeader class="pb-3">
          <CardDescription>运行任务</CardDescription>
          <CardTitle class="text-2xl">{stats.total_tasks}</CardTitle>
        </CardHeader>
      </Card>
    </div>
  </div>

  <!-- 搜索和筛选 -->
  <div class="flex items-center gap-4 mb-6">
    <div class="relative flex-1 max-w-md">
      <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
      <Input
        bind:value={searchQuery}
        placeholder="搜索项目名称、描述或目标..."
        class="pl-10"
      />
    </div>
    
    <Button variant="outline" class="flex items-center gap-2">
      <Filter class="h-4 w-4" />
      筛选
    </Button>
    
    <Button variant="outline" class="flex items-center gap-2">
      <ArrowUpDown class="h-4 w-4" />
      排序
    </Button>
  </div>

  <!-- 项目列表 -->
  {#if loading}
    <div class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
  {:else if projects.length === 0}
    <div class="text-center py-12">
      <div class="text-gray-500 text-lg mb-4">暂无项目</div>
      <Button href="/projects/create" class="flex items-center gap-2 mx-auto">
        <Plus class="h-4 w-4" />
        创建第一个项目
      </Button>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each projects as project (project.id)}
        <Card class="hover:shadow-lg transition-shadow">
          <CardHeader>
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <CardTitle class="text-lg mb-2">
                  <a href="/projects/{project.id}" class="text-blue-600 hover:text-blue-800">
                    {project.name}
                  </a>
                </CardTitle>
                {#if project.description}
                  <CardDescription class="line-clamp-2">
                    {project.description}
                  </CardDescription>
                {/if}
              </div>
              
              <div class="flex items-center gap-2">
                {#if project.scan_status}
                  <Badge class={getScanStatusColor(project.scan_status)}>
                    {project.scan_status}
                  </Badge>
                {/if}
                
                <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                  <MoreHorizontal class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardHeader>
          
          <CardContent>
            <div class="space-y-3">
              {#if project.target}
                <div class="text-sm">
                  <span class="text-gray-500">目标:</span>
                  <span class="ml-1 font-mono text-blue-600">{project.target}</span>
                </div>
              {/if}
              
              <div class="flex items-center justify-between text-sm">
                <div class="flex items-center gap-4">
                  <span class="text-gray-500">资产: <span class="font-semibold text-gray-900">{project.assets_count || 0}</span></span>
                  <span class="text-gray-500">漏洞: <span class="font-semibold text-red-600">{project.vulnerabilities_count || 0}</span></span>
                </div>
                
                {#if project.is_private}
                  <Badge variant="outline">私有</Badge>
                {/if}
              </div>
              
              <div class="text-xs text-gray-500">
                创建时间: {formatDate(project.created_at)}
              </div>
              
              <div class="flex items-center gap-2 pt-2">
                <Button size="sm" href="/projects/{project.id}">查看详情</Button>
                <Button variant="outline" size="sm" onclick={() => handleDuplicateProject(project)}>复制</Button>
                <Button variant="outline" size="sm" class="text-red-600 hover:text-red-800" onclick={() => handleDeleteProject(project.id)}>删除</Button>
              </div>
            </div>
          </CardContent>
        </Card>
      {/each}
    </div>

    <!-- 分页 -->
    {#if totalPages > 1}
      <div class="flex justify-center items-center gap-2 mt-8">
        <Button 
          variant="outline" 
          disabled={currentPage <= 1}
          onclick={() => handlePageChange(currentPage - 1)}
        >
          上一页
        </Button>
        
        <span class="px-4 py-2 text-sm text-gray-600">
          第 {currentPage} 页，共 {totalPages} 页
        </span>
        
        <Button 
          variant="outline" 
          disabled={currentPage >= totalPages}
          onclick={() => handlePageChange(currentPage + 1)}
        >
          下一页
        </Button>
      </div>
    {/if}
  {/if}
</div> 