/**
 * 测试文件检测功能
 * 验证敏感信息检测系统是否能够正确处理文件目标
 */

const FILE_DETECTION_API_BASE = 'http://localhost:8090/api/v1';

async function testFileDetection() {
	console.log('🔍 测试文件检测功能...\n');

	let authToken = '';

	try {
		// 0. 用户登录
		console.log('0. 用户登录...');
		const loginResponse = await fetch(`${FILE_DETECTION_API_BASE}/auth/login`, {
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

		// 1. 获取或创建项目
		console.log('1. 获取项目信息...');
		// 使用一个固定的测试项目ID，这样可以避免项目创建的复杂性
		const projectId = '507f1f77bcf86cd799439011'; // 固定的测试项目ID
		console.log(`✓ 使用测试项目ID: ${projectId}`);

		// 2. 获取或创建敏感规则
		console.log('\n2. 获取或创建敏感规则...');
		let rules = [];

		try {
			const rulesResponse = await fetch(`${FILE_DETECTION_API_BASE}/sensitive/sensitive/rules`, {
				headers: authHeaders
			});

			if (rulesResponse.ok) {
				rules = await rulesResponse.json();
			}
		} catch (error) {
			console.log('获取敏感规则失败，将创建测试规则');
		}

		if (!rules || rules.length === 0) {
			// 创建测试敏感规则
			console.log('创建测试敏感规则...');
			const testRules = [
				{
					name: '密码检测',
					description: '检测常见的密码配置',
					type: 'regex',
					pattern: '(password|passwd|pwd)\\s*[=:]\\s*["\']?[^\\s"\';,\\}]{6,}["\']?',
					category: 'password',
					riskLevel: 'high',
					enabled: true,
					tags: ['password', 'config'],
					context: 3,
					examples: ['password = "mysecret123"', 'passwd: admin123'],
					falsePositivePatterns: ['password.*example', 'password.*test']
				},
				{
					name: 'API密钥检测',
					description: '检测API密钥',
					type: 'regex',
					pattern:
						'(api[_-]?key|secret[_-]?key|access[_-]?key)\\s*[=:]\\s*["\']?[a-zA-Z0-9]{16,}["\']?',
					category: 'api_key',
					riskLevel: 'high',
					enabled: true,
					tags: ['api', 'key'],
					context: 3,
					examples: ['api_key = "sk-1234567890abcdef"'],
					falsePositivePatterns: ['api_key.*placeholder']
				}
			];

			for (const rule of testRules) {
				const createRuleResponse = await fetch(
					`${FILE_DETECTION_API_BASE}/sensitive/sensitive/rules`,
					{
						method: 'POST',
						headers: authHeaders,
						body: JSON.stringify(rule)
					}
				);

				if (createRuleResponse.ok) {
					const newRule = await createRuleResponse.json();
					if (rules === null) rules = [];
					rules.push(newRule);
					console.log(`✓ 创建敏感规则: ${rule.name}`);
				} else {
					console.log(`⚠️ 创建敏感规则失败: ${rule.name}`);
				}
			}
		}

		console.log(`✓ 找到 ${rules.length} 个敏感规则`);

		// 3. 创建测试文件检测任务（支持文件目标）
		console.log('3. 创建文件检测任务...');
		const detectionRequest = {
			projectId: projectId,
			name: '文件检测测试',
			description: '测试敏感信息检测系统的文件检测功能',
			targets: [
				'/tmp/test.txt',
				'/var/log/*.log',
				'./config/*.yaml',
				'file:///home/user/documents/**/*.txt'
			],
			rules: rules.slice(0, 3).map((rule: any) => ({ $oid: rule.id })),
			config: {
				concurrency: 5,
				timeout: 30,
				maxDepth: 3,
				contextLines: 3,
				followLinks: false,
				userAgent: 'Stellar File Scanner',
				ignoreRobots: true,
				maxFileSize: 1024,
				fileTypes: 'txt,log,yaml,json',
				excludeURLs: '',
				includeURLs: '',
				authentication: '',
				// 文件检测特定配置
				filePatterns: ['*.txt', '*.log', '*.yaml', '*.json'],
				excludePatterns: ['*.tmp', '*/node_modules/*'],
				recursiveSearch: true,
				followSymlinks: false,
				maxFileSizeBytes: 10485760, // 10MB
				scanArchives: false
			}
		};

		const createResponse = await fetch(`${FILE_DETECTION_API_BASE}/sensitive/sensitive/scan`, {
			method: 'POST',
			headers: authHeaders,
			body: JSON.stringify(detectionRequest)
		});

		if (!createResponse.ok) {
			const error = await createResponse.text();
			throw new Error(`创建检测任务失败: ${error}`);
		}

		const detectionResult = await createResponse.json();
		const detectionId = detectionResult.id;
		console.log(`✓ 创建检测任务成功: ${detectionId}\n`);

		// 4. 监控检测状态
		console.log('4. 监控检测状态...');
		let attempts = 0;
		const maxAttempts = 10;

		while (attempts < maxAttempts) {
			const statusResponse = await fetch(
				`${FILE_DETECTION_API_BASE}/sensitive/sensitive/${detectionId}`,
				{
					headers: authHeaders
				}
			);
			if (!statusResponse.ok) {
				throw new Error('获取检测状态失败');
			}

			const status = await statusResponse.json();
			console.log(`   状态: ${status.status}, 进度: ${status.progress.toFixed(1)}%`);

			if (status.status === 'completed' || status.status === 'failed') {
				console.log(`✓ 检测完成，状态: ${status.status}\n`);

				// 显示检测结果
				if (status.findings && status.findings.length > 0) {
					console.log('5. 检测结果:');
					status.findings.forEach((finding: any, index: number) => {
						console.log(`   ${index + 1}. 目标: ${finding.target}`);
						console.log(`      类型: ${finding.targetType}`);
						console.log(`      规则: ${finding.ruleName}`);
						console.log(`      风险等级: ${finding.riskLevel}`);
						console.log(`      匹配文本: ${finding.matchedText}`);
						if (finding.lineNumber) {
							console.log(`      行号: ${finding.lineNumber}`);
						}
						if (finding.fileSize) {
							console.log(`      文件大小: ${finding.fileSize} bytes`);
						}
						console.log('');
					});
				} else {
					console.log('✓ 未发现敏感信息\n');
				}

				// 显示摘要
				if (status.summary) {
					console.log('6. 检测摘要:');
					console.log(`   总发现数: ${status.summary.totalFindings}`);
					console.log(`   风险等级分布: ${JSON.stringify(status.summary.riskLevelCount)}`);
					console.log(`   分类分布: ${JSON.stringify(status.summary.categoryCount)}`);
				}

				break;
			}

			await new Promise((resolve) => setTimeout(resolve, 2000));
			attempts++;
		}

		if (attempts >= maxAttempts) {
			console.log('⚠️ 检测超时');
		}

		// 5. 测试文件模式检测
		console.log('\n7. 测试文件模式检测...');
		const patternRequest = {
			projectId: projectId,
			name: '文件模式检测测试',
			description: '测试使用文件模式进行敏感信息检测',
			targets: ['**/*.{txt,log,yaml,json,config}', '/etc/*.conf', './src/**/*.{js,ts,go,py}'],
			rules: rules.slice(0, 2).map((rule: any) => ({ $oid: rule.id })),
			config: {
				...detectionRequest.config,
				filePatterns: ['**/*.{txt,log,yaml,json}'],
				recursiveSearch: true
			}
		};

		const patternResponse = await fetch(`${FILE_DETECTION_API_BASE}/sensitive/sensitive/scan`, {
			method: 'POST',
			headers: authHeaders,
			body: JSON.stringify(patternRequest)
		});

		if (patternResponse.ok) {
			const patternResult = await patternResponse.json();
			console.log(`✓ 文件模式检测任务创建成功: ${patternResult.id}`);
		} else {
			console.log('⚠️ 文件模式检测任务创建失败');
		}

		console.log('\n🎉 文件检测功能测试完成！');
	} catch (error: unknown) {
		console.error('❌ 测试失败:', (error as Error).message);
		process.exit(1);
	}
}

// 运行测试
testFileDetection();
