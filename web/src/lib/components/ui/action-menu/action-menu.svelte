<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "$lib/components/ui/dropdown-menu";
  import { MoreHorizontal } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";

  export let actions: Array<{
    label: string;
    value: string;
    disabled?: boolean;
  }> = [];

  const dispatch = createEventDispatcher<{
    select: { value: string };
  }>();

  function handleSelect(value: string) {
    dispatch("select", { value });
  }
</script>

<DropdownMenu>
  <DropdownMenuTrigger asChild let:builder>
    <Button
      variant="ghost"
      size="icon"
      class="h-8 w-8 p-0"
      {...builder()}
    >
      <span class="sr-only">打开菜单</span>
      <MoreHorizontal class="h-4 w-4" />
    </Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent align="end">
    {#each actions as action}
      <DropdownMenuItem
        disabled={action.disabled}
        on:click={() => handleSelect(action.value)}
      >
        {action.label}
      </DropdownMenuItem>
    {/each}
  </DropdownMenuContent>
</DropdownMenu> 