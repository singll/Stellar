<!--
  @component
  Theme form component for creating and editing themes
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { cn } from '$lib/utils';
	import type {
		Theme,
		CreateThemeRequest,
		UpdateThemeRequest,
		ThemeColors
	} from '$lib/types/theme';
	import { createDefaultTheme, validateTheme, createTheme, updateTheme } from '$lib/utils/theme';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import ColorPicker from './ColorPicker.svelte';
	import ThemePreview from './ThemePreview.svelte';

	interface Props {
		theme?: Theme;
		mode?: 'create' | 'edit';
	}

	let { theme, mode = 'create' }: Props = $props();

	const dispatch = createEventDispatcher<{
		submit: Theme;
		cancel: void;
	}>();

	let name = $state(theme?.name ?? '');
	let description = $state(theme?.description ?? '');
	let author = $state(theme?.metadata.author ?? '');

	let lightColors = $state<ThemeColors>(
		theme?.colors.light ?? {
			primary: '#0284c7',
			secondary: '#64748b',
			accent: '#f59e0b',
			background: '#ffffff',
			foreground: '#020817',
			muted: '#f1f5f9',
			mutedForeground: '#64748b',
			border: '#e2e8f0',
			input: '#e2e8f0',
			ring: '#0284c7',
			destructive: '#ef4444',
			destructiveForeground: '#ffffff',
			success: '#22c55e',
			successForeground: '#ffffff',
			warning: '#f59e0b',
			warningForeground: '#ffffff',
			info: '#0ea5e9',
			infoForeground: '#ffffff',
			error: '#ef4444',
			header: '#ffffff',
			card: '#ffffff',
			sidebar: '#f1f5f9'
		}
	);

	let darkColors = $state<ThemeColors>(
		theme?.colors.dark ?? {
			primary: '#0ea5e9',
			secondary: '#94a3b8',
			accent: '#f59e0b',
			background: '#020817',
			foreground: '#ffffff',
			muted: '#1e293b',
			mutedForeground: '#94a3b8',
			border: '#1e293b',
			input: '#1e293b',
			ring: '#0ea5e9',
			destructive: '#ef4444',
			destructiveForeground: '#ffffff',
			success: '#22c55e',
			successForeground: '#ffffff',
			warning: '#f59e0b',
			warningForeground: '#ffffff',
			info: '#0ea5e9',
			infoForeground: '#ffffff',
			error: '#ef4444',
			header: '#020817',
			card: '#020817',
			sidebar: '#1e293b'
		}
	);

	let errors = $state<{ [key: string]: string }>({});

	function validateForm(): boolean {
		errors = {};

		if (!name) {
			errors.name = 'Theme name is required';
		}

		if (mode === 'create') {
			const request: CreateThemeRequest = {
				name,
				description,
				colors: {
					light: lightColors,
					dark: darkColors
				},
				metadata: {
					author
				}
			};

			const newTheme = createTheme(request);
			const validationErrors = validateTheme(newTheme);

			for (const error of validationErrors) {
				errors[error.field] = error.message;
			}
		} else if (theme) {
			const request: UpdateThemeRequest = {
				name,
				description,
				colors: {
					light: lightColors,
					dark: darkColors
				}
			};

			const updatedTheme = updateTheme(theme, request);
			const validationErrors = validateTheme(updatedTheme);

			for (const error of validationErrors) {
				errors[error.field] = error.message;
			}
		}

		return Object.keys(errors).length === 0;
	}

	function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		if (mode === 'create') {
			const request: CreateThemeRequest = {
				name,
				description,
				colors: {
					light: lightColors,
					dark: darkColors
				},
				metadata: {
					author
				}
			};

			const newTheme = createTheme(request);
			dispatch('submit', newTheme);
		} else if (theme) {
			const request: UpdateThemeRequest = {
				name,
				description,
				colors: {
					light: lightColors,
					dark: darkColors
				}
			};

			const updatedTheme = updateTheme(theme, request);
			dispatch('submit', updatedTheme);
		}
	}

	function handleCancel() {
		dispatch('cancel');
	}
</script>

<form
	class="grid gap-6"
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
>
	<div class="grid gap-4">
		<div class="grid gap-2">
			<Label for="name">Name</Label>
			<Input id="name" bind:value={name} class={cn(errors.name && 'border-destructive')} />
			{#if errors.name}
				<p class="text-sm text-destructive">{errors.name}</p>
			{/if}
		</div>

		<div class="grid gap-2">
			<Label for="description">Description</Label>
			<Textarea
				id="description"
				bind:value={description}
				class={cn(errors.description && 'border-destructive')}
			/>
			{#if errors.description}
				<p class="text-sm text-destructive">{errors.description}</p>
			{/if}
		</div>

		<div class="grid gap-2">
			<Label for="author">Author</Label>
			<Input id="author" bind:value={author} class={cn(errors.author && 'border-destructive')} />
			{#if errors.author}
				<p class="text-sm text-destructive">{errors.author}</p>
			{/if}
		</div>
	</div>

	<div class="grid gap-6">
		<div class="space-y-4">
			<h3 class="text-lg font-semibold">Light Mode Colors</h3>
			<div class="grid gap-4">
				<ColorPicker label="Primary" bind:value={lightColors.primary} />
				<ColorPicker label="Secondary" bind:value={lightColors.secondary} />
				<ColorPicker label="Accent" bind:value={lightColors.accent} />
				<ColorPicker label="Background" bind:value={lightColors.background} />
				<ColorPicker label="Foreground" bind:value={lightColors.foreground} />
				<ColorPicker label="Muted" bind:value={lightColors.muted} />
				<ColorPicker label="Muted Foreground" bind:value={lightColors.mutedForeground} />
				<ColorPicker label="Border" bind:value={lightColors.border} />
				<ColorPicker label="Input" bind:value={lightColors.input} />
				<ColorPicker label="Ring" bind:value={lightColors.ring} />
				<ColorPicker label="Destructive" bind:value={lightColors.destructive} />
				<ColorPicker
					label="Destructive Foreground"
					bind:value={lightColors.destructiveForeground}
				/>
				<ColorPicker label="Success" bind:value={lightColors.success} />
				<ColorPicker label="Success Foreground" bind:value={lightColors.successForeground} />
				<ColorPicker label="Warning" bind:value={lightColors.warning} />
				<ColorPicker label="Warning Foreground" bind:value={lightColors.warningForeground} />
				<ColorPicker label="Info" bind:value={lightColors.info} />
				<ColorPicker label="Info Foreground" bind:value={lightColors.infoForeground} />
			</div>
		</div>

		<div class="space-y-4">
			<h3 class="text-lg font-semibold">Dark Mode Colors</h3>
			<div class="grid gap-4">
				<ColorPicker label="Primary" bind:value={darkColors.primary} />
				<ColorPicker label="Secondary" bind:value={darkColors.secondary} />
				<ColorPicker label="Accent" bind:value={darkColors.accent} />
				<ColorPicker label="Background" bind:value={darkColors.background} />
				<ColorPicker label="Foreground" bind:value={darkColors.foreground} />
				<ColorPicker label="Muted" bind:value={darkColors.muted} />
				<ColorPicker label="Muted Foreground" bind:value={darkColors.mutedForeground} />
				<ColorPicker label="Border" bind:value={darkColors.border} />
				<ColorPicker label="Input" bind:value={darkColors.input} />
				<ColorPicker label="Ring" bind:value={darkColors.ring} />
				<ColorPicker label="Destructive" bind:value={darkColors.destructive} />
				<ColorPicker label="Destructive Foreground" bind:value={darkColors.destructiveForeground} />
				<ColorPicker label="Success" bind:value={darkColors.success} />
				<ColorPicker label="Success Foreground" bind:value={darkColors.successForeground} />
				<ColorPicker label="Warning" bind:value={darkColors.warning} />
				<ColorPicker label="Warning Foreground" bind:value={darkColors.warningForeground} />
				<ColorPicker label="Info" bind:value={darkColors.info} />
				<ColorPicker label="Info Foreground" bind:value={darkColors.infoForeground} />
			</div>
		</div>
	</div>

	<div class="flex gap-4">
		<Button type="submit">
			{mode === 'create' ? 'Create Theme' : 'Update Theme'}
		</Button>
		<Button type="button" variant="outline" onclick={handleCancel}>Cancel</Button>
	</div>

	<div class="grid gap-4">
		<h3 class="text-lg font-semibold">Preview</h3>
		<div class="grid gap-4 sm:grid-cols-2">
			<ThemePreview
				theme={{
					id: theme?.id ?? 'preview',
					name,
					description,
					colors: {
						light: lightColors,
						dark: darkColors
					},
					metadata: {
						author,
						createdAt: theme?.metadata.createdAt ?? new Date().toISOString(),
						updatedAt: new Date().toISOString()
					}
				}}
				mode="light"
			/>
			<ThemePreview
				theme={{
					id: theme?.id ?? 'preview',
					name,
					description,
					colors: {
						light: lightColors,
						dark: darkColors
					},
					metadata: {
						author,
						createdAt: theme?.metadata.createdAt ?? new Date().toISOString(),
						updatedAt: new Date().toISOString()
					}
				}}
				mode="dark"
			/>
		</div>
	</div>
</form>
