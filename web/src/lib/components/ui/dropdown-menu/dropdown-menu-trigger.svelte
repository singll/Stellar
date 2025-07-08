<script lang="ts">
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import { cn } from "$lib/utils";

  let className: string | undefined = undefined;
  export { className as class };

  export let asChild = false;
</script>

<DropdownMenuPrimitive.Trigger
  {asChild}
  class={cn(
    "inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=open]:bg-accent data-[state=open]:text-accent-foreground",
    className
  )}
  {...$$restProps}
>
  <slot />
</DropdownMenuPrimitive.Trigger> 