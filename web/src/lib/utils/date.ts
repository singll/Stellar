/**
 * 日期时间工具函数
 */

/**
 * 格式化相对时间（如：2分钟前、1小时前等）
 */
export function formatRelativeTime(dateString: string): string {
	const date = new Date(dateString);
	const now = new Date();
	const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

	if (diffInSeconds < 60) {
		return '刚刚';
	}

	const diffInMinutes = Math.floor(diffInSeconds / 60);
	if (diffInMinutes < 60) {
		return `${diffInMinutes}分钟前`;
	}

	const diffInHours = Math.floor(diffInMinutes / 60);
	if (diffInHours < 24) {
		return `${diffInHours}小时前`;
	}

	const diffInDays = Math.floor(diffInHours / 24);
	if (diffInDays < 30) {
		return `${diffInDays}天前`;
	}

	const diffInMonths = Math.floor(diffInDays / 30);
	if (diffInMonths < 12) {
		return `${diffInMonths}个月前`;
	}

	const diffInYears = Math.floor(diffInMonths / 12);
	return `${diffInYears}年前`;
}

/**
 * 格式化日期时间
 */
export function formatDateTime(
	dateString: string,
	options?: {
		includeTime?: boolean;
		includeSeconds?: boolean;
		use24Hour?: boolean;
	}
): string {
	const { includeTime = true, includeSeconds = false, use24Hour = true } = options || {};

	const date = new Date(dateString);

	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');

	let result = `${year}-${month}-${day}`;

	if (includeTime) {
		const hours = use24Hour
			? String(date.getHours()).padStart(2, '0')
			: String(date.getHours() % 12 || 12).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');

		result += ` ${hours}:${minutes}`;

		if (includeSeconds) {
			const seconds = String(date.getSeconds()).padStart(2, '0');
			result += `:${seconds}`;
		}

		if (!use24Hour) {
			result += date.getHours() >= 12 ? ' PM' : ' AM';
		}
	}

	return result;
}

/**
 * 格式化时间
 */
export function formatTime(
	dateString: string,
	options?: {
		includeSeconds?: boolean;
		use24Hour?: boolean;
	}
): string {
	const { includeSeconds = false, use24Hour = true } = options || {};

	const date = new Date(dateString);

	const hours = use24Hour
		? String(date.getHours()).padStart(2, '0')
		: String(date.getHours() % 12 || 12).padStart(2, '0');
	const minutes = String(date.getMinutes()).padStart(2, '0');

	let result = `${hours}:${minutes}`;

	if (includeSeconds) {
		const seconds = String(date.getSeconds()).padStart(2, '0');
		result += `:${seconds}`;
	}

	if (!use24Hour) {
		result += date.getHours() >= 12 ? ' PM' : ' AM';
	}

	return result;
}

/**
 * 格式化日期
 */
export function formatDate(dateString: string): string {
	const date = new Date(dateString);

	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');

	return `${year}-${month}-${day}`;
}

/**
 * 格式化持续时间
 */
export function formatDuration(startTime: string, endTime?: string): string {
	const start = new Date(startTime);
	const end = endTime ? new Date(endTime) : new Date();

	const diffInMs = end.getTime() - start.getTime();
	const diffInSeconds = Math.floor(diffInMs / 1000);

	if (diffInSeconds < 60) {
		return `${diffInSeconds}秒`;
	}

	const diffInMinutes = Math.floor(diffInSeconds / 60);
	if (diffInMinutes < 60) {
		const seconds = diffInSeconds % 60;
		return seconds > 0 ? `${diffInMinutes}分${seconds}秒` : `${diffInMinutes}分钟`;
	}

	const diffInHours = Math.floor(diffInMinutes / 60);
	if (diffInHours < 24) {
		const minutes = diffInMinutes % 60;
		return minutes > 0 ? `${diffInHours}小时${minutes}分钟` : `${diffInHours}小时`;
	}

	const diffInDays = Math.floor(diffInHours / 24);
	const hours = diffInHours % 24;
	return hours > 0 ? `${diffInDays}天${hours}小时` : `${diffInDays}天`;
}

/**
 * 检查日期是否是今天
 */
export function isToday(dateString: string): boolean {
	const date = new Date(dateString);
	const today = new Date();

	return (
		date.getDate() === today.getDate() &&
		date.getMonth() === today.getMonth() &&
		date.getFullYear() === today.getFullYear()
	);
}

/**
 * 检查日期是否是昨天
 */
export function isYesterday(dateString: string): boolean {
	const date = new Date(dateString);
	const yesterday = new Date();
	yesterday.setDate(yesterday.getDate() - 1);

	return (
		date.getDate() === yesterday.getDate() &&
		date.getMonth() === yesterday.getMonth() &&
		date.getFullYear() === yesterday.getFullYear()
	);
}

/**
 * 智能格式化日期时间
 */
export function formatSmartDateTime(dateString: string): string {
	if (isToday(dateString)) {
		return `今天 ${formatTime(dateString)}`;
	}

	if (isYesterday(dateString)) {
		return `昨天 ${formatTime(dateString)}`;
	}

	const date = new Date(dateString);
	const now = new Date();
	const diffInDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

	if (diffInDays < 7) {
		const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
		return `${weekdays[date.getDay()]} ${formatTime(dateString)}`;
	}

	if (date.getFullYear() === now.getFullYear()) {
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${month}-${day} ${formatTime(dateString)}`;
	}

	return formatDateTime(dateString, { includeSeconds: false });
}
