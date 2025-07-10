<!--
任务配置编辑器组件
根据任务类型显示相应的配置选项
-->
<script lang="ts">
	import type { TaskType } from '$lib/types/task';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Select } from '$lib/components/ui/select';
	import TagInput from '$lib/components/ui/TagInput.svelte';

	interface Props {
		config?: Record<string, any>;
		taskType: TaskType;
		disabled?: boolean;
		errors?: Record<string, string>;
	}

	let { config = $bindable({}), taskType, disabled = false, errors = {} }: Props = $props();

	// 根据任务类型返回配置字段
	function getConfigFields(type: TaskType) {
		switch (type) {
			case 'subdomain_enum':
				return [
					{
						key: 'target',
						label: '目标域名',
						type: 'text',
						placeholder: 'example.com',
						required: true,
						description: '要枚举子域名的目标域名'
					},
					{
						key: 'threads',
						label: '线程数',
						type: 'number',
						min: 1,
						max: 100,
						defaultValue: 10,
						description: '并发线程数量'
					},
					{
						key: 'timeout',
						label: '超时时间(秒)',
						type: 'number',
						min: 1,
						max: 300,
						defaultValue: 10,
						description: '单个请求的超时时间'
					},
					{
						key: 'wordlist',
						label: '字典文件',
						type: 'select',
						options: [
							{ value: 'small', label: '小字典 (1000条)' },
							{ value: 'medium', label: '中等字典 (10000条)' },
							{ value: 'large', label: '大字典 (100000条)' }
						],
						defaultValue: 'medium',
						description: '子域名爆破使用的字典大小'
					},
					{
						key: 'recursive',
						label: '递归扫描',
						type: 'checkbox',
						defaultValue: false,
						description: '是否对发现的子域名进行递归扫描'
					}
				];

			case 'port_scan':
				return [
					{
						key: 'target',
						label: '目标地址',
						type: 'text',
						placeholder: '192.168.1.1 或 example.com',
						required: true,
						description: '要扫描的目标IP地址或域名'
					},
					{
						key: 'port_range',
						label: '端口范围',
						type: 'text',
						placeholder: '1-65535 或 80,443,8080',
						defaultValue: '1-1000',
						description: '要扫描的端口范围或端口列表'
					},
					{
						key: 'scan_type',
						label: '扫描类型',
						type: 'select',
						options: [
							{ value: 'tcp', label: 'TCP扫描' },
							{ value: 'udp', label: 'UDP扫描' },
							{ value: 'syn', label: 'SYN扫描' }
						],
						defaultValue: 'tcp',
						description: '端口扫描的类型'
					},
					{
						key: 'threads',
						label: '线程数',
						type: 'number',
						min: 1,
						max: 1000,
						defaultValue: 100,
						description: '并发扫描线程数'
					},
					{
						key: 'timeout',
						label: '超时时间(秒)',
						type: 'number',
						min: 1,
						max: 60,
						defaultValue: 3,
						description: '端口连接超时时间'
					}
				];

			case 'vuln_scan':
				return [
					{
						key: 'target',
						label: '目标URL',
						type: 'text',
						placeholder: 'https://example.com',
						required: true,
						description: '要扫描漏洞的目标URL'
					},
					{
						key: 'scan_level',
						label: '扫描等级',
						type: 'select',
						options: [
							{ value: 'low', label: '低 - 快速扫描' },
							{ value: 'medium', label: '中 - 平衡扫描' },
							{ value: 'high', label: '高 - 深度扫描' }
						],
						defaultValue: 'medium',
						description: '扫描的深度和覆盖范围'
					},
					{
						key: 'modules',
						label: '扫描模块',
						type: 'tags',
						defaultValue: ['sql_injection', 'xss', 'directory_traversal'],
						description: '要启用的漏洞扫描模块'
					},
					{
						key: 'user_agent',
						label: 'User-Agent',
						type: 'text',
						placeholder: '自定义User-Agent',
						description: '请求使用的User-Agent字符串'
					}
				];

			case 'dir_scan':
				return [
					{
						key: 'target',
						label: '目标URL',
						type: 'text',
						placeholder: 'https://example.com',
						required: true,
						description: '要扫描目录的目标URL'
					},
					{
						key: 'wordlist',
						label: '字典文件',
						type: 'select',
						options: [
							{ value: 'common', label: '常用路径' },
							{ value: 'medium', label: '中等字典' },
							{ value: 'large', label: '大字典' }
						],
						defaultValue: 'common',
						description: '目录扫描使用的字典'
					},
					{
						key: 'extensions',
						label: '文件扩展名',
						type: 'tags',
						defaultValue: ['php', 'html', 'js', 'css'],
						description: '要扫描的文件扩展名'
					},
					{
						key: 'threads',
						label: '线程数',
						type: 'number',
						min: 1,
						max: 50,
						defaultValue: 10,
						description: '并发扫描线程数'
					},
					{
						key: 'status_codes',
						label: '状态码过滤',
						type: 'text',
						placeholder: '200,301,302,403',
						defaultValue: '200,301,302,403',
						description: '要记录的HTTP状态码'
					}
				];

			default:
				return [];
		}
	}

	let configFields = $derived(getConfigFields(taskType));

	// 使用 derived 计算初始化后的配置，避免 effect 无限循环
	let initializedConfig = $derived(() => {
		const newConfig = { ...config };
		
		configFields.forEach((field) => {
			if (newConfig[field.key] === undefined && field.defaultValue !== undefined) {
				newConfig[field.key] = field.defaultValue;
			}
		});
		
		return newConfig;
	});

	// 当初始化配置改变时，更新 config
	$effect(() => {
		const initialized = initializedConfig();
		const hasNewDefaults = configFields.some(field => 
			config[field.key] === undefined && field.defaultValue !== undefined
		);
		
		if (hasNewDefaults) {
			config = initialized;
		}
	});

	// 处理字段值变化
	function handleFieldChange(key: string, value: any) {
		config = { ...config, [key]: value };
	}
</script>

<div class="space-y-6">
	{#if configFields.length === 0}
		<div class="text-center py-8 text-gray-500 dark:text-gray-400">
			<i class="fas fa-cog text-2xl mb-2"></i>
			<p>该任务类型暂无配置选项</p>
		</div>
	{:else}
		{#each configFields as field}
			<div class="space-y-2">
				<label
					for="config-{field.key}"
					class="block text-sm font-medium text-gray-700 dark:text-gray-300"
				>
					{field.label}
					{#if field.required}
						<span class="text-red-500">*</span>
					{/if}
				</label>

				{#if field.description}
					<p class="text-sm text-gray-600 dark:text-gray-400">{field.description}</p>
				{/if}

				<div>
					{#if field.type === 'text'}
						<Input
							id="config-{field.key}"
							value={config[field.key] || ''}
							placeholder={field.placeholder}
							{disabled}
							class={errors[`config.${field.key}`] ? 'border-red-500' : ''}
							onchange={(e) => handleFieldChange(field.key, e.target.value)}
						/>
					{:else if field.type === 'number'}
						<Input
							id="config-{field.key}"
							type="number"
							value={config[field.key] || field.defaultValue || ''}
							min={field.min}
							max={field.max}
							{disabled}
							class={errors[`config.${field.key}`] ? 'border-red-500' : ''}
							onchange={(e) => handleFieldChange(field.key, parseInt(e.target.value))}
						/>
					{:else if field.type === 'textarea'}
						<Textarea
							id="config-{field.key}"
							value={config[field.key] || ''}
							placeholder={field.placeholder}
							{disabled}
							class={errors[`config.${field.key}`] ? 'border-red-500' : ''}
							onchange={(e) => handleFieldChange(field.key, e.target.value)}
						/>
					{:else if field.type === 'select'}
						<Select
							id="config-{field.key}"
							value={config[field.key] || field.defaultValue || ''}
							options={field.options}
							{disabled}
							class={errors[`config.${field.key}`] ? 'border-red-500' : ''}
							onchange={(value) => handleFieldChange(field.key, value)}
						/>
					{:else if field.type === 'checkbox'}
						<div class="flex items-center">
							<input
								id="config-{field.key}"
								type="checkbox"
								checked={config[field.key] || field.defaultValue || false}
								{disabled}
								class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
								onchange={(e) => handleFieldChange(field.key, e.target.checked)}
							/>
							<span class="ml-2 text-sm text-gray-600 dark:text-gray-400">启用</span>
						</div>
					{:else if field.type === 'tags'}
						<TagInput
							tags={config[field.key] || field.defaultValue || []}
							{disabled}
							placeholder="添加标签"
							class={errors[`config.${field.key}`] ? 'border-red-500' : ''}
							onchange={(tags) => handleFieldChange(field.key, tags)}
						/>
					{/if}
				</div>

				{#if errors[`config.${field.key}`]}
					<p class="text-sm text-red-600 dark:text-red-400">
						{errors[`config.${field.key}`]}
					</p>
				{/if}
			</div>
		{/each}
	{/if}
</div>
