<!--
  @component
  Theme toggle component for switching between light and dark mode
-->
<script lang="ts">
	import Icon from '$lib/components/ui/Icon.svelte';
	import { Button } from '$lib/components/ui/button';
	import { themeStore, themeMode } from '$lib/stores/theme';

	let isDark = $state(false);
	let ariaLabel = $derived(isDark ? '切换到亮色模式' : '切换到暗色模式');

	// 订阅主题模式变化
	$effect(() => {
		const unsubscribe = themeMode.subscribe((mode) => {
			isDark = mode === 'dark';
		});
		return unsubscribe;
	});
</script>

<Button variant="ghost" size="icon" onclick={() => themeStore.toggleMode()} aria-label={ariaLabel}>
	{#if isDark}
		<Icon name="sun" class="h-5 w-5" />
	{:else}
		<Icon name="moon" class="h-5 w-5" />
	{/if}
</Button>
