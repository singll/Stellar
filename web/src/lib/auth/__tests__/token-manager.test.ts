/// <reference types="vitest" />
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { TokenManager } from '../token-manager';
import axios from 'axios';
import type { Mock } from 'vitest';

vi.mock('axios');
vi.mock('$app/navigation', () => ({
	goto: vi.fn()
}));

describe('TokenManager', () => {
	let tokenManager: TokenManager;
	const mockStorage = new Map<string, string>();

	beforeEach(() => {
		// Mock localStorage methods
		vi.spyOn(window.localStorage, 'getItem').mockImplementation(
			(key: string) => mockStorage.get(key) || null
		);
		vi.spyOn(window.localStorage, 'setItem').mockImplementation((key: string, value: string) => {
			mockStorage.set(key, value);
		});
		vi.spyOn(window.localStorage, 'removeItem').mockImplementation((key: string) => {
			mockStorage.delete(key);
		});
		vi.spyOn(window.localStorage, 'clear').mockImplementation(() => {
			mockStorage.clear();
		});

		// Reset TokenManager instance
		(TokenManager as any)['instance'] = undefined;
		tokenManager = TokenManager.getInstance();
	});

	afterEach(() => {
		vi.clearAllMocks();
		mockStorage.clear();
	});

	it('should be a singleton', () => {
		const instance1 = TokenManager.getInstance();
		const instance2 = TokenManager.getInstance();
		expect(instance1).toBe(instance2);
	});

	it('should save and load tokens', () => {
		const tokens = {
			accessToken: 'test-access-token',
			refreshToken: 'test-refresh-token',
			expiresIn: 3600
		};

		tokenManager.setTokens(tokens);

		expect(window.localStorage.setItem).toHaveBeenCalledWith(
			'stellar_auth_tokens',
			JSON.stringify(tokens)
		);

		expect(tokenManager.getAccessToken()).toBe(tokens.accessToken);
	});

	it('should clear tokens', () => {
		tokenManager.clearTokens();
		expect(window.localStorage.removeItem).toHaveBeenCalledWith('stellar_auth_tokens');
		expect(tokenManager.getAccessToken()).toBeNull();
	});

	it('should refresh tokens', async () => {
		const mockResponse = {
			data: {
				access_token: 'new-access-token',
				refresh_token: 'new-refresh-token',
				expires_in: 3600
			}
		};

		(axios.post as Mock).mockResolvedValueOnce(mockResponse);

		const initialTokens = {
			accessToken: 'old-access-token',
			refreshToken: 'old-refresh-token',
			expiresIn: 3600
		};

		tokenManager.setTokens(initialTokens);
		await tokenManager.refreshTokens();

		expect(axios.post).toHaveBeenCalledWith('/api/v1/auth/refresh', {
			refresh_token: initialTokens.refreshToken
		});

		expect(tokenManager.getAccessToken()).toBe(mockResponse.data.access_token);
	});

	it('should handle refresh token failure', async () => {
		(axios.post as Mock).mockRejectedValueOnce(new Error('Refresh failed'));

		const initialTokens = {
			accessToken: 'old-access-token',
			refreshToken: 'old-refresh-token',
			expiresIn: 3600
		};

		tokenManager.setTokens(initialTokens);
		await expect(tokenManager.refreshTokens()).rejects.toThrow();
		expect(tokenManager.getAccessToken()).toBeNull();
	});
});
