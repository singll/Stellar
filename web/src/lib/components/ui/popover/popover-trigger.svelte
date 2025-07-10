<script lang="ts">
	let { children } = $props();
</script>

<div class="relative inline-block">
	{#if children}
		{@render children()}
	{/if}
</div>
