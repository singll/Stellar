<!-- 主页 -->
<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";

  onMount(() => {
    // 检查用户是否已登录
    if (!$auth.isAuthenticated) {
      goto("/login");
    }
  });
</script>

<div class="min-h-screen bg-gray-50">
  <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">
      <div class="border-4 border-dashed border-gray-200 rounded-lg h-96 flex items-center justify-center">
        <div class="text-center">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">
            欢迎来到 Stellar 系统
          </h1>
          <p class="text-lg text-gray-600 mb-8">
            这是一个基于 ScopeSentry 的安全资产管理和漏洞扫描平台
          </p>
          <div class="space-x-4">
            <button
              class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-2 px-4 rounded"
              onclick={() => goto("/assets")}
            >
              资产管理
            </button>
            <button
              class="bg-green-600 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
              onclick={() => goto("/scan")}
            >
              安全扫描
            </button>
            <button
              class="bg-red-600 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
              onclick={() => auth.logout()}
            >
              退出登录
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
