<!-- (auth)/+layout.svelte -->
<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="min-h-screen bg-background flex items-center justify-center">
	<main class="w-full max-w-md mx-auto p-6">
		{@render children()}
	</main>
</div>
