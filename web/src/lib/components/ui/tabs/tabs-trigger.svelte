<script lang="ts">
  import { getContext } from 'svelte';

  interface Props {
    value: string;
    children?: import('svelte').Snippet;
    class?: string;
  }

  let { value, children, class: className = '' }: Props = $props();

  const { selectedValue, setValue } = getContext<{ 
    selectedValue: import('svelte/store').Writable<string>; 
    setValue: (val: string) => void; 
  }>('tabs');

  let isSelected = $state(false);

  $effect(() => {
    const unsubscribe = selectedValue.subscribe(selected => {
      isSelected = selected === value;
    });
    return unsubscribe;
  });
</script>

<button
  type="button"
  role="tab"
  aria-selected={isSelected}
  onclick={() => setValue(value)}
  class="inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 {isSelected ? 'bg-background text-foreground shadow-sm' : 'hover:bg-background/50'} {className}"
>
  {@render children?.()}
</button>
