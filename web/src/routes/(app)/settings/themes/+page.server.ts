import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return {
		title: '主题管理',
		description: '自定义和管理应用主题'
	};
};
