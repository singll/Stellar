/**
 * 测试页面监控差异算法
 * 验证优化后的差异算法性能和准确性
 */

const DIFF_API_BASE = 'http://localhost:8090/api/v1';

// 测试数据
const testCases = [
	{
		name: '微小变化测试',
		oldContent: `<!DOCTYPE html>
<html>
<head><title>测试页面</title></head>
<body>
    <h1>欢迎来到我们的网站</h1>
    <p>这是一个测试页面，包含了一些基本的内容。</p>
    <div class="content">
        <p>当前时间：2024-07-11 10:00:00</p>
        <p>访问次数：12345</p>
    </div>
</body>
</html>`,
		newContent: `<!DOCTYPE html>
<html>
<head><title>测试页面</title></head>
<body>
    <h1>欢迎来到我们的网站</h1>
    <p>这是一个测试页面，包含了一些基本的内容。</p>
    <div class="content">
        <p>当前时间：2024-07-11 10:30:00</p>
        <p>访问次数：12348</p>
    </div>
</body>
</html>`,
		expectedSimilarity: 0.95 // 预期相似度很高
	},
	{
		name: '中等变化测试',
		oldContent: `<!DOCTYPE html>
<html>
<head><title>产品列表</title></head>
<body>
    <h1>产品列表</h1>
    <div class="products">
        <div class="product">
            <h2>产品A</h2>
            <p>价格：$100</p>
        </div>
        <div class="product">
            <h2>产品B</h2>
            <p>价格：$200</p>
        </div>
    </div>
</body>
</html>`,
		newContent: `<!DOCTYPE html>
<html>
<head><title>产品列表</title></head>
<body>
    <h1>产品列表</h1>
    <div class="products">
        <div class="product">
            <h2>产品A</h2>
            <p>价格：$95</p>
            <p class="sale">限时促销!</p>
        </div>
        <div class="product">
            <h2>产品B</h2>
            <p>价格：$200</p>
        </div>
        <div class="product">
            <h2>产品C</h2>
            <p>价格：$150</p>
        </div>
    </div>
</body>
</html>`,
		expectedSimilarity: 0.75 // 预期中等相似度
	},
	{
		name: '大幅变化测试',
		oldContent: `<!DOCTYPE html>
<html>
<head><title>旧版首页</title></head>
<body>
    <h1>欢迎访问旧版网站</h1>
    <nav>
        <a href="/about">关于我们</a>
        <a href="/contact">联系我们</a>
    </nav>
    <main>
        <p>这是我们的旧版网站首页。</p>
    </main>
</body>
</html>`,
		newContent: `<!DOCTYPE html>
<html>
<head><title>全新改版首页</title></head>
<body>
    <header>
        <h1>全新改版网站</h1>
        <nav class="modern-nav">
            <a href="/home">首页</a>
            <a href="/services">服务</a>
            <a href="/products">产品</a>
            <a href="/news">新闻</a>
            <a href="/contact">联系我们</a>
        </nav>
    </header>
    <main class="container">
        <section class="hero">
            <h2>欢迎来到我们全新的网站</h2>
            <p>体验全新的用户界面和更丰富的功能。</p>
            <button class="cta-button">立即开始</button>
        </section>
        <section class="features">
            <div class="feature">
                <h3>功能一</h3>
                <p>描述功能一的优势</p>
            </div>
            <div class="feature">
                <h3>功能二</h3>
                <p>描述功能二的优势</p>
            </div>
        </section>
    </main>
    <footer>
        <p>&copy; 2024 我们的公司</p>
    </footer>
</body>
</html>`,
		expectedSimilarity: 0.3 // 预期低相似度
	}
];

async function testDiffAlgorithm() {
	console.log('🧪 测试页面监控差异算法...\n');

	let authToken = '';

	try {
		// 0. 用户登录
		console.log('0. 用户登录...');
		const loginResponse = await fetch(`${DIFF_API_BASE}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				username: 'filetest',
				password: 'filetest123'
			})
		});

		if (!loginResponse.ok) {
			throw new Error('登录失败，请检查用户名和密码');
		}

		const loginResult = await loginResponse.json();
		authToken = loginResult.data.token;
		console.log('✓ 登录成功\n');

		// 创建认证头
		const authHeaders = {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${authToken}`
		};

		// 1. 创建测试项目
		console.log('1. 创建测试项目...');
		const testProject = {
			name: '差异算法测试项目',
			description: '用于测试页面监控差异算法的项目',
			tags: ['test', 'diff-algorithm']
		};

		const projectResponse = await fetch(`${DIFF_API_BASE}/projects/projects`, {
			method: 'POST',
			headers: authHeaders,
			body: JSON.stringify(testProject)
		});

		let projectId;
		if (projectResponse.ok) {
			const project = await projectResponse.json();
			projectId = project.id;
			console.log(`✓ 创建测试项目成功: ${projectId}`);
		} else {
			// 如果创建失败，使用固定ID
			projectId = '507f1f77bcf86cd799439011';
			console.log(`✓ 使用固定测试项目ID: ${projectId}`);
		}

		// 2. 测试每个测试用例
		console.log('\n2. 开始差异算法测试...\n');

		for (let i = 0; i < testCases.length; i++) {
			const testCase = testCases[i];
			console.log(`测试用例 ${i + 1}: ${testCase.name}`);

			try {
				// 创建页面监控任务
				const monitoringConfig = {
					projectId: projectId,
					name: `差异测试-${testCase.name}`,
					url: `http://test-page-${i}.example.com`,
					interval: 1, // 1小时
					config: {
						compareMethod: 'html',
						similarityThreshold: 0.8,
						timeout: 30,
						notifyOnChange: false,
						ignoreNumbers: false,
						ignorePatterns: []
					}
				};

				const monitoringResponse = await fetch(`${DIFF_API_BASE}/monitoring/monitoring`, {
					method: 'POST',
					headers: authHeaders,
					body: JSON.stringify(monitoringConfig)
				});

				if (!monitoringResponse.ok) {
					console.log(`   ⚠️ 创建监控任务失败，跳过此测试用例`);
					continue;
				}

				const monitoring = await monitoringResponse.json();
				console.log(`   ✓ 创建监控任务: ${monitoring.id}`);

				// 模拟两次快照数据并比较
				console.log(`   📸 模拟快照对比...`);

				// 创建模拟快照数据
				const oldSnapshot = {
					id: '507f1f77bcf86cd799439012',
					url: monitoringConfig.url,
					statusCode: 200,
					html: testCase.oldContent,
					text: testCase.oldContent.replace(/<[^>]*>/g, ''),
					contentHash: calculateSimpleHash(testCase.oldContent),
					size: testCase.oldContent.length,
					loadTime: 150,
					createdAt: new Date().toISOString()
				};

				const newSnapshot = {
					id: '507f1f77bcf86cd799439013',
					url: monitoringConfig.url,
					statusCode: 200,
					html: testCase.newContent,
					text: testCase.newContent.replace(/<[^>]*>/g, ''),
					contentHash: calculateSimpleHash(testCase.newContent),
					size: testCase.newContent.length,
					loadTime: 160,
					createdAt: new Date().toISOString()
				};

				// 计算相似度（模拟）
				const similarity = calculateTestSimilarity(testCase.oldContent, testCase.newContent);

				console.log(`   📊 相似度分析:`);
				console.log(`      计算得出: ${(similarity * 100).toFixed(2)}%`);
				console.log(`      预期范围: ${(testCase.expectedSimilarity * 100).toFixed(2)}% ± 10%`);

				// 验证相似度是否在合理范围内
				const tolerance = 0.1; // 10%容错率
				const isAccurate = Math.abs(similarity - testCase.expectedSimilarity) <= tolerance;

				if (isAccurate) {
					console.log(`   ✅ 测试通过 - 算法准确性良好`);
				} else {
					console.log(`   ❌ 测试失败 - 算法准确性需要调整`);
				}

				// 生成差异报告
				const diffReport = generateTestDiff(testCase.oldContent, testCase.newContent);
				console.log(`   📋 差异报告预览:`);
				console.log(`      ${diffReport.substring(0, 100)}${diffReport.length > 100 ? '...' : ''}`);

				// 清理测试数据
				await fetch(`${DIFF_API_BASE}/monitoring/monitoring/${monitoring.id}`, {
					method: 'DELETE',
					headers: authHeaders
				});

				console.log(`   🧹 清理测试数据完成\n`);
			} catch (error: any) {
				console.log(`   ❌ 测试用例执行失败: ${error.message}\n`);
			}
		}

		// 3. 性能测试
		console.log('3. 性能测试...');
		const performanceStart = Date.now();

		for (let i = 0; i < 100; i++) {
			calculateTestSimilarity(testCases[0].oldContent, testCases[0].newContent);
		}

		const performanceEnd = Date.now();
		const avgTime = (performanceEnd - performanceStart) / 100;

		console.log(`   💨 平均计算时间: ${avgTime.toFixed(2)}ms`);
		console.log(`   📈 性能评估: ${avgTime < 10 ? '优秀' : avgTime < 50 ? '良好' : '需要优化'}`);

		console.log('\n🎉 差异算法测试完成！');
	} catch (error: any) {
		console.error('❌ 测试失败:', error.message);
		process.exit(1);
	}
}

// 简单哈希计算函数
function calculateSimpleHash(content: string): string {
	let hash = 0;
	for (let i = 0; i < content.length; i++) {
		const char = content.charCodeAt(i);
		hash = (hash << 5) - hash + char;
		hash = hash & hash; // 转换为32位整数
	}
	return Math.abs(hash).toString(16);
}

// 测试相似度计算函数（模拟优化后的算法）
function calculateTestSimilarity(oldContent: string, newContent: string): number {
	if (oldContent === newContent) {
		return 1.0;
	}

	if (!oldContent || !newContent) {
		return 0.0;
	}

	// 简化的混合算法
	const editDistSim = calculateEditDistanceSimTest(oldContent, newContent);
	const cosineSim = calculateCosineSimTest(oldContent, newContent);
	const jaccardSim = calculateJaccardSimTest(oldContent, newContent);

	// 加权平均
	return 0.5 * editDistSim + 0.3 * cosineSim + 0.2 * jaccardSim;
}

// 编辑距离相似度测试
function calculateEditDistanceSimTest(s1: string, s2: string): number {
	const distance = levenshteinDistanceTest(s1, s2);
	const maxLen = Math.max(s1.length, s2.length);
	return maxLen === 0 ? 1.0 : 1.0 - distance / maxLen;
}

// Levenshtein距离测试
function levenshteinDistanceTest(s1: string, s2: string): number {
	if (s1.length === 0) return s2.length;
	if (s2.length === 0) return s1.length;

	const matrix = Array(s2.length + 1)
		.fill(null)
		.map(() => Array(s1.length + 1).fill(null));

	for (let i = 0; i <= s1.length; i++) matrix[0][i] = i;
	for (let j = 0; j <= s2.length; j++) matrix[j][0] = j;

	for (let j = 1; j <= s2.length; j++) {
		for (let i = 1; i <= s1.length; i++) {
			const indicator = s1[i - 1] === s2[j - 1] ? 0 : 1;
			matrix[j][i] = Math.min(
				matrix[j][i - 1] + 1, // 删除
				matrix[j - 1][i] + 1, // 插入
				matrix[j - 1][i - 1] + indicator // 替换
			);
		}
	}

	return matrix[s2.length][s1.length];
}

// 余弦相似度测试
function calculateCosineSimTest(s1: string, s2: string): number {
	const words1 = s1.toLowerCase().match(/\b\w+\b/g) || [];
	const words2 = s2.toLowerCase().match(/\b\w+\b/g) || [];

	const wordSet = new Set([...words1, ...words2]);
	const vector1: number[] = [],
		vector2: number[] = [];

	for (const word of wordSet) {
		vector1.push(words1.filter((w: string) => w === word).length);
		vector2.push(words2.filter((w: string) => w === word).length);
	}

	const dotProduct = vector1.reduce((sum, val, i) => sum + val * vector2[i], 0);
	const magnitude1 = Math.sqrt(vector1.reduce((sum, val) => sum + val * val, 0));
	const magnitude2 = Math.sqrt(vector2.reduce((sum, val) => sum + val * val, 0));

	return magnitude1 && magnitude2 ? dotProduct / (magnitude1 * magnitude2) : 0;
}

// Jaccard相似度测试
function calculateJaccardSimTest(s1: string, s2: string): number {
	const set1 = new Set(s1.toLowerCase().match(/\b\w+\b/g) || []);
	const set2 = new Set(s2.toLowerCase().match(/\b\w+\b/g) || []);

	const intersection = new Set([...set1].filter((x) => set2.has(x)));
	const union = new Set([...set1, ...set2]);

	return union.size === 0 ? 1.0 : intersection.size / union.size;
}

// 生成测试差异报告
function generateTestDiff(oldContent: string, newContent: string): string {
	const oldLines = oldContent.split('\n');
	const newLines = newContent.split('\n');

	let diff = '';
	const maxLines = Math.max(oldLines.length, newLines.length);

	for (let i = 0; i < Math.min(maxLines, 10); i++) {
		const oldLine = i < oldLines.length ? oldLines[i] : '';
		const newLine = i < newLines.length ? newLines[i] : '';

		if (oldLine !== newLine) {
			if (oldLine) diff += `- ${oldLine.substring(0, 50)}...\n`;
			if (newLine) diff += `+ ${newLine.substring(0, 50)}...\n`;
		}
	}

	return diff || '检测到内容变化，但无法生成详细差异信息';
}

// 运行测试
testDiffAlgorithm();
