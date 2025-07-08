import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import type { User, AuthState } from '$lib/types/auth';
import { authApi } from '$lib/api/auth';
import { notifications } from './notifications';
import { writable } from 'svelte/store';

const STORAGE_KEY = 'auth_state';

const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
};

function getInitialState(): AuthState {
  if (!browser) {
    return initialState;
  }
  const storedState = localStorage.getItem(STORAGE_KEY);
  if (storedState) {
    try {
      const parsedState = JSON.parse(storedState);
      return { ...initialState, ...parsedState, isAuthenticated: !!parsedState.token };
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
      token: state.token,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(stateToStore));
  });
}

export const auth = {
  subscribe: store.subscribe,
  get state() {
    let currentState: AuthState;
    store.subscribe(state => { currentState = state; })();
    return currentState!;
  },
  login(data: { token: string; user: User }) {
    store.set({
      ...initialState,
      isAuthenticated: true,
      token: data.token,
      user: data.user,
    });
  },
  async logout() {
    try {
      await authApi.logout();
    } catch (error) {
      console.error("API logout call failed, forcing client logout:", error);
    } finally {
      store.set(initialState);
      if (browser) {
        localStorage.removeItem(STORAGE_KEY);
        if (window.location.pathname !== '/login') {
          goto('/login');
        }
      }
      notifications.add({
        type: 'info',
        message: '您已成功退出登录'
      });
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
          isAuthenticated: true,
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
  setCurrentUser(user: User) {
    store.update(state => ({ ...state, user }));
  }
};

// End of file. No other exports or functions should exist. 