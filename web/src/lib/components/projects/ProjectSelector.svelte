<!--
项目选择器组件
提供项目选择功能
-->
<script lang="ts">
	import type { Project } from '$lib/types/project';
	import { Select } from '$lib/components/ui/select';

	interface Props {
		value?: string;
		projects: Project[];
		disabled?: boolean;
		class?: string;
		placeholder?: string;
	}

	let {
		value = $bindable(''),
		projects,
		disabled = false,
		class: className = '',
		placeholder = '请选择项目'
	}: Props = $props();

	// 转换项目数据为选择器选项
	let projectOptions = $derived([
		{ value: '', label: placeholder },
		...projects.map((project) => ({
			value: project.id,
			label: project.name
		}))
	]);

	function handleChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		value = target.value;
	}
</script>

<Select {value} options={projectOptions} {disabled} class={className} onselect={handleChange} />
