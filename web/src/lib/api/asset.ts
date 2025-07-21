import api from './axios-config';
import type {
	Asset,
	AssetQueryParams,
	AssetListResult,
	CreateAssetRequest,
	UpdateAssetRequest,
	BatchCreateAssetsRequest,
	BatchDeleteAssetsRequest,
	CreateAssetRelationRequest,
	AssetRelation
} from '$lib/types/asset';
import type { APIResponse } from '$lib/types/api';

export const assetApi = {
	// 获取资产列表 - 修复资产类型为必填参数的问题
	getAssets: async (params?: AssetQueryParams): Promise<APIResponse<AssetListResult>> => {
		// 确保资产类型参数不为空，默认使用domain
		const queryParams = {
			...params,
			type: params?.type || 'domain', // 确保type参数不为空
			page: params?.page || 1,
			pageSize: params?.pageSize || 20
		};
		const response = await api.get<APIResponse<AssetListResult>>('/assets', {
			params: queryParams
		});
		return response.data;
	},

	// 获取所有类型的资产列表
	getAllAssets: async (params?: Omit<AssetQueryParams, 'type'>): Promise<APIResponse<Asset[]>> => {
		try {
			// 使用 type="all" 参数一次性获取所有类型的资产
			const queryParams = {
				...params,
				type: 'all',
				page: params?.page || 1,
				pageSize: params?.pageSize || 100
			};
			
			const response = await api.get<APIResponse<AssetListResult>>('/assets', {
				params: queryParams
			});
			
			// 直接返回资产数组
			const assets = response.data.data?.items || [];

			return {
				code: 200,
				message: 'success',
				data: assets
			};
		} catch (error) {
			throw new Error('获取资产列表失败');
		}
	},

	// 获取单个资产
	getAssetById: async (id: string, type: string): Promise<APIResponse<Asset>> => {
		const response = await api.get<APIResponse<Asset>>(`/assets/${id}`, {
			params: { type }
		});
		return response.data;
	},

	// 创建资产
	createAsset: async (data: CreateAssetRequest): Promise<APIResponse<Asset>> => {
		const response = await api.post<APIResponse<Asset>>('/assets', data);
		return response.data;
	},

	// 更新资产
	updateAsset: async (id: string, data: UpdateAssetRequest): Promise<APIResponse<void>> => {
		const response = await api.put<APIResponse<void>>(`/assets/${id}`, data);
		return response.data;
	},

	// 删除资产
	deleteAsset: async (id: string, type: string): Promise<APIResponse<void>> => {
		const response = await api.delete<APIResponse<void>>(`/assets/${id}`, {
			params: { type }
		});
		return response.data;
	},

	// 批量创建资产
	batchCreateAssets: async (
		data: BatchCreateAssetsRequest
	): Promise<APIResponse<{ insertedCount: number; insertedIds: string[] }>> => {
		const response = await api.post<APIResponse<{ insertedCount: number; insertedIds: string[] }>>(
			'/assets/batch',
			data
		);
		return response.data;
	},

	// 批量删除资产
	batchDeleteAssets: async (
		data: BatchDeleteAssetsRequest
	): Promise<APIResponse<{ deletedCount: number }>> => {
		const response = await api.delete<APIResponse<{ deletedCount: number }>>('/assets/batch', {
			data
		});
		return response.data;
	},

	// 导入资产
	importAssets: async (
		file: File,
		projectId: string,
		type: string
	): Promise<APIResponse<{ importedCount: number }>> => {
		const formData = new FormData();
		formData.append('file', file);
		formData.append('projectId', projectId);
		formData.append('type', type);

		const response = await api.post<APIResponse<{ importedCount: number }>>(
			'/assets/import',
			formData,
			{
				headers: {
					'Content-Type': 'multipart/form-data'
				}
			}
		);
		return response.data;
	},

	// 导出资产
	exportAssets: async (
		projectId: string,
		type: string,
		format: 'json' | 'csv' = 'json',
		filename?: string
	): Promise<void> => {
		const params: Record<string, string> = {
			projectId,
			type,
			format
		};
		if (filename) {
			params.filename = filename;
		}

		const response = await api.get('/assets/export', {
			params,
			responseType: 'blob'
		});

		// 创建下载链接
		const url = window.URL.createObjectURL(new Blob([response.data]));
		const link = document.createElement('a');
		link.href = url;
		link.setAttribute('download', filename || `assets-export.${format}`);
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		window.URL.revokeObjectURL(url);
	},

	// 创建资产关系
	createAssetRelation: async (
		data: CreateAssetRelationRequest
	): Promise<APIResponse<AssetRelation>> => {
		const response = await api.post<APIResponse<AssetRelation>>('/assets/relations', data);
		return response.data;
	},

	// 获取资产关系
	getAssetRelations: async (
		projectId: string,
		assetId: string
	): Promise<APIResponse<AssetRelation[]>> => {
		const response = await api.get<APIResponse<AssetRelation[]>>('/assets/relations', {
			params: { projectId, assetId }
		});
		return response.data;
	},

	// 获取资产类型列表
	getAssetTypes: async (): Promise<APIResponse<{ types: Array<{ value: string; label: string; description: string }> }>> => {
		const response = await api.get('/assets/types');
		console.log('Raw API response:', response);
		return response.data;
	},

	// 获取资产统计
	getAssetStats: async (projectId?: string): Promise<APIResponse<Record<string, number>>> => {
		// 使用正确的统计API路径
		const params: Record<string, string> = {};
		if (projectId) {
			params.projectId = projectId;
		}

		const response = await api.get<APIResponse<Record<string, number>>>(
			'/statistics/asset/relationship',
			{ params }
		);
		return response.data;
	}
};
