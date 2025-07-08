export interface User {
	id: string; // 用户唯一标识
	username: string; // 用户名
	email: string; // 邮箱
	roles: string[]; // 角色数组
	created: string; // 创建时间
	lastLogin: string; // 最后登录时间
}

export interface LoginCredentials {
	username: string;
	password: string;
}

export interface RegisterData {
	username: string;
	email: string;
	password: string;
	confirmPassword: string;
	agreeToTerms: boolean;
}

export interface ResetPasswordData {
	email: string;
}

export interface NewPasswordData {
	token: string;
	password: string;
	confirmPassword: string;
}

export interface AuthResponse {
	token: string;
	user: User;
}

export interface AuthError {
	code: number;
	message: string;
	details?: string;
}

export interface AuthState {
	user: User | null;
	token: string | null;
	isAuthenticated: boolean;
}

export interface APIResponse<T> {
	code: number;
	message: string;
	data: T;
}
