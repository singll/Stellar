<script lang="ts">
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  export let open: boolean = false;
</script>

<DropdownMenuPrimitive.Root bind:open {...$$restProps}>
  <slot />
</DropdownMenuPrimitive.Root> 