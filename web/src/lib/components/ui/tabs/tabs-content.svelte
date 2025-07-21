<script lang="ts">
  import { getContext } from 'svelte';

  interface Props {
    value: string;
    children?: import('svelte').Snippet;
    class?: string;
  }

  let { value, children, class: className = '' }: Props = $props();

  const { selectedValue } = getContext<{ selectedValue: import('svelte/store').Writable<string> }>('tabs');

  let isVisible = $state(false);

  $effect(() => {
    const unsubscribe = selectedValue.subscribe(selected => {
      isVisible = selected === value;
    });
    return unsubscribe;
  });
</script>

{#if isVisible}
  <div role="tabpanel" class="mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 {className}">
    {@render children?.()}
  </div>
{/if}
