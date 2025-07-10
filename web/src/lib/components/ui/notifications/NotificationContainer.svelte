<script lang="ts">
	import { notifications, type NotificationData } from '$lib/stores/notifications';
	import { fly } from 'svelte/transition';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { Button } from '$lib/components/ui/button';

	let notificationList = $state<NotificationData[]>([]);

	// 订阅通知状态
	$effect(() => {
		const unsubscribe = notifications.subscribe((items) => {
			notificationList = items;
		});
		return unsubscribe;
	});

	function getNotificationIcon(type: NotificationData['type']) {
		switch (type) {
			case 'success':
				return 'check-circle';
			case 'error':
				return 'x-circle';
			case 'warning':
				return 'alert-triangle';
			case 'info':
				return 'info';
			default:
				return 'info';
		}
	}

	function getNotificationClass(type: NotificationData['type']) {
		switch (type) {
			case 'success':
				return 'notification-success';
			case 'error':
				return 'notification-error';
			case 'warning':
				return 'notification-warning';
			case 'info':
				return 'notification-info';
			default:
				return 'notification-info';
		}
	}
</script>

<div class="notification-container">
	{#each notificationList as notification (notification.id)}
		<div
			class="notification {getNotificationClass(notification.type)}"
			transition:fly={{ y: -50, duration: 300 }}
		>
			<div class="flex items-start gap-3">
				<Icon name={getNotificationIcon(notification.type)} class="h-5 w-5 flex-shrink-0" />
				<div class="flex-1">
					<p class="text-sm font-medium">{notification.message}</p>
				</div>
			</div>
			<Button
				variant="ghost"
				size="icon"
				class="absolute right-2 top-2 h-6 w-6"
				onclick={() => notifications.remove(notification.id)}
			>
				<Icon name="x" class="h-4 w-4" />
			</Button>
		</div>
	{/each}
</div>
