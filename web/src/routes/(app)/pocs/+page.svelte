<script lang="ts">
	import { onMount } from 'svelte';
	import POCManagement from '$lib/components/vulnerability/POCManagement.svelte';
	import POCDialog from '$lib/components/vulnerability/POCDialog.svelte';
	import { vulnStore, selectedPOC } from '$lib/stores/vulnerability';
	import type { POC } from '$lib/types/vulnerability';

	// 组件状态
	let showCreateDialog = $state(false);
	let showEditDialog = $state(false);
	let currentPOC = $state<POC | null>(null);

	// 监听选中的POC变化
	$effect(() => {
		const unsubscribe = selectedPOC.subscribe((poc) => {
			currentPOC = poc;
		});
		return unsubscribe;
	});

	// 创建POC
	function createPOC() {
		currentPOC = null;
		vulnStore.selectPOC(null);
		showCreateDialog = true;
	}

	// 编辑POC
	function editPOC(poc: POC) {
		currentPOC = poc;
		vulnStore.selectPOC(poc);
		showEditDialog = true;
	}

	// 保存成功回调
	function handleSave(poc: POC) {
		showCreateDialog = false;
		showEditDialog = false;
		currentPOC = null;
		vulnStore.selectPOC(null);
		// 重新加载列表
		vulnStore.loadPOCs();
	}

	// 取消回调
	function handleCancel() {
		showCreateDialog = false;
		showEditDialog = false;
		currentPOC = null;
		vulnStore.selectPOC(null);
	}
</script>

<svelte:head>
	<title>POC管理 - Stellar</title>
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<POCManagement onCreatePOC={createPOC} onEditPOC={editPOC} />

	<!-- 创建POC对话框 -->
	<POCDialog
		poc={null}
		bind:isOpen={showCreateDialog}
		onSave={handleSave}
		onCancel={handleCancel}
	/>

	<!-- 编辑POC对话框 -->
	<POCDialog
		poc={currentPOC}
		bind:isOpen={showEditDialog}
		onSave={handleSave}
		onCancel={handleCancel}
	/>
</div>
