import { cn } from '$lib/utils/theme';

export function Label(node: HTMLLabelElement, props: { class?: string } = {}) {
	node.className = cn(
		'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
		props.class
	);

	return {
		update(newProps: { class?: string }) {
			node.className = cn(
				'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
				newProps.class
			);
		}
	};
}
