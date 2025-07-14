import api from './axios-config';
import type { User } from '$lib/types/auth';

export interface LoginRequest {
	username?: string;
	email?: string;
	password: string;
}

export interface LoginResponse {
	code: number;
	message: string;
	data?: {
		token: string;
		user: User;
	};
}

export interface RegisterRequest {
	username: string;
	email: string;
	password: string;
}

export interface RegisterResponse {
	code: number;
	message: string;
	data?: {
		token: string;
		user: User;
	};
}

export interface RefreshTokenResponse {
	code: number;
	message: string;
	data?: {
		token: string;
		user: User;
	};
}

export class APIError extends Error {
	code: number;
	details?: string;
	constructor(code: number, message: string, details?: string) {
		super(message);
		this.code = code;
		this.details = details;
		this.name = 'APIError';
	}
}

export const authApi = {
	/**
	 * 用户登录
	 * @param credentials { email: string, password: string }
	 * @returns { code, message, data: { token, user } }
	 */
	async login(credentials: LoginRequest): Promise<LoginResponse> {
		try {
			const response = await api.post('/auth/login', credentials);
			return response.data;
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 用户注册
	 * @param userData { email, password, name }
	 * @returns { code, message, data: { token, user } }
	 */
	async register(userData: RegisterRequest): Promise<RegisterResponse> {
		try {
			const response = await api.post('/auth/register', userData);
			return response.data;
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 用户登出
	 * @returns void
	 */
	async logout(): Promise<void> {
		try {
			// 对于logout请求，我们可能需要发送token，但如果失败也无关紧要
			await api.post('/auth/logout');
		} catch (error: any) {
			// logout失败是正常的，可能token已经过期
			// 不抛出错误，让客户端能正常清理状态
			console.warn('Logout API call failed:', error.message);
		}
	},

	/**
	 * 刷新Token
	 * @returns { code, message, data: { token, user } }
	 */
	async refreshToken(): Promise<RefreshTokenResponse> {
		try {
			const response = await api.post('/auth/refresh');
			return response.data;
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 获取当前用户信息
	 * @returns User
	 */
	async getCurrentUser(): Promise<User> {
		try {
			const response = await api.get('/auth/me');
			return response.data.data;
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 更新用户信息
	 * @param userData Partial<User>
	 * @returns User
	 */
	async updateProfile(userData: Partial<User>): Promise<User> {
		try {
			const response = await api.put('/auth/profile', userData);
			return response.data.data;
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 修改密码
	 * @param data { oldPassword, newPassword }
	 * @returns void
	 */
	async changePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
		try {
			await api.put('/auth/password', data);
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 重置密码
	 * @param email string
	 * @returns void
	 */
	async resetPassword(email: string): Promise<void> {
		try {
			await api.post('/auth/reset-password', { email });
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 验证重置Token
	 * @param token string
	 * @param newPassword string
	 * @returns void
	 */
	async verifyResetToken(token: string, newPassword: string): Promise<void> {
		try {
			await api.post('/auth/verify-reset', { token, newPassword });
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	},

	/**
	 * 更新密码
	 * @param data { oldPassword: string, newPassword: string }
	 * @returns void
	 */
	async updatePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
		try {
			await api.post('/auth/update-password', data);
		} catch (error: any) {
			if (error.response?.data) {
				const { code, message, details } = error.response.data;
				throw new APIError(code, message, details);
			}
			throw error;
		}
	}
};
