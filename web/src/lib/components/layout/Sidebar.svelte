<script lang="ts">
  import { page } from "$app/stores";
</script>

<aside class="fixed left-0 top-0 z-30 flex h-screen w-64 flex-col border-r bg-sidebar">
  <div class="flex h-14 items-center border-b px-6">
    <a href="/" class="flex items-center space-x-2">
      <span class="font-bold">Stellar</span>
    </a>
  </div>
  
  <nav class="flex-1 space-y-1 p-4">
    <a
      href="/dashboard"
      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm {$page.url.pathname === '/dashboard' ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground'}"
    >
      <span>仪表盘</span>
    </a>
    
    <a
      href="/assets"
      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm {$page.url.pathname.startsWith('/assets') ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground'}"
    >
      <span>资产管理</span>
    </a>
    
    <a
      href="/tasks"
      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm {$page.url.pathname.startsWith('/tasks') ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground'}"
    >
      <span>任务管理</span>
    </a>
    
    <a
      href="/nodes"
      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm {$page.url.pathname.startsWith('/nodes') ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-sidebar-foreground/60 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground'}"
    >
      <span>节点管理</span>
    </a>
  </nav>
</aside> 