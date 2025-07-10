import { authApi, APIError } from '../auth';
import api from '../axios-config';
import type { User } from '$lib/types/auth';
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest';

vi.mock('../axios-config');
const mockedApi = api as any;

describe('authApi', () => {
	const mockUser: User = {
		id: '1',
		username: 'testuser',
		email: 'test@example.com',
		roles: ['user'],
		role: 'user',
		created: '2024-01-01T00:00:00Z',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-01T00:00:00Z',
		lastLogin: '2024-01-01T00:00:00Z'
	};

	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	describe('login', () => {
		it('登录成功返回 token 和 user', async () => {
			mockedApi.post.mockResolvedValueOnce({
				data: {
					code: 200,
					message: 'success',
					data: { token: 'jwt-token', user: mockUser }
				}
			});
			const res = await authApi.login({ email: 'test@example.com', password: '123456' });
			expect(res.code).toBe(200);
			expect(res.data?.token).toBe('jwt-token');
			expect(res.data?.user).toEqual(mockUser);
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/login', {
				email: 'test@example.com',
				password: '123456'
			});
		});

		it('支持用户名登录', async () => {
			mockedApi.post.mockResolvedValueOnce({
				data: {
					code: 200,
					message: 'success',
					data: { token: 'jwt-token', user: mockUser }
				}
			});
			const res = await authApi.login({ username: 'testuser', password: '123456' });
			expect(res.code).toBe(200);
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/login', {
				username: 'testuser',
				password: '123456'
			});
		});

		it('登录失败抛出 APIError', async () => {
			mockedApi.post.mockRejectedValueOnce({
				response: { data: { code: 401, message: '认证失败', details: '密码错误' } }
			});
			try {
				await authApi.login({ email: 'test@example.com', password: 'wrong' });
				expect.fail('应该抛出错误');
			} catch (error) {
				expect(error).toBeInstanceOf(APIError);
				expect((error as APIError).code).toBe(401);
				expect((error as APIError).message).toBe('认证失败');
				expect((error as APIError).details).toBe('密码错误');
			}
		});

		it('网络错误时抛出原始错误', async () => {
			const networkError = new Error('Network Error');
			mockedApi.post.mockRejectedValueOnce(networkError);
			try {
				await authApi.login({ email: 'test@example.com', password: '123456' });
				expect.fail('应该抛出错误');
			} catch (error) {
				expect(error).toBe(networkError);
			}
		});
	});

	describe('register', () => {
		it('注册成功返回 token 和 user', async () => {
			mockedApi.post.mockResolvedValueOnce({
				data: {
					code: 200,
					message: '注册成功',
					data: { token: 'jwt-token', user: mockUser }
				}
			});
			const res = await authApi.register({
				username: 'newuser',
				email: 'new@example.com',
				password: 'password123'
			});
			expect(res.code).toBe(200);
			expect(res.data?.token).toBe('jwt-token');
			expect(res.data?.user).toEqual(mockUser);
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/register', {
				username: 'newuser',
				email: 'new@example.com',
				password: 'password123'
			});
		});

		it('注册失败抛出 APIError', async () => {
			mockedApi.post.mockRejectedValueOnce({
				response: { data: { code: 400, message: '用户名已存在', details: '请选择其他用户名' } }
			});
			try {
				await authApi.register({
					username: 'existinguser',
					email: 'existing@example.com',
					password: 'password123'
				});
				expect.fail('应该抛出错误');
			} catch (error) {
				expect(error).toBeInstanceOf(APIError);
				expect((error as APIError).code).toBe(400);
				expect((error as APIError).message).toBe('用户名已存在');
			}
		});
	});

	describe('logout', () => {
		it('logout 成功', async () => {
			mockedApi.post.mockResolvedValueOnce({});
			await expect(authApi.logout()).resolves.toBeUndefined();
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/logout');
		});

		it('logout 失败时不抛出错误', async () => {
			mockedApi.post.mockRejectedValueOnce({
				response: { data: { code: 500, message: '服务器错误' } }
			});
			// logout 失败时不应该抛出错误，只是警告
			await expect(authApi.logout()).resolves.toBeUndefined();
		});

		it('网络错误时不抛出错误', async () => {
			mockedApi.post.mockRejectedValueOnce(new Error('Network Error'));
			await expect(authApi.logout()).resolves.toBeUndefined();
		});
	});

	describe('getCurrentUser', () => {
		it('getCurrentUser 成功返回 user', async () => {
			mockedApi.get.mockResolvedValueOnce({
				data: { code: 200, message: 'ok', data: mockUser }
			});
			const user = await authApi.getCurrentUser();
			expect(user).toEqual(mockUser);
			expect(mockedApi.get).toHaveBeenCalledWith('/auth/me');
		});

		it('getCurrentUser 失败抛出 APIError', async () => {
			mockedApi.get.mockRejectedValueOnce({
				response: { data: { code: 401, message: '未认证' } }
			});
			try {
				await authApi.getCurrentUser();
				expect.fail('应该抛出错误');
			} catch (error) {
				expect(error).toBeInstanceOf(APIError);
				expect((error as APIError).code).toBe(401);
			}
		});
	});

	describe('其他认证功能', () => {
		it('refreshToken 成功', async () => {
			mockedApi.post.mockResolvedValueOnce({
				data: {
					code: 200,
					message: '刷新成功',
					data: { token: 'new-jwt-token', user: mockUser }
				}
			});
			const res = await authApi.refreshToken();
			expect(res.code).toBe(200);
			expect(res.data?.token).toBe('new-jwt-token');
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/refresh');
		});

		it('updateProfile 成功', async () => {
			const updatedUser = { ...mockUser, username: 'updateduser' };
			mockedApi.put.mockResolvedValueOnce({
				data: { code: 200, message: '更新成功', data: updatedUser }
			});
			const user = await authApi.updateProfile({ username: 'updateduser' });
			expect(user).toEqual(updatedUser);
			expect(mockedApi.put).toHaveBeenCalledWith('/auth/profile', { username: 'updateduser' });
		});

		it('changePassword 成功', async () => {
			mockedApi.put.mockResolvedValueOnce({ data: { code: 200, message: '密码修改成功' } });
			await expect(
				authApi.changePassword({
					oldPassword: 'oldpass',
					newPassword: 'newpass'
				})
			).resolves.toBeUndefined();
			expect(mockedApi.put).toHaveBeenCalledWith('/auth/password', {
				oldPassword: 'oldpass',
				newPassword: 'newpass'
			});
		});

		it('resetPassword 成功', async () => {
			mockedApi.post.mockResolvedValueOnce({ data: { code: 200, message: '重置邮件已发送' } });
			await expect(authApi.resetPassword('test@example.com')).resolves.toBeUndefined();
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/reset-password', {
				email: 'test@example.com'
			});
		});

		it('verifyResetToken 成功', async () => {
			mockedApi.post.mockResolvedValueOnce({ data: { code: 200, message: '密码重置成功' } });
			await expect(authApi.verifyResetToken('reset-token', 'newpass')).resolves.toBeUndefined();
			expect(mockedApi.post).toHaveBeenCalledWith('/auth/verify-reset', {
				token: 'reset-token',
				newPassword: 'newpass'
			});
		});
	});

	describe('APIError 类', () => {
		it('正确创建 APIError 实例', () => {
			const error = new APIError(400, '验证失败', '密码格式错误');
			expect(error.code).toBe(400);
			expect(error.message).toBe('验证失败');
			expect(error.details).toBe('密码格式错误');
			expect(error.name).toBe('APIError');
			expect(error).toBeInstanceOf(Error);
		});

		it('可以不传 details 参数', () => {
			const error = new APIError(404, '资源不存在');
			expect(error.code).toBe(404);
			expect(error.message).toBe('资源不存在');
			expect(error.details).toBeUndefined();
		});
	});
});
