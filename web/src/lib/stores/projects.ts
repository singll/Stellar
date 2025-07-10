// 项目管理状态存储
import { writable } from 'svelte/store';
import { ProjectAPI } from '$lib/api/projects';
import type { Project, ProjectsResponse } from '$lib/types/project';

interface ProjectState {
	projects: Project[];
	loading: boolean;
	error: string | null;
	currentProject: Project | null;
	pagination: {
		page: number;
		pageSize: number;
		total: number;
		totalPages: number;
	};
}

const initialState: ProjectState = {
	projects: [],
	loading: false,
	error: null,
	currentProject: null,
	pagination: {
		page: 1,
		pageSize: 20,
		total: 0,
		totalPages: 0
	}
};

export const projectStore = writable<ProjectState>(initialState);

// 项目操作
export const projectActions = {
	// 加载项目列表
	async loadProjects(page: number = 1, pageSize: number = 20) {
		projectStore.update((state) => ({ ...state, loading: true, error: null }));

		try {
			const response = await ProjectAPI.getProjects({ page, limit: pageSize });

			projectStore.update((state) => ({
				...state,
				projects: response.data || [],
				pagination: {
					page: response.page || page,
					pageSize: response.limit || pageSize,
					total: response.total || 0,
					totalPages: Math.ceil((response.total || 0) / pageSize)
				},
				loading: false
			}));

			return response;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : '加载项目失败';
			projectStore.update((state) => ({
				...state,
				loading: false,
				error: errorMessage
			}));
			throw error;
		}
	},

	// 获取单个项目
	async getProject(id: string) {
		projectStore.update((state) => ({ ...state, loading: true, error: null }));

		try {
			const project = await ProjectAPI.getProject(id);

			projectStore.update((state) => ({
				...state,
				currentProject: project,
				loading: false
			}));

			return project;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : '获取项目失败';
			projectStore.update((state) => ({
				...state,
				loading: false,
				error: errorMessage
			}));
			throw error;
		}
	},

	// 创建项目
	async createProject(projectData: Omit<Project, 'id' | 'created_at' | 'updated_at'>) {
		projectStore.update((state) => ({ ...state, loading: true, error: null }));

		try {
			const newProject = await ProjectAPI.createProject(projectData);

			projectStore.update((state) => ({
				...state,
				projects: [newProject, ...state.projects],
				loading: false
			}));

			return newProject;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : '创建项目失败';
			projectStore.update((state) => ({
				...state,
				loading: false,
				error: errorMessage
			}));
			throw error;
		}
	},

	// 更新项目
	async updateProject(id: string, projectData: Partial<Project>) {
		projectStore.update((state) => ({ ...state, loading: true, error: null }));

		try {
			const updatedProject = await ProjectAPI.updateProject(id, projectData);

			projectStore.update((state) => ({
				...state,
				projects: state.projects.map((p) => (p.id === id ? updatedProject : p)),
				currentProject: state.currentProject?.id === id ? updatedProject : state.currentProject,
				loading: false
			}));

			return updatedProject;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : '更新项目失败';
			projectStore.update((state) => ({
				...state,
				loading: false,
				error: errorMessage
			}));
			throw error;
		}
	},

	// 删除项目
	async deleteProject(id: string) {
		projectStore.update((state) => ({ ...state, loading: true, error: null }));

		try {
			await ProjectAPI.deleteProject(id);

			projectStore.update((state) => ({
				...state,
				projects: state.projects.filter((p) => p.id !== id),
				currentProject: state.currentProject?.id === id ? null : state.currentProject,
				loading: false
			}));
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : '删除项目失败';
			projectStore.update((state) => ({
				...state,
				loading: false,
				error: errorMessage
			}));
			throw error;
		}
	},

	// 重置状态
	reset() {
		projectStore.set(initialState);
	},

	// 清除错误
	clearError() {
		projectStore.update((state) => ({ ...state, error: null }));
	}
};

// 导出便捷方法
export default projectStore;
