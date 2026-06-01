<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { nextTick, onUnmounted, ref, watch } from 'vue';

// ---- 类型定义 ----
type EffectMode = 'dot' | 'line';

// ---- 状态 ----
const mode = ref<EffectMode>('dot');
const spacing = ref(10);
const dotScale = ref(0.9);
const lineWidth = ref(2);
const angle = ref(45);
const contrast = ref(150);
const invert = ref(false);
const threshold = ref(128); // 灰度阈值，低于此值视为"暗"区

const isDragging = ref(false);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const previewCanvasRef = ref<HTMLCanvasElement | null>(null);
const uploadRef = ref<HTMLInputElement | null>(null);
const originalImage = ref<HTMLImageElement | null>(null);
const hasImage = ref(false);

// ---- 图像加载 ----
function loadImageFile(file: File) {
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请上传图片文件');
    return;
  }
  const reader = new FileReader();
  reader.onload = (e) => {
    const img = new Image();
    img.onload = () => {
      originalImage.value = img;
      hasImage.value = true;
      nextTick(() => {
        renderHalftone();
      });
    };
    img.src = e.target?.result as string;
  };
  reader.readAsDataURL(file);
}

function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (file) loadImageFile(file);
}

function onDrop(e: DragEvent) {
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file) loadImageFile(file);
}

// ---- 灰度采样：带预处理 ----
function getGrayscaleData(img: HTMLImageElement): { data: Uint8ClampedArray; w: number; h: number } {
  const offscreen = document.createElement('canvas');
  offscreen.width = img.naturalWidth;
  offscreen.height = img.naturalHeight;
  const ctx = offscreen.getContext('2d')!;

  // 1. 绘制原图
  ctx.drawImage(img, 0, 0);

  // 2. 对比度增强 + 灰度（CSS filter 应用到第二次绘制）
  ctx.filter = `grayscale(100%) contrast(${contrast.value}%)`;
  ctx.drawImage(offscreen, 0, 0);
  ctx.filter = 'none';

  const imageData = ctx.getImageData(0, 0, offscreen.width, offscreen.height);
  return { data: imageData.data, w: offscreen.width, h: offscreen.height };
}

function sampleGray(data: Uint8ClampedArray, w: number, x: number, y: number): number {
  const idx = (y * w + x) * 4;
  return data[idx]; // 已灰度化，R=G=B
}

// ---- 点阵算法 ----
function renderDot(ctx: CanvasRenderingContext2D, grayData: Uint8ClampedArray, w: number, h: number) {
  const sp = spacing.value;
  const scale = dotScale.value;
  const maxR = (sp / 2) * scale;

  ctx.fillStyle = invert.value ? '#000000' : '#ffffff';
  ctx.fillRect(0, 0, w, h);
  ctx.fillStyle = invert.value ? '#ffffff' : '#000000';

  for (let y = sp / 2; y < h; y += sp) {
    for (let x = sp / 2; x < w; x += sp) {
      const sx = Math.min(Math.round(x), w - 1);
      const sy = Math.min(Math.round(y), h - 1);
      const gray = sampleGray(grayData, w, sx, sy);
      // 暗区 → 大圆（gray 低 → 半径大）
      const brightness = invert.value ? gray / 255 : 1 - gray / 255;
      const r = maxR * brightness;
      if (r < 0.5) continue;
      ctx.beginPath();
      ctx.arc(x, y, r, 0, Math.PI * 2);
      ctx.fill();
    }
  }
}

// ---- 斜纹算法 ----
function renderLine(ctx: CanvasRenderingContext2D, grayData: Uint8ClampedArray, w: number, h: number) {
  const sp = spacing.value;
  const lw = lineWidth.value;
  const deg = (angle.value * Math.PI) / 180;
  const diagLen = Math.sqrt(w * w + h * h) * 1.2;

  ctx.fillStyle = invert.value ? '#000000' : '#ffffff';
  ctx.fillRect(0, 0, w, h);
  ctx.strokeStyle = invert.value ? '#ffffff' : '#000000';
  ctx.lineCap = 'round';

  ctx.save();
  ctx.translate(w / 2, h / 2);
  ctx.rotate(deg);

  const halfLen = diagLen / 2;
  const rows = Math.ceil(diagLen / sp);

  for (let rowIdx = -rows; rowIdx <= rows; rowIdx++) {
    const perpPos = rowIdx * sp;
    // 采样该扫描线中点对应的原图坐标（反变换）
    const sampleX = Math.round(w / 2 + perpPos * Math.sin(deg));
    const sampleY = Math.round(h / 2 - perpPos * Math.cos(deg));
    const clampX = Math.max(0, Math.min(w - 1, sampleX));
    const clampY = Math.max(0, Math.min(h - 1, sampleY));
    const gray = sampleGray(grayData, w, clampX, clampY);
    const brightness = invert.value ? gray / 255 : 1 - gray / 255;

    // 线宽按亮度映射（暗区线更粗）
    const computedLW = lw * brightness * 2;
    if (computedLW < 0.3) continue;

    ctx.lineWidth = computedLW;
    ctx.beginPath();
    ctx.moveTo(-halfLen, perpPos);
    ctx.lineTo(halfLen, perpPos);
    ctx.stroke();
  }

  ctx.restore();
}

// ---- 主渲染函数 ----
function renderHalftone() {
  const img = originalImage.value;
  const canvas = canvasRef.value;
  if (!img || !canvas) return;

  const { data, w, h } = getGrayscaleData(img);
  canvas.width = w;
  canvas.height = h;

  const ctx = canvas.getContext('2d')!;
  ctx.clearRect(0, 0, w, h);

  if (mode.value === 'dot') {
    renderDot(ctx, data, w, h);
  } else {
    renderLine(ctx, data, w, h);
  }
}

// ---- 防抖 ----
let debounceTimer: ReturnType<typeof setTimeout> | null = null;
function scheduleRender() {
  if (!hasImage.value) return;
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    renderHalftone();
  }, 250);
}

watch([mode, spacing, dotScale, lineWidth, angle, contrast, invert], scheduleRender);

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer);
});

// ---- 导出 PNG ----
function exportPng() {
  const canvas = canvasRef.value;
  if (!canvas) {
    ElMessage.warning('请先上传图片');
    return;
  }
  const url = canvas.toDataURL('image/png');
  const a = document.createElement('a');
  a.href = url;
  a.download = `halftone-stamp-${Date.now()}.png`;
  a.click();
  ElMessage.success('导出成功');
}
</script>

<template>
  <div class="halftone-page">
    <div class="page-header">
      <h2 class="page-title">印章生成</h2>
      <p class="page-desc">将照片处理为点阵或斜纹半调效果，可直接用于制作印章</p>
    </div>

    <div class="halftone-layout">
      <!-- 左侧控制面板 -->
      <div class="control-panel">
        <!-- 上传区 -->
        <el-card class="panel-card">
          <template #header><span class="card-title">上传图片</span></template>
          <div
            class="upload-zone"
            :class="{ dragging: isDragging }"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @drop.prevent="onDrop"
            @click="uploadRef?.click()"
          >
            <el-icon :size="36" class="upload-icon"><i-ep-picture /></el-icon>
            <p class="upload-hint">点击或拖拽图片到此处</p>
            <p class="upload-sub">支持 JPG / PNG / WEBP</p>
            <input ref="uploadRef" type="file" accept="image/*" style="display:none" @change="onFileChange" />
          </div>
        </el-card>

        <!-- 效果参数 -->
        <el-card class="panel-card">
          <template #header><span class="card-title">效果参数</span></template>

          <div class="param-row">
            <span class="param-label">效果类型</span>
            <el-radio-group v-model="mode" size="small">
              <el-radio-button value="dot">点阵</el-radio-button>
              <el-radio-button value="line">斜纹</el-radio-button>
            </el-radio-group>
          </div>

          <div class="param-row">
            <span class="param-label">网格间距 <em>{{ spacing }}px</em></span>
            <el-slider v-model="spacing" :min="4" :max="40" :step="1" />
          </div>

          <div class="param-row">
            <span class="param-label">对比度 <em>{{ contrast }}%</em></span>
            <el-slider v-model="contrast" :min="100" :max="400" :step="10" />
          </div>

          <template v-if="mode === 'dot'">
            <div class="param-row">
              <span class="param-label">点大小系数 <em>{{ dotScale }}</em></span>
              <el-slider v-model="dotScale" :min="0.1" :max="1.0" :step="0.05" />
            </div>
          </template>

          <template v-if="mode === 'line'">
            <div class="param-row">
              <span class="param-label">线条宽度 <em>{{ lineWidth }}px</em></span>
              <el-slider v-model="lineWidth" :min="1" :max="8" :step="0.5" />
            </div>
            <div class="param-row">
              <span class="param-label">斜纹角度 <em>{{ angle }}°</em></span>
              <el-slider v-model="angle" :min="0" :max="180" :step="5" />
            </div>
          </template>

          <div class="param-row">
            <span class="param-label">反色（黑底白纹）</span>
            <el-switch v-model="invert" />
          </div>
        </el-card>

        <!-- 操作按钮 -->
        <div class="action-row">
          <el-button type="primary" :disabled="!hasImage" @click="renderHalftone" style="flex:1">
            重新生成
          </el-button>
          <el-button type="success" :disabled="!hasImage" @click="exportPng" style="flex:1">
            导出 PNG
          </el-button>
        </div>
      </div>

      <!-- 右侧预览区 -->
      <div class="preview-panel">
        <el-card class="preview-card">
          <template #header>
            <span class="card-title">预览</span>
            <span v-if="hasImage" class="canvas-size-hint">
              {{ canvasRef?.width }} × {{ canvasRef?.height }} px
            </span>
          </template>
          <div class="canvas-scroll">
            <div v-if="!hasImage" class="placeholder">
              <el-icon :size="64" color="#dcdfe6"><i-ep-picture /></el-icon>
              <p>上传图片后在此预览效果</p>
            </div>
            <canvas v-show="hasImage" ref="canvasRef" class="output-canvas" />
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.halftone-page {
  padding: 24px;
  min-height: 100%;
  background: #f5f7fa;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 700;
  color: #1d2129;
  margin: 0 0 4px;
}

.page-desc {
  font-size: 13px;
  color: #86909c;
  margin: 0;
}

.halftone-layout {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

/* 控制面板 */
.control-panel {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel-card :deep(.el-card__header) {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-title {
  font-size: 13px;
  font-weight: 600;
  color: #1d2129;
}

.upload-zone {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 24px 12px;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
}

.upload-zone:hover,
.upload-zone.dragging {
  border-color: #409eff;
  background: #ecf5ff;
}

.upload-icon {
  color: #c0c4cc;
}

.upload-hint {
  font-size: 13px;
  color: #606266;
  margin: 0;
}

.upload-sub {
  font-size: 11px;
  color: #c0c4cc;
  margin: 0;
}

.param-row {
  margin-bottom: 16px;
}

.param-label {
  display: block;
  font-size: 12px;
  color: #606266;
  margin-bottom: 6px;
}

.param-label em {
  font-style: normal;
  font-weight: 600;
  color: #409eff;
  margin-left: 4px;
}

.action-row {
  display: flex;
  gap: 10px;
}

/* 预览区 */
.preview-panel {
  flex: 1;
  min-width: 0;
}

.preview-card {
  height: 100%;
}

.preview-card :deep(.el-card__header) {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0f0f0;
}

.canvas-size-hint {
  font-size: 11px;
  color: #c0c4cc;
}

.canvas-scroll {
  overflow: auto;
  max-height: 70vh;
  background: #f5f5f5;
  border-radius: 4px;
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #c0c4cc;
  font-size: 13px;
  padding: 40px;
}

.placeholder p {
  margin: 0;
}

.output-canvas {
  max-width: 100%;
  display: block;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .halftone-page {
    padding: 12px;
  }

  .halftone-layout {
    flex-direction: column;
  }

  .control-panel {
    width: 100%;
  }

  .canvas-scroll {
    max-height: 50vh;
  }
}
</style>
