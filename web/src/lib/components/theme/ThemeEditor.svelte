<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Theme } from '$lib/types/theme';
  import ThemeForm from './ThemeForm.svelte';
  import ThemePreview from './ThemePreview.svelte';

  export let theme: Theme | undefined = undefined;
  export let baseTheme: Theme | undefined = undefined;

  const dispatch = createEventDispatcher<{
    save: Theme;
    cancel: void;
  }>();

  function handleSave(event: CustomEvent<Theme>) {
    dispatch('save', event.detail);
  }

  function handleCancel() {
    dispatch('cancel');
  }
</script>

<div class="grid grid-cols-2 gap-8">
  <!-- 左侧：主题表单 -->
  <div class="space-y-6">
    <h2 class="text-2xl font-bold">
      {theme ? '编辑主题' : '创建主题'}
    </h2>
    
    <ThemeForm
      {theme}
      {baseTheme}
      on:save={handleSave}
      on:cancel={handleCancel}
    />
  </div>

  <!-- 右侧：主题预览 -->
  <div class="space-y-6">
    <h2 class="text-2xl font-bold">预览</h2>
    
    <div class="sticky top-6">
      <ThemePreview theme={theme || baseTheme || getDefaultTheme()} />
    </div>
  </div>
</div> 