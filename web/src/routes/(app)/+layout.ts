import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	if (browser) {
		const authState = get(auth);
		if (!authState.isAuthenticated) {
			goto('/login');
			return {};
		}
	}

	return {};
};

export const ssr = false;
