<script lang="ts">
	let { open = $bindable(false), children } = $props();
</script>

<div>
	{#if children}
		{@render children()}
	{/if}
</div>
