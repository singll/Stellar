import api from './axios-config';
import type {
	Project,
	CreateProjectRequest,
	UpdateProjectRequest,
	ProjectQueryParams,
	ProjectListResponse,
	ProjectResponse,
	ProjectStats
} from '$lib/types/project';

/**
 * 项目管理API客户端
 */
export class ProjectAPI {
	/**
	 * 获取项目列表
	 * @param params 查询参数
	 * @returns 项目列表
	 */
	static async getProjects(params?: ProjectQueryParams): Promise<ProjectListResponse> {
		const query = {
			search: params?.search || '',
			pageIndex: params?.page || 1,
			pageSize: params?.limit || 20
		};
		
		console.log('🔍 [API] 请求项目列表参数:', query);
		
		try {
			const response = await api.get('/projects', { params: query });
			console.log('📦 [API] 项目列表API原始响应:', response);
			console.log('📊 [API] 项目列表API响应数据:', response.data);
			
			// 适配不同的响应格式
			const data = response.data;
			
			// 记录响应结构
			console.log('📋 [API] 响应数据类型:', typeof data);
			console.log('📋 [API] 响应数据结构:', {
				isArray: Array.isArray(data),
				hasDataField: data && 'data' in data,
				hasCodeField: data && 'code' in data,
				keys: data ? Object.keys(data) : [],
				dataType: data ? typeof data : 'null'
			});
			
			// 检查标准格式
			if (data && data.data && Array.isArray(data.data)) {
				console.log('✅ [API] 返回标准格式:', {
					data: data.data,
					total: data.total,
					page: data.page,
					limit: data.limit
				});
				return data;
			}
			
			// 检查嵌套格式
			if (data && data.code === 200 && data.data && Array.isArray(data.data.data)) {
				console.log('✅ [API] 返回嵌套格式:', data.data);
				return data.data;
			}
			
			// 如果直接返回数组
			if (Array.isArray(data)) {
				console.log('✅ [API] 返回数组格式:', data);
				return {
					data: data,
					total: data.length,
					page: params?.page || 1,
					limit: params?.limit || 20
				};
			}
			
			// 如果返回的是对象但没有data字段
			if (data && typeof data === 'object' && !Array.isArray(data)) {
				console.log('✅ [API] 返回对象格式:', data);
				const result = {
					data: data.items || data.list || data.projects || data.records || [],
					total: data.total || data.count || data.totalRecords || 0,
					page: params?.page || 1,
					limit: params?.limit || 20
				};
				console.log('✅ [API] 转换后的格式:', result);
				return result;
			}
			
			console.error('❌ [API] 未知响应格式:', data);
			return {
				data: [],
				total: 0,
				page: params?.page || 1,
				limit: params?.limit || 20
			};
			
		} catch (error) {
			console.error('❌ [API] 获取项目列表API错误:', error);
			return {
				data: [],
				total: 0,
				page: params?.page || 1,
				limit: params?.limit || 20
			};
		}
	}

	/**
	 * 获取单个项目
	 * @param id 项目ID
	 * @returns 项目详情
	 */
	static async getProject(id: string): Promise<ProjectResponse> {
		const response = await api.get(`/projects/${id}`);
		return response.data;
	}

	/**
	 * 创建项目
	 * @param project 项目创建数据
	 * @returns 创建的项目
	 */
	static async createProject(project: CreateProjectRequest): Promise<{ id: string }> {
		const response = await api.post('/projects', project);
		return response.data.data;
	}

	/**
	 * 更新项目
	 * @param id 项目ID
	 * @param project 更新数据
	 * @returns 更新后的项目
	 */
	static async updateProject(id: string, project: UpdateProjectRequest): Promise<ProjectResponse> {
		const response = await api.put(`/projects/${id}`, project);
		return response.data;
	}

	/**
	 * 删除项目
	 * @param id 项目ID
	 */
	static async deleteProject(id: string): Promise<void> {
		await api.delete(`/projects/${id}`);
	}

	/**
	 * 获取项目统计信息
	 * @returns 项目统计
	 */
	static async getProjectStats(): Promise<ProjectStats> {
		try {
			// 使用正确的统计API路径
			const response = await api.get('/statistics/dashboard/stats');
			const data = response.data;
			
			// 适配不同的响应格式
			if (data.data) {
				return data.data;
			}
			return data;
		} catch (error) {
			console.error('获取项目统计失败:', error);
			// 返回默认统计信息
			return {
				total_projects: 0,
				active_projects: 0,
				total_assets: 0,
				total_vulnerabilities: 0,
				total_tasks: 0
			};
		}
	}

	/**
	 * 复制项目
	 * @param id 项目ID
	 * @param name 新项目名称
	 * @returns 复制的项目
	 */
	static async duplicateProject(id: string, name: string): Promise<Project> {
		const response = await api.post(`/projects/${id}/duplicate`, { name });
		return response.data.data;
	}

	/**
	 * 导出项目数据
	 * @param id 项目ID
	 * @param format 导出格式
	 * @returns 导出数据的下载链接
	 */
	static async exportProject(
		id: string,
		format: 'json' | 'csv' | 'xlsx' = 'json'
	): Promise<string> {
		const response = await api.get(`/projects/${id}/export`, {
			params: { format },
			responseType: 'blob'
		});

		// 创建下载链接
		const blob = new Blob([response.data], {
			type:
				format === 'json'
					? 'application/json'
					: format === 'csv'
						? 'text/csv'
						: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
		});

		return URL.createObjectURL(blob);
	}

	/**
	 * 获取项目成员列表
	 * @param id 项目ID
	 * @returns 成员列表
	 */
	static async getProjectMembers(id: string): Promise<any[]> {
		const response = await api.get(`/projects/${id}/members`);
		return response.data.data;
	}

	/**
	 * 添加项目成员
	 * @param id 项目ID
	 * @param userIds 用户ID列表
	 * @param role 角色
	 */
	static async addProjectMembers(
		id: string,
		userIds: string[],
		role: string = 'member'
	): Promise<void> {
		await api.post(`/projects/${id}/members`, { user_ids: userIds, role });
	}

	/**
	 * 移除项目成员
	 * @param id 项目ID
	 * @param userId 用户ID
	 */
	static async removeProjectMember(id: string, userId: string): Promise<void> {
		await api.delete(`/projects/${id}/members/${userId}`);
	}

	/**
	 * 更新项目成员角色
	 * @param id 项目ID
	 * @param userId 用户ID
	 * @param role 新角色
	 */
	static async updateProjectMemberRole(id: string, userId: string, role: string): Promise<void> {
		await api.put(`/projects/${id}/members/${userId}`, { role });
	}

	/**
	 * 获取项目活动日志
	 * @param id 项目ID
	 * @param params 查询参数
	 * @returns 活动日志
	 */
	static async getProjectActivities(
		id: string,
		params?: { page?: number; limit?: number }
	): Promise<any> {
		const response = await api.get(`/projects/${id}/activities`, { params });
		return response.data;
	}

	/**
	 * 归档项目
	 * @param id 项目ID
	 */
	static async archiveProject(id: string): Promise<void> {
		await api.put(`/projects/${id}/archive`);
	}

	/**
	 * 取消归档项目
	 * @param id 项目ID
	 */
	static async unarchiveProject(id: string): Promise<void> {
		await api.put(`/projects/${id}/unarchive`);
	}

	/**
	 * 获取项目资产
	 * @param id 项目ID
	 * @param params 查询参数
	 * @returns 项目资产
	 */
	static async getProjectAssets(id: string, params?: any): Promise<any> {
		const response = await api.get(`/projects/${id}/assets`, { params });
		return response.data;
	}

	/**
	 * 获取项目任务
	 * @param id 项目ID
	 * @param params 查询参数
	 * @returns 项目任务
	 */
	static async getProjectTasks(id: string, params?: any): Promise<any> {
		const response = await api.get(`/projects/${id}/tasks`, { params });
		return response.data;
	}
}
