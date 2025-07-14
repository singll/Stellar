/**
 * Toast 通知系统状态管理
 * 提供全局通知功能
 */

import { writable, get } from 'svelte/store';

export interface Toast {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	title: string;
	message?: string;
	duration?: number; // 毫秒，0 表示不自动关闭
	action?: {
		label: string;
		handler: () => void;
	};
}

interface ToastState {
	toasts: Toast[];
}

const initialState: ToastState = {
	toasts: []
};

// 创建基础 store
const baseStore = writable<ToastState>(initialState);

// Toast 操作
class ToastActions {
	// 添加 toast
	add(toast: Omit<Toast, 'id'>): string {
		const id = generateId();
		const newToast: Toast = {
			id,
			duration: 5000, // 默认 5 秒
			...toast
		};

		baseStore.update((state) => ({
			...state,
			toasts: [...state.toasts, newToast]
		}));

		// 自动移除
		if (newToast.duration && newToast.duration > 0) {
			setTimeout(() => {
				this.remove(id);
			}, newToast.duration);
		}

		return id;
	}

	// 移除 toast
	remove(id: string): void {
		baseStore.update((state) => ({
			...state,
			toasts: state.toasts.filter((toast) => toast.id !== id)
		}));
	}

	// 清空所有 toast
	clear(): void {
		baseStore.update((state) => ({
			...state,
			toasts: []
		}));
	}

	// 便捷方法
	success(title: string, message?: string, options?: Partial<Toast>): string {
		return this.add({
			type: 'success',
			title,
			message,
			...options
		});
	}

	error(title: string, message?: string, options?: Partial<Toast>): string {
		return this.add({
			type: 'error',
			title,
			message,
			duration: 0, // 错误消息默认不自动关闭
			...options
		});
	}

	warning(title: string, message?: string, options?: Partial<Toast>): string {
		return this.add({
			type: 'warning',
			title,
			message,
			...options
		});
	}

	info(title: string, message?: string, options?: Partial<Toast>): string {
		return this.add({
			type: 'info',
			title,
			message,
			...options
		});
	}

	// 获取当前状态
	get toasts(): Toast[] {
		return get(baseStore).toasts;
	}

	// Store 订阅方法
	subscribe = baseStore.subscribe;
	set = baseStore.set;
	update = baseStore.update;
}

// 生成唯一 ID
function generateId(): string {
	return `toast_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

// 导出操作实例
export const toastActions = new ToastActions();

// 为了向后兼容，将方法直接附加到 store 上
export const toastStore = Object.assign(baseStore, toastActions);

// 导出默认值
export default toastStore;
