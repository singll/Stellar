import type { Handle } from '@sveltejs/kit';

// 公开路由列表
const publicRoutes = [
	'/login',
	'/register',
	'/forgot-password',
	'/reset-password',
	'/verify-email',
	'/test-auth',
	'/test-basic',
	'/test-minimal',
	'/test-simple'
];

// 检查是否是公开路由
function isPublicRoute(path: string): boolean {
	return publicRoutes.some((route) => path.startsWith(route));
}

export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;
	
	// 如果是公开路由，直接放行
	if (isPublicRoute(path)) {
		return resolve(event);
	}
	
	// 对于受保护的路由，在客户端处理认证
	// 这里我们不进行服务器端认证，让客户端处理
	return resolve(event);
};