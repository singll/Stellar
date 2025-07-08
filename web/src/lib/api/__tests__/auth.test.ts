import { authApi, APIError } from '../auth';
import api from '../axios-config';
import type { User } from '$lib/types/auth';
import { describe, it, expect, afterEach, vi } from 'vitest';

vi.mock('../axios-config');
const mockedApi = api as any;

describe('authApi', () => {
	const mockUser: User = {
		id: '1',
		username: 'testuser',
		email: 'test@example.com',
		role: 'user',
		created_at: '2024-01-01T00:00:00Z',
		last_login: '2024-01-01T00:00:00Z'
	};

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('login 成功返回 token 和 user', async () => {
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
	});

	it('login 失败抛出 APIError', async () => {
		mockedApi.post.mockRejectedValueOnce({
			response: { data: { code: 401, message: '认证失败', details: '密码错误' } }
		});
		await expect(authApi.login({ email: 'test@example.com', password: 'wrong' })).rejects.toThrow(
			APIError
		);
	});

	it('logout 成功', async () => {
		mockedApi.post.mockResolvedValueOnce({});
		await expect(authApi.logout()).resolves.toBeUndefined();
	});

	it('logout 失败抛出 APIError', async () => {
		mockedApi.post.mockRejectedValueOnce({
			response: { data: { code: 500, message: '服务器错误' } }
		});
		await expect(authApi.logout()).rejects.toThrow(APIError);
	});

	it('getCurrentUser 成功返回 user', async () => {
		mockedApi.get.mockResolvedValueOnce({
			data: { code: 200, message: 'ok', data: mockUser }
		});
		const user = await authApi.getCurrentUser();
		expect(user).toEqual(mockUser);
	});

	it('getCurrentUser 失败抛出 APIError', async () => {
		mockedApi.get.mockRejectedValueOnce({
			response: { data: { code: 401, message: '未认证' } }
		});
		await expect(authApi.getCurrentUser()).rejects.toThrow(APIError);
	});
});
