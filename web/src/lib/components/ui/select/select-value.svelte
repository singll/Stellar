<script lang="ts">
  import { cn } from "$lib/utils";

  let className: string | undefined = undefined;
  export { className as class };

  export let placeholder: string | undefined = undefined;
</script>

<span class={cn("block truncate", className)}>
  <slot>
    {#if placeholder}
      <span class="text-muted-foreground">{placeholder}</span>
    {/if}
  </slot>
</span> 