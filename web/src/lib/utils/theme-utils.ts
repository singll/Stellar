import type { Theme, ThemeValidationResult, CreateThemeRequest, ThemeColors, ColorSet } from '$lib/types/theme';
import { v4 as uuidv4 } from 'uuid';

/**
 * 验证主题配置是否有效
 */
export function validateTheme(theme: Theme): ThemeValidationResult {
  const errors: string[] = [];

  // 基本属性验证
  if (!theme.name) errors.push('主题名称不能为空');
  if (!theme.colors) errors.push('主题颜色不能为空');
  if (!theme.colors.light) errors.push('缺少亮色主题配置');
  if (!theme.colors.dark) errors.push('缺少暗色主题配置');

  // 必需的颜色验证
  const requiredColors = [
    'primary',
    'secondary',
    'accent',
    'success',
    'warning',
    'error',
    'info',
    'background',
    'foreground',
    'border',
    'card',
    'sidebar',
    'header'
  ] as const;

  // 验证亮色主题
  for (const color of requiredColors) {
    if (!theme.colors.light[color]) {
      errors.push(`缺少亮色主题必需的颜色配置: ${color}`);
    }
  }

  // 验证暗色主题
  for (const color of requiredColors) {
    if (!theme.colors.dark[color]) {
      errors.push(`缺少暗色主题必需的颜色配置: ${color}`);
    }
  }

  return {
    isValid: errors.length === 0,
    errors
  };
}

/**
 * 创建新主题
 */
export function createTheme(options: CreateThemeRequest): Theme {
  const now = new Date().toISOString();
  
  return {
    id: uuidv4(),
    name: options.name,
    description: options.description,
    colors: options.colors,
    metadata: {
      ...options.metadata,
      createdAt: now,
      updatedAt: now,
      isBuiltin: false,
      tags: []
    }
  };
}

/**
 * 获取默认主题配置
 */
export function getDefaultTheme(): Theme {
  return {
    id: 'default',
    name: 'Default Theme',
    description: '默认主题',
    colors: {
      light: getDefaultLightColors(),
      dark: getDefaultDarkColors()
    },
    metadata: {
      author: 'Stellar',
      version: '1.0.0',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      isBuiltin: true,
      tags: ['default', 'light']
    }
  };
}

function getDefaultLightColors(): ThemeColors {
  return {
    primary: createColorSet('#3b82f6', '#ffffff'),
    secondary: createColorSet('#6b7280', '#ffffff'),
    accent: createColorSet('#8b5cf6', '#ffffff'),
    success: createColorSet('#22c55e', '#ffffff'),
    warning: createColorSet('#f59e0b', '#ffffff'),
    error: createColorSet('#ef4444', '#ffffff'),
    info: createColorSet('#3b82f6', '#ffffff'),
    background: createColorSet('#ffffff', '#000000'),
    foreground: createColorSet('#000000', '#ffffff'),
    border: createColorSet('#e5e7eb', '#000000'),
    card: createColorSet('#ffffff', '#000000'),
    sidebar: createColorSet('#f8fafc', '#000000'),
    header: createColorSet('#ffffff', '#000000')
  };
}

function getDefaultDarkColors(): ThemeColors {
  return {
    primary: createColorSet('#60a5fa', '#000000'),
    secondary: createColorSet('#9ca3af', '#000000'),
    accent: createColorSet('#a78bfa', '#000000'),
    success: createColorSet('#4ade80', '#000000'),
    warning: createColorSet('#fbbf24', '#000000'),
    error: createColorSet('#f87171', '#000000'),
    info: createColorSet('#60a5fa', '#000000'),
    background: createColorSet('#0f172a', '#ffffff'),
    foreground: createColorSet('#ffffff', '#000000'),
    border: createColorSet('#1e293b', '#ffffff'),
    card: createColorSet('#1e293b', '#ffffff'),
    sidebar: createColorSet('#0f172a', '#ffffff'),
    header: createColorSet('#1e293b', '#ffffff')
  };
}

function createColorSet(value: string, foreground: string): ColorSet {
  return { value, foreground };
}

/**
 * 转换颜色格式
 */
export function parseColor(color: string): ColorSet {
  return {
    value: color,
    foreground: getContrastColor(color)
  };
}

/**
 * 获取对比色
 */
function getContrastColor(hexColor: string): string {
  // 移除#号
  const hex = hexColor.replace('#', '');
  
  // 转换为RGB
  const r = parseInt(hex.substr(0, 2), 16);
  const g = parseInt(hex.substr(2, 2), 16);
  const b = parseInt(hex.substr(4, 2), 16);
  
  // 计算亮度
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;
  
  // 根据亮度返回黑色或白色
  return brightness > 128 ? '#000000' : '#ffffff';
}

/**
 * 序列化主题为JSON字符串
 */
export function serializeTheme(theme: Theme): string {
  return JSON.stringify(theme, null, 2);
}

/**
 * 从JSON字符串解析主题
 */
export function deserializeTheme(json: string): Theme {
  try {
    const theme = JSON.parse(json) as Theme;
    return theme;
  } catch (error) {
    throw new Error('Invalid theme JSON');
  }
}

export function generateCssVariables(colors: ThemeColors): string {
  const variables: string[] = [];

  for (const [key, color] of Object.entries(colors)) {
    variables.push(`--${key}: ${color.value};`);
    variables.push(`--${key}-foreground: ${color.foreground};`);
  }

  return variables.join('\n');
}

export function applyTheme(theme: Theme, mode: 'light' | 'dark'): void {
  const colors = mode === 'light' ? theme.colors.light : theme.colors.dark;
  const root = document.documentElement;
  
  for (const [key, color] of Object.entries(colors)) {
    root.style.setProperty(`--${key}`, color.value);
    root.style.setProperty(`--${key}-foreground`, color.foreground);
  }
} 