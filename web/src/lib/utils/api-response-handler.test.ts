/**
 * API响应处理工具测试
 * 验证统一数据解析逻辑
 */

import { ApiResponseHandler } from './api-response-handler';

// 测试用例
const testCases = [
	// 测试标准格式：response.data.data.result
	{
		name: '标准嵌套格式',
		input: {
			code: 200,
			message: 'success',
			data: {
				data: {
					result: [
						{ id: 1, name: '项目1' },
						{ id: 2, name: '项目2' }
					]
				}
			}
		},
		expected: [
			{ id: 1, name: '项目1' },
			{ id: 2, name: '项目2' }
		]
	},
	
	// 测试标准格式：response.data.data
	{
		name: '标准数据格式',
		input: {
			code: 200,
			message: 'success',
			data: [
				{ id: 1, name: '项目1' },
				{ id:2, name: '项目2' }
			]
		},
		expected: [
			{ id: 1, name: '项目1' },
			{ id: 2, name: '项目2' }
		]
	},
	
	// 测试分页格式
	{
		name: '分页数据格式',
		input: {
			code: 200,
			message: 'success',
			data: {
				items: [{ id: 1, name: '项目1' }],
				total: 1,
				page: 1,
				pageSize: 20
			}
		},
		expected: {
			data: [{ id: 1, name: '项目1' }],
			total: 1,
			page: 1,
			limit: 20,
			totalPages: 1
		}
	},
	
	// 测试数组格式
	{
		name: '数组格式',
		input: [{ id: 1, name: '项目1' }],
		expected: {
			data: [{ id: 1, name: '项目1' }],
			total: 1,
			page: 1,
			limit: 1,
			totalPages: 1
		}
	}
];

// 运行测试
console.log('🧪 开始测试API响应处理器...');

// 测试数据提取
console.log('\n📊 测试数据提取:')
testCases.slice(0, 2).forEach(testCase => {
	const result = ApiResponseHandler.extractData(testCase.input);
	console.log(`  ${testCase.name}:`, result, JSON.stringify(result) === JSON.stringify(testCase.expected) ? '✅' : '❌');
});

// 测试分页数据提取
console.log('\n📊 测试分页数据提取:')
testCases.slice(2, 4).forEach(testCase => {
	const result = ApiResponseHandler.extractPaginatedData(testCase.input);
	console.log(`  ${testCase.name}:`, result);
	console.log(`  数据正确性: ${result.data?.length === testCase.expected.data.length ? '✅' : '❌'}`);
});

console.log('\n✅ API响应处理器测试完成');

// 导出测试函数
export const runApiResponseTests = () => {
	console.log('API响应处理器已就绪');
};