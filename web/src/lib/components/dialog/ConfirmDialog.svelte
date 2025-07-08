<!-- ConfirmDialog.svelte -->
<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "$lib/components/ui/dialog";
  import { createEventDispatcher } from "svelte";

  export let open = false;
  export let title = "确认";
  export let message = "您确定要执行此操作吗？";
  export let confirmText = "确认";
  export let cancelText = "取消";
  export let type: "danger" | "warning" | "info" = "info";

  const dispatch = createEventDispatcher();

  function handleConfirm() {
    dispatch("confirm");
    open = false;
  }

  function handleCancel() {
    dispatch("cancel");
    open = false;
  }
</script>

<Dialog bind:open>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>{title}</DialogTitle>
    </DialogHeader>
    <div class="py-4">
      <p class="text-sm text-gray-500 dark:text-gray-400">{message}</p>
    </div>
    <DialogFooter>
      <div class="flex justify-end gap-2">
        <Button
          variant="outline"
          on:click={handleCancel}
        >
          {cancelText}
        </Button>
        <Button
          variant={type === "danger" ? "destructive" : type === "warning" ? "warning" : "primary"}
          on:click={handleConfirm}
        >
          {confirmText}
        </Button>
      </div>
    </DialogFooter>
  </DialogContent>
</Dialog> 