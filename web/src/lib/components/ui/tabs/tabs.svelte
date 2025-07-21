<script lang="ts">
  import { setContext } from 'svelte';
  import { writable } from 'svelte/store';

  interface Props {
    value?: string;
    children?: import('svelte').Snippet;
    class?: string;
  }

  let { value = $bindable(), children, class: className = '' }: Props = $props();

  const selectedValue = writable(value);
  
  setContext('tabs', {
    selectedValue,
    setValue: (val: string) => {
      value = val;
      selectedValue.set(val);
    }
  });

  $effect(() => {
    selectedValue.set(value);
  });
</script>

<div class="w-full {className}">
  {@render children?.()}
</div>