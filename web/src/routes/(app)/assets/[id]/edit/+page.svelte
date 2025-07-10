<!-- 资产编辑页面 -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { assetApi } from '$lib/api/asset';
	import type { Asset, AssetType } from '$lib/types/asset';
	import { notifications } from '$lib/stores/notifications';
	import { goto } from '$app/navigation';

	let asset: Asset | null = $state(null);
	let loading = $state(true);
	let saving = $state(false);

	let name = $state('');
	let type: AssetType = $state('domain');
	let url = $state('');
	let ip = $state('');
	let description = $state('');
	let tags = $state('');

	async function loadAsset() {
		try {
			loading = true;
			const response = await assetApi.getAssetById($page.params.id, 'domain');
			asset = response.data;

			// 填充表单
			type = asset.type;
			tags = asset.tags?.join(', ') || '';

			// 根据资产类型填充对应字段
			switch (asset.type) {
				case 'domain':
					name = (asset as any).domain || '';
					ip = (asset as any).ips?.[0] || '';
					break;
				case 'ip':
					name = (asset as any).ip || '';
					ip = (asset as any).ip || '';
					break;
				case 'url':
					name = (asset as any).url || '';
					url = (asset as any).url || '';
					break;
				default:
					name = (asset as any).appName || (asset as any).host || '';
					url = (asset as any).url || '';
					ip = (asset as any).ip || '';
			}
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
			await assetApi.updateAsset(asset.id, {
				type,
				data: {
					...(type === 'domain' && { domain: name }),
					...(type === 'ip' && { ip: name }),
					...(type === 'url' && { url: name }),
					...(url && { url }),
					...(ip && { ip }),
					...(description && { description }),
					...(tags && {
						tags: tags
							.split(',')
							.map((t) => t.trim())
							.filter(Boolean)
					})
				}
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
				<form
					onsubmit={(e) => {
						e.preventDefault();
						handleSubmit();
					}}
					class="space-y-4"
				>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<!-- 名称 -->
						<div class="space-y-2">
							<Label for="name">名称</Label>
							<Input id="name" bind:value={name} required placeholder="输入资产名称" />
						</div>

						<!-- 类型 -->
						<div class="space-y-2">
							<Label for="type">类型</Label>
							<select
								id="type"
								bind:value={type}
								class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
							>
								<option value="domain">域名</option>
								<option value="subdomain">子域名</option>
								<option value="ip">IP</option>
								<option value="port">端口</option>
								<option value="url">URL</option>
								<option value="http">HTTP服务</option>
								<option value="app">应用</option>
								<option value="miniapp">小程序</option>
							</select>
						</div>

						<!-- URL -->
						<div class="space-y-2">
							<Label for="url">URL</Label>
							<Input id="url" bind:value={url} placeholder="输入资产URL" />
						</div>

						<!-- IP -->
						<div class="space-y-2">
							<Label for="ip">IP</Label>
							<Input id="ip" bind:value={ip} placeholder="输入资产IP" />
						</div>

						<!-- 标签 -->
						<div class="space-y-2">
							<Label for="tags">标签</Label>
							<Input id="tags" bind:value={tags} placeholder="输入标签，用逗号分隔" />
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
						<Button type="button" variant="outline" on:click={() => goto(`/assets/${asset?.id}`)}>
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
