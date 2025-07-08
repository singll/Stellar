<!-- 资产编辑页面 -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '$lib/components/ui/select';
  import { Textarea } from '$lib/components/ui/textarea';
  import { assetApi } from '$lib/api/asset';
  import type { Asset, AssetType } from '$lib/types/asset';
  import { notifications } from '$lib/stores/notifications';
  import { goto } from '$app/navigation';

  let asset: Asset | null = null;
  let loading = true;
  let saving = false;

  let name = '';
  let type: AssetType = 'domain';
  let url = '';
  let ip = '';
  let description = '';
  let tags = '';

  async function loadAsset() {
    try {
      loading = true;
      const response = await assetApi.getAssetById($page.params.id);
      asset = response.data;
      
      // 填充表单
      name = asset.name;
      type = asset.type;
      url = asset.url || '';
      ip = asset.ip || '';
      description = asset.description || '';
      tags = asset.tags?.join(', ') || '';
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

  async function handleSubmit() {
    if (!asset) return;

    try {
      saving = true;
      await assetApi.updateAsset({
        id: asset.id,
        name,
        type,
        url: url || undefined,
        ip: ip || undefined,
        description: description || undefined,
        tags: tags ? tags.split(',').map(t => t.trim()).filter(Boolean) : undefined
      });

      notifications.add({
        type: 'success',
        message: '资产已更新'
      });
      goto(`/assets/${asset.id}`);
    } catch (error) {
      notifications.add({
        type: 'error',
        message: '更新资产失败'
      });
    } finally {
      saving = false;
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
    <Card>
      <CardHeader>
        <CardTitle>编辑资产</CardTitle>
        <CardDescription>修改资产信息</CardDescription>
      </CardHeader>
      <CardContent>
        <form on:submit|preventDefault={handleSubmit} class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- 名称 -->
            <div class="space-y-2">
              <Label for="name">名称</Label>
              <Input
                id="name"
                bind:value={name}
                required
                placeholder="输入资产名称"
              />
            </div>

            <!-- 类型 -->
            <div class="space-y-2">
              <Label for="type">类型</Label>
              <Select bind:value={type} required>
                <SelectTrigger>
                  <SelectValue placeholder="选择资产类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="domain">域名</SelectItem>
                  <SelectItem value="ip">IP</SelectItem>
                  <SelectItem value="web">网站</SelectItem>
                  <SelectItem value="app">应用</SelectItem>
                  <SelectItem value="other">其他</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <!-- URL -->
            <div class="space-y-2">
              <Label for="url">URL</Label>
              <Input
                id="url"
                bind:value={url}
                placeholder="输入资产URL"
              />
            </div>

            <!-- IP -->
            <div class="space-y-2">
              <Label for="ip">IP</Label>
              <Input
                id="ip"
                bind:value={ip}
                placeholder="输入资产IP"
              />
            </div>

            <!-- 标签 -->
            <div class="space-y-2">
              <Label for="tags">标签</Label>
              <Input
                id="tags"
                bind:value={tags}
                placeholder="输入标签，用逗号分隔"
              />
            </div>
          </div>

          <!-- 描述 -->
          <div class="space-y-2">
            <Label for="description">描述</Label>
            <Textarea
              id="description"
              bind:value={description}
              placeholder="输入资产描述"
              rows={4}
            />
          </div>

          <!-- 按钮 -->
          <div class="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              on:click={() => goto(`/assets/${asset.id}`)}
            >
              取消
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? '保存中...' : '保存'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  {/if}
</div> 