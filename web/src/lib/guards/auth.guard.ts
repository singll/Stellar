import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import { auth } from '$lib/stores/auth';
import { browser } from '$app/environment';

// 公开路由列表
const publicRoutes = [
	'/login',
	'/register',
	'/forgot-password',
	'/reset-password',
	'/verify-email'
];

// 检查是否是公开路由
function isPublicRoute(path: string): boolean {
	return publicRoutes.some((route) => path.startsWith(route));
}

// 认证路由守卫
export const authGuard: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;
	
	// 只在客户端执行认证检查
	if (!browser) {
		return resolve(event);
	}

	// 初始化认证状态（如果还没有初始化）
	await auth.initialize();

	// 检查认证状态
	const { isAuthenticated, token } = auth.state;

	console.log('路由守卫检查:', { path, isAuthenticated, hasToken: !!token });

	// 如果已认证且访问登录页，重定向到dashboard
	if (isAuthenticated && token && path === '/login') {
		console.log('已认证用户访问登录页，重定向到dashboard');
		throw redirect(303, '/dashboard');
	}

	// 如果是公开路由，直接放行
	if (isPublicRoute(path)) {
		console.log('公开路由，直接放行');
		return resolve(event);
	}

	// 如果未认证且不是公开路由，需要验证会话
	if (!isAuthenticated || !token) {
		console.log('未认证，尝试从localStorage恢复状态');
		
		// 尝试从localStorage恢复状态
		const storedState = localStorage.getItem('auth_state');
		if (storedState) {
			try {
				const parsedState = JSON.parse(storedState);
				if (parsedState.token && parsedState.user) {
					console.log('发现存储的认证状态，验证会话');
					// 验证会话状态
					const isValid = await auth.verifySession();
					if (isValid) {
						console.log('会话验证成功，继续处理请求');
						// 会话有效，继续处理请求
						return resolve(event);
					} else {
						console.log('会话验证失败');
					}
				}
			} catch (error) {
				console.error('解析存储的认证状态失败:', error);
			}
		}
		
		console.log('会话无效，重定向到登录页');
		// 会话无效，重定向到登录页
		throw redirect(303, `/login?redirect=${encodeURIComponent(path)}`);
	}

	console.log('已认证，继续处理请求');
	// 已认证，继续处理请求
	return resolve(event);
};

// 角色守卫
export function requireRole(allowedRoles: string[]) {
	return async ({ event, resolve }: { event: any; resolve: any }) => {
		const { user } = auth.state;

		if (!user || !allowedRoles.includes(user.role)) {
			throw redirect(303, '/403');
		}

		return resolve(event);
	};
}

// 权限守卫
export function requirePermission(permission: string) {
	return async ({ event, resolve }: { event: any; resolve: any }) => {
		const { user } = auth.state;

		// TODO: 实现权限检查逻辑
		if (!user || !user.permissions?.includes(permission)) {
			throw redirect(303, '/403');
		}

		return resolve(event);
	};
}
