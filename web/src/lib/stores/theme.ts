import { writable, derived } from 'svelte/store';
import type { Theme, ColorMode } from '$lib/types/theme';
import { getDefaultTheme, generateCssVariables, applyTheme } from '$lib/utils/theme-utils';
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

function createThemeStore() {
  const defaultTheme = getDefaultTheme();
  const initialState: ThemeState = {
    currentTheme: defaultTheme,
    availableThemes: [defaultTheme],
    mode: 'light'
  };

  const { subscribe, set, update } = writable<ThemeState>(initialState);

  return {
    subscribe,

    // 设置当前主题
    setTheme: (theme: Theme) => {
      update(state => {
        applyTheme(theme, state.mode);
        return { ...state, currentTheme: theme };
      });
    },

    // 添加新主题
    addTheme: (theme: Theme) => {
      update(state => ({
        ...state,
        availableThemes: [...state.availableThemes, theme]
      }));
    },

    // 更新主题
    updateTheme: (theme: Theme) => {
      update(state => {
        const index = state.availableThemes.findIndex(t => t.id === theme.id);
        if (index === -1) return state;

        const newThemes = [...state.availableThemes];
        newThemes[index] = theme;

        // 如果更新的是当前主题，也需要更新当前主题
        const newState = {
          ...state,
          availableThemes: newThemes
        };

        if (state.currentTheme.id === theme.id) {
          newState.currentTheme = theme;
          applyTheme(theme, state.mode);
        }

        return newState;
      });
    },

    // 删除主题
    deleteTheme: (themeId: string) => {
      update(state => {
        const newThemes = state.availableThemes.filter(t => t.id !== themeId);
        
        // 如果删除的是当前主题，切换到默认主题
        const newState = {
          ...state,
          availableThemes: newThemes
        };

        if (state.currentTheme.id === themeId) {
          const defaultTheme = getDefaultTheme();
          newState.currentTheme = defaultTheme;
          applyTheme(defaultTheme, state.mode);
        }

        return newState;
      });
    },

    // 切换主题模式
    toggleMode: () => {
      update(state => {
        const newMode = state.mode === 'light' ? 'dark' : 'light';
        applyTheme(state.currentTheme, newMode);
        return { ...state, mode: newMode };
      });
    },

    // 设置主题模式
    setMode: (mode: ColorMode) => {
      update(state => {
        applyTheme(state.currentTheme, mode);
        return { ...state, mode };
      });
    },

    // 重置为默认主题
    reset: () => {
      const defaultTheme = getDefaultTheme();
      update(state => {
        applyTheme(defaultTheme, state.mode);
        return {
          ...state,
          currentTheme: defaultTheme,
          availableThemes: [defaultTheme]
        };
      });
    }
  };
}

// 创建主题存储
export const themeStore = createThemeStore();

// Subscribe to changes and persist to localStorage
if (browser) {
  themeStore.subscribe(state => {
    localStorage.setItem('stellar-theme', JSON.stringify(state.currentTheme));
    localStorage.setItem('stellar-theme-mode', state.mode);
    localStorage.setItem('stellar-available-themes', JSON.stringify(state.availableThemes));
    
    // Apply theme to document
    applyTheme(state.currentTheme, state.mode);
  });
}

// 派生存储：主题列表
export const themes = derived(themeStore, $store => $store.availableThemes);

// 派生存储：当前主题
export const currentTheme = derived(themeStore, $store => $store.currentTheme);

// 派生存储：暗色模式状态
export const darkMode = derived(themeStore, $store => $store.mode === 'dark'); 