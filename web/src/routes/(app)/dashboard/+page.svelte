<script lang="ts">
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
</script>

<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
	<Card>
		<CardHeader>
			<CardTitle>资产</CardTitle>
			<CardDescription>已发现的资产总数</CardDescription>
		</CardHeader>
		<CardContent>
			<p class="text-2xl font-bold">0</p>
		</CardContent>
	</Card>

	<Card>
		<CardHeader>
			<CardTitle>任务</CardTitle>
			<CardDescription>活跃的扫描任务</CardDescription>
		</CardHeader>
		<CardContent>
			<p class="text-2xl font-bold">0</p>
		</CardContent>
	</Card>

	<Card>
		<CardHeader>
			<CardTitle>节点</CardTitle>
			<CardDescription>已连接的扫描节点</CardDescription>
		</CardHeader>
		<CardContent>
			<p class="text-2xl font-bold">0</p>
		</CardContent>
	</Card>
</div>
