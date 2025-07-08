import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import { auth } from '$lib/stores/auth';

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
  return publicRoutes.some(route => path.startsWith(route));
}

// 认证路由守卫
export const authGuard: Handle = async ({ event, resolve }) => {
  const path = event.url.pathname;

  // 如果是公开路由，直接放行
  if (isPublicRoute(path)) {
    return resolve(event);
  }

  // 检查认证状态
  const { isAuthenticated, token } = auth.state;

  // 如果未认证且不是公开路由，重定向到登录页
  if (!isAuthenticated || !token) {
    throw redirect(303, `/login?redirect=${encodeURIComponent(path)}`);
  }

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