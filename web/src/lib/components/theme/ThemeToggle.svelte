<!--
  @component
  Theme toggle component for switching between light and dark mode
-->
<script lang="ts">
  import { Moon, Sun } from "lucide-svelte";
  import { Button } from "$lib/components/ui/button";
  import { themeStore } from "$lib/stores/theme";

  let isDark = $derived($themeStore.mode === 'dark');
  let ariaLabel = $derived(isDark ? 'Switch to light mode' : 'Switch to dark mode');
</script>

<Button
  variant="ghost"
  size="icon"
  on:click={() => themeStore.toggleMode()}
  aria-label={ariaLabel}
>
  {#if isDark}
    <Sun class="h-5 w-5" />
  {:else}
    <Moon class="h-5 w-5" />
  {/if}
</Button> 