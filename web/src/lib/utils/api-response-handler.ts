/**
 * API响应数据处理工具
 * 统一处理不同API返回的数据结构格式
 */

export interface StandardApiResponse<T = any> {
  code?: number;
  message?: string;
  data: T;
  success?: boolean;
}

export interface PaginatedApiResponse<T = any> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages?: number;
}

export interface ApiResponseWrapper<T = any> {
  code: number;
  message: string;
  data: T;
  success: boolean;
}

/**
 * 统一API响应格式处理器
 */
export class ApiResponseHandler {
  /**
   * 提取标准响应数据
   * 支持多种响应格式：
   * 1. 标准格式: { code: 200, message: 'success', data: { items: [], total: 0 } }
   * 2. 嵌套格式: { code: 200, data: { data: { items: [], total: 0 } } }
   * 3. 简化格式: { items: [], total: 0 }
   * 4. 数组格式: []
   */
  static extractData<T = any>(response: any): T {
    if (!response) {
      return response;
    }

    // 处理 null 或 undefined
    if (response === null || response === undefined) {
      return response;
    }

    // 处理字符串或数字等基本类型
    if (typeof response !== 'object') {
      return response;
    }

    // 处理后端项目列表API的特殊格式: response.data.data.result
    if (response.data && response.data.data && response.data.data.result !== undefined) {
      // 如果是项目列表格式，直接返回result
      if (typeof response.data.data.result === 'object') {
        return response.data.data.result as T;
      }
      return response.data.data.result;
    }

    // 处理标准格式: response.data.data.result
    if (response.code !== undefined && response.data && response.data.data && response.data.data.result !== undefined) {
      return response.data.data.result;
    }

    // 标准格式: response.data.data
    if (response.data && response.data.data !== undefined) {
      return response.data.data;
    }

    // 标准格式: response.data
    if (response.data !== undefined) {
      return response.data;
    }

    // 如果是数组或对象，直接返回
    return response;
  }

  /**
   * 提取分页数据
   */
  static extractPaginatedData<T = any>(response: any): PaginatedApiResponse<T> {
    const data = ApiResponseHandler.extractData(response);
    
    // 如果已经是分页格式
    if (data && typeof data === 'object' && 'data' in data && Array.isArray(data.data)) {
      return {
        data: data.data,
        total: data.total || data.totalCount || data.count || 0,
        page: data.page || data.pageIndex || data.currentPage || 1,
        limit: data.limit || data.pageSize || data.size || 20,
        totalPages: data.totalPages || Math.ceil((data.total || 0) / (data.limit || 20))
      };
    }

    // 处理后端按标签分组的格式: { result: { tag1: [...], tag2: [...] }, tag: {...} }
    if (data && typeof data === 'object' && 'result' in data && typeof data.result === 'object') {
      // 合并所有标签下的项目到一个数组，并去重
      const allProjects: T[] = [];
      let totalCount = 0;
      const seenIds = new Set<string>();
      
      for (const tag of Object.keys(data.result)) {
        const projectsInTag = data.result[tag];
        if (Array.isArray(projectsInTag)) {
          for (const project of projectsInTag) {
            // 检查项目是否有ID并且是否重复
            const projectId = (project as any).id || (project as any)._id;
            if (projectId && !seenIds.has(projectId)) {
              seenIds.add(projectId);
              allProjects.push(project);
            } else if (!projectId) {
              // 如果没有ID，仍然添加（可能是测试数据）
              allProjects.push(project);
            }
          }
        }
        // 如果有 tag 统计信息，使用它来计算总数
        if (data.tag && typeof data.tag === 'object' && data.tag[tag]) {
          totalCount += data.tag[tag];
        } else {
          totalCount += projectsInTag.length;
        }
      }
      
      return {
        data: allProjects,
        total: totalCount,
        page: 1,
        limit: Math.max(allProjects.length, 20),
        totalPages: Math.max(1, Math.ceil(totalCount / 20))
      };
    }

    // 如果是数组，创建分页格式
    if (Array.isArray(data)) {
      return {
        data,
        total: data.length,
        page: 1,
        limit: data.length || 20,
        totalPages: 1
      };
    }

    // 如果是对象，尝试提取数据
    if (data && typeof data === 'object') {
      const items = data.items || data.list || data.records || data.result || [];
      return {
        data: Array.isArray(items) ? items : [],
        total: data.total || data.totalCount || data.count || (Array.isArray(items) ? items.length : 0),
        page: data.page || data.pageIndex || data.currentPage || 1,
        limit: data.limit || data.pageSize || data.size || 20,
        totalPages: data.totalPages || Math.ceil((data.total || 0) / (data.limit || 20))
      };
    }

    // 默认返回空分页
    return {
      data: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0
    };
  }

  /**
   * 检查响应是否成功
   */
  static isSuccess(response: any): boolean {
    if (!response) return false;
    
    // 检查 code 字段
    if (response.code !== undefined) {
      return response.code === 200 || response.code === 0 || response.code === '200';
    }
    
    // 检查 success 字段
    if (response.success !== undefined) {
      return response.success === true;
    }
    
    // 检查 HTTP 状态码
    if (response.status) {
      return response.status >= 200 && response.status < 300;
    }
    
    return true;
  }

  /**
   * 提取错误信息
   */
  static extractError(response: any): string {
    if (!response) return '未知错误';
    
    if (response.message) return response.message;
    if (response.msg) return response.msg;
    if (response.error) return response.error;
    if (response.error_message) return response.error_message;
    if (response.description) return response.description;
    if (response.detail) return response.detail;
    
    return '请求处理失败';
  }

  /**
   * 包装响应数据为标准格式
   */
  static wrapResponse<T = any>(data: any): StandardApiResponse<T> {
    return {
      code: 200,
      message: 'success',
      data: data,
      success: true
    };
  }

  /**
   * 标准化错误响应
   */
  static wrapError(message: string, code: number = 500): StandardApiResponse<null> {
    return {
      code,
      message,
      data: null,
      success: false
    };
  }
}

/**
 * 快捷函数 - 非静态版本
 */
export function handleApiResponse<T = any>(response: any): T {
  return ApiResponseHandler.extractData(response);
}

export function handlePaginatedResponse<T = any>(response: any) {
  return ApiResponseHandler.extractPaginatedData<T>(response);
}

export function isApiSuccess(response: any): boolean {
  return ApiResponseHandler.isSuccess(response);
}

export function getApiError(response: any): string {
  return ApiResponseHandler.extractError(response);
}

// 默认导出
export default ApiResponseHandler;