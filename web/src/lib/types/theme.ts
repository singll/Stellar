/**
 * Theme system type definitions
 */

/**
 * Base color set interface
 */
export interface ColorSet {
	value: string;
	foreground: string;
}

/**
 * Simple theme colors interface
 */
export interface ThemeColors {
	primary: string;
	secondary: string;
	accent: string;
	background: string;
	foreground: string;
	muted: string;
	mutedForeground: string;
	border: string;
	input: string;
	ring: string;
	destructive: string;
	destructiveForeground: string;
	success: string;
	successForeground: string;
	warning: string;
	warningForeground: string;
	info: string;
	infoForeground: string;
}

/**
 * Color mode type
 */
export type ColorMode = 'light' | 'dark';

/**
 * Theme metadata interface
 */
export interface ThemeMetadata {
	author?: string;
	version?: string;
	tags?: string[];
	createdAt?: string;
	updatedAt?: string;
}

/**
 * Complete theme interface
 */
export interface Theme {
	id: string;
	name: string;
	description?: string;
	colors: {
		light: ThemeColors;
		dark: ThemeColors;
	};
	metadata: ThemeMetadata;
}

/**
 * Theme creation request interface
 */
export interface CreateThemeRequest {
	name: string;
	description?: string;
	colors: {
		light: Partial<ThemeColors>;
		dark: Partial<ThemeColors>;
	};
	metadata?: Partial<ThemeMetadata>;
}

/**
 * Theme update request interface
 */
export interface UpdateThemeRequest {
	name?: string;
	description?: string;
	colors?: {
		light?: Partial<ThemeColors>;
		dark?: Partial<ThemeColors>;
	};
	metadata?: Partial<ThemeMetadata>;
}

/**
 * Theme validation error interface
 */
export interface ThemeValidationError {
	field: string;
	message: string;
}

/**
 * Theme response interface
 */
export interface ThemeResponse {
	themes: Theme[];
	total: number;
	page: number;
	limit: number;
}

/**
 * Theme import/export interface
 */
export interface ThemeExport {
	version: string;
	themes: Theme[];
	exportedAt: string;
}

/**
 * Theme constants
 */
export const THEME_COLORS = [
	'primary',
	'secondary',
	'accent',
	'background',
	'foreground',
	'muted',
	'mutedForeground',
	'border',
	'input',
	'ring',
	'destructive',
	'destructiveForeground',
	'success',
	'successForeground',
	'warning',
	'warningForeground',
	'info',
	'infoForeground'
] as const;

export type ThemeColorName = (typeof THEME_COLORS)[number];

/**
 * Default theme colors
 */
export const DEFAULT_THEME_COLORS = {
	light: {
		primary: '221.2 83.2% 53.3%',
		secondary: '210 40% 96.1%',
		accent: '210 40% 96.1%',
		background: '0 0% 100%',
		foreground: '222.2 84% 4.9%',
		muted: '210 40% 96.1%',
		mutedForeground: '215.4 16.3% 46.9%',
		border: '214.3 31.8% 91.4%',
		input: '214.3 31.8% 91.4%',
		ring: '221.2 83.2% 53.3%',
		destructive: '0 84.2% 60.2%',
		destructiveForeground: '210 40% 98%',
		success: '142.1 76.2% 36.3%',
		successForeground: '355.7 100% 97.3%',
		warning: '32.6 94.6% 43.7%',
		warningForeground: '210 40% 98%',
		info: '221.2 83.2% 53.3%',
		infoForeground: '210 40% 98%'
	},
	dark: {
		primary: '217.2 91.2% 59.8%',
		secondary: '217.2 32.6% 17.5%',
		accent: '217.2 32.6% 17.5%',
		background: '222.2 84% 4.9%',
		foreground: '210 40% 98%',
		muted: '217.2 32.6% 17.5%',
		mutedForeground: '215 20.2% 65.1%',
		border: '217.2 32.6% 17.5%',
		input: '217.2 32.6% 17.5%',
		ring: '224.3 76.3% 48%',
		destructive: '0 62.8% 30.6%',
		destructiveForeground: '210 40% 98%',
		success: '142.1 70.6% 45.3%',
		successForeground: '144.9 80.4% 10%',
		warning: '32.6 94.6% 43.7%',
		warningForeground: '210 40% 98%',
		info: '217.2 91.2% 59.8%',
		infoForeground: '222.2 47.4% 11.2%'
	}
} as const;
