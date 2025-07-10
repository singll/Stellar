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
	// 获取资产列表 - 使用正确的路径和参数名
	getAssets: async (params?: AssetQueryParams): Promise<APIResponse<AssetListResult>> => {
		// 确保参数名称与后端一致
		const queryParams = {
			...params,
			page: params?.page || 1,
			pageSize: params?.pageSize || 20
		};
		const response = await api.get<APIResponse<AssetListResult>>('/assets/assets', { 
			params: queryParams 
		});
		return response.data;
	},

	// 获取单个资产
	getAssetById: async (id: string, type: string): Promise<APIResponse<Asset>> => {
		const response = await api.get<APIResponse<Asset>>(`/assets/assets/${id}`, {
			params: { type }
		});
		return response.data;
	},

	// 创建资产
	createAsset: async (data: CreateAssetRequest): Promise<APIResponse<Asset>> => {
		const response = await api.post<APIResponse<Asset>>('/assets/assets', data);
		return response.data;
	},

	// 更新资产
	updateAsset: async (id: string, data: UpdateAssetRequest): Promise<APIResponse<void>> => {
		const response = await api.put<APIResponse<void>>(`/assets/assets/${id}`, data);
		return response.data;
	},

	// 删除资产
	deleteAsset: async (id: string, type: string): Promise<APIResponse<void>> => {
		const response = await api.delete<APIResponse<void>>(`/assets/assets/${id}`, {
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

	// 获取资产统计
	getAssetStats: async (projectId?: string): Promise<APIResponse<Record<string, number>>> => {
		// 使用正确的统计API路径
		const requestData: Record<string, string> = {};
		if (projectId) {
			requestData.projectId = projectId;
		}

		const response = await api.post<APIResponse<Record<string, number>>>('/statistics/asset/relationship', requestData);
		return response.data;
	}
};
