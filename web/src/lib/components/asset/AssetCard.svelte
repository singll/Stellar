<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import type {
		Asset,
		DomainAsset,
		SubdomainAsset,
		IPAsset,
		PortAsset,
		URLAsset,
		HTTPAsset,
		AppAsset,
		MiniAppAsset
	} from '$lib/types/asset';

	let { asset }: { asset: Asset } = $props();

	const getStatusVariant = (status?: Asset['status']) => {
		switch (status) {
			case 'active':
				return 'default';
			case 'inactive':
				return 'secondary';
			case 'deleted':
				return 'destructive';
			default:
				return 'default';
		}
	};

	const getRiskVariant = (risk?: Asset['riskLevel']) => {
		switch (risk) {
			case 'high':
				return 'destructive';
			case 'medium':
				return 'secondary';
			case 'low':
				return 'default';
			default:
				return 'default';
		}
	};

	// 获取资产名称
	const getAssetName = (asset: Asset): string => {
		if (asset.name) return asset.name;

		// 使用类型断言确保所有情况都被处理
		const assetAny = asset as any;

		switch (asset.type) {
			case 'domain':
				return (asset as DomainAsset).domain;
			case 'subdomain':
				return (asset as SubdomainAsset).host;
			case 'ip':
				return (asset as IPAsset).ip;
			case 'port':
				return `${(asset as PortAsset).ip}:${(asset as PortAsset).port}`;
			case 'url':
				return (asset as URLAsset).url;
			case 'http':
				return (asset as HTTPAsset).url;
			case 'app':
				return (asset as AppAsset).appName;
			case 'miniapp':
				return (asset as MiniAppAsset).appName;
			default:
				// 对于其他类型，尝试使用value或id
				return assetAny.value || assetAny.id || 'Unknown';
		}
	};

	// 获取资产IP地址
	const getAssetIP = (asset: Asset): string => {
		switch (asset.type) {
			case 'domain':
				return (asset as DomainAsset).ips?.join(', ') || '无';
			case 'subdomain':
				return (asset as SubdomainAsset).ips?.join(', ') || '无';
			case 'ip':
				return (asset as IPAsset).ip;
			case 'port':
				return (asset as PortAsset).ip;
			case 'http':
				return (asset as HTTPAsset).ip;
			default:
				return '无';
		}
	};

	// 获取资产域名
	const getAssetDomain = (asset: Asset): string => {
		switch (asset.type) {
			case 'domain':
				return (asset as DomainAsset).domain;
			case 'subdomain':
				return (asset as SubdomainAsset).host;
			case 'url':
				return (asset as URLAsset).host;
			case 'http':
				return (asset as HTTPAsset).host;
			case 'port':
				return (asset as PortAsset).host || '无';
			default:
				return '无';
		}
	};
</script>

<Card class="hover:bg-muted/50 transition-colors">
	<CardHeader>
		<CardTitle class="flex items-center justify-between">
			<span>{getAssetName(asset)}</span>
			<Badge variant={getStatusVariant(asset.status)}>
				{asset.status || 'unknown'}
			</Badge>
		</CardTitle>
		<CardDescription>{asset.description || '无描述'}</CardDescription>
	</CardHeader>
	<CardContent>
		<div class="grid gap-2">
			<div class="flex items-center justify-between text-sm">
				<span class="text-muted-foreground">IP地址</span>
				<span>{getAssetIP(asset)}</span>
			</div>
			<div class="flex items-center justify-between text-sm">
				<span class="text-muted-foreground">域名</span>
				<span>{getAssetDomain(asset)}</span>
			</div>
			<div class="flex items-center justify-between text-sm">
				<span class="text-muted-foreground">最后扫描</span>
				<span
					>{asset.lastScan
						? new Date(asset.lastScan).toLocaleString()
						: new Date(asset.lastScanTime).toLocaleString()}</span
				>
			</div>
			<div class="flex items-center justify-between text-sm">
				<span class="text-muted-foreground">风险等级</span>
				<Badge variant={getRiskVariant(asset.riskLevel)}>
					{asset.riskLevel || 'unknown'}
				</Badge>
			</div>
		</div>
	</CardContent>
</Card>
