import type {
	Theme,
	ThemeValidationResult,
	CreateThemeRequest,
	ThemeColors,
	ColorSet
} from '$lib/types/theme';
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

	// 合并默认颜色和用户提供的颜色
	const defaultLight = getDefaultLightColors();
	const defaultDark = getDefaultDarkColors();

	return {
		id: uuidv4(),
		name: options.name,
		description: options.description,
		colors: {
			light: { ...defaultLight, ...options.colors.light },
			dark: { ...defaultDark, ...options.colors.dark }
		},
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
		primary: '#3b82f6',
		secondary: '#6b7280',
		accent: '#8b5cf6',
		success: '#22c55e',
		warning: '#f59e0b',
		error: '#ef4444',
		info: '#3b82f6',
		background: '#ffffff',
		foreground: '#000000',
		border: '#e5e7eb',
		card: '#ffffff',
		sidebar: '#f8fafc',
		header: '#ffffff',
		muted: '#f9fafb',
		mutedForeground: '#6b7280',
		input: '#f9fafb',
		ring: '#3b82f6',
		destructive: '#ef4444',
		destructiveForeground: '#ffffff',
		successForeground: '#ffffff',
		warningForeground: '#ffffff',
		infoForeground: '#ffffff'
	};
}

function getDefaultDarkColors(): ThemeColors {
	return {
		primary: '#60a5fa',
		secondary: '#9ca3af',
		accent: '#a78bfa',
		success: '#4ade80',
		warning: '#fbbf24',
		error: '#f87171',
		info: '#60a5fa',
		background: '#0f172a',
		foreground: '#ffffff',
		border: '#1e293b',
		card: '#1e293b',
		sidebar: '#0f172a',
		header: '#1e293b',
		muted: '#1e293b',
		mutedForeground: '#9ca3af',
		input: '#1e293b',
		ring: '#60a5fa',
		destructive: '#f87171',
		destructiveForeground: '#000000',
		successForeground: '#000000',
		warningForeground: '#000000',
		infoForeground: '#000000'
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
		variables.push(`--${key}: ${color};`);
	}

	return variables.join('\n');
}

export function applyTheme(theme: Theme, mode: 'light' | 'dark'): void {
	const colors = mode === 'light' ? theme.colors.light : theme.colors.dark;
	const root = document.documentElement;

	for (const [key, color] of Object.entries(colors)) {
		root.style.setProperty(`--${key}`, color);
	}
}
