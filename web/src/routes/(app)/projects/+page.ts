import { ProjectAPI } from '$lib/api/projects';
import type { ProjectListResponse, ProjectStats } from '$lib/types/project';
import type { LoadEvent } from '@sveltejs/kit';

export const load = async ({ url }: LoadEvent) => {
	try {
		// 从URL获取查询参数
		const page = parseInt(url.searchParams.get('page') || '1');
		const limit = parseInt(url.searchParams.get('limit') || '20');
		const search = url.searchParams.get('search') || undefined;
		const sort_by = url.searchParams.get('sort_by') || 'created_at';
		const sort_order = (url.searchParams.get('sort_order') as 'asc' | 'desc') || 'desc';

		// 并行获取项目列表和统计信息
		const [projectsData, statsData] = await Promise.allSettled([
			ProjectAPI.getProjects({
				page,
				limit,
				search,
				sort_by,
				sort_order
			}),
			ProjectAPI.getProjectStats()
		]);

		const projects: ProjectListResponse =
			projectsData.status === 'fulfilled'
				? projectsData.value
				: { data: [], total: 0, page: 1, limit: 20 };

		const stats: ProjectStats =
			statsData.status === 'fulfilled'
				? statsData.value
				: {
						total_projects: 0,
						active_projects: 0,
						total_assets: 0,
						total_vulnerabilities: 0,
						total_tasks: 0
					};

		return {
			projects,
			stats,
			searchParams: {
				page,
				limit,
				search,
				sort_by,
				sort_order
			}
		};
	} catch (error) {
		console.error('Failed to load projects:', error);

		// 返回默认数据
		return {
			projects: { data: [], total: 0, page: 1, limit: 20 },
			stats: {
				total_projects: 0,
				active_projects: 0,
				total_assets: 0,
				total_vulnerabilities: 0,
				total_tasks: 0
			},
			searchParams: {
				page: 1,
				limit: 20,
				search: undefined,
				sort_by: 'created_at',
				sort_order: 'desc' as const
			},
			error: error instanceof Error ? error.message : 'Unknown error occurred'
		};
	}
};
