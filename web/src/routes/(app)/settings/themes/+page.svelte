<!--
  @component
  Theme settings page for managing themes
-->
<script lang="ts">
	import { themeStore, themeActions } from '$lib/stores/theme';
	import { Button } from '$lib/components/ui/button';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
		DialogTrigger
	} from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import ThemePreview from '$lib/components/theme/ThemePreview.svelte';
	import ThemeForm from '$lib/components/theme/ThemeForm.svelte';
	import type { Theme } from '$lib/types/theme';
	import { exportTheme, importTheme } from '$lib/utils/theme';

	let createDialogOpen = $state(false);
	let editDialogOpen = $state(false);
	let importDialogOpen = $state(false);
	let selectedTheme: Theme | undefined = $state();
	let importError: string | undefined = $state();

	function handleCreateSubmit(event: CustomEvent<Theme>) {
		const theme = event.detail;
		themeActions.addTheme(theme);
		createDialogOpen = false;
	}

	function handleEditSubmit(event: CustomEvent<Theme>) {
		const theme = event.detail;
		themeActions.updateTheme(theme);
		editDialogOpen = false;
		selectedTheme = undefined;
	}

	function handleDelete(theme: Theme) {
		if (confirm('Are you sure you want to delete this theme?')) {
			themeActions.deleteTheme(theme.id);
		}
	}

	function handleExport(theme: Theme) {
		const json = exportTheme(theme);
		const blob = new Blob([json], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${theme.name.toLowerCase().replace(/\s+/g, '-')}.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	async function handleImport(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		try {
			const text = await file.text();
			const themes = importTheme(text);
			// 添加所有导入的主题
			themes.forEach((theme) => themeActions.addTheme(theme));
			importDialogOpen = false;
			importError = undefined;
			input.value = '';
		} catch (error) {
			importError = error instanceof Error ? error.message : 'Failed to import theme';
		}
	}
</script>

<div class="container py-10">
	<div class="flex items-center justify-between">
		<div class="space-y-1">
			<h2 class="text-2xl font-semibold tracking-tight">Theme Settings</h2>
			<p class="text-sm text-muted-foreground">
				Manage your themes and customize the appearance of your application.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<Dialog bind:open={importDialogOpen}>
				<DialogTrigger>
					<Button variant="outline">Import Theme</Button>
				</DialogTrigger>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Import Theme</DialogTitle>
						<DialogDescription>Import a theme from a JSON file.</DialogDescription>
					</DialogHeader>
					<div class="grid gap-4 py-4">
						<div class="grid gap-2">
							<Label for="import-file">Theme File</Label>
							<Input id="import-file" type="file" accept=".json" onchange={handleImport} />
							{#if importError}
								<p class="text-sm text-destructive">{importError}</p>
							{/if}
						</div>
					</div>
				</DialogContent>
			</Dialog>

			<Dialog bind:open={createDialogOpen}>
				<DialogTrigger>
					<Button>Create Theme</Button>
				</DialogTrigger>
				<DialogContent class="max-w-4xl">
					<DialogHeader>
						<DialogTitle>Create Theme</DialogTitle>
						<DialogDescription>
							Create a new theme by customizing colors and other properties.
						</DialogDescription>
					</DialogHeader>
					<ThemeForm
						mode="create"
						on:submit={handleCreateSubmit}
						on:cancel={() => (createDialogOpen = false)}
					/>
				</DialogContent>
			</Dialog>
		</div>
	</div>

	<div class="mt-10 grid gap-6">
		{#each $themeStore.availableThemes as theme (theme.id)}
			<div class="flex items-start gap-6 rounded-lg border p-4">
				<div class="flex-1">
					<ThemePreview {theme} mode={$themeStore.mode} />
				</div>

				<div class="flex flex-col items-end gap-2">
					<Button
						variant="ghost"
						disabled={theme.id === 'default'}
						on:click={() => {
							selectedTheme = theme;
							editDialogOpen = true;
						}}
					>
						Edit
					</Button>
					<Button variant="ghost" on:click={() => handleExport(theme)}>Export</Button>
					<Button
						variant="ghost"
						class="text-destructive"
						disabled={theme.id === 'default'}
						on:click={() => handleDelete(theme)}
					>
						Delete
					</Button>
					<Button
						variant="outline"
						disabled={$themeStore.currentTheme.id === theme.id}
						on:click={() => themeActions.setTheme(theme)}
					>
						{$themeStore.currentTheme.id === theme.id ? 'Active' : 'Activate'}
					</Button>
				</div>
			</div>
		{/each}
	</div>

	<Dialog bind:open={editDialogOpen}>
		<DialogContent class="max-w-4xl">
			<DialogHeader>
				<DialogTitle>Edit Theme</DialogTitle>
				<DialogDescription>Edit theme colors and properties.</DialogDescription>
			</DialogHeader>
			{#if selectedTheme}
				<ThemeForm
					mode="edit"
					theme={selectedTheme}
					on:submit={handleEditSubmit}
					on:cancel={() => {
						editDialogOpen = false;
						selectedTheme = undefined;
					}}
				/>
			{/if}
		</DialogContent>
	</Dialog>
</div>
