import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	if (browser) {
		const authState = get(auth);
		if (!authState.isAuthenticated) {
			// 如果未认证，重定向到登录页，并记录当前路径
			const currentPath = window.location.pathname;
			goto(`/login?redirect=${encodeURIComponent(currentPath)}`);
			return {};
		}
	}

	return {};
};

export const ssr = false;
