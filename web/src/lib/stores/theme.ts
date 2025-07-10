import { writable, derived } from 'svelte/store';
import type { Theme, ColorMode } from '$lib/types/theme';
import { createDefaultTheme, applyThemeToCSS } from '$lib/utils/theme';
import { browser } from '$app/environment';

// 本地存储键
const THEME_STORAGE_KEY = 'stellar:themes';
const CURRENT_THEME_KEY = 'stellar:currentTheme';
const DARK_MODE_KEY = 'stellar:darkMode';

interface ThemeState {
	currentTheme: Theme;
	availableThemes: Theme[];
	mode: ColorMode;
}

function loadFromStorage<T>(key: string, defaultValue: T): T {
	if (!browser) return defaultValue;

	try {
		const stored = localStorage.getItem(key);
		return stored ? JSON.parse(stored) : defaultValue;
	} catch {
		return defaultValue;
	}
}

function saveToStorage<T>(key: string, value: T): void {
	if (!browser) return;

	try {
		localStorage.setItem(key, JSON.stringify(value));
	} catch {
		// 忽略存储错误
	}
}

function createThemeStore() {
	const defaultTheme = createDefaultTheme('Default', 'Default theme');
	const initialState: ThemeState = {
		currentTheme: loadFromStorage(CURRENT_THEME_KEY, defaultTheme),
		availableThemes: loadFromStorage(THEME_STORAGE_KEY, [defaultTheme]),
		mode: loadFromStorage(DARK_MODE_KEY, 'light' as ColorMode)
	};

	const { subscribe, set, update } = writable<ThemeState>(initialState);

	// 应用初始主题
	if (browser) {
		applyThemeToCSS(initialState.currentTheme, initialState.mode);
	}

	return {
		subscribe,

		// 设置当前主题
		setTheme: (theme: Theme) => {
			update((state) => {
				const newState = { ...state, currentTheme: theme };
				if (browser) {
					applyThemeToCSS(theme, state.mode);
					saveToStorage(CURRENT_THEME_KEY, theme);
				}
				return newState;
			});
		},

		// 切换颜色模式
		toggleMode: () => {
			update((state) => {
				const newMode: ColorMode = state.mode === 'light' ? 'dark' : 'light';
				const newState = { ...state, mode: newMode };
				if (browser) {
					applyThemeToCSS(state.currentTheme, newMode);
					saveToStorage(DARK_MODE_KEY, newMode);
				}
				return newState;
			});
		},

		// 设置颜色模式
		setMode: (mode: ColorMode) => {
			update((state) => {
				const newState = { ...state, mode };
				if (browser) {
					applyThemeToCSS(state.currentTheme, mode);
					saveToStorage(DARK_MODE_KEY, mode);
				}
				return newState;
			});
		},

		// 添加新主题
		addTheme: (theme: Theme) => {
			update((state) => {
				const newThemes = [...state.availableThemes, theme];
				const newState = { ...state, availableThemes: newThemes };
				if (browser) {
					saveToStorage(THEME_STORAGE_KEY, newThemes);
				}
				return newState;
			});
		},

		// 更新主题
		updateTheme: (theme: Theme) => {
			update((state) => {
				const index = state.availableThemes.findIndex((t) => t.id === theme.id);
				if (index === -1) return state;

				const newThemes = [...state.availableThemes];
				newThemes[index] = theme;

				const newState = {
					...state,
					availableThemes: newThemes,
					currentTheme: state.currentTheme.id === theme.id ? theme : state.currentTheme
				};

				if (browser) {
					saveToStorage(THEME_STORAGE_KEY, newThemes);
					if (state.currentTheme.id === theme.id) {
						applyThemeToCSS(theme, state.mode);
						saveToStorage(CURRENT_THEME_KEY, theme);
					}
				}

				return newState;
			});
		},

		// 删除主题
		deleteTheme: (themeId: string) => {
			update((state) => {
				const newThemes = state.availableThemes.filter((t) => t.id !== themeId);
				let newCurrentTheme = state.currentTheme;

				// 如果删除的是当前主题，切换到默认主题
				if (state.currentTheme.id === themeId) {
					newCurrentTheme = newThemes[0] || defaultTheme;
				}

				const newState = {
					...state,
					availableThemes: newThemes,
					currentTheme: newCurrentTheme
				};

				if (browser) {
					saveToStorage(THEME_STORAGE_KEY, newThemes);
					if (state.currentTheme.id === themeId) {
						applyThemeToCSS(newCurrentTheme, state.mode);
						saveToStorage(CURRENT_THEME_KEY, newCurrentTheme);
					}
				}

				return newState;
			});
		}
	};
}

export const themeStore = createThemeStore();

// 便捷的 derived stores
export const currentTheme = derived(themeStore, (state) => state.currentTheme);
export const availableThemes = derived(themeStore, (state) => state.availableThemes);
export const themeMode = derived(themeStore, (state) => state.mode);
export const isDarkMode = derived(themeStore, (state) => state.mode === 'dark');

// 主题操作actions
export const themeActions = {
	setTheme: themeStore.setTheme,
	addTheme: themeStore.addTheme,
	updateTheme: themeStore.updateTheme,
	deleteTheme: themeStore.deleteTheme,
	toggleMode: themeStore.toggleMode,
	setMode: themeStore.setMode
};
