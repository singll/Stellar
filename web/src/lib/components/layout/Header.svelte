<script lang="ts">
  import { auth } from "$lib/stores/auth";
  import { Button } from "$lib/components/ui/button";
  import { authApi } from "$lib/api/auth";
  import { Moon, Sun } from "lucide-svelte";
  import { themeStore } from "$lib/stores/theme";

  let isDark = $derived($themeStore.mode === 'dark');

  async function handleLogout() {
    try {
      await authApi.logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  }
</script>

<header class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
  <div class="container flex h-14 items-center">
    <div class="mr-4 flex">
      <a href="/" class="mr-6 flex items-center space-x-2">
        <span class="font-bold">Stellar</span>
      </a>
    </div>
    <div class="flex flex-1 items-center justify-between space-x-2 md:justify-end">
      <nav class="flex items-center space-x-6">
        <a href="/dashboard" class="font-bold">仪表盘</a>
        <a href="/assets" class="text-foreground/60 transition-colors hover:text-foreground">资产</a>
        <a href="/tasks" class="text-foreground/60 transition-colors hover:text-foreground">任务</a>
        <a href="/nodes" class="text-foreground/60 transition-colors hover:text-foreground">节点</a>
      </nav>
      
      <div class="flex items-center space-x-2">
        {#if $auth.user}
          <span class="text-sm text-foreground/60">{$auth.user.username}</span>
          <Button variant="ghost" size="sm" on:click={handleLogout}>退出</Button>
        {/if}
      </div>
    </div>
  </div>
</header> 