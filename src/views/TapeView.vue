<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { ArrowDown, ArrowUp, Delete, Download, Picture } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

import TapeCanvas from '@/components/tape/TapeCanvas.vue';
import type { TapeDesign, TapePatternType } from '@/types';
import {
  TAPE_LENGTH_PRESETS,
  TAPE_PALETTE,
  TAPE_PATTERNS,
  TAPE_WIDTH_PRESETS,
  createTapeLayer,
  exportTapePng,
  exportTapePsd,
} from '@/utils/tape';

function createDefaultDesign(): TapeDesign {
  return {
    name: '未命名胶带',
    widthMm: 20,
    lengthMm: 100,
    background: '#fff8ee',
    transparent: false,
    palette: [...TAPE_PALETTE],
    layers: [
      { ...createTapeLayer('stars', 0), rows: 2, cols: 8, sizeMm: 5, color: '#e9a05e' },
      {
        ...createTapeLayer('hearts', 1),
        rows: 1,
        cols: 5,
        sizeMm: 4,
        color: '#e97a8a',
        stagger: true,
      },
    ],
  };
}

const design = reactive<TapeDesign>(createDefaultDesign());
const dpi = ref(300);
const exporting = ref<'' | 'png' | 'psd'>('');

const estimatedPixels = computed(() => {
  const pxPerMm = dpi.value / 25.4;
  return {
    width: Math.round(design.lengthMm * pxPerMm),
    height: Math.round(design.widthMm * pxPerMm),
  };
});

function addLayer(type: TapePatternType) {
  design.layers.push(createTapeLayer(type, design.layers.length, design.palette));
}

function removeLayer(index: number) {
  if (design.layers.length <= 1) {
    ElMessage.warning('至少保留一个图案');
    return;
  }
  design.layers.splice(index, 1);
}

function moveLayer(index: number, direction: -1 | 1) {
  const target = index + direction;
  if (target < 0 || target >= design.layers.length) return;
  const [layer] = design.layers.splice(index, 1);
  design.layers.splice(target, 0, layer);
}

function shufflePalette() {
  design.palette = [...design.palette].sort(() => Math.random() - 0.5);
}

function resetDesign() {
  Object.assign(design, createDefaultDesign());
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function exportFileName(extension: string) {
  return `tape-${design.widthMm}mm-${Math.round(design.lengthMm)}mm-${dpi.value}dpi.${extension}`;
}

function reportError(error: unknown) {
  ElMessage.error(error instanceof Error ? error.message : '导出失败，请稍后重试');
}

async function handleExportPng() {
  exporting.value = 'png';
  try {
    const blob = await exportTapePng(design, dpi.value);
    downloadBlob(blob, exportFileName('png'));
    ElMessage.success('PNG 已导出');
  } catch (error) {
    reportError(error);
  } finally {
    exporting.value = '';
  }
}

async function handleExportPsd() {
  exporting.value = 'psd';
  try {
    const blob = await exportTapePsd(design, dpi.value);
    downloadBlob(blob, exportFileName('psd'));
    ElMessage.success('PSD 已导出');
  } catch (error) {
    reportError(error);
  } finally {
    exporting.value = '';
  }
}
</script>

<template>
  <div class="page-stack tape-page">
    <section class="page-hero">
      <div>
        <p class="hero-kicker">胶带生成 · DECO TAPE</p>
        <h1 class="hero-title">把喜欢的图案卷成一条胶带</h1>
        <p class="hero-description">
          自定义宽长、图案排布与配色，叠出属于自己的 deco 胶带；调好后导出 PNG 或分层 PSD，直接收进手帐素材库。
        </p>
        <div class="hero-badges">
          <span class="hero-badge">{{ design.widthMm }} mm 宽</span>
          <span class="hero-badge">{{ (design.lengthMm / 10).toFixed(1) }} cm 长</span>
          <span class="hero-badge">{{ design.layers.length }} 个图案</span>
        </div>
      </div>
    </section>

    <div class="tape-layout">
      <div class="tape-controls">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">胶带尺寸</span>
              <span class="muted">{{ (design.lengthMm / 10).toFixed(0) }} cm</span>
            </div>
          </template>

          <div class="field-row">
            <label>宽度</label>
            <el-input-number
              v-model="design.widthMm"
              :min="10"
              :max="50"
              :step="5"
              controls-position="right"
            />
            <span class="unit">mm</span>
          </div>
          <div class="field-row">
            <label>长度</label>
            <el-input-number
              v-model="design.lengthMm"
              :min="50"
              :max="600"
              :step="50"
              controls-position="right"
            />
            <span class="unit">mm</span>
          </div>

          <div class="preset-block">
            <span class="preset-label">常用宽度</span>
            <div class="preset-row">
              <button
                v-for="width in TAPE_WIDTH_PRESETS"
                :key="width"
                type="button"
                class="preset-chip"
                :class="{ 'is-active': design.widthMm === width }"
                @click="design.widthMm = width"
              >
                {{ width }} mm
              </button>
            </div>
          </div>
          <div class="preset-block">
            <span class="preset-label">常用长度</span>
            <div class="preset-row">
              <button
                v-for="length in TAPE_LENGTH_PRESETS"
                :key="length"
                type="button"
                class="preset-chip"
                :class="{ 'is-active': design.lengthMm === length * 10 }"
                @click="design.lengthMm = length * 10"
              >
                {{ length }} cm
              </button>
            </div>
          </div>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">图案库</span>
              <span class="muted">点一下加入一层</span>
            </div>
          </template>

          <div class="pattern-grid">
            <button
              v-for="pattern in TAPE_PATTERNS"
              :key="pattern.type"
              type="button"
              class="pattern-button"
              @click="addLayer(pattern.type)"
            >
              {{ pattern.label }}
            </button>
          </div>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">图案图层</span>
              <div class="head-actions">
                <span class="muted">从下往上叠加</span>
                <el-button link type="primary" @click="resetDesign">重置</el-button>
              </div>
            </div>
          </template>

          <div v-for="(layer, index) in design.layers" :key="layer.id" class="layer-card">
            <div class="layer-head">
              <el-select v-model="layer.type" size="small" class="layer-type">
                <el-option
                  v-for="pattern in TAPE_PATTERNS"
                  :key="pattern.type"
                  :label="pattern.label"
                  :value="pattern.type"
                />
              </el-select>
              <div class="layer-actions">
                <el-button
                  :icon="ArrowUp"
                  circle
                  size="small"
                  :disabled="index === 0"
                  aria-label="上移图层"
                  @click="moveLayer(index, -1)"
                />
                <el-button
                  :icon="ArrowDown"
                  circle
                  size="small"
                  :disabled="index === design.layers.length - 1"
                  aria-label="下移图层"
                  @click="moveLayer(index, 1)"
                />
                <el-button
                  :icon="Delete"
                  circle
                  size="small"
                  aria-label="删除图层"
                  @click="removeLayer(index)"
                />
              </div>
            </div>

            <div class="layer-grid">
              <label class="mini-field">
                <span>排数</span>
                <el-input-number
                  v-model="layer.rows"
                  :min="1"
                  :max="8"
                  size="small"
                  controls-position="right"
                />
              </label>
              <label class="mini-field">
                <span>列数</span>
                <el-input-number
                  v-model="layer.cols"
                  :min="1"
                  :max="20"
                  size="small"
                  controls-position="right"
                />
              </label>
              <label class="mini-field">
                <span>大小 mm</span>
                <el-input-number
                  v-model="layer.sizeMm"
                  :min="1"
                  :max="20"
                  :step="0.5"
                  :precision="1"
                  size="small"
                  controls-position="right"
                />
              </label>
              <label class="mini-field">
                <span>旋转 °</span>
                <el-input-number
                  v-model="layer.rotation"
                  :min="0"
                  :max="345"
                  :step="15"
                  size="small"
                  controls-position="right"
                />
              </label>
              <label class="mini-field">
                <span>透明度</span>
                <el-input-number
                  v-model="layer.opacity"
                  :min="0.1"
                  :max="1"
                  :step="0.1"
                  :precision="1"
                  size="small"
                  controls-position="right"
                />
              </label>
              <div class="mini-field mini-field--inline">
                <span>颜色</span>
                <el-color-picker v-model="layer.color" size="small" />
                <span class="mini-check">多彩 <el-switch v-model="layer.multi" size="small" /></span>
              </div>
              <div class="mini-field mini-field--inline">
                <span>错位</span>
                <el-switch v-model="layer.stagger" size="small" />
              </div>
            </div>
          </div>

          <p v-if="design.layers.length === 0" class="hint">还没有图案，从上面的图案库加入一层。</p>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">颜色与底纸</span>
              <el-button link type="primary" @click="shufflePalette">换一组顺序</el-button>
            </div>
          </template>

          <div class="palette-row">
            <el-color-picker
              v-for="(color, index) in design.palette"
              :key="index"
              v-model="design.palette[index]"
              size="small"
            />
          </div>
          <p class="hint">图案开启「多彩」后会按这里的顺序轮换颜色。</p>

          <div class="field-row field-row--switch">
            <label>透明底</label>
            <el-switch v-model="design.transparent" />
            <el-color-picker v-if="!design.transparent" v-model="design.background" size="small" />
          </div>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">导出</span>
              <span class="muted">约 {{ estimatedPixels.width }} × {{ estimatedPixels.height }} px</span>
            </div>
          </template>

          <div class="field-row">
            <label>DPI</label>
            <el-radio-group v-model="dpi" size="small">
              <el-radio-button :value="150">150</el-radio-button>
              <el-radio-button :value="300">300</el-radio-button>
            </el-radio-group>
          </div>

          <div class="export-actions">
            <el-button
              type="primary"
              :icon="Download"
              :loading="exporting === 'png'"
              @click="handleExportPng"
            >
              导出 PNG
            </el-button>
            <el-button :icon="Picture" :loading="exporting === 'psd'" @click="handleExportPsd">
              导出 PSD
            </el-button>
          </div>
          <p class="hint">
            PSD 会按图案拆成独立图层，并在有底纸时保留底色层；透明底导出会保留透明通道。
          </p>
        </el-card>
      </div>

      <div class="tape-preview-panel">
        <el-card class="section-card preview-card" shadow="never">
          <template #header>
            <div class="preview-head">
              <span class="panel-title">实时预览</span>
              <span class="muted">{{ design.widthMm }}×{{ Math.round(design.lengthMm) }} mm</span>
            </div>
          </template>

          <div class="tape-canvas-wrap">
            <TapeCanvas :design="design" />
          </div>
          <p class="preview-note">
            预览按容器宽度缩放；导出时按所选 DPI 输出 1:1 像素。长度越长，单个图案在预览里越小，属正常现象。
          </p>
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tape-layout {
  display: grid;
  grid-template-columns: minmax(0, 440px) minmax(0, 1fr);
  gap: 20px;
  align-items: start;
}

.tape-controls {
  display: grid;
  gap: 16px;
}

.panel-head,
.preview-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.panel-title {
  color: #2f2b3a;
  font-weight: 600;
}

.muted {
  color: #8a93a6;
  font-size: 12px;
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.field-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.field-row label {
  width: 52px;
  color: #5b6472;
  font-size: 13px;
}

.field-row--switch {
  margin-top: 14px;
  margin-bottom: 0;
}

.unit {
  color: #8a93a6;
  font-size: 12px;
}

.preset-block {
  margin-top: 14px;
}

.preset-label {
  display: block;
  margin-bottom: 8px;
  color: #8a93a6;
  font-size: 12px;
}

.preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-chip {
  padding: 5px 10px;
  border: 1px solid #e4e7f0;
  border-radius: 999px;
  background: #fbfcff;
  color: #5b6472;
  font-size: 12px;
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    color 0.18s ease,
    background 0.18s ease;
}

.preset-chip:hover,
.preset-chip.is-active {
  border-color: #8f9bea;
  background: #f1f3ff;
  color: #4d5ac4;
}

.pattern-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(78px, 1fr));
  gap: 8px;
}

.pattern-button {
  padding: 9px 6px;
  border: 1px dashed #d9deea;
  border-radius: 12px;
  background: #fbfcff;
  color: #4a5364;
  font-size: 13px;
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    color 0.18s ease,
    background 0.18s ease;
}

.pattern-button:hover {
  border-color: #8f9bea;
  background: #f4f5ff;
  color: #5563c9;
}

.layer-card {
  margin-bottom: 10px;
  padding: 12px;
  border: 1px solid #eef0f6;
  border-radius: 14px;
  background: #fbfcff;
}

.layer-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.layer-type {
  width: 118px;
}

.layer-actions {
  display: flex;
  gap: 4px;
}

.layer-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 9px 12px;
}

.mini-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  color: #6b7383;
  font-size: 12px;
}

.mini-field :deep(.el-input-number) {
  width: 92px;
}

.mini-field--inline {
  justify-content: flex-start;
}

.mini-check {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
  white-space: nowrap;
}

.palette-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.hint {
  margin: 10px 0 0;
  color: #8a93a6;
  font-size: 12px;
  line-height: 1.6;
}

.export-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 12px;
}

.tape-preview-panel {
  position: sticky;
  top: 18px;
}

.tape-canvas-wrap {
  padding: 18px;
  border: 1px solid #eef0f6;
  border-radius: 16px;
  background: repeating-conic-gradient(#f4f4f8 0% 25%, #ffffff 0% 50%) 50% / 14px 14px;
}

.preview-note {
  margin: 12px 0 0;
  color: #8a93a6;
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 1024px) {
  .tape-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .tape-preview-panel {
    position: static;
  }
}

@media (max-width: 768px) {
  .layer-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .field-row label {
    width: auto;
  }
}
</style>
