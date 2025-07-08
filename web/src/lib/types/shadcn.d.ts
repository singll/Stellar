/**
 * shadcn-svelte 组件类型定义
 */

import type { SvelteComponent } from 'svelte';
import type { HTMLAttributes } from 'svelte/elements';

interface CommonProps {
	children?: any;
	class?: string;
}

declare module '$lib/components/ui/button' {
	interface ButtonProps extends HTMLAttributes<HTMLButtonElement>, CommonProps {
		variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
		size?: 'default' | 'sm' | 'lg' | 'icon';
		disabled?: boolean;
		type?: 'button' | 'submit' | 'reset';
		href?: string;
	}

	export class Button extends SvelteComponent<ButtonProps> {}
}

declare module '$lib/components/ui/card' {
	interface CardProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}
	interface CardHeaderProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}
	interface CardTitleProps extends HTMLAttributes<HTMLHeadingElement>, CommonProps {}
	interface CardDescriptionProps extends HTMLAttributes<HTMLParagraphElement>, CommonProps {}
	interface CardContentProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}
	interface CardFooterProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}

	export class Card extends SvelteComponent<CardProps> {}
	export class CardHeader extends SvelteComponent<CardHeaderProps> {}
	export class CardTitle extends SvelteComponent<CardTitleProps> {}
	export class CardDescription extends SvelteComponent<CardDescriptionProps> {}
	export class CardContent extends SvelteComponent<CardContentProps> {}
	export class CardFooter extends SvelteComponent<CardFooterProps> {}
}

declare module '$lib/components/ui/dialog' {
	interface DialogProps extends CommonProps {
		open?: boolean;
	}

	interface DialogContentProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}
	interface DialogHeaderProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}
	interface DialogFooterProps extends HTMLAttributes<HTMLDivElement>, CommonProps {}
	interface DialogTitleProps extends HTMLAttributes<HTMLHeadingElement>, CommonProps {}
	interface DialogDescriptionProps extends HTMLAttributes<HTMLParagraphElement>, CommonProps {}
	interface DialogTriggerProps extends CommonProps {
		asChild?: boolean;
	}

	export class Dialog extends SvelteComponent<DialogProps> {}
	export class DialogContent extends SvelteComponent<DialogContentProps> {}
	export class DialogHeader extends SvelteComponent<DialogHeaderProps> {}
	export class DialogFooter extends SvelteComponent<DialogFooterProps> {}
	export class DialogTitle extends SvelteComponent<DialogTitleProps> {}
	export class DialogDescription extends SvelteComponent<DialogDescriptionProps> {}
	export class DialogTrigger extends SvelteComponent<DialogTriggerProps> {}
}

declare module '$lib/components/ui/dropdown-menu' {
	import type { SvelteComponent } from 'svelte';

	export class DropdownMenu extends SvelteComponent<{
		class?: string;
	}> {}

	export class DropdownMenuTrigger extends SvelteComponent<{
		asChild?: boolean;
		class?: string;
	}> {}

	export class DropdownMenuContent extends SvelteComponent<{
		align?: 'start' | 'end';
		class?: string;
	}> {}

	export class DropdownMenuItem extends SvelteComponent<{
		class?: string;
	}> {}
}

declare module '$lib/components/ui/input' {
	interface InputProps extends HTMLAttributes<HTMLInputElement>, CommonProps {
		type?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		accept?: string;
	}

	export class Input extends SvelteComponent<InputProps> {}
}

declare module '$lib/components/ui/label' {
	interface LabelProps extends HTMLAttributes<HTMLLabelElement>, CommonProps {
		for?: string;
	}

	export class Label extends SvelteComponent<LabelProps> {}
}

declare module '$lib/components/ui/popover' {
	import type { SvelteComponent } from 'svelte';

	export class Popover extends SvelteComponent<{
		open?: boolean;
		class?: string;
	}> {}

	export class PopoverTrigger extends SvelteComponent<{
		asChild?: boolean;
		class?: string;
	}> {}

	export class PopoverContent extends SvelteComponent<{
		align?: 'start' | 'end';
		class?: string;
	}> {}
}

declare module '$lib/components/ui/switch' {
	import type { SvelteComponent } from 'svelte';

	export class Switch extends SvelteComponent<{
		checked?: boolean;
		class?: string;
	}> {}
}

declare module '$lib/components/ui/textarea' {
	import type { SvelteComponent } from 'svelte';

	export class Textarea extends SvelteComponent<{
		value?: string;
		placeholder?: string;
		rows?: number;
		id?: string;
		required?: boolean;
		class?: string;
	}> {}
}
