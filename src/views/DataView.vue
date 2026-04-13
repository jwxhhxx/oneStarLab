<script setup lang="ts">
import { ref } from 'vue';
import { exportDB, importDB } from 'dexie-export-import';
import { ElMessage, ElMessageBox } from 'element-plus';

import { db } from '@/db/appDb';
import { useShopStore } from '@/stores/useShopStore';

const store = useShopStore();

const importing = ref(false);
const exporting = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);

async function handleExport() {
  exporting.value = true;
  try {
    const blob = await exportDB(db);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    const date = new Date().toISOString().slice(0, 10);
    a.href = url;
    a.download = `onestarlab-backup-${date}.json`;
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
      '导入会覆盖当前所有本地数据，建议先导出备份。确认继续？',
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
    await db.delete();
    await db.open();
    await importDB(file);
    await store.loadAll();
    ElMessage.success('数据导入成功，页面已刷新');
  } catch (e) {
    ElMessage.error('导入失败，请检查文件格式');
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
        <h2 class="hero-title">掌握你的本地数据</h2>
        <p class="hero-description">
          所有数据仅保存在当前浏览器，建议定期导出备份，在更换设备或清理浏览器前请先导出。
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
        从之前导出的 JSON 备份文件中恢复数据。<strong>导入会覆盖当前所有本地数据</strong>，请谨慎操作。
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

    <!-- 云同步（预留） -->
    <el-card class="section-card">
      <template #header>
        <span class="card-section-title">云端同步</span>
        <el-tag size="small" type="info" style="margin-left: 8px">即将推出</el-tag>
      </template>
      <p class="section-desc" style="color: var(--el-text-color-secondary)">
        多设备数据同步功能正在规划中，敬请期待。
      </p>
      <el-button disabled>连接云端</el-button>
    </el-card>
  </div>
</template>
