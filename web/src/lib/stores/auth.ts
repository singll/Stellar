import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import type { User, AuthState } from '$lib/types/auth';
import { authApi } from '$lib/api/auth';
import { notifications } from './notifications';
import { writable, derived } from 'svelte/store';

const STORAGE_KEY = 'auth_state';

// 防止无限循环的标志
let isLoggingOut = false;

const initialState: AuthState = {
	user: null,
	token: null,
	isAuthenticated: false
};

function getInitialState(): AuthState {
	if (!browser) {
		return initialState;
	}
	const storedState = localStorage.getItem(STORAGE_KEY);
	if (storedState) {
		try {
			const parsedState = JSON.parse(storedState);
			// 检查token是否存在且有效
			const hasValidToken = !!parsedState.token && typeof parsedState.token === 'string' && parsedState.token.length > 0;
			return { 
				...initialState, 
				...parsedState, 
				isAuthenticated: hasValidToken 
			};
		} catch {
			localStorage.removeItem(STORAGE_KEY);
		}
	}
	return initialState;
}

const store = writable<AuthState>(getInitialState());

// 监听状态变化并保存到 localStorage
if (browser) {
	store.subscribe((state) => {
		const stateToStore = {
			user: state.user,
			token: state.token
		};
		localStorage.setItem(STORAGE_KEY, JSON.stringify(stateToStore));
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
	async login(credentials: { username?: string; email?: string; password: string }) {
		try {
			const response = await authApi.login(credentials);
			if (response.code === 200 && response.data) {
				store.set({
					...initialState,
					isAuthenticated: true,
					token: response.data.token,
					user: response.data.user
				});
				if (browser) {
					localStorage.setItem(
						STORAGE_KEY,
						JSON.stringify({
							token: response.data.token,
							user: response.data.user
						})
					);
				}
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
			store.set(initialState);
			if (browser) {
				localStorage.removeItem(STORAGE_KEY);
				localStorage.removeItem('token'); // 确保清理登录时设置的token
				if (window.location.pathname !== '/login') {
					goto('/login');
				}
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
				store.set({
					...initialState,
					token: response.data.token,
					user: response.data.user,
					isAuthenticated: true
				});
				return response.data;
			}
			throw new Error('Token刷新失败');
		} catch (error) {
			console.error('Token刷新失败:', error);
			store.set(initialState);
			if (browser) {
				localStorage.removeItem(STORAGE_KEY);
				if (window.location.pathname !== '/login') {
					goto('/login');
				}
			}
			notifications.add({
				type: 'error',
				message: '会话已过期，请重新登录'
			});
			throw error;
		}
	},

	async verifySession() {
		try {
			const response = await authApi.verifySession();
			if (response.code === 200 && response.valid) {
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
				return true;
			}
			return false;
		} catch (error) {
			console.error('会话验证失败:', error);
			return false;
		}
	},
	setCurrentUser(user: User) {
		store.update((state) => ({ ...state, user }));
	},
	async register(userData: any) {
		try {
			const response = await authApi.register(userData);
			if (response.code === 200 && response.data) {
				store.set({
					...initialState,
					token: response.data.token,
					user: response.data.user,
					isAuthenticated: true
				});
				if (browser) {
					localStorage.setItem(
						STORAGE_KEY,
						JSON.stringify({
							token: response.data.token,
							user: response.data.user
						})
					);
				}
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
