import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { auth } from '../auth';
import { tokenManager } from '$lib/auth/token-manager';
import api from '$lib/api/axios-config';

vi.mock('$lib/auth/token-manager', () => ({
	tokenManager: {
		getAccessToken: vi.fn(),
		setTokens: vi.fn(),
		clearTokens: vi.fn(),
		subscribe: vi.fn()
	}
}));

vi.mock('$lib/api/axios-config', () => ({
	default: {
		post: vi.fn(),
		get: vi.fn()
	}
}));

describe('Auth Store', () => {
	const mockUser = {
		id: 'test-id',
		username: 'testuser',
		email: 'test@example.com',
		role: 'user'
	};

	beforeEach(() => {
		vi.clearAllMocks();
		// Reset the store to initial state
		auth.logout();
	});

	describe('login', () => {
		it('should handle successful login', async () => {
			const credentials = {
				username: 'testuser',
				password: 'password123'
			};

			const mockResponse = {
				data: {
					code: 200,
					data: {
						user: mockUser,
						token: 'test-access-token'
					}
				}
			};

			(api.post as any).mockResolvedValueOnce(mockResponse);

			await auth.login(credentials);

			expect(api.post).toHaveBeenCalledWith('/auth/login', credentials);
			expect(auth.state.isAuthenticated).toBe(true);
			expect(auth.state.user).toEqual(mockUser);
		});

		it('should handle login failure', async () => {
			const credentials = {
				username: 'testuser',
				password: 'wrong-password'
			};

			(api.post as any).mockRejectedValueOnce(new Error('Login failed'));

			await expect(auth.login(credentials)).rejects.toThrow('Login failed');
			expect(auth.state.isAuthenticated).toBe(false);
			expect(auth.state.user).toBeNull();
		});
	});

	describe('logout', () => {
		it('should handle successful logout', async () => {
			(api.post as any).mockResolvedValueOnce({});

			await auth.logout();

			expect(auth.state.isAuthenticated).toBe(false);
			expect(auth.state.user).toBeNull();
		});
	});

	describe('register', () => {
		it('should handle successful registration', async () => {
			const userData = {
				username: 'newuser',
				email: 'new@example.com',
				password: 'password123'
			};

			const mockResponse = {
				data: {
					code: 200,
					data: {
						user: mockUser,
						token: 'test-access-token'
					}
				}
			};

			(api.post as any).mockResolvedValueOnce(mockResponse);

			await auth.register(userData);

			expect(api.post).toHaveBeenCalledWith('/auth/register', userData);
			expect(auth.state.isAuthenticated).toBe(true);
			expect(auth.state.user).toEqual(mockUser);
		});

		it('should handle registration failure', async () => {
			const userData = {
				username: 'newuser',
				email: 'new@example.com',
				password: 'password123'
			};

			(api.post as any).mockRejectedValueOnce(new Error('Registration failed'));

			await expect(auth.register(userData)).rejects.toThrow('Registration failed');
			expect(auth.state.isAuthenticated).toBe(false);
			expect(auth.state.user).toBeNull();
		});
	});

	describe('resetPassword', () => {
		it('should handle successful password reset request', async () => {
			const email = 'test@example.com';
			(api.post as any).mockResolvedValueOnce({});

			await auth.resetPassword(email);

			expect(api.post).toHaveBeenCalledWith('/auth/reset-password', { email });
		});

		it('should handle password reset request failure', async () => {
			const email = 'test@example.com';
			(api.post as any).mockRejectedValueOnce(new Error('Reset failed'));

			await expect(auth.resetPassword(email)).rejects.toThrow('Reset failed');
		});
	});

	describe('updatePassword', () => {
		it('should handle successful password update', async () => {
			const data = {
				oldPassword: 'old-password',
				newPassword: 'new-password'
			};
			(api.post as any).mockResolvedValueOnce({});

			await auth.updatePassword(data);

			expect(api.post).toHaveBeenCalledWith('/auth/update-password', data);
		});

		it('should handle password update failure', async () => {
			const data = {
				oldPassword: 'old-password',
				newPassword: 'new-password'
			};
			(api.post as any).mockRejectedValueOnce(new Error('Update failed'));

			await expect(auth.updatePassword(data)).rejects.toThrow('Update failed');
		});
	});
});
