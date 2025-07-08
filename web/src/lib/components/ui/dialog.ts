import { cn } from '$lib/utils/theme';

export function Dialog(node: HTMLDivElement, props: { class?: string; open?: boolean } = {}) {
  node.setAttribute('role', 'dialog');
  node.setAttribute('aria-modal', 'true');
  node.className = cn(
    'relative z-50',
    props.class
  );

  if (props.open) {
    node.style.display = 'block';
  } else {
    node.style.display = 'none';
  }

  return {
    update(newProps: { class?: string; open?: boolean }) {
      node.className = cn(
        'relative z-50',
        newProps.class
      );
      if (newProps.open) {
        node.style.display = 'block';
      } else {
        node.style.display = 'none';
      }
    }
  };
}

export function DialogTrigger(node: HTMLButtonElement, props: { class?: string } = {}) {
  node.className = cn(props.class);

  return {
    update(newProps: { class?: string }) {
      node.className = cn(newProps.class);
    }
  };
}

export function DialogContent(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'fixed inset-0 z-50 flex items-center justify-center',
    props.class
  );

  // 创建背景遮罩
  const overlay = document.createElement('div');
  overlay.className = 'fixed inset-0 bg-black/50';
  node.appendChild(overlay);

  // 创建内容容器
  const content = document.createElement('div');
  content.className = cn(
    'relative bg-background rounded-lg shadow-lg w-full max-w-lg p-6',
    'max-h-[85vh] overflow-y-auto'
  );
  node.appendChild(content);

  // 移动原有子元素到内容容器
  Array.from(node.children).forEach(child => {
    if (child !== overlay && child !== content) {
      content.appendChild(child);
    }
  });

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'fixed inset-0 z-50 flex items-center justify-center',
        newProps.class
      );
    }
  };
}

export function DialogHeader(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'flex flex-col space-y-1.5 text-center sm:text-left',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'flex flex-col space-y-1.5 text-center sm:text-left',
        newProps.class
      );
    }
  };
}

export function DialogTitle(node: HTMLHeadingElement, props: { class?: string } = {}) {
  node.className = cn(
    'text-lg font-semibold leading-none tracking-tight',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'text-lg font-semibold leading-none tracking-tight',
        newProps.class
      );
    }
  };
}

export function DialogDescription(node: HTMLParagraphElement, props: { class?: string } = {}) {
  node.className = cn(
    'text-sm text-muted-foreground',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'text-sm text-muted-foreground',
        newProps.class
      );
    }
  };
} 