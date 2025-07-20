<script lang="ts" module>
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';
	import { type VariantProps, tv } from 'tailwind-variants';

	export const separatorVariants = tv({
		base: 'shrink-0 bg-border',
		variants: {
			orientation: {
				horizontal: 'h-px w-full',
				vertical: 'h-full w-px'
			}
		},
		defaultVariants: {
			orientation: 'horizontal'
		}
	});

	export type SeparatorOrientation = VariantProps<typeof separatorVariants>['orientation'];

	export type SeparatorProps = WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		orientation?: SeparatorOrientation;
		decorative?: boolean;
	};
</script>

<script lang="ts">
	let {
		class: className,
		orientation = 'horizontal',
		decorative = true,
		ref = $bindable(null),
		...restProps
	}: SeparatorProps = $props();
</script>

<div
	bind:this={ref}
	role={decorative ? 'none' : 'separator'}
	aria-orientation={orientation}
	class={cn(separatorVariants({ orientation }), className)}
	{...restProps}
></div> 