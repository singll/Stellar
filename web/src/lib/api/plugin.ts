import type {
	PluginMetadata,
	PluginRunRecord,
	PluginConfig,
	PluginInstallRequest,
	PluginInstallResponse
} from '$lib/types/plugin';

// 获取插件列表
export async function getPlugins(): Promise<PluginMetadata[]> {
	const response = await fetch('/api/v1/plugins');
	if (!response.ok) {
		throw new Error('获取插件列表失败');
	}
	return response.json();
}

// 根据类型获取插件列表
export async function getPluginsByType(type: string): Promise<PluginMetadata[]> {
	const response = await fetch(`/api/v1/plugins?type=${type}`);
	if (!response.ok) {
		throw new Error('获取插件列表失败');
	}
	return response.json();
}

// 根据分类获取插件列表
export async function getPluginsByCategory(category: string): Promise<PluginMetadata[]> {
	const response = await fetch(`/api/v1/plugins?category=${category}`);
	if (!response.ok) {
		throw new Error('获取插件列表失败');
	}
	return response.json();
}

// 获取插件详情
export async function getPlugin(id: string): Promise<PluginMetadata> {
	const response = await fetch(`/api/v1/plugins/${id}`);
	if (!response.ok) {
		throw new Error('获取插件详情失败');
	}
	return response.json();
}

// 安装插件
export async function installPlugin(request: PluginInstallRequest): Promise<PluginInstallResponse> {
	const formData = new FormData();

	switch (request.method) {
		case 'file':
			if (request.file) {
				formData.append('file', request.file);
			}
			break;
		case 'url':
			if (request.url) {
				formData.append('url', request.url);
			}
			break;
		case 'yaml':
			if (request.yaml) {
				formData.append('yaml', request.yaml);
			}
			break;
	}

	const response = await fetch('/api/v1/plugins/install', {
		method: 'POST',
		body: formData
	});

	const result = await response.json();
	return {
		success: response.ok,
		message: result.message || (response.ok ? '安装成功' : '安装失败'),
		plugin: result.plugin,
		error: result.error
	};
}

// 卸载插件
export async function uninstallPlugin(id: string): Promise<void> {
	const response = await fetch(`/api/v1/plugins/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('卸载插件失败');
	}
}

// 启用/禁用插件
export async function togglePlugin(id: string): Promise<void> {
	const response = await fetch(`/api/v1/plugins/${id}/toggle`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' }
	});
	if (!response.ok) {
		throw new Error('切换插件状态失败');
	}
}

// 更新插件
export async function updatePlugin(
	id: string,
	metadata: Partial<PluginMetadata>
): Promise<PluginMetadata> {
	const response = await fetch(`/api/v1/plugins/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(metadata)
	});
	if (!response.ok) {
		throw new Error('更新插件失败');
	}
	return response.json();
}

// 执行插件
export async function executePlugin(id: string, params: Record<string, any>): Promise<any> {
	const response = await fetch(`/api/v1/plugins/${id}/execute`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(params)
	});
	if (!response.ok) {
		throw new Error('执行插件失败');
	}
	return response.json();
}

// 获取插件运行记录
export async function getPluginRunRecords(id: string): Promise<PluginRunRecord[]> {
	const response = await fetch(`/api/v1/plugins/${id}/records`);
	if (!response.ok) {
		throw new Error('获取运行记录失败');
	}
	return response.json();
}

// 获取插件配置
export async function getPluginConfig(id: string): Promise<PluginConfig[]> {
	const response = await fetch(`/api/v1/plugins/${id}/configs`);
	if (!response.ok) {
		throw new Error('获取插件配置失败');
	}
	return response.json();
}

// 保存插件配置
export async function savePluginConfig(
	id: string,
	config: Partial<PluginConfig>
): Promise<PluginConfig> {
	const response = await fetch(`/api/v1/plugins/${id}/configs`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(config)
	});
	if (!response.ok) {
		throw new Error('保存插件配置失败');
	}
	return response.json();
}

// 删除插件配置
export async function deletePluginConfig(pluginId: string, configId: string): Promise<void> {
	const response = await fetch(`/api/v1/plugins/${pluginId}/configs/${configId}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('删除插件配置失败');
	}
}

// 获取插件市场列表
export async function getPluginMarket(params?: {
	category?: string;
	type?: string;
	search?: string;
	limit?: number;
	offset?: number;
}): Promise<any> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				searchParams.append(key, String(value));
			}
		});
	}

	const response = await fetch(`/api/v1/plugins/market?${searchParams}`);
	if (!response.ok) {
		throw new Error('获取插件市场失败');
	}
	return response.json();
}

// 从市场安装插件
export async function installFromMarket(id: string): Promise<PluginInstallResponse> {
	const response = await fetch(`/api/v1/plugins/market/${id}/install`, {
		method: 'POST'
	});

	const result = await response.json();
	return {
		success: response.ok,
		message: result.message || (response.ok ? '安装成功' : '安装失败'),
		plugin: result.plugin,
		error: result.error
	};
}

// 验证YAML插件配置
export async function validateYAMLPlugin(
	yaml: string
): Promise<{ valid: boolean; error?: string }> {
	const response = await fetch('/api/v1/plugins/validate-yaml', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ yaml })
	});

	if (!response.ok) {
		return { valid: false, error: '验证请求失败' };
	}

	return response.json();
}
