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
		let projects: ProjectListResponse;
		let stats: ProjectStats;

		try {
			console.log('🚀 [Server] 开始加载项目数据...');
			console.log('📝 [Server] 查询参数:', { page, limit, search, sort_by, sort_order });
			[projects, stats] = await Promise.all([
				ProjectAPI.getProjects({
					page,
					limit,
					search,
					sort_by,
					sort_order
				}),
				ProjectAPI.getProjectStats()
			]);
			console.log('✅ [Server] 项目数据加载成功:', { 
				projectsCount: projects.data?.length || 0, 
				projectsTotal: projects.total,
				stats 
			});
		} catch (error) {
			console.error('加载项目数据失败:', error);
			// 返回默认数据
			projects = { data: [], total: 0, page: 1, limit: 20 };
			stats = {
				total_projects: 0,
				active_projects: 0,
				total_assets: 0,
				total_vulnerabilities: 0,
				total_tasks: 0
			};
		}

		console.log('服务端返回数据:', { projects, stats });
		
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
