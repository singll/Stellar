import api from './axios-config';
import type {
	Asset,
	AssetQueryParams,
	CreateAssetRequest,
	UpdateAssetRequest
} from '$lib/types/asset';
import { handleApiResponse, handlePaginatedResponse } from '$lib/utils/api-response-handler';

export interface AssetListResponse {
	code: number;
	message: string;
	data?: {
		items: Asset[];
		total: number;
		page: number;
		pageSize: number;
	};
}

export interface AssetResponse {
	code: number;
	message: string;
	data?: Asset;
}

export const assetsApi = {
	async getAssets(params?: AssetQueryParams): Promise<AssetListResponse> {
		const response = await api.get('/assets', { params: { ...params, type: params?.type || 'domain' } });
		return handleApiResponse(response.data);
	},

	async getAsset(id: string): Promise<AssetResponse> {
		const response = await api.get(`/assets/${id}`);
		return handleApiResponse(response.data);
	},

	async createAsset(asset: CreateAssetRequest): Promise<AssetResponse> {
		const response = await api.post('/assets', asset);
		return handleApiResponse(response.data);
	},

	async updateAsset(id: string, asset: UpdateAssetRequest): Promise<AssetResponse> {
		const response = await api.put(`/assets/${id}`, asset);
		return handleApiResponse(response.data);
	},

	async deleteAsset(id: string): Promise<void> {
		await api.delete(`/assets/${id}`);
	},

	async bulkDeleteAssets(ids: string[]): Promise<void> {
		await api.delete('/assets/bulk', { data: { ids } });
	},

	async searchAssets(query: string): Promise<AssetListResponse> {
		const response = await api.get('/assets/search', { params: { q: query } });
		return handleApiResponse(response.data);
	},

	async getAssetHistory(id: string): Promise<any> {
		const response = await api.get(`/assets/${id}/history`);
		return handleApiResponse(response.data);
	},

	async exportAssets(params?: AssetQueryParams): Promise<Blob> {
		const response = await api.get('/assets/export', {
			params,
			responseType: 'blob'
		});
		return response.data;
	}
};
