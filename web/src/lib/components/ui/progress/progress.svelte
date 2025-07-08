<script lang="ts">
  import { cn } from "$lib/utils";
  import { Progress as ProgressPrimitive } from "bits-ui";

  let className: string | undefined = undefined;
  export { className as class };
  export let value: number | undefined = undefined;
</script>

<ProgressPrimitive.Root
  class={cn(
    "relative h-2 w-full overflow-hidden rounded-full bg-primary/20",
    className
  )}
  {...$$restProps}
>
  <div
    class="h-full w-full flex-1 bg-primary transition-all"
    style="transform: translateX(-{100 - (value || 0)}%)"
  />
</ProgressPrimitive.Root> 