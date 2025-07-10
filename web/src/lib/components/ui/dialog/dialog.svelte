<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	export type DialogProps = {
		open?: boolean;
		title?: string;
		description?: string;
		class?: string;
		children?: Snippet;
		onOpenChange?: (open: boolean) => void;
	};

	let {
		open = $bindable(false),
		title,
		description,
		class: className,
		children,
		onOpenChange
	}: DialogProps = $props();

	// 处理点击背景关闭
	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			open = false;
			onOpenChange?.(false);
		}
	}

	// 处理ESC键关闭
	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			onOpenChange?.(false);
		}
	}

	// 关闭对话框
	function closeDialog() {
		open = false;
		onOpenChange?.(false);
	}
</script>

<!-- 背景遮罩 -->
{#if open}
	<div
		class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
		aria-labelledby={title ? 'dialog-title' : undefined}
		aria-describedby={description ? 'dialog-description' : undefined}
	>
		<!-- 对话框内容 -->
		<div
			class={cn(
				'fixed left-1/2 top-1/2 z-50 grid w-full max-w-lg -translate-x-1/2 -translate-y-1/2 gap-4 border bg-white p-6 shadow-lg duration-200 sm:rounded-lg',
				className
			)}
			onclick={(e) => e.stopPropagation()}
		>
			<!-- 标题和关闭按钮 -->
			{#if title}
				<div class="flex items-center justify-between">
					<h2 id="dialog-title" class="text-lg font-semibold leading-none tracking-tight">
						{title}
					</h2>
					<button
						onclick={closeDialog}
						class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
						aria-label="关闭"
					>
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>
			{/if}

			<!-- 描述 -->
			{#if description}
				<p id="dialog-description" class="text-sm text-gray-500">
					{description}
				</p>
			{/if}

			<!-- 内容 -->
			{#if children}
				{@render children()}
			{/if}
		</div>
	</div>
{/if}
