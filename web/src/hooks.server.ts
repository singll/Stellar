import type { Handle } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import { authGuard } from '$lib/guards/auth.guard';

// 应用路由守卫
export const handle: Handle = sequence(authGuard);
