import type { Theme, ThemeValidationError, ColorSet, ThemeColors, CreateThemeRequest, UpdateThemeRequest, ColorMode } from '$lib/types/theme';
import { nanoid } from 'nanoid';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * 合并 Tailwind 类名
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * 验证颜色值是否合法
 */
export function isValidColor(color: string): boolean {
  const colorRegex = /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$|^rgb\(\d{1,3},\s*\d{1,3},\s*\d{1,3}\)$|^rgba\(\d{1,3},\s*\d{1,3},\s*\d{1,3},\s*[0-1](\.\d+)?\)$/;
  return colorRegex.test(color);
}

/**
 * 验证颜色集合
 */
function validateColorSet(colorSet: Partial<ColorSet>): ThemeValidationError[] {
  const errors: ThemeValidationError[] = [];
  
  for (const [key, value] of Object.entries(colorSet)) {
    if (!isValidColor(value)) {
      errors.push({
        field: key,
        message: `Invalid color value: ${value}`
      });
    }
  }
  
  return errors;
}

/**
 * 验证主题颜色
 */
function validateThemeColors(colors: Partial<ThemeColors>): ThemeValidationError[] {
  const errors: ThemeValidationError[] = [];
  const requiredColors = [
    'primary',
    'secondary',
    'accent',
    'background',
    'foreground',
    'muted',
    'mutedForeground',
    'border',
    'input',
    'ring'
  ] as const;

  // 检查必需的颜色
  for (const color of requiredColors) {
    if (!colors[color]) {
      errors.push({
        field: `colors.${color}`,
        message: `Missing required color: ${color}`
      });
    }
  }

  // 验证所有颜色值
  for (const [key, value] of Object.entries(colors)) {
    if (value && !isValidColor(value)) {
      errors.push({
        field: `colors.${key}`,
        message: `Invalid color value: ${value}`
      });
    }
  }

  return errors;
}

/**
 * 验证主题配置
 */
export function validateTheme(theme: Theme): ThemeValidationError[] {
  const errors: ThemeValidationError[] = [];

  // 验证基本属性
  if (!theme.name?.trim()) {
    errors.push({
      field: 'name',
      message: 'Theme name is required'
    });
  }

  // 验证颜色
  if (!theme.colors) {
    errors.push({
      field: 'colors',
      message: 'Theme colors are required'
    });
  } else {
    // 验证亮色主题
    if (theme.colors.light) {
      const lightErrors = validateThemeColors(theme.colors.light);
      errors.push(...lightErrors.map(error => ({
        field: `colors.light.${error.field}`,
        message: error.message
      })));
    } else {
      errors.push({
        field: 'colors.light',
        message: 'Light theme colors are required'
      });
    }

    // 验证暗色主题
    if (theme.colors.dark) {
      const darkErrors = validateThemeColors(theme.colors.dark);
      errors.push(...darkErrors.map(error => ({
        field: `colors.dark.${error.field}`,
        message: error.message
      })));
    } else {
      errors.push({
        field: 'colors.dark',
        message: 'Dark theme colors are required'
      });
    }
  }

  return errors;
}

/**
 * 创建新主题
 */
export function createTheme(request: CreateThemeRequest): Theme {
  const now = new Date().toISOString();
  
  const theme: Theme = {
    id: nanoid(),
    name: request.name,
    description: request.description || '',
    colors: {
      light: {
        primary: '#0066cc',
        secondary: '#666666',
        accent: '#ff4081',
        background: '#ffffff',
        foreground: '#000000',
        muted: '#f5f5f5',
        mutedForeground: '#666666',
        border: '#e0e0e0',
        input: '#ffffff',
        ring: '#0066cc',
        destructive: '#dc2626',
        destructiveForeground: '#ffffff',
        success: '#22c55e',
        successForeground: '#ffffff',
        warning: '#f59e0b',
        warningForeground: '#ffffff',
        info: '#3b82f6',
        infoForeground: '#ffffff',
        ...request.colors.light
      },
      dark: {
        primary: '#60a5fa',
        secondary: '#9ca3af',
        accent: '#f472b6',
        background: '#1a1a1a',
        foreground: '#ffffff',
        muted: '#374151',
        mutedForeground: '#9ca3af',
        border: '#374151',
        input: '#1a1a1a',
        ring: '#60a5fa',
        destructive: '#ef4444',
        destructiveForeground: '#ffffff',
        success: '#34d399',
        successForeground: '#ffffff',
        warning: '#fbbf24',
        warningForeground: '#ffffff',
        info: '#60a5fa',
        infoForeground: '#ffffff',
        ...request.colors.dark
      }
    },
    metadata: {
      author: request.author,
      version: '1.0.0',
      createdAt: now,
      updatedAt: now
    }
  };

  return theme;
}

/**
 * 更新主题
 */
export function updateTheme(theme: Theme, request: UpdateThemeRequest): Theme {
  return {
    ...theme,
    name: request.name ?? theme.name,
    description: request.description ?? theme.description,
    colors: {
      light: {
        ...theme.colors.light,
        ...request.colors?.light
      },
      dark: {
        ...theme.colors.dark,
        ...request.colors?.dark
      }
    },
    metadata: {
      ...theme.metadata,
      updatedAt: new Date().toISOString()
    }
  };
}

/**
 * 生成CSS变量
 */
export function generateCssVariables(theme: Theme, mode: ColorMode): string {
  const colors = mode === 'light' ? theme.colors.light : theme.colors.dark;
  let css = ':root {\n';
  
  for (const [key, value] of Object.entries(colors)) {
    css += `  --${key}: ${value};\n`;
  }
  
  css += '}';
  return css;
}

/**
 * 应用主题
 */
export function applyTheme(theme: Theme, mode: ColorMode): void {
  const style = document.createElement('style');
  style.textContent = generateCssVariables(theme, mode);
  
  // 移除旧的主题样式
  const oldStyle = document.getElementById('theme-variables');
  if (oldStyle) {
    oldStyle.remove();
  }
  
  // 添加新的主题样式
  style.id = 'theme-variables';
  document.head.appendChild(style);
  
  // 更新 data-theme 属性
  document.documentElement.setAttribute('data-theme', mode);
}

/**
 * 导出主题为JSON
 */
export function exportTheme(theme: Theme): string {
  return JSON.stringify(theme, null, 2);
}

/**
 * 导入主题
 */
export function importTheme(json: string): Theme {
  try {
    const theme = JSON.parse(json) as Theme;
    const errors = validateTheme(theme);
    if (errors.length > 0) {
      throw new Error(`Invalid theme: ${errors.map(e => e.message).join(', ')}`);
    }
    return theme;
  } catch (error: unknown) {
    if (error instanceof Error) {
      throw new Error(`Failed to import theme: ${error.message}`);
    }
    throw new Error('Failed to import theme: Unknown error');
  }
} 