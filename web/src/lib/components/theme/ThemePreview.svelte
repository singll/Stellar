<!--
  @component
  Theme preview component that shows how the theme looks
-->
<script lang="ts">
	import type { Theme, ColorMode } from '$lib/types/theme';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		theme: Theme;
		mode?: ColorMode;
	}

	let { theme, mode = 'light' }: Props = $props();

	let colors = $derived(mode === 'light' ? theme.colors.light : theme.colors.dark);
	let colorList = $derived([
		{ label: 'Primary', color: colors.primary },
		{ label: 'Secondary', color: colors.secondary },
		{ label: 'Accent', color: colors.accent },
		{ label: 'Background', color: colors.background },
		{ label: 'Foreground', color: colors.foreground },
		{ label: 'Border', color: colors.border },
		{ label: 'Card', color: colors.card },
		{ label: 'Success', color: colors.success },
		{ label: 'Warning', color: colors.warning },
		{ label: 'Error', color: colors.error },
		{ label: 'Info', color: colors.info }
	]);
</script>

<Card>
	<CardContent class="p-6">
		<div class="grid gap-4">
			<div class="flex items-center justify-between">
				<div class="space-y-1">
					<h3 class="text-lg font-semibold">{theme.name}</h3>
					{#if theme.description}
						<p class="text-sm text-muted-foreground">{theme.description}</p>
					{/if}
				</div>
				<Badge variant="outline">{mode}</Badge>
			</div>

			<div class="grid grid-cols-2 gap-2">
				{#each colorList as { label, color }}
					<div class="flex items-center gap-2">
						<div class="h-5 w-5 rounded-full border" style:background-color={color}></div>
						<span class="text-sm">{label}</span>
						<span class="ml-auto text-sm text-muted-foreground">{color}</span>
					</div>
				{/each}
			</div>

			<div class="space-y-4">
				<div class="grid gap-2">
					<Button variant="default">Primary Button</Button>
					<Button variant="secondary">Secondary Button</Button>
					<Button variant="outline">Outline Button</Button>
					<Button variant="ghost">Ghost Button</Button>
					<Button variant="destructive">Destructive Button</Button>
				</div>

				<div class="grid gap-2">
					<Badge variant="default">Default Badge</Badge>
					<Badge variant="secondary">Secondary Badge</Badge>
					<Badge variant="outline">Outline Badge</Badge>
					<Badge variant="destructive">Destructive Badge</Badge>
				</div>
			</div>
		</div>
	</CardContent>
</Card>
