<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Theme } from '$lib/types/theme';
	import { createDefaultTheme } from '$lib/utils/theme';
	import ThemeForm from './ThemeForm.svelte';
	import ThemePreview from './ThemePreview.svelte';

	let {
		theme = undefined,
		baseTheme = undefined
	}: {
		theme?: Theme | undefined;
		baseTheme?: Theme | undefined;
	} = $props();

	const dispatch = createEventDispatcher<{
		save: Theme;
		cancel: void;
	}>();

	function handleSave(event: CustomEvent<Theme>) {
		dispatch('save', event.detail);
	}

	function handleCancel() {
		dispatch('cancel');
	}

	// 创建默认主题用于预览
	let defaultTheme = $derived(createDefaultTheme('预览主题', '用于预览的默认主题'));
</script>

<div class="grid grid-cols-2 gap-8">
	<!-- 左侧：主题表单 -->
	<div class="space-y-6">
		<h2 class="text-2xl font-bold">
			{theme ? '编辑主题' : '创建主题'}
		</h2>

		<ThemeForm {theme} on:save={handleSave} on:cancel={handleCancel} />
	</div>

	<!-- 右侧：主题预览 -->
	<div class="space-y-6">
		<h2 class="text-2xl font-bold">预览</h2>

		<div class="sticky top-6">
			<ThemePreview theme={theme || baseTheme || defaultTheme} />
		</div>
	</div>
</div>
