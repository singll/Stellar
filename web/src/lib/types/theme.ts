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
 * Extended theme colors interface with status colors
 */
export interface ThemeColors {
	primary: ColorSet;
	secondary: ColorSet;
	accent: ColorSet;
	success: ColorSet;
	warning: ColorSet;
	error: ColorSet;
	info: ColorSet;
	background: ColorSet;
	foreground: ColorSet;
	border: ColorSet;
	card: ColorSet;
	sidebar: ColorSet;
	header: ColorSet;
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
		light: ThemeColors;
		dark: ThemeColors;
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
		light: Partial<ThemeColors>;
		dark: Partial<ThemeColors>;
	};
	metadata?: Partial<ThemeMetadata>;
}

/**
 * Color mode type
 */
export type ColorMode = 'light' | 'dark';

/**
 * Theme validation error interface
 */
export interface ThemeValidationError {
	field: string;
	message: string;
}

/**
 * 主题事件类型
 */
export type ThemeEvent =
	| { type: 'create'; theme: Theme }
	| { type: 'update'; id: string; theme: Theme }
	| { type: 'delete'; id: string }
	| { type: 'import'; theme: Theme }
	| { type: 'export'; id: string }
	| { type: 'setActive'; id: string };

/**
 * 主题元数据
 */
export interface ThemeMetadata {
	author?: string;
	version?: string;
	createdAt: string;
	updatedAt: string;
	isBuiltin?: boolean;
	tags?: string[];
}

/**
 * 主题创建选项
 */
export interface CreateThemeOptions {
	name: string;
	author: string;
	description?: string;
	baseTheme?: Theme;
}

export interface ThemeValidationResult {
	isValid: boolean;
	errors: string[];
}
