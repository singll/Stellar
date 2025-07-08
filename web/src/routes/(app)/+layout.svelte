<!-- 基本布局结构 -->
<script lang="ts">
  import { page } from '$app/stores';
  import { cn } from '$lib/utils';
  import { auth } from '$lib/stores/auth';
  import { Button } from '$lib/components/ui/button/index';
</script>

<div class="min-h-screen bg-background">
  <!-- 顶部导航栏 -->
  <header class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <div class="container flex h-14 items-center">
      <div class="mr-4 flex">
        <a href="/" class="mr-6 flex items-center space-x-2">
          <span class="font-bold">Stellar</span>
        </a>
      </div>
      
      <nav class="flex items-center space-x-6 text-sm font-medium">
        <a href="/dashboard" class={cn("transition-colors hover:text-foreground/80", $page.url.pathname === "/dashboard" ? "text-foreground" : "text-foreground/60")}>
          仪表盘
        </a>
        <a href="/projects" class={cn("transition-colors hover:text-foreground/80", $page.url.pathname.startsWith("/projects") ? "text-foreground" : "text-foreground/60")}>
          项目管理
        </a>
        <a href="/assets" class={cn("transition-colors hover:text-foreground/80", $page.url.pathname.startsWith("/assets") ? "text-foreground" : "text-foreground/60")}>
          资产管理
        </a>
        <a href="/tasks" class={cn("transition-colors hover:text-foreground/80", $page.url.pathname.startsWith("/tasks") ? "text-foreground" : "text-foreground/60")}>
          任务管理
        </a>
        <a href="/nodes" class={cn("transition-colors hover:text-foreground/80", $page.url.pathname.startsWith("/nodes") ? "text-foreground" : "text-foreground/60")}>
          节点管理
        </a>
      </nav>
      <div class="flex-1"></div>
      <Button on:click={() => auth.logout()} class="ml-auto" size="sm" variant="outline">登出</Button>
    </div>
  </header>

  <!-- 主要内容区域 -->
  <main class="container py-6">
    <slot />
  </main>
</div> 