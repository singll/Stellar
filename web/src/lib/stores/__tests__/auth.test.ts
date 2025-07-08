import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import {
  isAuthenticated,
  currentUser,
  isLoading,
  login,
  logout,
  register,
  resetPassword,
  updatePassword
} from '../auth';
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
    isAuthenticated.set(false);
    currentUser.set(null);
    isLoading.set(false);
  });

  describe('login', () => {
    it('should handle successful login', async () => {
      const credentials = {
        username: 'testuser',
        password: 'password123'
      };

      const mockResponse = {
        data: {
          user: mockUser,
          access_token: 'test-access-token',
          refresh_token: 'test-refresh-token',
          expires_in: 3600
        }
      };

      (api.post as any).mockResolvedValueOnce(mockResponse);

      await login(credentials);

      expect(api.post).toHaveBeenCalledWith('/auth/login', credentials);
      expect(tokenManager.setTokens).toHaveBeenCalledWith({
        accessToken: mockResponse.data.access_token,
        refreshToken: mockResponse.data.refresh_token,
        expiresIn: mockResponse.data.expires_in
      });
      expect(get(currentUser)).toEqual(mockUser);
      expect(get(isAuthenticated)).toBe(true);
      expect(get(isLoading)).toBe(false);
    });

    it('should handle login failure', async () => {
      const credentials = {
        username: 'testuser',
        password: 'wrong-password'
      };

      (api.post as any).mockRejectedValueOnce(new Error('Login failed'));

      await expect(login(credentials)).rejects.toThrow('Login failed');
      expect(get(currentUser)).toBeNull();
      expect(get(isAuthenticated)).toBe(false);
      expect(get(isLoading)).toBe(false);
    });
  });

  describe('logout', () => {
    it('should handle successful logout', async () => {
      (api.post as any).mockResolvedValueOnce({});

      await logout();

      expect(api.post).toHaveBeenCalledWith('/auth/logout');
      expect(tokenManager.clearTokens).toHaveBeenCalled();
      expect(get(currentUser)).toBeNull();
      expect(get(isAuthenticated)).toBe(false);
    });

    it('should handle logout failure gracefully', async () => {
      (api.post as any).mockRejectedValueOnce(new Error('Logout failed'));

      await logout();

      expect(tokenManager.clearTokens).toHaveBeenCalled();
      expect(get(currentUser)).toBeNull();
      expect(get(isAuthenticated)).toBe(false);
    });
  });

  describe('register', () => {
    it('should handle successful registration', async () => {
      const userData = {
        username: 'newuser',
        email: 'new@example.com',
        password: 'password123'
      };

      (api.post as any).mockResolvedValueOnce({});
      (api.post as any).mockResolvedValueOnce({
        data: {
          user: mockUser,
          access_token: 'test-access-token',
          refresh_token: 'test-refresh-token',
          expires_in: 3600
        }
      });

      await register(userData);

      expect(api.post).toHaveBeenCalledWith('/auth/register', userData);
      expect(get(currentUser)).toEqual(mockUser);
      expect(get(isAuthenticated)).toBe(true);
      expect(get(isLoading)).toBe(false);
    });

    it('should handle registration failure', async () => {
      const userData = {
        username: 'newuser',
        email: 'new@example.com',
        password: 'password123'
      };

      (api.post as any).mockRejectedValueOnce(new Error('Registration failed'));

      await expect(register(userData)).rejects.toThrow('Registration failed');
      expect(get(currentUser)).toBeNull();
      expect(get(isAuthenticated)).toBe(false);
      expect(get(isLoading)).toBe(false);
    });
  });

  describe('resetPassword', () => {
    it('should handle successful password reset request', async () => {
      const email = 'test@example.com';
      (api.post as any).mockResolvedValueOnce({});

      await resetPassword(email);

      expect(api.post).toHaveBeenCalledWith('/auth/reset-password', { email });
      expect(get(isLoading)).toBe(false);
    });

    it('should handle password reset request failure', async () => {
      const email = 'test@example.com';
      (api.post as any).mockRejectedValueOnce(new Error('Reset failed'));

      await expect(resetPassword(email)).rejects.toThrow('Reset failed');
      expect(get(isLoading)).toBe(false);
    });
  });

  describe('updatePassword', () => {
    it('should handle successful password update', async () => {
      const currentPassword = 'old-password';
      const newPassword = 'new-password';
      (api.post as any).mockResolvedValueOnce({});

      await updatePassword(currentPassword, newPassword);

      expect(api.post).toHaveBeenCalledWith('/auth/update-password', {
        current_password: currentPassword,
        new_password: newPassword
      });
      expect(get(isLoading)).toBe(false);
    });

    it('should handle password update failure', async () => {
      const currentPassword = 'old-password';
      const newPassword = 'new-password';
      (api.post as any).mockRejectedValueOnce(new Error('Update failed'));

      await expect(
        updatePassword(currentPassword, newPassword)
      ).rejects.toThrow('Update failed');
      expect(get(isLoading)).toBe(false);
    });
  });
}); 