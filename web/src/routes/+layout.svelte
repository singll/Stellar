<script lang="ts">
	import "../app.css";
	import NotificationContainer from "$lib/components/notification/NotificationContainer.svelte";
	import type { Snippet } from "svelte";

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="min-h-screen bg-background font-sans antialiased">
	{@render children()}
</div>

<NotificationContainer />

<style>
	:global(html) {
		height: 100%;
	}
	:global(body) {
		height: 100%;
		margin: 0;
	}
</style>
