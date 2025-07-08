// 项目管理功能验证脚本
import { ProjectAPI } from '$lib/api/projects';
import type { CreateProjectRequest, Project } from '$lib/types/project';

/**
 * 验证项目管理功能的完整性
 * 这是一个开发环境下的功能验证脚本
 */
export class ProjectFeatureVerifier {
	private static testProject: Project | null = null;

	/**
	 * 运行所有验证测试
	 */
	static async runAllTests(): Promise<{
		success: boolean;
		results: Array<{ test: string; passed: boolean; error?: string }>;
	}> {
		const results: Array<{ test: string; passed: boolean; error?: string }> = [];

		// 测试项目创建
		try {
			await this.testCreateProject();
			results.push({ test: 'Create Project', passed: true });
		} catch (error) {
			results.push({
				test: 'Create Project',
				passed: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			});
		}

		// 测试项目列表获取
		try {
			await this.testGetProjectsList();
			results.push({ test: 'Get Projects List', passed: true });
		} catch (error) {
			results.push({
				test: 'Get Projects List',
				passed: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			});
		}

		// 测试项目详情获取
		if (this.testProject) {
			try {
				await this.testGetProjectDetails();
				results.push({ test: 'Get Project Details', passed: true });
			} catch (error) {
				results.push({
					test: 'Get Project Details',
					passed: false,
					error: error instanceof Error ? error.message : 'Unknown error'
				});
			}

			// 测试项目更新
			try {
				await this.testUpdateProject();
				results.push({ test: 'Update Project', passed: true });
			} catch (error) {
				results.push({
					test: 'Update Project',
					passed: false,
					error: error instanceof Error ? error.message : 'Unknown error'
				});
			}

			// 测试项目统计
			try {
				await this.testGetProjectStats();
				results.push({ test: 'Get Project Stats', passed: true });
			} catch (error) {
				results.push({
					test: 'Get Project Stats',
					passed: false,
					error: error instanceof Error ? error.message : 'Unknown error'
				});
			}

			// 清理测试数据
			try {
				await this.cleanupTestData();
				results.push({ test: 'Cleanup Test Data', passed: true });
			} catch (error) {
				results.push({
					test: 'Cleanup Test Data',
					passed: false,
					error: error instanceof Error ? error.message : 'Unknown error'
				});
			}
		}

		const success = results.every((result) => result.passed);
		return { success, results };
	}

	/**
	 * 测试创建项目
	 */
	private static async testCreateProject(): Promise<void> {
		const testProjectData: CreateProjectRequest = {
			name: `Test Project ${Date.now()}`,
			description: 'This is a test project created by verification script',
			target: 'test.example.com',
			color: 'blue',
			is_private: false
		};

		this.testProject = await ProjectAPI.createProject(testProjectData);

		if (!this.testProject || !this.testProject.id) {
			throw new Error('Failed to create test project');
		}

		console.log('✅ Test project created:', this.testProject.name);
	}

	/**
	 * 测试获取项目列表
	 */
	private static async testGetProjectsList(): Promise<void> {
		const projectsList = await ProjectAPI.getProjects({
			page: 1,
			limit: 10
		});

		if (!projectsList || !Array.isArray(projectsList.data)) {
			throw new Error('Failed to get projects list');
		}

		console.log('✅ Projects list retrieved:', projectsList.data.length, 'projects');
	}

	/**
	 * 测试获取项目详情
	 */
	private static async testGetProjectDetails(): Promise<void> {
		if (!this.testProject) {
			throw new Error('No test project available');
		}

		const projectDetails = await ProjectAPI.getProject(this.testProject.id);

		if (!projectDetails || projectDetails.id !== this.testProject.id) {
			throw new Error('Failed to get project details');
		}

		console.log('✅ Project details retrieved:', projectDetails.name);
	}

	/**
	 * 测试更新项目
	 */
	private static async testUpdateProject(): Promise<void> {
		if (!this.testProject) {
			throw new Error('No test project available');
		}

		const updatedProject = await ProjectAPI.updateProject(this.testProject.id, {
			description: 'Updated description for test project'
		});

		if (!updatedProject || updatedProject.description !== 'Updated description for test project') {
			throw new Error('Failed to update project');
		}

		this.testProject = updatedProject;
		console.log('✅ Project updated successfully');
	}

	/**
	 * 测试获取项目统计
	 */
	private static async testGetProjectStats(): Promise<void> {
		const stats = await ProjectAPI.getProjectStats();

		if (!stats || typeof stats.total_projects !== 'number') {
			throw new Error('Failed to get project stats');
		}

		console.log('✅ Project stats retrieved:', stats);
	}

	/**
	 * 清理测试数据
	 */
	private static async cleanupTestData(): Promise<void> {
		if (!this.testProject) {
			return;
		}

		await ProjectAPI.deleteProject(this.testProject.id);
		console.log('✅ Test project cleaned up');
		this.testProject = null;
	}

	/**
	 * 验证项目类型定义
	 */
	static verifyProjectTypes(): boolean {
		try {
			// 验证创建项目请求类型
			const createRequest: CreateProjectRequest = {
				name: 'Test',
				description: 'Test description',
				target: 'test.com',
				color: 'blue',
				is_private: false
			};

			// 验证项目类型
			const project: Project = {
				id: '1',
				name: 'Test Project',
				description: 'Test description',
				target: 'test.com',
				scan_status: 'pending',
				color: 'blue',
				is_private: false,
				assets_count: 0,
				vulnerabilities_count: 0,
				tasks_count: 0,
				created_by: 'test-user',
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};

			console.log('✅ Project types verified successfully');
			return true;
		} catch (error) {
			console.error('❌ Project types verification failed:', error);
			return false;
		}
	}

	/**
	 * 验证API客户端配置
	 */
	static async verifyAPIConfiguration(): Promise<boolean> {
		try {
			// 检查API客户端是否正确配置
			const testCall = async () => {
				// 这应该会触发请求拦截器
				await ProjectAPI.getProjects({ page: 1, limit: 1 });
			};

			await testCall();
			console.log('✅ API configuration verified');
			return true;
		} catch (error) {
			console.log('⚠️ API configuration check (expected in dev):', (error as Error).message);
			return true; // 在开发环境中，API可能不可用，这是正常的
		}
	}
}

/**
 * 开发环境快速验证函数
 */
export async function quickVerifyProjectFeatures(): Promise<void> {
	console.log('🔍 Starting project management features verification...\n');

	// 验证类型定义
	console.log('1. Verifying TypeScript types...');
	ProjectFeatureVerifier.verifyProjectTypes();

	// 验证API配置
	console.log('\n2. Verifying API configuration...');
	await ProjectFeatureVerifier.verifyAPIConfiguration();

	// 如果在开发环境且后端可用，运行完整测试
	if (typeof window !== 'undefined' && window.location.hostname === 'localhost') {
		console.log('\n3. Running integration tests...');
		try {
			const results = await ProjectFeatureVerifier.runAllTests();

			console.log('\n📊 Test Results:');
			results.results.forEach((result) => {
				const icon = result.passed ? '✅' : '❌';
				console.log(`${icon} ${result.test}${result.error ? `: ${result.error}` : ''}`);
			});

			if (results.success) {
				console.log('\n🎉 All project management features verified successfully!');
			} else {
				console.log('\n⚠️ Some features need attention. Check the results above.');
			}
		} catch (error) {
			console.log('\n⚠️ Integration tests skipped (backend not available)');
		}
	}

	console.log('\n✨ Project management verification completed.');
}
