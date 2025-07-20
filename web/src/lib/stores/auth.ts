import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import type { User, AuthState } from '$lib/types/auth';
import { authApi } from '$lib/api/auth';
import { notifications } from './notifications';
import { writable, derived } from 'svelte/store';

const STORAGE_KEY = 'auth_state';

// 防止无限循环的标志
let isLoggingOut = false;
let isVerifyingSession = false;
let isInitialized = false;

const initialState: AuthState = {
	user: null,
	token: null,
	isAuthenticated: false
};

function getInitialState(): AuthState {
	if (!browser) {
		return initialState;
	}
	
	try {
		const storedState = localStorage.getItem(STORAGE_KEY);
		if (storedState) {
			const parsedState = JSON.parse(storedState);
			// 检查token是否存在且有效
			const hasValidToken = !!parsedState.token && typeof parsedState.token === 'string' && parsedState.token.length > 0;
			const hasValidUser = !!parsedState.user && typeof parsedState.user === 'object';
			
			if (hasValidToken && hasValidUser) {
				console.log('从localStorage恢复认证状态');
				return { 
					...initialState, 
					...parsedState, 
					isAuthenticated: true 
				};
			}
		}
	} catch (error) {
		console.error('解析存储的认证状态失败:', error);
		localStorage.removeItem(STORAGE_KEY);
	}
	
	return initialState;
}

const store = writable<AuthState>(getInitialState());

// 监听状态变化并保存到 localStorage（只在浏览器环境下）
if (browser) {
	store.subscribe((state) => {
		// 防止在初始化过程中触发保存
		if (!isInitialized) return;
		
		const stateToStore = {
			user: state.user,
			token: state.token
		};
		
		if (state.isAuthenticated && state.token && state.user) {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(stateToStore));
		} else {
			localStorage.removeItem(STORAGE_KEY);
		}
	});
}

export const auth = {
	subscribe: store.subscribe,
	get state() {
		let currentState: AuthState = initialState;
		const unsubscribe = store.subscribe((state) => {
			currentState = state;
		});
		unsubscribe();
		return currentState;
	},
	
	/**
	 * 初始化认证状态（在应用启动时调用）
	 */
	async initialize() {
		if (!browser || isInitialized) return;
		
		console.log('初始化认证状态...');
		
		const { token, user } = auth.state;
		if (token && user) {
			console.log('发现存储的认证状态，验证会话...');
			try {
				// 验证会话状态
				const isValid = await this.verifySession();
				if (!isValid) {
					console.log('会话验证失败，清理状态');
					// 会话无效，清理状态
					this.clearState();
				} else {
					console.log('会话验证成功');
				}
			} catch (error) {
				console.error('会话验证过程中发生错误:', error);
				// 验证过程中发生错误，清理状态
				this.clearState();
			}
		} else {
			console.log('未发现存储的认证状态');
		}
		
		isInitialized = true;
	},
	
	/**
	 * 清理认证状态
	 */
	clearState() {
		console.log('清理认证状态');
		store.set(initialState);
		if (browser) {
			localStorage.removeItem(STORAGE_KEY);
		}
	},
	
	/**
	 * 设置认证状态
	 */
	setAuthState(token: string, user: User) {
		console.log('设置认证状态:', { token: token.substring(0, 20) + '...', user: user.username });
		store.set({
			...initialState,
			isAuthenticated: true,
			token,
			user
		});
	},
	
	async login(credentials: { username?: string; email?: string; password: string }) {
		try {
			const response = await authApi.login(credentials);
			if (response.code === 200 && response.data) {
				this.setAuthState(response.data.token, response.data.user);
				return response.data;
			}
			throw new Error('登录失败');
		} catch (error) {
			console.error('登录失败:', error);
			throw error;
		}
	},
	
	async logout() {
		// 防止重复调用
		if (isLoggingOut) return;
		isLoggingOut = true;

		try {
			// 尝试调用logout API，但不让错误阻止客户端清理
			await authApi.logout();
		} catch (error) {
			// logout API失败是常见的（token过期等），不应该阻止客户端清理
			console.warn('Logout API call failed (this is normal):', error);
		} finally {
			// 无论API调用成功与否，都清理客户端状态
			this.clearState();
			if (browser && window.location.pathname !== '/login') {
				goto('/login');
			}

			// 重置标志
			setTimeout(() => {
				isLoggingOut = false;
			}, 100);
		}
	},
	
	async refreshToken() {
		try {
			const response = await authApi.refreshToken();
			if (response.code === 200 && response.data) {
				this.setAuthState(response.data.token, response.data.user);
				return response.data;
			}
			throw new Error('Token刷新失败');
		} catch (error) {
			console.error('Token刷新失败:', error);
			this.clearState();
			if (browser && window.location.pathname !== '/login') {
				goto('/login');
			}
			notifications.add({
				type: 'error',
				message: '会话已过期，请重新登录'
			});
			throw error;
		}
	},

	async verifySession() {
		// 防止重复验证
		if (isVerifyingSession) return false;
		isVerifyingSession = true;
		
		try {
			const { token } = auth.state;
			if (!token) {
				console.log('验证会话：无token');
				return false;
			}
			
			console.log('验证会话...');
			const response = await authApi.verifySession();
			if (response.code === 200 && response.valid) {
				console.log('会话验证成功');
				// 会话有效，更新用户信息
				if (response.user && auth.state.user) {
					store.update((state) => ({
						...state,
						user: {
							...state.user!,
							username: response.user!.username,
							roles: response.user!.roles,
						},
					}));
				}
				
				// 检查是否需要刷新会话
				if (response.session_status?.needs_refresh) {
					console.log('会话需要刷新');
					await this.refreshSession();
				}
				
				return true;
			}
			console.log('会话验证失败');
			return false;
		} catch (error) {
			console.error('会话验证失败:', error);
			return false;
		} finally {
			isVerifyingSession = false;
		}
	},
	
	async refreshSession() {
		try {
			const response = await authApi.refreshSession();
			if (response.code === 200 && response.data) {
				console.log('会话刷新成功:', response.data);
				return true;
			}
			return false;
		} catch (error) {
			console.error('会话刷新失败:', error);
			return false;
		}
	},
	
	async getSessionStatus() {
		try {
			const response = await authApi.getSessionStatus();
			return response.data;
		} catch (error) {
			console.error('获取会话状态失败:', error);
			return null;
		}
	},
	
	setCurrentUser(user: User) {
		store.update((state) => ({ ...state, user }));
	},
	
	async register(userData: any) {
		try {
			const response = await authApi.register(userData);
			if (response.code === 200 && response.data) {
				this.setAuthState(response.data.token, response.data.user);
				return response.data;
			}
			throw new Error('注册失败');
		} catch (error) {
			console.error('注册失败:', error);
			throw error;
		}
	},
	
	async resetPassword(email: string) {
		try {
			await authApi.resetPassword(email);
			notifications.add({
				type: 'success',
				message: '密码重置邮件已发送'
			});
		} catch (error) {
			console.error('密码重置失败:', error);
			throw error;
		}
	},
	
	async updatePassword(data: any) {
		try {
			await authApi.updatePassword(data);
			notifications.add({
				type: 'success',
				message: '密码更新成功'
			});
		} catch (error) {
			console.error('密码更新失败:', error);
			throw error;
		}
	}
};

// 为了向后兼容测试文件，导出单独的函数和状态
export const isAuthenticated = derived(store, (state) => state.isAuthenticated);
export const currentUser = derived(store, (state) => state.user);
export const isLoading = writable(false);

export const login = auth.login.bind(auth);
export const logout = auth.logout.bind(auth);
export const register = auth.register.bind(auth);
export const resetPassword = auth.resetPassword.bind(auth);
export const updatePassword = auth.updatePassword.bind(auth);

// End of file. No other exports or functions should exist.
