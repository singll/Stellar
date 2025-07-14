/**
 * shadcn-svelte 组件类型定义
 *
 * 注意：这些类型定义已经被实际的组件实现替代，保留此文件仅作为参考
 */

import type { SvelteComponent } from 'svelte';
import type { HTMLAttributes } from 'svelte/elements';

interface CommonProps {
	children?: any;
	class?: string;
}

// 移除了所有的 declare module 声明以避免与实际组件导出冲突
// 实际的组件类型定义现在位于各自的 .svelte 文件中
