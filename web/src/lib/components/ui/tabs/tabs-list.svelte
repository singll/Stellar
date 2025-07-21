<script lang="ts">
  interface Props {
    children?: import('svelte').Snippet;
    class?: string;
  }

  let { children, class: className = '' }: Props = $props();
</script>

<div class="inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground {className}" role="tablist">
  {@render children?.()}
</div>
