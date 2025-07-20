import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	// 应用布局不需要重复初始化认证状态，根布局已经处理了
	return {};
};

export const ssr = false;
