import { cn } from '$lib/utils/theme';

export function DropdownMenu(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'relative inline-block text-left',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'relative inline-block text-left',
        newProps.class
      );
    }
  };
}

export function DropdownMenuTrigger(node: HTMLButtonElement, props: { class?: string } = {}) {
  node.className = cn(
    'inline-flex w-full justify-center rounded-md bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'inline-flex w-full justify-center rounded-md bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100',
        newProps.class
      );
    }
  };
}

export function DropdownMenuContent(node: HTMLDivElement, props: { class?: string; align?: 'start' | 'end' } = {}) {
  node.className = cn(
    'absolute z-10 mt-2 w-56 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none',
    props.align === 'end' ? 'right-0' : 'left-0',
    props.class
  );

  return {
    update(newProps: { class?: string; align?: 'start' | 'end' }) {
      node.className = cn(
        'absolute z-10 mt-2 w-56 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none',
        newProps.align === 'end' ? 'right-0' : 'left-0',
        newProps.class
      );
    }
  };
}

export function DropdownMenuItem(node: HTMLButtonElement, props: { class?: string } = {}) {
  node.className = cn(
    'text-gray-700 block w-full px-4 py-2 text-left text-sm hover:bg-gray-100 focus:outline-none focus:bg-gray-100',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'text-gray-700 block w-full px-4 py-2 text-left text-sm hover:bg-gray-100 focus:outline-none focus:bg-gray-100',
        newProps.class
      );
    }
  };
} 