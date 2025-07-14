import type {
	PageMonitoring,
	PageSnapshot,
	PageChange,
	PageMonitoringCreateRequest,
	PageMonitoringUpdateRequest,
	PageMonitoringQueryRequest
} from '$lib/types/monitoring';

// 获取页面监控列表
export async function getPageMonitoring(
	params?: PageMonitoringQueryRequest
): Promise<PageMonitoring[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				if (Array.isArray(value)) {
					value.forEach((v) => searchParams.append(key, String(v)));
				} else {
					searchParams.append(key, String(value));
				}
			}
		});
	}

	const response = await fetch(`/api/v1/monitoring?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取页面监控列表失败');
	}
	return response.json();
}

// 获取页面监控详情
export async function getPageMonitoringById(id: string): Promise<PageMonitoring> {
	const response = await fetch(`/api/v1/monitoring/${id}`);
	if (!response.ok) {
		throw new Error('获取页面监控详情失败');
	}
	return response.json();
}

// 创建页面监控
export async function createPageMonitoring(
	request: PageMonitoringCreateRequest
): Promise<PageMonitoring> {
	const response = await fetch('/api/v1/monitoring', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '创建页面监控失败');
	}
	return response.json();
}

// 更新页面监控
export async function updatePageMonitoring(
	id: string,
	request: PageMonitoringUpdateRequest
): Promise<PageMonitoring> {
	const response = await fetch(`/api/v1/monitoring/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '更新页面监控失败');
	}
	return response.json();
}

// 删除页面监控
export async function deletePageMonitoring(id: string): Promise<void> {
	const response = await fetch(`/api/v1/monitoring/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('删除页面监控失败');
	}
}

// 手动检查页面监控
export async function manualCheckMonitoring(id: string): Promise<any> {
	const response = await fetch(`/api/v1/monitoring/${id}/check`, {
		method: 'POST'
	});
	if (!response.ok) {
		throw new Error('手动检查失败');
	}
	return response.json();
}

// 测试页面监控连接
export async function testMonitoringConnection(url: string, config: any): Promise<any> {
	const response = await fetch('/api/v1/monitoring/test', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ url, config })
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '连接测试失败');
	}
	return response.json();
}

// 获取页面快照列表
export async function getPageSnapshots(
	monitoringId: string,
	params?: {
		limit?: number;
		offset?: number;
		sortBy?: string;
		sortOrder?: string;
	}
): Promise<PageSnapshot[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/monitoring/${monitoringId}/snapshots?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取页面快照失败');
	}
	return response.json();
}

// 获取页面变化记录
export async function getPageChanges(
	monitoringId: string,
	params?: {
		limit?: number;
		offset?: number;
		sortBy?: string;
		sortOrder?: string;
	}
): Promise<PageChange[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/monitoring/${monitoringId}/changes?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取页面变化记录失败');
	}
	return response.json();
}

// 获取页面快照详情
export async function getPageSnapshot(
	monitoringId: string,
	snapshotId: string
): Promise<PageSnapshot> {
	const response = await fetch(`/api/v1/monitoring/${monitoringId}/snapshots/${snapshotId}`);
	if (!response.ok) {
		throw new Error('获取页面快照详情失败');
	}
	return response.json();
}

// 比较两个快照
export async function compareSnapshots(
	monitoringId: string,
	oldSnapshotId: string,
	newSnapshotId: string
): Promise<any> {
	const response = await fetch(`/api/v1/monitoring/${monitoringId}/compare`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ oldSnapshotId, newSnapshotId })
	});
	if (!response.ok) {
		throw new Error('比较快照失败');
	}
	return response.json();
}

// 导出页面监控数据
export async function exportMonitoringData(
	id: string,
	format: 'csv' | 'json' | 'pdf' = 'json'
): Promise<Blob> {
	const response = await fetch(`/api/v1/monitoring/${id}/export?format=${format}`);
	if (!response.ok) {
		throw new Error('导出数据失败');
	}
	return response.blob();
}

// 获取页面监控统计信息
export async function getMonitoringStats(params?: {
	projectId?: string;
	dateRange?: string;
}): Promise<any> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/monitoring/stats?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取统计信息失败');
	}
	return response.json();
}

// 暂停/恢复页面监控
export async function toggleMonitoring(id: string): Promise<void> {
	const response = await fetch(`/api/v1/monitoring/${id}/toggle`, {
		method: 'POST'
	});
	if (!response.ok) {
		throw new Error('切换监控状态失败');
	}
}

// 批量操作页面监控
export async function batchOperateMonitoring(
	ids: string[],
	operation: 'delete' | 'pause' | 'resume'
): Promise<void> {
	const response = await fetch('/api/v1/monitoring/batch', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ ids, operation })
	});
	if (!response.ok) {
		throw new Error('批量操作失败');
	}
}

// 获取页面监控模板
export async function getMonitoringTemplates(): Promise<any[]> {
	const response = await fetch('/api/v1/monitoring/templates');
	if (!response.ok) {
		throw new Error('获取监控模板失败');
	}
	return response.json();
}

// 从模板创建页面监控
export async function createMonitoringFromTemplate(
	templateId: string,
	params: any
): Promise<PageMonitoring> {
	const response = await fetch(`/api/v1/monitoring/templates/${templateId}/create`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(params)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '从模板创建监控失败');
	}
	return response.json();
}
