import api from './axios-config';
import type {
	Asset,
	CreateAssetRequest,
	UpdateAssetRequest,
	AssetQueryParams,
	AssetResponse
} from '$lib/types/asset';
import type { APIResponse } from '$lib/types/api';

export const assetApi = {
	getAssets: async (params?: AssetQueryParams): Promise<APIResponse<Asset[]>> => {
		const response = await api.get<APIResponse<Asset[]>>('/assets', { params });
		return response.data;
	},

	getAssetById: async (id: string): Promise<APIResponse<Asset>> => {
		const response = await api.get<APIResponse<Asset>>(`/assets/${id}`);
		return response.data;
	},

	createAsset: async (data: CreateAssetRequest): Promise<APIResponse<Asset>> => {
		const response = await api.post<APIResponse<Asset>>('/assets', data);
		return response.data;
	},

	updateAsset: async (data: UpdateAssetRequest): Promise<APIResponse<Asset>> => {
		const response = await api.put<APIResponse<Asset>>(`/assets/${data.id}`, data);
		return response.data;
	},

	deleteAsset: async (id: string): Promise<APIResponse<void>> => {
		const response = await api.delete<APIResponse<void>>(`/assets/${id}`);
		return response.data;
	},

	scanAsset: async (id: string): Promise<APIResponse<void>> => {
		const response = await api.post<APIResponse<void>>(`/assets/${id}/scan`);
		return response.data;
	}
};
