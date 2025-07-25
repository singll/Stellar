<!-- 通用的删除确认对话框 -->
<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { DialogContent, DialogHeader, DialogTitle, DialogFooter } from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import Icon from '$lib/components/ui/Icon.svelte';

	interface Props {
		open: boolean;
		title?: string;
		description?: string;
		itemName?: string;
		itemType?: string;
		loading?: boolean;
		onConfirm: () => void;
		onCancel: () => void;
	}

	let {
		open = $bindable(),
		title,
		description,
		itemName,
		itemType = '项目',
		loading = false,
		onConfirm,
		onCancel
	}: Props = $props();

	const handleConfirm = () => {
		onConfirm();
	};

	const handleCancel = () => {
		open = false;
		onCancel();
	};

	const defaultTitle = $derived(title || `删除${itemType}`);
	const defaultDescription = $derived(
		description || 
		(itemName ? `确定要删除${itemType}「${itemName}」吗？此操作不可逆，删除后所有相关数据将永久丢失。` : `确定要删除这个${itemType}吗？此操作不可逆。`)
	);
</script>

<DialogPrimitive.Root bind:open>
	<DialogContent class="sm:max-w-md">
		<DialogHeader>
			<div class="flex items-center gap-3">
				<div class="flex items-center justify-center w-12 h-12 rounded-full bg-red-100">
					<Icon name="alert-triangle" class="h-6 w-6 text-red-600" />
				</div>
				<div>
					<DialogTitle class="text-left">{defaultTitle}</DialogTitle>
				</div>
			</div>
		</DialogHeader>

		<div class="py-4">
			<p class="text-sm text-gray-600 leading-relaxed">
				{defaultDescription}
			</p>
		</div>

		<DialogFooter>
			<div class="flex justify-end gap-3 w-full">
				<Button
					variant="outline"
					onclick={handleCancel}
					disabled={loading}
				>
					取消
				</Button>
				<Button
					variant="destructive"
					onclick={handleConfirm}
					disabled={loading}
					class="min-w-[80px]"
				>
					{#if loading}
						<div class="flex items-center gap-2">
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
							删除中...
						</div>
					{:else}
						<div class="flex items-center gap-2">
							<Icon name="trash" class="h-4 w-4" />
							确认删除
						</div>
					{/if}
				</Button>
			</div>
		</DialogFooter>
	</DialogContent>
</DialogPrimitive.Root>