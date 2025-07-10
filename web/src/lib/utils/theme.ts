import type {
	Theme,
	ThemeValidationError,
	ColorSet,
	ThemeColors,
	CreateThemeRequest,
	UpdateThemeRequest,
	ColorMode,
	ThemeColorName
} from '$lib/types/theme';
import { THEME_COLORS, DEFAULT_THEME_COLORS } from '$lib/types/theme';
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
	const colorRegex =
		/^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$|^rgb\(\d{1,3},\s*\d{1,3},\s*\d{1,3}\)$|^rgba\(\d{1,3},\s*\d{1,3},\s*\d{1,3},\s*[0-1](\.\d+)?\)$/;
	return colorRegex.test(color);
}

/**
 * 验证主题颜色
 */
function validateThemeColors(colors: Partial<ThemeColors>): ThemeValidationError[] {
	const errors: ThemeValidationError[] = [];

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
 * 验证主题
 */
export function validateTheme(theme: Partial<Theme>): ThemeValidationError[] {
	const errors: ThemeValidationError[] = [];

	// 验证基础字段
	if (!theme.name || theme.name.trim() === '') {
		errors.push({
			field: 'name',
			message: 'Theme name is required'
		});
	}

	// 验证颜色
	if (theme.colors) {
		if (theme.colors.light) {
			errors.push(...validateThemeColors(theme.colors.light));
		}
		if (theme.colors.dark) {
			errors.push(...validateThemeColors(theme.colors.dark));
		}
	}

	return errors;
}

/**
 * 创建默认主题
 */
export function createDefaultTheme(name: string, description?: string): Theme {
	return {
		id: nanoid(),
		name,
		description,
		colors: {
			light: { ...DEFAULT_THEME_COLORS.light },
			dark: { ...DEFAULT_THEME_COLORS.dark }
		},
		metadata: {
			author: 'System',
			version: '1.0.0',
			tags: ['default'],
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString()
		}
	};
}

/**
 * 合并主题颜色
 */
export function mergeThemeColors(base: ThemeColors, override: Partial<ThemeColors>): ThemeColors {
	return {
		...base,
		...override
	};
}

/**
 * 获取主题颜色值
 */
export function getThemeColorValue(
	theme: Theme,
	mode: ColorMode,
	colorName: ThemeColorName
): string {
	return theme.colors[mode][colorName];
}

/**
 * 应用主题到CSS变量
 */
export function applyThemeToCSS(theme: Theme, mode: ColorMode): void {
	const colors = theme.colors[mode];
	const root = document.documentElement;

	// 为根元素添加或移除 dark 类
	if (mode === 'dark') {
		root.classList.add('dark');
	} else {
		root.classList.remove('dark');
	}

	// 转换颜色值为HSL格式并应用到CSS变量
	for (const [key, value] of Object.entries(colors)) {
		const hslValue = convertToHSL(value);
		root.style.setProperty(`--${key}`, hslValue);
	}
}

/**
 * 将颜色值转换为HSL格式（无hsl()包装）
 */
function convertToHSL(color: string): string {
	// 如果已经是HSL格式，直接返回数值部分
	if (color.includes('hsl')) {
		return color.replace(/hsl\(|\)/g, '');
	}

	// 如果是十六进制，转换为HSL
	if (color.startsWith('#')) {
		const hsl = hexToHSL(color);
		return `${hsl.h} ${hsl.s}% ${hsl.l}%`;
	}

	// 如果是RGB，转换为HSL
	if (color.startsWith('rgb')) {
		const hsl = rgbToHSL(color);
		return `${hsl.h} ${hsl.s}% ${hsl.l}%`;
	}

	// 默认返回原始值
	return color;
}

/**
 * 十六进制转HSL
 */
function hexToHSL(hex: string): { h: number; s: number; l: number } {
	const r = parseInt(hex.slice(1, 3), 16) / 255;
	const g = parseInt(hex.slice(3, 5), 16) / 255;
	const b = parseInt(hex.slice(5, 7), 16) / 255;

	return rgbToHSLComponents(r, g, b);
}

/**
 * RGB字符串转HSL
 */
function rgbToHSL(rgb: string): { h: number; s: number; l: number } {
	const matches = rgb.match(/\d+/g);
	if (!matches || matches.length < 3) {
		return { h: 0, s: 0, l: 0 };
	}

	const r = parseInt(matches[0]) / 255;
	const g = parseInt(matches[1]) / 255;
	const b = parseInt(matches[2]) / 255;

	return rgbToHSLComponents(r, g, b);
}

/**
 * RGB组件转HSL
 */
function rgbToHSLComponents(r: number, g: number, b: number): { h: number; s: number; l: number } {
	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	const diff = max - min;
	const sum = max + min;
	const l = sum / 2;

	let h = 0;
	let s = 0;

	if (diff !== 0) {
		s = l > 0.5 ? diff / (2 - sum) : diff / sum;

		switch (max) {
			case r:
				h = (g - b) / diff + (g < b ? 6 : 0);
				break;
			case g:
				h = (b - r) / diff + 2;
				break;
			case b:
				h = (r - g) / diff + 4;
				break;
		}
		h /= 6;
	}

	return {
		h: Math.round(h * 360),
		s: Math.round(s * 100),
		l: Math.round(l * 100)
	};
}

/**
 * 从CSS变量获取主题
 */
export function getThemeFromCSS(mode: ColorMode): Partial<ThemeColors> {
	const root = document.documentElement;
	const colors: Partial<ThemeColors> = {};

	// 从CSS变量读取颜色值
	for (const colorName of THEME_COLORS) {
		const value = root.style.getPropertyValue(`--${colorName}`);
		if (value) {
			colors[colorName] = value;
		}
	}

	return colors;
}

/**
 * 导出主题
 */
export function exportTheme(theme: Theme): string {
	return JSON.stringify(
		{
			version: '1.0.0',
			themes: [theme],
			exportedAt: new Date().toISOString()
		},
		null,
		2
	);
}

/**
 * 导入主题
 */
export function importTheme(themeData: string): Theme[] {
	try {
		const parsed = JSON.parse(themeData);

		// 验证导入的数据格式
		if (!parsed.themes || !Array.isArray(parsed.themes)) {
			throw new Error('Invalid theme data format');
		}

		// 验证每个主题
		const themes: Theme[] = [];
		for (const theme of parsed.themes) {
			const errors = validateTheme(theme);
			if (errors.length > 0) {
				throw new Error(`Invalid theme: ${errors.map((e) => e.message).join(', ')}`);
			}
			themes.push(theme);
		}

		return themes;
	} catch (error) {
		throw new Error(
			`Failed to import theme: ${error instanceof Error ? error.message : 'Unknown error'}`
		);
	}
}

/**
 * 生成主题预览
 */
export function generateThemePreview(theme: Theme, mode: ColorMode): string {
	const colors = theme.colors[mode];

	return `
		<div style="
			background: ${colors.background};
			color: ${colors.foreground};
			padding: 16px;
			border-radius: 8px;
			border: 1px solid ${colors.border};
			font-family: system-ui, -apple-system, sans-serif;
		">
			<h3 style="margin: 0 0 12px 0; color: ${colors.primary};">${theme.name}</h3>
			<div style="display: flex; gap: 8px; margin-bottom: 12px;">
				<div style="
					background: ${colors.primary};
					color: ${colors.background};
					padding: 4px 8px;
					border-radius: 4px;
					font-size: 12px;
				">Primary</div>
				<div style="
					background: ${colors.secondary};
					color: ${colors.background};
					padding: 4px 8px;
					border-radius: 4px;
					font-size: 12px;
				">Secondary</div>
				<div style="
					background: ${colors.accent};
					color: ${colors.background};
					padding: 4px 8px;
					border-radius: 4px;
					font-size: 12px;
				">Accent</div>
			</div>
			<p style="margin: 0; color: ${colors.mutedForeground}; font-size: 14px;">
				${theme.description || 'No description'}
			</p>
		</div>
	`;
}
