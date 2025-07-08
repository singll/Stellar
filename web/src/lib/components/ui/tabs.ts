import { cn } from '$lib/utils/theme';

export function Tabs(node: HTMLDivElement, props: { class?: string; defaultValue?: string } = {}) {
	node.setAttribute('role', 'tablist');
	node.className = cn('w-full', props.class);

	if (props.defaultValue) {
		node.dataset.defaultValue = props.defaultValue;
	}

	return {
		update(newProps: { class?: string; defaultValue?: string }) {
			node.className = cn('w-full', newProps.class);
			if (newProps.defaultValue) {
				node.dataset.defaultValue = newProps.defaultValue;
			}
		}
	};
}

export function TabsList(node: HTMLDivElement, props: { class?: string } = {}) {
	node.setAttribute('role', 'tablist');
	node.className = cn(
		'inline-flex h-10 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground',
		props.class
	);

	return {
		update(newProps: { class?: string }) {
			node.className = cn(
				'inline-flex h-10 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground',
				newProps.class
			);
		}
	};
}

export function TabsTrigger(
	node: HTMLButtonElement,
	props: { class?: string; value: string } = { value: '' }
) {
	node.setAttribute('role', 'tab');
	node.type = 'button';
	node.className = cn(
		'inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm',
		props.class
	);

	if (props.value) {
		node.dataset.value = props.value;
	}

	return {
		update(newProps: { class?: string; value: string }) {
			node.className = cn(
				'inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm',
				newProps.class
			);
			if (newProps.value) {
				node.dataset.value = newProps.value;
			}
		}
	};
}

export function TabsContent(
	node: HTMLDivElement,
	props: { class?: string; value: string } = { value: '' }
) {
	node.setAttribute('role', 'tabpanel');
	node.className = cn(
		'mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
		props.class
	);

	if (props.value) {
		node.dataset.value = props.value;
	}

	return {
		update(newProps: { class?: string; value: string }) {
			node.className = cn(
				'mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
				newProps.class
			);
			if (newProps.value) {
				node.dataset.value = newProps.value;
			}
		}
	};
}
