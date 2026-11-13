<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import { api } from '@/api/client';
import { useShopStore } from '@/stores/useShopStore';

const store = useShopStore();

const importing = ref(false);
const exporting = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);

async function handleExport() {
  exporting.value = true;
  try {
    const blob = await api.exportSnapshot();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    const now = new Date();
    const datePart = now.toISOString().slice(0, 10);
    const timePart = `${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`;
    a.href = url;
    a.download = `onestarlab-backup-${datePart}-${timePart}.json`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success('数据已导出');
  } catch (e) {
    ElMessage.error('导出失败，请重试');
  } finally {
    exporting.value = false;
  }
}

async function handleImport(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file) return;

  try {
    await ElMessageBox.confirm(
      '导入会覆盖服务端数据库中的全部数据，建议先导出备份。确认继续？',
      '覆盖确认',
      { type: 'warning', confirmButtonText: '确认导入', cancelButtonText: '取消' },
    );
  } catch {
    // 用户取消
    if (fileInputRef.value) fileInputRef.value.value = '';
    return;
  }

  importing.value = true;
  try {
    const text = await file.text();
    const snapshot = JSON.parse(text) as unknown;
    const result = await api.importSnapshot(snapshot);
    await store.loadAll();
    ElMessage.success(
      result.source === 'dexie'
        ? '旧版备份已自动转换并导入，页面已刷新'
        : '数据导入成功，页面已刷新',
    );
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '导入失败，请检查文件格式');
  } finally {
    importing.value = false;
    if (fileInputRef.value) fileInputRef.value.value = '';
  }
}
</script>

<template>
  <div class="page-stack">
    <section class="page-hero">
      <div>
        <div class="hero-kicker">数据中心 · 备份与恢复</div>
        <h2 class="hero-title">掌握你的服务端数据</h2>
        <p class="hero-description">
          所有数据保存在 Go 服务内置的 SQLite 数据库中，建议定期导出备份，在更换设备或重建容器前请先导出。
        </p>
      </div>
    </section>

    <!-- 数据导出 -->
    <el-card class="section-card">
      <template #header>
        <span class="card-section-title">导出数据</span>
      </template>
      <p class="section-desc">将所有商品、订单、定价规则、研究所数据导出为 JSON 文件备份。</p>
      <el-button type="primary" :loading="exporting" @click="handleExport">
        导出全部数据
      </el-button>
    </el-card>

    <!-- 数据导入 -->
    <el-card class="section-card">
      <template #header>
        <span class="card-section-title">导入数据</span>
      </template>
      <p class="section-desc">
        支持本系统导出的快照，也支持旧版纯前端版本（IndexedDB / Dexie）在「数据中心」导出的 JSON
        备份，旧备份会自动转换为新版结构。<strong>导入会覆盖服务端数据库中的全部数据</strong>，请谨慎操作。
      </p>
      <el-button type="danger" :loading="importing" @click="fileInputRef?.click()">
        选择备份文件导入
      </el-button>
      <input
        ref="fileInputRef"
        type="file"
        accept="application/json,.json"
        style="display: none"
        @change="handleImport"
      />
    </el-card>

    <!-- 服务端存储 -->
    <el-card class="section-card">
      <template #header>
        <span class="card-section-title">服务端存储</span>
        <el-tag size="small" type="success" style="margin-left: 8px">已启用</el-tag>
      </template>
      <p class="section-desc" style="color: var(--el-text-color-secondary)">
        数据保存在 Go 服务的 SQLite 文件中。使用 Docker 部署时把 data 目录挂载到宿主机即可持久化。
      </p>
      <el-button disabled>由服务器统一持久化</el-button>
    </el-card>
  </div>
</template>
