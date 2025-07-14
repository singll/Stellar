import type {
	SensitiveRule,
	SensitiveRuleGroup,
	SensitiveWhitelist,
	SensitiveDetectionResult,
	SensitiveFinding,
	SensitiveDetectionRequest,
	SensitiveRuleCreateRequest,
	SensitiveRuleUpdateRequest,
	SensitiveRuleGroupCreateRequest,
	SensitiveRuleGroupUpdateRequest
} from '$lib/types/sensitive';

// 获取敏感信息检测结果列表
export async function getSensitiveDetectionResults(params?: {
	projectId?: string;
	status?: string;
	limit?: number;
	offset?: number;
}): Promise<SensitiveDetectionResult[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/sensitive/results?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取敏感信息检测结果失败');
	}
	return response.json();
}

// 获取敏感信息检测结果详情
export async function getSensitiveDetectionResult(id: string): Promise<SensitiveDetectionResult> {
	const response = await fetch(`/api/v1/sensitive/results/${id}`);
	if (!response.ok) {
		throw new Error('获取检测结果详情失败');
	}
	return response.json();
}

// 获取检测结果的发现列表
export async function getSensitiveFindings(
	resultId: string,
	params?: {
		category?: string;
		riskLevel?: string;
		limit?: number;
		offset?: number;
	}
): Promise<SensitiveFinding[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/sensitive/results/${resultId}/findings?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取敏感信息发现失败');
	}
	return response.json();
}

// 创建敏感信息检测任务
export async function createSensitiveDetection(
	request: SensitiveDetectionRequest
): Promise<SensitiveDetectionResult> {
	const response = await fetch('/api/v1/sensitive/detect', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '创建检测任务失败');
	}
	return response.json();
}

// 删除敏感信息检测结果
export async function deleteSensitiveDetectionResult(id: string): Promise<void> {
	const response = await fetch(`/api/v1/sensitive/results/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('删除检测结果失败');
	}
}

// 导出检测结果
export async function exportSensitiveDetectionResult(
	id: string,
	format: 'csv' | 'json' | 'pdf' = 'json'
): Promise<Blob> {
	const response = await fetch(`/api/v1/sensitive/results/${id}/export?format=${format}`);
	if (!response.ok) {
		throw new Error('导出检测结果失败');
	}
	return response.blob();
}

// 测试检测规则
export async function testSensitiveRules(params: {
	targets: string[];
	ruleGroups: { $oid: string }[];
	rules: { $oid: string }[];
}): Promise<any> {
	const response = await fetch('/api/v1/sensitive/test', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(params)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '测试规则失败');
	}
	return response.json();
}

// ========== 规则管理 ==========

// 获取敏感信息规则列表
export async function getSensitiveRules(params?: {
	category?: string;
	riskLevel?: string;
	enabled?: boolean;
	limit?: number;
	offset?: number;
}): Promise<SensitiveRule[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/sensitive/rules?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取敏感信息规则失败');
	}
	return response.json();
}

// 获取敏感信息规则详情
export async function getSensitiveRule(id: string): Promise<SensitiveRule> {
	const response = await fetch(`/api/v1/sensitive/rules/${id}`);
	if (!response.ok) {
		throw new Error('获取规则详情失败');
	}
	return response.json();
}

// 创建敏感信息规则
export async function createSensitiveRule(
	request: SensitiveRuleCreateRequest
): Promise<SensitiveRule> {
	const response = await fetch('/api/v1/sensitive/rules', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '创建规则失败');
	}
	return response.json();
}

// 更新敏感信息规则
export async function updateSensitiveRule(
	id: string,
	request: SensitiveRuleUpdateRequest
): Promise<SensitiveRule> {
	const response = await fetch(`/api/v1/sensitive/rules/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '更新规则失败');
	}
	return response.json();
}

// 删除敏感信息规则
export async function deleteSensitiveRule(id: string): Promise<void> {
	const response = await fetch(`/api/v1/sensitive/rules/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('删除规则失败');
	}
}

// ========== 规则组管理 ==========

// 获取敏感信息规则组列表
export async function getSensitiveRuleGroups(params?: {
	enabled?: boolean;
	limit?: number;
	offset?: number;
}): Promise<SensitiveRuleGroup[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/sensitive/rule-groups?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取规则组失败');
	}
	return response.json();
}

// 获取敏感信息规则组详情
export async function getSensitiveRuleGroup(id: string): Promise<SensitiveRuleGroup> {
	const response = await fetch(`/api/v1/sensitive/rule-groups/${id}`);
	if (!response.ok) {
		throw new Error('获取规则组详情失败');
	}
	return response.json();
}

// 创建敏感信息规则组
export async function createSensitiveRuleGroup(
	request: SensitiveRuleGroupCreateRequest
): Promise<SensitiveRuleGroup> {
	const response = await fetch('/api/v1/sensitive/rule-groups', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '创建规则组失败');
	}
	return response.json();
}

// 更新敏感信息规则组
export async function updateSensitiveRuleGroup(
	id: string,
	request: SensitiveRuleGroupUpdateRequest
): Promise<SensitiveRuleGroup> {
	const response = await fetch(`/api/v1/sensitive/rule-groups/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '更新规则组失败');
	}
	return response.json();
}

// 删除敏感信息规则组
export async function deleteSensitiveRuleGroup(id: string): Promise<void> {
	const response = await fetch(`/api/v1/sensitive/rule-groups/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('删除规则组失败');
	}
}

// ========== 白名单管理 ==========

// 获取敏感信息白名单列表
export async function getSensitiveWhitelists(params?: {
	type?: string;
	limit?: number;
	offset?: number;
}): Promise<SensitiveWhitelist[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/sensitive/whitelists?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取白名单失败');
	}
	return response.json();
}

// 获取敏感信息白名单详情
export async function getSensitiveWhitelist(id: string): Promise<SensitiveWhitelist> {
	const response = await fetch(`/api/v1/sensitive/whitelists/${id}`);
	if (!response.ok) {
		throw new Error('获取白名单详情失败');
	}
	return response.json();
}

// 创建敏感信息白名单
export async function createSensitiveWhitelist(request: any): Promise<SensitiveWhitelist> {
	const response = await fetch('/api/v1/sensitive/whitelists', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '创建白名单失败');
	}
	return response.json();
}

// 更新敏感信息白名单
export async function updateSensitiveWhitelist(
	id: string,
	request: any
): Promise<SensitiveWhitelist> {
	const response = await fetch(`/api/v1/sensitive/whitelists/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(request)
	});
	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.message || '更新白名单失败');
	}
	return response.json();
}

// 删除敏感信息白名单
export async function deleteSensitiveWhitelist(id: string): Promise<void> {
	const response = await fetch(`/api/v1/sensitive/whitelists/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('删除白名单失败');
	}
}

// ========== 统计和报告 ==========

// 获取敏感信息检测统计
export async function getSensitiveDetectionStats(params?: {
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

	const response = await fetch(`/api/v1/sensitive/stats?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取统计信息失败');
	}
	return response.json();
}

// 生成敏感信息检测报告
export async function generateSensitiveDetectionReport(
	resultId: string,
	format: 'pdf' | 'html' = 'pdf'
): Promise<Blob> {
	const response = await fetch(`/api/v1/sensitive/results/${resultId}/report?format=${format}`, {
		method: 'POST'
	});
	if (!response.ok) {
		throw new Error('生成报告失败');
	}
	return response.blob();
}
