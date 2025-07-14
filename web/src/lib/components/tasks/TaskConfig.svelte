<!--
任务配置组件
显示任务的配置信息
-->
<script lang="ts">
	import type { Task } from '$lib/types/task';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		task: Task;
	}

	interface ConfigItem {
		key: string;
		label: string;
		value: string;
		isSensitive: boolean;
		isObject: boolean;
	}

	let { task }: Props = $props();

	// 格式化配置值
	function formatConfigValue(value: any): string {
		if (value === null || value === undefined) {
			return '-';
		}

		if (typeof value === 'boolean') {
			return value ? '是' : '否';
		}

		if (Array.isArray(value)) {
			return value.length > 0 ? value.join(', ') : '空数组';
		}

		if (typeof value === 'object') {
			return JSON.stringify(value, null, 2);
		}

		return String(value);
	}

	// 获取配置字段的中文名称
	function getConfigFieldName(key: string): string {
		const fieldNames: Record<string, string> = {
			// 通用字段
			target: '目标',
			targets: '目标列表',
			domains: '域名列表',
			urls: 'URL列表',
			threads: '线程数',
			timeout: '超时时间',
			retry: '重试次数',
			output: '输出格式',
			wordlist: '字典文件',
			recursive: '递归扫描',
			depth: '扫描深度',
			delay: '延迟时间',
			user_agent: '用户代理',
			headers: '请求头',
			cookies: 'Cookie',
			proxy: '代理设置',
			follow_redirects: '跟随重定向',
			verify_ssl: '验证SSL',
			save_responses: '保存响应',
			extensions: '文件扩展名',
			status_codes: '状态码过滤',
			size_filter: '大小过滤',
			exclude_paths: '排除路径',
			include_paths: '包含路径',
			rate_limit: '速率限制',
			max_pages: '最大页面数',
			max_depth: '最大深度',
			crawl_forms: '爬取表单',
			crawl_javascript: '爬取JavaScript',
			screenshot: '截图',
			technologies: '技术识别',
			waf_detection: 'WAF检测',
			subdomain_takeover: '子域接管检测',
			dns_resolver: 'DNS解析器',
			wildcards: '通配符检测',
			brute_force: '暴力破解',
			port_range: '端口范围',
			top_ports: '常用端口',
			service_detection: '服务检测',
			version_detection: '版本检测',
			os_detection: '系统检测',
			script_scan: '脚本扫描',
			aggressive_scan: '激进扫描',
			ping_scan: 'Ping扫描',
			tcp_scan: 'TCP扫描',
			udp_scan: 'UDP扫描',
			syn_scan: 'SYN扫描',
			ack_scan: 'ACK扫描',
			window_scan: 'Window扫描',
			maimon_scan: 'Maimon扫描',
			null_scan: 'Null扫描',
			fin_scan: 'FIN扫描',
			xmas_scan: 'XMAS扫描',
			scan_delay: '扫描延迟',
			max_rate: '最大速率',
			max_parallelism: '最大并行度',
			max_hostgroup: '最大主机组',
			max_scan_delay: '最大扫描延迟',
			max_retries: '最大重试',
			host_timeout: '主机超时',
			initial_rtt_timeout: '初始RTT超时',
			max_rtt_timeout: '最大RTT超时',
			min_rtt_timeout: '最小RTT超时',
			min_parallelism: '最小并行度',
			min_hostgroup: '最小主机组',
			min_rate: '最小速率'
		};

		return fieldNames[key as keyof typeof fieldNames] || key;
	}

	// 判断是否为敏感信息
	function isSensitiveField(key: string): boolean {
		const sensitiveKeys = ['password', 'token', 'key', 'secret', 'auth', 'credential'];
		return sensitiveKeys.some((sensitive) => key.toLowerCase().includes(sensitive));
	}

	// 获取配置项列表
	let configItems = $derived(() => {
		if (!task.config || typeof task.config !== 'object') {
			return [] as ConfigItem[];
		}

		return Object.entries(task.config)
			.filter(([_, value]) => value !== null && value !== undefined)
			.map(
				([key, value]) =>
					({
						key,
						label: getConfigFieldName(key),
						value: formatConfigValue(value),
						isSensitive: isSensitiveField(key),
						isObject: typeof value === 'object' && !Array.isArray(value)
					}) as ConfigItem
			);
	});

	// 获取配置项数组，用于模板中的 each 循环
	let configItemsArray = $derived(configItems());
</script>

<Card>
	<CardHeader>
		<CardTitle>任务配置</CardTitle>
	</CardHeader>
	<CardContent>
		{#if configItemsArray.length === 0}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<i class="fas fa-cog text-2xl mb-2"></i>
				<p>暂无配置信息</p>
			</div>
		{:else}
			<div class="space-y-4">
				{#each configItemsArray as item}
					<div class="border-b border-gray-200 dark:border-gray-700 pb-3 last:border-b-0 last:pb-0">
						<div class="flex items-start justify-between">
							<div class="flex-1">
								<div class="flex items-center gap-2 mb-1">
									<span class="font-medium text-gray-900 dark:text-white">
										{item.label}
									</span>
									{#if item.isSensitive}
										<Badge variant="outline" class="text-xs">
											<i class="fas fa-lock mr-1"></i>
											敏感
										</Badge>
									{/if}
								</div>

								<div class="text-sm text-gray-600 dark:text-gray-400 mb-1">
									{item.key}
								</div>

								<div class="mt-2">
									{#if item.isSensitive}
										<div
											class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-2"
										>
											<div
												class="text-yellow-800 dark:text-yellow-200 text-sm flex items-center gap-2"
											>
												<i class="fas fa-eye-slash"></i>
												<span>敏感信息已隐藏</span>
											</div>
										</div>
									{:else if item.isObject}
										<div
											class="bg-gray-50 dark:bg-gray-900/20 border border-gray-200 dark:border-gray-700 rounded-md p-3"
										>
											<pre
												class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap overflow-x-auto font-mono">{item.value}</pre>
										</div>
									{:else}
										<div class="text-gray-900 dark:text-white">
											{#if item.value === '是'}
												<Badge variant="default">
													<i class="fas fa-check mr-1"></i>
													{item.value}
												</Badge>
											{:else if item.value === '否'}
												<Badge variant="secondary">
													<i class="fas fa-times mr-1"></i>
													{item.value}
												</Badge>
											{:else if item.value === '-'}
												<span class="text-gray-500 dark:text-gray-400">未设置</span>
											{:else}
												<span class="font-mono text-sm">{item.value}</span>
											{/if}
										</div>
									{/if}
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</CardContent>
</Card>
