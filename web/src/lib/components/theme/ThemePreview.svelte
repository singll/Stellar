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

	export let theme: Theme;
	export let mode: ColorMode = 'light';

	$: colors = mode === 'light' ? theme.colors.light : theme.colors.dark;
	$: colorList = [
		{ label: 'Primary', color: colors.primary.value },
		{ label: 'Secondary', color: colors.secondary.value },
		{ label: 'Accent', color: colors.accent.value },
		{ label: 'Background', color: colors.background.value },
		{ label: 'Foreground', color: colors.foreground.value },
		{ label: 'Border', color: colors.border.value },
		{ label: 'Card', color: colors.card.value },
		{ label: 'Success', color: colors.success.value },
		{ label: 'Warning', color: colors.warning.value },
		{ label: 'Error', color: colors.error.value },
		{ label: 'Info', color: colors.info.value }
	];
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
						<div class="h-5 w-5 rounded-full border" style:background-color={color} />
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
