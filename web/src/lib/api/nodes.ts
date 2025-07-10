/**
 * 节点管理 API 服务
 * 与后端 internal/api/node.go 保持一致
 */

import api from './axios-config';
import type {
	Node,
	NodeQueryParams,
	NodeRegistrationRequest,
	NodeRegistrationResponse,
	NodeUpdateRequest,
	NodeStatusUpdateRequest,
	NodeHeartbeat,
	NodeConfig,
	NodeStats,
	NodeHealth,
	NodeTaskStats,
	NodeListResponse,
	NodeMaintenanceRequest,
	NodeBatchOperationRequest,
	NodeCleanupRequest,
	NodeCleanupResponse,
	NodeStatusType,
	NodeRoleType
} from '../types/node';

export class NodeAPI {
	/**
	 * 获取节点列表
	 * @param params 查询参数
	 * @returns 节点列表响应
	 */
	async getNodes(params: NodeQueryParams = {}): Promise<NodeListResponse> {
		// 确保参数名称与后端一致：page和pageSize
		const queryParams = {
			...params,
			page: params?.page || 1,
			pageSize: params?.pageSize || 20
		};
		const response = await api.get('/nodes/nodes', { params: queryParams });
		return response.data.data;
	}

	/**
	 * 创建节点(手动添加)
	 * @param data 节点注册数据
	 * @returns 节点注册响应
	 */
	async createNode(data: NodeRegistrationRequest): Promise<NodeRegistrationResponse> {
		const response = await api.post('/nodes', data);
		return response.data.data;
	}

	/**
	 * 获取节点详情
	 * @param id 节点ID
	 * @returns 节点详情
	 */
	async getNode(id: string): Promise<Node> {
		const response = await api.get(`/nodes/${id}`);
		return response.data.data;
	}

	/**
	 * 更新节点信息
	 * @param id 节点ID
	 * @param data 更新数据
	 * @returns 更新后的节点信息
	 */
	async updateNode(id: string, data: NodeUpdateRequest): Promise<Node> {
		const response = await api.put(`/nodes/${id}`, data);
		return response.data.data;
	}

	/**
	 * 删除节点
	 * @param id 节点ID
	 */
	async deleteNode(id: string): Promise<void> {
		await api.delete(`/nodes/${id}`);
	}

	/**
	 * 更新节点状态
	 * @param id 节点ID
	 * @param data 状态数据
	 */
	async updateNodeStatus(id: string, data: NodeStatusUpdateRequest): Promise<void> {
		await api.put(`/nodes/${id}/status`, data);
	}

	/**
	 * 节点心跳
	 * @param id 节点ID
	 * @param heartbeat 心跳数据
	 */
	async nodeHeartbeat(id: string, heartbeat: NodeHeartbeat): Promise<void> {
		await api.post(`/nodes/${id}/heartbeat`, heartbeat);
	}

	/**
	 * 获取节点健康状态
	 * @param id 节点ID
	 * @returns 节点健康状态
	 */
	async getNodeHealth(id: string): Promise<NodeHealth> {
		const response = await api.get(`/nodes/${id}/health`);
		return response.data.data;
	}

	/**
	 * 更新节点配置
	 * @param id 节点ID
	 * @param config 节点配置
	 */
	async updateNodeConfig(id: string, config: NodeConfig): Promise<void> {
		await api.put(`/nodes/${id}/config`, config);
	}

	/**
	 * 获取节点配置
	 * @param id 节点ID
	 * @returns 节点配置
	 */
	async getNodeConfig(id: string): Promise<NodeConfig> {
		const response = await api.get(`/nodes/${id}/config`);
		return response.data.data;
	}

	/**
	 * 获取节点任务
	 * @param id 节点ID
	 * @returns 节点任务列表
	 */
	async getNodeTasks(id: string): Promise<any[]> {
		const response = await api.get(`/nodes/${id}/tasks`);
		return response.data.data;
	}

	/**
	 * 获取节点任务统计
	 * @param id 节点ID
	 * @returns 节点任务统计
	 */
	async getNodeTaskStats(id: string): Promise<NodeTaskStats> {
		const response = await api.get(`/nodes/${id}/task-stats`);
		return response.data.data;
	}

	/**
	 * 节点注册
	 * @param data 注册数据
	 * @returns 注册响应
	 */
	async registerNode(data: NodeRegistrationRequest): Promise<NodeRegistrationResponse> {
		const response = await api.post('/nodes/register', data);
		return response.data.data;
	}

	/**
	 * 节点注销
	 * @param id 节点ID
	 */
	async unregisterNode(id: string): Promise<void> {
		await api.post(`/nodes/unregister/${id}`);
	}

	/**
	 * 按状态获取节点
	 * @param status 节点状态
	 * @returns 节点列表
	 */
	async getNodesByStatus(status: NodeStatusType): Promise<{ nodes: Node[]; total: number }> {
		const response = await api.get(`/nodes/status/${status}`);
		return response.data.data;
	}

	/**
	 * 按角色获取节点
	 * @param role 节点角色
	 * @returns 节点列表
	 */
	async getNodesByRole(role: NodeRoleType): Promise<{ nodes: Node[]; total: number }> {
		const response = await api.get(`/nodes/role/${role}`);
		return response.data.data;
	}

	/**
	 * 按标签获取节点
	 * @param tag 标签
	 * @returns 节点列表
	 */
	async getNodesByTag(tag: string): Promise<{ nodes: Node[]; total: number }> {
		const response = await api.get(`/nodes/tags/${tag}`);
		return response.data.data;
	}

	/**
	 * 获取节点统计
	 * @returns 节点统计信息
	 */
	async getNodeStats(): Promise<NodeStats> {
		const response = await api.get('/nodes/nodes/stats');
		return response.data.data;
	}

	/**
	 * 批量操作节点
	 * @param data 批量操作数据
	 */
	async batchOperation(data: NodeBatchOperationRequest): Promise<void> {
		await api.post('/nodes/batch', data);
	}

	/**
	 * 获取节点监控信息
	 * @returns 节点监控信息
	 */
	async getNodeMonitor(): Promise<any> {
		const response = await api.get('/nodes/monitor');
		return response.data.data;
	}

	/**
	 * 获取节点事件
	 * @returns 节点事件列表
	 */
	async getNodeEvents(): Promise<any[]> {
		const response = await api.get('/nodes/events');
		return response.data.data;
	}

	/**
	 * 设置维护模式
	 * @param id 节点ID
	 * @param data 维护模式数据
	 */
	async setMaintenanceMode(id: string, data: NodeMaintenanceRequest): Promise<void> {
		await api.post(`/nodes/maintenance/${id}`, data);
	}

	/**
	 * 清理离线节点
	 * @param data 清理参数
	 * @returns 清理结果
	 */
	async cleanupOfflineNodes(data: NodeCleanupRequest): Promise<NodeCleanupResponse> {
		const response = await api.post('/nodes/cleanup', data);
		return response.data.data;
	}

	/**
	 * 批量删除节点
	 * @param nodeIds 节点ID列表
	 */
	async batchDeleteNodes(nodeIds: string[]): Promise<void> {
		await this.batchOperation({
			action: 'delete',
			nodeIds
		});
	}

	/**
	 * 批量更新节点状态
	 * @param nodeIds 节点ID列表
	 * @param status 新状态
	 */
	async batchUpdateNodeStatus(nodeIds: string[], status: NodeStatusType): Promise<void> {
		await this.batchOperation({
			action: 'updateStatus',
			nodeIds,
			data: { status }
		});
	}

	/**
	 * 启用节点
	 * @param id 节点ID
	 */
	async enableNode(id: string): Promise<void> {
		await this.updateNodeStatus(id, { status: 'online' });
	}

	/**
	 * 禁用节点
	 * @param id 节点ID
	 */
	async disableNode(id: string): Promise<void> {
		await this.updateNodeStatus(id, { status: 'disabled' });
	}

	/**
	 * 设置节点为维护模式
	 * @param id 节点ID
	 * @param reason 维护原因
	 */
	async setNodeMaintenance(id: string, reason?: string): Promise<void> {
		await this.setMaintenanceMode(id, {
			maintenance: true,
			reason
		});
	}

	/**
	 * 取消节点维护模式
	 * @param id 节点ID
	 */
	async cancelNodeMaintenance(id: string): Promise<void> {
		await this.setMaintenanceMode(id, {
			maintenance: false
		});
	}

	/**
	 * 获取在线节点
	 * @returns 在线节点列表
	 */
	async getOnlineNodes(): Promise<{ nodes: Node[]; total: number }> {
		return this.getNodesByStatus('online');
	}

	/**
	 * 获取离线节点
	 * @returns 离线节点列表
	 */
	async getOfflineNodes(): Promise<{ nodes: Node[]; total: number }> {
		return this.getNodesByStatus('offline');
	}

	/**
	 * 获取主节点
	 * @returns 主节点列表
	 */
	async getMasterNodes(): Promise<{ nodes: Node[]; total: number }> {
		return this.getNodesByRole('master');
	}

	/**
	 * 获取工作节点
	 * @returns 工作节点列表
	 */
	async getWorkerNodes(): Promise<{ nodes: Node[]; total: number }> {
		return this.getNodesByRole('worker');
	}

	/**
	 * 搜索节点
	 * @param query 搜索关键词
	 * @param filters 过滤条件
	 * @returns 搜索结果
	 */
	async searchNodes(
		query: string,
		filters: Partial<NodeQueryParams> = {}
	): Promise<NodeListResponse> {
		return this.getNodes({
			search: query,
			...filters
		});
	}

	/**
	 * 获取节点详细监控数据
	 * @param id 节点ID
	 * @param timeRange 时间范围
	 * @returns 监控数据
	 */
	async getNodeMonitorData(id: string, timeRange: string = '24h'): Promise<any> {
		const response = await api.get(`/nodes/${id}/monitor`, {
			params: { timeRange }
		});
		return response.data.data;
	}

	/**
	 * 获取节点日志
	 * @param id 节点ID
	 * @param params 日志查询参数
	 * @returns 日志数据
	 */
	async getNodeLogs(id: string, params: any = {}): Promise<any> {
		const response = await api.get(`/nodes/${id}/logs`, { params });
		return response.data.data;
	}

	/**
	 * 重启节点
	 * @param id 节点ID
	 */
	async restartNode(id: string): Promise<void> {
		await api.post(`/nodes/${id}/restart`);
	}

	/**
	 * 停止节点
	 * @param id 节点ID
	 */
	async stopNode(id: string): Promise<void> {
		await api.post(`/nodes/${id}/stop`);
	}

	/**
	 * 启动节点
	 * @param id 节点ID
	 */
	async startNode(id: string): Promise<void> {
		await api.post(`/nodes/${id}/start`);
	}

	/**
	 * 获取节点性能报告
	 * @param id 节点ID
	 * @param timeRange 时间范围
	 * @returns 性能报告
	 */
	async getNodePerformanceReport(id: string, timeRange: string = '24h'): Promise<any> {
		const response = await api.get(`/nodes/${id}/performance`, {
			params: { timeRange }
		});
		return response.data.data;
	}

	/**
	 * 测试节点连接
	 * @param id 节点ID
	 * @returns 连接测试结果
	 */
	async testNodeConnection(id: string): Promise<any> {
		const response = await api.post(`/nodes/${id}/test`);
		return response.data.data;
	}

	/**
	 * 获取节点版本信息
	 * @param id 节点ID
	 * @returns 版本信息
	 */
	async getNodeVersion(id: string): Promise<any> {
		const response = await api.get(`/nodes/${id}/version`);
		return response.data.data;
	}

	/**
	 * 更新节点版本
	 * @param id 节点ID
	 * @param version 目标版本
	 */
	async updateNodeVersion(id: string, version: string): Promise<void> {
		await api.post(`/nodes/${id}/update`, { version });
	}

	/**
	 * 获取节点配置模板
	 * @param role 节点角色
	 * @returns 配置模板
	 */
	async getNodeConfigTemplate(role: NodeRoleType): Promise<NodeConfig> {
		const response = await api.get(`/nodes/config-template/${role}`);
		return response.data.data;
	}

	/**
	 * 验证节点配置
	 * @param config 节点配置
	 * @returns 验证结果
	 */
	async validateNodeConfig(config: NodeConfig): Promise<any> {
		const response = await api.post('/nodes/validate-config', config);
		return response.data.data;
	}

	/**
	 * 导出节点配置
	 * @param id 节点ID
	 * @returns 配置文件内容
	 */
	async exportNodeConfig(id: string): Promise<string> {
		const response = await api.get(`/nodes/${id}/export-config`);
		return response.data.data;
	}

	/**
	 * 导入节点配置
	 * @param id 节点ID
	 * @param configData 配置数据
	 */
	async importNodeConfig(id: string, configData: string): Promise<void> {
		await api.post(`/nodes/${id}/import-config`, { configData });
	}

	/**
	 * 获取节点告警
	 * @param id 节点ID
	 * @returns 告警列表
	 */
	async getNodeAlerts(id: string): Promise<any[]> {
		const response = await api.get(`/nodes/${id}/alerts`);
		return response.data.data;
	}

	/**
	 * 确认节点告警
	 * @param id 节点ID
	 * @param alertId 告警ID
	 */
	async acknowledgeNodeAlert(id: string, alertId: string): Promise<void> {
		await api.post(`/nodes/${id}/alerts/${alertId}/acknowledge`);
	}

	/**
	 * 获取节点资源使用历史
	 * @param id 节点ID
	 * @param timeRange 时间范围
	 * @returns 资源使用历史
	 */
	async getNodeResourceHistory(id: string, timeRange: string = '24h'): Promise<any> {
		const response = await api.get(`/nodes/${id}/resource-history`, {
			params: { timeRange }
		});
		return response.data.data;
	}

	/**
	 * 获取节点网络信息
	 * @param id 节点ID
	 * @returns 网络信息
	 */
	async getNodeNetworkInfo(id: string): Promise<any> {
		const response = await api.get(`/nodes/${id}/network`);
		return response.data.data;
	}

	/**
	 * 获取节点系统信息
	 * @param id 节点ID
	 * @returns 系统信息
	 */
	async getNodeSystemInfo(id: string): Promise<any> {
		const response = await api.get(`/nodes/${id}/system`);
		return response.data.data;
	}
}

// 创建节点API实例
export const nodeAPI = new NodeAPI();

// 导出默认实例
export default nodeAPI;
