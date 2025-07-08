<!-- NotificationContainer.svelte -->
<script lang="ts">
  import { notifications } from "$lib/stores/notifications";
  import Notification from "./Notification.svelte";

  function handleClose(id: string) {
    notifications.remove(id);
  }
</script>

<div class="pointer-events-none fixed inset-0 z-50 flex items-start justify-end p-6 gap-2">
  {#each $notifications as notification (notification.id)}
    <Notification
      type={notification.type}
      message={notification.message}
      duration={notification.duration}
      id={notification.id}
      onClose={handleClose}
    />
  {/each}
</div> 