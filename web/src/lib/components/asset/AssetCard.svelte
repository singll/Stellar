<script lang="ts">
  import { Badge } from "$lib/components/ui/badge";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "$lib/components/ui/card";
  import type { Asset } from "$lib/types/asset";

  export let asset: Asset;

  const getStatusVariant = (status: Asset['status']) => {
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

  const getRiskVariant = (risk: Asset['riskLevel']) => {
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
</script>

<Card class="hover:bg-muted/50 transition-colors">
  <CardHeader>
    <CardTitle class="flex items-center justify-between">
      <span>{asset.name}</span>
      <Badge variant={getStatusVariant(asset.status)}>
        {asset.status}
      </Badge>
    </CardTitle>
    <CardDescription>{asset.description}</CardDescription>
  </CardHeader>
  <CardContent>
    <div class="grid gap-2">
      <div class="flex items-center justify-between text-sm">
        <span class="text-muted-foreground">IP地址</span>
        <span>{asset.ip}</span>
      </div>
      <div class="flex items-center justify-between text-sm">
        <span class="text-muted-foreground">域名</span>
        <span>{asset.domain || '无'}</span>
      </div>
      <div class="flex items-center justify-between text-sm">
        <span class="text-muted-foreground">最后扫描</span>
        <span>{new Date(asset.lastScan).toLocaleString()}</span>
      </div>
      <div class="flex items-center justify-between text-sm">
        <span class="text-muted-foreground">风险等级</span>
        <Badge variant={getRiskVariant(asset.riskLevel)}>
          {asset.riskLevel}
        </Badge>
      </div>
    </div>
  </CardContent>
</Card> 