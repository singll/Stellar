import Root, { type DialogProps } from './dialog.svelte';

// 为了兼容性，创建简单的组件替代品
const DialogContent = Root;
const DialogDescription = Root;
const DialogFooter = Root;
const DialogHeader = Root;
const DialogTitle = Root;
const DialogTrigger = Root;

export {
	Root,
	Root as Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger
};

export type { DialogProps };
