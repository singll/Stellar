<!-- 资产详情组件 -->
<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import type { Asset } from '$lib/types/asset';

	let { asset }: { asset: Asset } = $props();

	// 获取资产类型的中文名称
	function getAssetTypeName(type: string): string {
		const typeMap: Record<string, string> = {
			domain: '域名',
			subdomain: '子域名',
			ip: 'IP地址',
			port: '端口',
			url: 'URL',
			http: 'HTTP服务',
			app: '应用程序',
			miniapp: '小程序',
			other: '其他'
		};
		return typeMap[type] || type;
	}

	// 渲染资产特定字段
	function renderAssetSpecificFields(
		asset: Asset
	): Array<{ label: string; value: any; type?: string }> {
		const fields: Array<{ label: string; value: any; type?: string }> = [];

		switch (asset.type) {
			case 'domain':
				if ('domain' in asset) {
					fields.push({ label: '域名', value: asset.domain });
					fields.push({ label: 'IP地址', value: asset.ips?.join(', ') || '-' });
					fields.push({ label: 'Whois信息', value: asset.whois || '-' });
					if (asset.icpInfo) {
						fields.push({ label: 'ICP备案号', value: asset.icpInfo.icpNo || '-' });
						fields.push({ label: '备案公司', value: asset.icpInfo.companyName || '-' });
						fields.push({ label: '公司类型', value: asset.icpInfo.companyType || '-' });
					}
				}
				break;

			case 'subdomain':
				if ('host' in asset) {
					fields.push({ label: '主机名', value: asset.host });
					fields.push({ label: 'IP地址', value: asset.ips?.join(', ') || '-' });
					fields.push({ label: 'CNAME', value: asset.cname || '-' });
					fields.push({ label: 'DNS类型', value: asset.dnsType || '-' });
					fields.push({ label: 'DNS值', value: asset.value?.join(', ') || '-' });
					fields.push({ label: '可接管', value: asset.takeOver ? '是' : '否', type: 'boolean' });
				}
				break;

			case 'ip':
				if ('ip' in asset) {
					fields.push({ label: 'IP地址', value: asset.ip });
					fields.push({ label: 'ASN', value: asset.asn || '-' });
					fields.push({ label: 'ISP', value: asset.isp || '-' });
					if (asset.location) {
						fields.push({ label: '国家', value: asset.location.country || '-' });
						fields.push({ label: '地区', value: asset.location.region || '-' });
						fields.push({ label: '城市', value: asset.location.city || '-' });
					}
				}
				break;

			case 'port':
				if ('port' in asset) {
					fields.push({ label: 'IP地址', value: asset.ip });
					fields.push({ label: '端口', value: asset.port });
					fields.push({ label: '服务', value: asset.service || '-' });
					fields.push({ label: '协议', value: asset.protocol || '-' });
					fields.push({ label: '版本', value: asset.version || '-' });
					fields.push({ label: 'Banner', value: asset.banner || '-' });
					fields.push({ label: 'TLS', value: asset.tls ? '是' : '否', type: 'boolean' });
					fields.push({ label: '状态', value: asset.status || '-' });
				}
				break;

			case 'url':
				if ('url' in asset) {
					fields.push({ label: 'URL', value: asset.url });
					fields.push({ label: '主机名', value: asset.host });
					fields.push({ label: '路径', value: asset.path || '-' });
					fields.push({ label: '状态码', value: asset.statusCode || '-' });
					fields.push({ label: '标题', value: asset.title || '-' });
					fields.push({ label: '内容类型', value: asset.contentType || '-' });
					fields.push({ label: '内容长度', value: asset.contentLength || '-' });
					fields.push({ label: '技术栈', value: asset.technologies?.join(', ') || '-' });
				}
				break;

			case 'http':
				if ('url' in asset) {
					fields.push({ label: 'URL', value: asset.url });
					fields.push({ label: '主机名', value: asset.host });
					fields.push({ label: 'IP地址', value: asset.ip });
					fields.push({ label: '端口', value: asset.port });
					fields.push({ label: '状态码', value: asset.statusCode || '-' });
					fields.push({ label: '标题', value: asset.title || '-' });
					fields.push({ label: 'Web服务器', value: asset.webServer || '-' });
					fields.push({ label: 'TLS', value: asset.tls ? '是' : '否', type: 'boolean' });
					fields.push({ label: 'CDN', value: asset.cdn ? '是' : '否', type: 'boolean' });
					fields.push({ label: 'CDN名称', value: asset.cdnName || '-' });
				}
				break;

			case 'app':
				if ('appName' in asset) {
					fields.push({ label: '应用名称', value: asset.appName });
					fields.push({ label: '包名', value: asset.packageName });
					fields.push({ label: '平台', value: asset.platform });
					fields.push({ label: '版本', value: asset.version || '-' });
					fields.push({ label: '开发者', value: asset.developer || '-' });
					fields.push({ label: '描述', value: asset.description || '-' });
					fields.push({ label: '权限', value: asset.permissions?.join(', ') || '-' });
				}
				break;

			case 'miniapp':
				if ('appName' in asset) {
					fields.push({ label: '小程序名称', value: asset.appName });
					fields.push({ label: '小程序ID', value: asset.appId });
					fields.push({ label: '平台', value: asset.platform });
					fields.push({ label: '开发者', value: asset.developer || '-' });
					fields.push({ label: '描述', value: asset.description || '-' });
				}
				break;
		}

		return fields;
	}

	// 从响应式语句转换为函数
	let assetFields = $derived(renderAssetSpecificFields(asset));
</script>

<div class="space-y-6">
	<!-- 基本信息 -->
	<Card>
		<CardHeader>
			<CardTitle>基本信息</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<div class="text-sm text-muted-foreground">资产类型</div>
					<div class="flex items-center gap-2">
						<Badge variant="outline">{getAssetTypeName(asset.type)}</Badge>
					</div>
				</div>
				<div>
					<div class="text-sm text-muted-foreground">项目ID</div>
					<div class="font-mono text-sm">{asset.projectId}</div>
				</div>
				<div>
					<div class="text-sm text-muted-foreground">根域名</div>
					<div>{asset.rootDomain || '-'}</div>
				</div>
				<div>
					<div class="text-sm text-muted-foreground">任务名称</div>
					<div>{asset.taskName || '-'}</div>
				</div>
				<div>
					<div class="text-sm text-muted-foreground">创建时间</div>
					<div>{new Date(asset.createdAt).toLocaleString()}</div>
				</div>
				<div>
					<div class="text-sm text-muted-foreground">更新时间</div>
					<div>{new Date(asset.updatedAt).toLocaleString()}</div>
				</div>
				<div>
					<div class="text-sm text-muted-foreground">最后扫描时间</div>
					<div>{new Date(asset.lastScanTime).toLocaleString()}</div>
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

	<!-- 资产特定字段 -->
	{#if assetFields.length > 0}
		<Card>
			<CardHeader>
				<CardTitle>详细信息</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each assetFields as field}
						<div>
							<div class="text-sm text-muted-foreground">{field.label}</div>
							<div class="break-all">
								{#if field.type === 'boolean'}
									<Badge variant={field.value === '是' ? 'default' : 'secondary'}>
										{field.value}
									</Badge>
								{:else if field.value && typeof field.value === 'string' && field.value.length > 100}
									<div class="max-h-32 overflow-y-auto bg-muted p-2 rounded text-sm font-mono">
										{field.value}
									</div>
								{:else}
									<div class={field.value === '-' ? 'text-muted-foreground' : ''}>
										{field.value}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	{/if}

	<!-- 变更历史 -->
	{#if asset.changeHistory && asset.changeHistory.length > 0}
		<Card>
			<CardHeader>
				<CardTitle>变更历史</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#each asset.changeHistory as change}
						<div class="border-l-2 border-muted pl-4">
							<div class="flex items-center gap-2 mb-2">
								<Badge
									variant={change.changeType === 'add'
										? 'default'
										: change.changeType === 'update'
											? 'secondary'
											: 'destructive'}
								>
									{change.changeType === 'add'
										? '新增'
										: change.changeType === 'update'
											? '更新'
											: '删除'}
								</Badge>
								<span class="text-sm text-muted-foreground"
									>{new Date(change.time).toLocaleString()}</span
								>
							</div>
							<div class="text-sm">
								<div class="font-medium">{change.fieldName}</div>
								{#if change.changeType === 'update'}
									<div class="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2">
										<div>
											<div class="text-muted-foreground">旧值:</div>
											<div class="bg-red-50 p-2 rounded text-sm">
												{JSON.stringify(change.oldValue)}
											</div>
										</div>
										<div>
											<div class="text-muted-foreground">新值:</div>
											<div class="bg-green-50 p-2 rounded text-sm">
												{JSON.stringify(change.newValue)}
											</div>
										</div>
									</div>
								{:else}
									<div class="bg-muted p-2 rounded text-sm mt-2">
										{JSON.stringify(change.newValue)}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	{/if}
</div>
"
