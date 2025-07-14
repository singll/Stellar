import Root, { type DialogProps } from './dialog.svelte';
import DialogClose from './dialog-close.svelte';
import DialogContent from './dialog-content.svelte';
import DialogDescription from './dialog-description.svelte';
import DialogFooter from './dialog-footer.svelte';
import DialogHeader from './dialog-header.svelte';
import DialogOverlay from './dialog-overlay.svelte';
import DialogTitle from './dialog-title.svelte';
import DialogTrigger from './dialog-trigger.svelte';

export {
	Root as Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogOverlay,
	DialogTitle,
	DialogTrigger,
	// 添加 Portal 和 Overlay 别名以保持向后兼容
	DialogOverlay as Portal,
	DialogOverlay as Overlay
};

export type { DialogProps };
