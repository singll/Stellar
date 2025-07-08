<!-- Notification.svelte -->
<script lang="ts">
  import { fade, slide } from "svelte/transition";
  import { cn } from "$lib/utils";

  interface Props {
    type?: "success" | "error" | "warning" | "info";
    message: string;
    duration?: number;
    id: string;
    onClose?: (id: string) => void;
  }

  let { type = "info", message, duration = 3000, id, onClose }: Props = $props();

  let timeoutId: number;

  function close() {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    onClose?.(id);
  }

  $effect(() => {
    if (duration > 0) {
      timeoutId = window.setTimeout(close, duration);
    }
    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    };
  });

  const bgColors = {
    success: "bg-green-100 dark:bg-green-800/20",
    error: "bg-destructive/15 dark:bg-destructive/20",
    warning: "bg-yellow-100 dark:bg-yellow-800/20",
    info: "bg-primary/15 dark:bg-primary/20"
  };

  const textColors = {
    success: "text-green-800 dark:text-green-100",
    error: "text-destructive dark:text-destructive-foreground",
    warning: "text-yellow-800 dark:text-yellow-100",
    info: "text-primary dark:text-primary-foreground"
  };

  const icons = {
    success: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
    </svg>`,
    error: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
    </svg>`,
    warning: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>`,
    info: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>`
  };
</script>

<div
  class={cn(
    "pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-md border p-6 pr-8 shadow-lg transition-all",
    type === "info" && "bg-background text-foreground",
    type === "success" && "bg-success text-success-foreground",
    type === "error" && "bg-error text-error-foreground",
    type === "warning" && "bg-warning text-warning-foreground"
  )}
  role="alert"
  transition:slide={{ duration: 150, axis: "x" }}
>
  <div class="grid gap-1">
    <p class="text-sm font-semibold">
      {message}
    </p>
  </div>
  <button
    class="absolute right-2 top-2 rounded-md p-1 text-foreground/50 opacity-0 transition-opacity hover:text-foreground focus:opacity-100 focus:outline-none focus:ring-2 group-hover:opacity-100"
    onclick={close}
  >
    <span class="sr-only">Close</span>
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="h-4 w-4"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <line x1="18" y1="6" x2="6" y2="18"></line>
      <line x1="6" y1="6" x2="18" y2="18"></line>
    </svg>
  </button>
</div> 