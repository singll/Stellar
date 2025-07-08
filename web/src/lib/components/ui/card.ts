import { cn } from '$lib/utils/theme';

export function Card(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'rounded-lg border bg-card text-card-foreground shadow-sm',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'rounded-lg border bg-card text-card-foreground shadow-sm',
        newProps.class
      );
    }
  };
}

export function CardHeader(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'flex flex-col space-y-1.5 p-6',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'flex flex-col space-y-1.5 p-6',
        newProps.class
      );
    }
  };
}

export function CardTitle(node: HTMLHeadingElement, props: { class?: string } = {}) {
  node.className = cn(
    'text-2xl font-semibold leading-none tracking-tight',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'text-2xl font-semibold leading-none tracking-tight',
        newProps.class
      );
    }
  };
}

export function CardDescription(node: HTMLParagraphElement, props: { class?: string } = {}) {
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

export function CardContent(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'p-6 pt-0',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'p-6 pt-0',
        newProps.class
      );
    }
  };
}

export function CardFooter(node: HTMLDivElement, props: { class?: string } = {}) {
  node.className = cn(
    'flex items-center p-6 pt-0',
    props.class
  );

  return {
    update(newProps: { class?: string }) {
      node.className = cn(
        'flex items-center p-6 pt-0',
        newProps.class
      );
    }
  };
} 