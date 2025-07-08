import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';

export interface NotificationData {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	message: string;
	duration?: number;
}

interface NotificationStore {
	subscribe: Writable<NotificationData[]>['subscribe'];
	add: (notification: Omit<NotificationData, 'id'>) => void;
	remove: (id: string) => void;
	clear: () => void;
}

function createNotificationStore(): NotificationStore {
	const { subscribe, set, update } = writable<NotificationData[]>([]);

	return {
		subscribe,
		add: (notification) => {
			const id = crypto.randomUUID();
			update((notifications) => [...notifications, { ...notification, id }]);

			if (notification.duration !== 0) {
				setTimeout(() => {
					update((notifications) => notifications.filter((n) => n.id !== id));
				}, notification.duration || 3000);
			}
		},
		remove: (id) => {
			update((notifications) => notifications.filter((n) => n.id !== id));
		},
		clear: () => set([])
	};
}

export const notifications = createNotificationStore();
