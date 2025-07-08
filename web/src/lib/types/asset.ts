export type AssetType = 'domain' | 'ip' | 'web' | 'app' | 'other';
export type AssetStatus = 'active' | 'inactive' | 'deleted';
export type RiskLevel = 'low' | 'medium' | 'high';

export interface Asset {
	id: string;
	name: string;
	type: AssetType;
	url?: string;
	ip: string;
	domain?: string;
	status: AssetStatus;
	lastScan: string;
	riskLevel: RiskLevel;
	description: string;
	tags?: string[];
	createdAt: string;
	updatedAt: string;
}

export interface CreateAssetRequest {
	name: string;
	type: AssetType;
	url?: string;
	ip?: string;
	description?: string;
	tags?: string[];
}

export interface UpdateAssetRequest extends Partial<CreateAssetRequest> {
	id: string;
}

export interface AssetQueryParams {
	type?: AssetType;
	status?: AssetStatus;
	riskLevel?: RiskLevel;
	search?: string;
	page?: number;
	limit?: number;
}

export interface AssetResponse {
	code: number;
	message: string;
	data: Asset[];
	total?: number;
	page?: number;
	limit?: number;
}
