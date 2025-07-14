<!--
  @component
  Color picker component with presets and custom color input
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/ui/popover';
	import { cn } from '$lib/utils';

	interface Props {
		value?: string;
		label?: string;
		presets?: string[];
	}

	let {
		value = $bindable('#000000'),
		label,
		presets = [
			'#0284c7', // primary
			'#64748b', // secondary
			'#f59e0b', // accent
			'#ffffff', // background
			'#020817', // foreground
			'#f1f5f9', // muted
			'#64748b', // mutedForeground
			'#e2e8f0', // border
			'#ef4444', // destructive
			'#22c55e', // success
			'#f59e0b', // warning
			'#0ea5e9' // info
		]
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		change: string;
	}>();

	let isOpen = $state(false);
	let inputValue = $state(value);

	function handlePresetClick(preset: string) {
		value = preset;
		inputValue = preset;
		isOpen = false;
	}

	function handleChange(event: Event) {
		const target = event.target as HTMLInputElement;
		dispatch('change', target.value);
	}

	// 同步外部值到内部状态
	$effect(() => {
		if (value !== inputValue) {
			inputValue = value;
		}
	});
</script>

<div class="grid gap-2">
	{#if label}
		<Label>{label}</Label>
	{/if}
	<Popover bind:open={isOpen}>
		<PopoverTrigger>
			{#snippet children()}
				<Button
					variant="outline"
					class={cn(
						'w-[220px] justify-start text-left font-normal',
						!value && 'text-muted-foreground'
					)}
				>
					<div class="mr-2 h-4 w-4 rounded" style:background-color={value}></div>
					<span>{value}</span>
				</Button>
			{/snippet}
		</PopoverTrigger>
		<PopoverContent>
			{#snippet children()}
				<div class="grid gap-2">
					<div class="flex items-center justify-center">
						<div class="h-9 w-9 rounded" style:background-color={value}></div>
					</div>
					<Input id="color" type="color" {value} oninput={handleChange} />
				</div>
			{/snippet}
		</PopoverContent>
	</Popover>
</div>
