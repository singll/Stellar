<script lang="ts">
	let {
		tags = $bindable([]),
		placeholder = '添加标签...',
		disabled = false,
		class: className = ''
	} = $props();
</script>

<div class={className}>
	<input type="text" {placeholder} {disabled} class="w-full p-2 border rounded-md" />
</div>
