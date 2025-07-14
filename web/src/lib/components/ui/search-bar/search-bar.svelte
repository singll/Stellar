<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';

	interface Props {
		class?: string;
		value?: string;
		placeholder?: string;
		debounce?: number;
		[key: string]: any;
	}

	let {
		class: className,
		value = $bindable(''),
		placeholder = '搜索...',
		debounce = 300,
		...restProps
	}: Props = $props();

	let timeoutId: ReturnType<typeof setTimeout>;
	const dispatch = createEventDispatcher<{
		search: { value: string };
	}>();

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		value = target.value;

		clearTimeout(timeoutId);
		timeoutId = setTimeout(() => {
			dispatch('search', { value });
		}, debounce);
	}
</script>

<div class={cn('relative', className)}>
	<Icon
		name="search"
		class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
	/>
	<Input type="search" {placeholder} class="pl-10" {value} oninput={handleInput} {...restProps} />
</div>
