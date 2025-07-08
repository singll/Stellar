<!-- 资产详情页面 -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { assetApi } from '$lib/api/asset';
  import type { Asset } from '$lib/types/asset';
  import { notifications } from '$lib/stores/notifications';
  import { goto } from '$app/navigation';

  let asset: Asset | null = null;
  let loading = true;

  async function loadAsset() {
    try {
      loading = true;
      const response = await assetApi.getAssetById($page.params.id);
      asset = response.data;
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '加载资产详情失败'
      });
      goto('/assets');
    } finally {
      loading = false;
    }
  }

  async function handleScan() {
    try {
      await assetApi.scanAsset($page.params.id);
      notifications.add({
        type: 'success',
        message: '扫描任务已启动'
      });
      loadAsset();
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '启动扫描失败'
      });
    }
  }

  async function handleDelete() {
    if (!confirm('确定要删除此资产吗？')) {
      return;
    }

    try {
      await assetApi.deleteAsset($page.params.id);
      notifications.add({
        type: 'success',
        message: '资产已删除'
      });
      goto('/assets');
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '删除资产失败'
      });
    }
  }

  onMount(() => {
    loadAsset();
  });
</script>

<div class="container mx-auto p-4 space-y-4">
  {#if loading}
    <div class="text-center py-8">加载中...</div>
  {:else if asset}
    <div class="flex justify-between items-start mb-4">
      <div>
        <h1 class="text-2xl font-bold">{asset.name}</h1>
        <p class="text-muted-foreground">{asset.description || '暂无描述'}</p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" on:click={() => goto(`/assets/${asset.id}/edit`)}>
          编辑
        </Button>
        <Button variant="outline" on:click={handleScan}>
          扫描
        </Button>
        <Button variant="destructive" on:click={handleDelete}>
          删除
        </Button>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- 基本信息 -->
      <Card>
        <CardHeader>
          <CardTitle>基本信息</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-muted-foreground">类型</div>
              <div>{asset.type}</div>
            </div>
            <div>
              <div class="text-sm text-muted-foreground">状态</div>
              <div>
                <Badge variant={asset.status === 'online' ? 'default' : 'destructive'}>
                  {asset.status === 'online' ? '在线' : '离线'}
                </Badge>
              </div>
            </div>
            <div>
              <div class="text-sm text-muted-foreground">URL</div>
              <div>{asset.url || '-'}</div>
            </div>
            <div>
              <div class="text-sm text-muted-foreground">IP</div>
              <div>{asset.ip || '-'}</div>
            </div>
            <div>
              <div class="text-sm text-muted-foreground">风险等级</div>
              <div>
                <Badge
                  variant={
                    asset.riskLevel === 'high'
                      ? 'destructive'
                      : asset.riskLevel === 'medium'
                      ? 'secondary'
                      : 'default'
                  }
                >
                  {asset.riskLevel === 'high'
                    ? '高'
                    : asset.riskLevel === 'medium'
                    ? '中'
                    : '低'}
                </Badge>
              </div>
            </div>
            <div>
              <div class="text-sm text-muted-foreground">最后扫描</div>
              <div>{new Date(asset.lastScan).toLocaleString()}</div>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- 标签 -->
      <Card>
        <CardHeader>
          <CardTitle>标签</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="flex flex-wrap gap-2">
            {#if asset.tags && asset.tags.length > 0}
              {#each asset.tags as tag}
                <Badge variant="outline">{tag}</Badge>
              {/each}
            {:else}
              <div class="text-muted-foreground">暂无标签</div>
            {/if}
          </div>
        </CardContent>
      </Card>

      <!-- 时间信息 -->
      <Card>
        <CardHeader>
          <CardTitle>时间信息</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-muted-foreground">创建时间</div>
              <div>{new Date(asset.createdAt).toLocaleString()}</div>
            </div>
            <div>
              <div class="text-sm text-muted-foreground">更新时间</div>
              <div>{new Date(asset.updatedAt).toLocaleString()}</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  {/if}
</div> 