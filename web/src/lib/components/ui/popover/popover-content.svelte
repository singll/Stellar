<script lang="ts">
	let { children } = $props();
</script>

<div class="absolute z-50 w-72 rounded-md border bg-white p-4 shadow-md">
	{#if children}
		{@render children()}
	{/if}
</div>
