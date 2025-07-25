// 模拟API响应处理器的问题调试

// 模拟后端返回的数据（根据后端日志）
const mockResponse = {
  code: 200,
  data: {
    data: [
      {id: '1', name: '项目1'},
      {id: '2', name: '项目2'},
      {id: '3', name: '项目3'},
      {id: '4', name: '项目4'},
      {id: '5', name: '项目5'},
      {id: '6', name: '项目6'},
      {id: '7', name: '项目7'},
      {id: '8', name: '项目8'},
      {id: '9', name: '项目9'},
      {id: '10', name: '项目10'}
    ],
    total: 12,
    page: 1,
    limit: 10
  }
};

// 模拟extractData方法
function extractData(response) {
  if (!response) {
    return response;
  }

  if (response === null || response === undefined) {
    return response;
  }

  if (typeof response !== 'object') {
    return response;
  }

  // 标准格式: response.data
  if (response.data !== undefined) {
    return response.data;
  }

  return response;
}

// 模拟extractPaginatedData方法
function extractPaginatedData(response) {
  const data = extractData(response);
  
  console.log('🎯 数据提取结果:', data);
  console.log('🔍 判断条件检查:', {
    是对象: data && typeof data === 'object',
    有data字段: data && 'data' in data,
    data是数组: data && 'data' in data && Array.isArray(data.data),
    有total字段: data && ('total' in data || 'totalCount' in data || 'count' in data)
  });
  
  // 标准分页格式判断
  if (data && typeof data === 'object' && 'data' in data && Array.isArray(data.data) && 
      ('total' in data || 'totalCount' in data || 'count' in data)) {
    console.log('✅ 匹配标准分页格式');
    const total = data.total || data.totalCount || data.count || 0;
    const limit = data.limit || data.pageSize || data.size || 20;
    const result = {
      data: data.data,
      total: total,
      page: data.page || data.pageIndex || data.currentPage || 1,
      limit: limit,
      totalPages: data.totalPages || Math.ceil(total / limit)
    };
    console.log('🔧 标准分页处理结果:', result);
    return result;
  }
  
  console.log('❌ 未匹配标准分页格式');
  return { data: [], total: 0, page: 1, limit: 20, totalPages: 0 };
}

// 测试
console.log('=== 测试分页问题 ===');
const result = extractPaginatedData(mockResponse);
console.log('最终结果:', result);
console.log('总数应该是12，实际是:', result.total);
console.log('页数应该是2 (12/10=1.2向上取整)，实际是:', result.totalPages);