<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { nextTick, onUnmounted, ref, watch } from 'vue';

// ---- 类型定义 ----
type EffectMode = 'dot' | 'line';

const MM_PER_INCH = 25.4;
const EXPORT_DPI = 300;
const MAX_RENDER_SIDE = 3200;

// ---- 状态 ----
const mode = ref<EffectMode>('dot');
const spacing = ref(10);
const dotScale = ref(0.9);
const lineWidth = ref(2);
const angle = ref(45);
const contrast = ref(150);
const brightness = ref(100);
const invert = ref(false);
const threshold = ref(128); // 灰度阈值，低于此值视为"暗"区
const previewMode = ref<'fit' | 'actual'>('fit');
const fitScale = ref(82);
const preset = ref<'soft' | 'balanced' | 'stamp'>('balanced');
const widthMm = ref(60);
const heightMm = ref(60);
const lockAspect = ref(true);

const isDragging = ref(false);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const uploadRef = ref<HTMLInputElement | null>(null);
const originalImage = ref<HTMLImageElement | null>(null);
const hasImage = ref(false);
let sourceAspect = 1;
let syncingSize = false;

function clampMm(value: number): number {
  return Math.max(10, Math.min(300, value));
}

function mmToPx(mm: number): number {
  return Math.max(32, Math.round((clampMm(mm) / MM_PER_INCH) * EXPORT_DPI));
}

function pxToMm(px: number): number {
  return (px / EXPORT_DPI) * MM_PER_INCH;
}

function formatMm(px: number | undefined): string {
  if (!px) return '0.0';
  return pxToMm(px).toFixed(1);
}

function initSizeFromImage(img: HTMLImageElement) {
  sourceAspect = img.naturalWidth > 0 && img.naturalHeight > 0 ? img.naturalWidth / img.naturalHeight : 1;
  syncingSize = true;
  widthMm.value = 60;
  heightMm.value = Number((60 / sourceAspect).toFixed(1));
  syncingSize = false;
}

function getTargetSizePx(): { w: number; h: number } {
  const rawW = mmToPx(widthMm.value);
  const rawH = mmToPx(heightMm.value);
  const maxSide = Math.max(rawW, rawH);
  if (maxSide <= MAX_RENDER_SIDE) {
    return { w: rawW, h: rawH };
  }
  const scale = MAX_RENDER_SIDE / maxSide;
  return {
    w: Math.max(32, Math.round(rawW * scale)),
    h: Math.max(32, Math.round(rawH * scale)),
  };
}

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
      initSizeFromImage(img);
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
function getGrayscaleData(img: HTMLImageElement, targetW: number, targetH: number): { data: Uint8ClampedArray; w: number; h: number } {
  const offscreen = document.createElement('canvas');
  offscreen.width = targetW;
  offscreen.height = targetH;
  const ctx = offscreen.getContext('2d')!;

  // 1) 绘制原图
  ctx.drawImage(img, 0, 0, targetW, targetH);

  // 2) 手动灰度 + 对比度，避免 canvas 自绘滤镜造成阈值偏移
  const imageData = ctx.getImageData(0, 0, offscreen.width, offscreen.height);
  const data = imageData.data;
  const contrastFactor = contrast.value / 100;
  const brightnessFactor = brightness.value / 100;

  for (let i = 0; i < data.length; i += 4) {
    const r = data[i];
    const g = data[i + 1];
    const b = data[i + 2];
    const gray = 0.299 * r + 0.587 * g + 0.114 * b;
    const contrasted = (gray - 128) * contrastFactor + 128;
    const adjusted = Math.max(0, Math.min(255, contrasted * brightnessFactor));
    data[i] = adjusted;
    data[i + 1] = adjusted;
    data[i + 2] = adjusted;
  }

  return { data, w: offscreen.width, h: offscreen.height };
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
  const diagLen = Math.sqrt(w * w + h * h) * 1.4;

  // 1) 先画完整斜纹图层（均匀线宽与间距）
  const hatchCanvas = document.createElement('canvas');
  hatchCanvas.width = w;
  hatchCanvas.height = h;
  const hatchCtx = hatchCanvas.getContext('2d')!;

  hatchCtx.fillStyle = invert.value ? '#000000' : '#ffffff';
  hatchCtx.fillRect(0, 0, w, h);
  hatchCtx.strokeStyle = invert.value ? '#ffffff' : '#000000';
  hatchCtx.lineCap = 'round';
  hatchCtx.lineWidth = lw;

  hatchCtx.save();
  hatchCtx.translate(w / 2, h / 2);
  hatchCtx.rotate(deg);

  const halfLen = diagLen / 2;
  const rows = Math.ceil(diagLen / sp);
  for (let rowIdx = -rows; rowIdx <= rows; rowIdx++) {
    const perpPos = rowIdx * sp;
    hatchCtx.beginPath();
    hatchCtx.moveTo(-halfLen, perpPos);
    hatchCtx.lineTo(halfLen, perpPos);
    hatchCtx.stroke();
  }
  hatchCtx.restore();

  // 2) 构建主体掩膜（阈值二值化）
  const maskCanvas = document.createElement('canvas');
  maskCanvas.width = w;
  maskCanvas.height = h;
  const maskCtx = maskCanvas.getContext('2d')!;
  const maskImage = maskCtx.createImageData(w, h);
  const maskData = maskImage.data;

  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      const idx = (y * w + x) * 4;
      const gray = grayData[idx];
      const keep = invert.value ? gray > threshold.value : gray < threshold.value;
      const v = keep ? 255 : 0;
      maskData[idx] = 255;
      maskData[idx + 1] = 255;
      maskData[idx + 2] = 255;
      maskData[idx + 3] = v;
    }
  }
  maskCtx.putImageData(maskImage, 0, 0);

  // 3) 用掩膜裁切斜纹，仅保留主体内部
  hatchCtx.globalCompositeOperation = 'destination-in';
  hatchCtx.drawImage(maskCanvas, 0, 0);
  hatchCtx.globalCompositeOperation = 'source-over';

  ctx.drawImage(hatchCanvas, 0, 0);
}

// ---- 主渲染函数 ----
function renderHalftone() {
  const img = originalImage.value;
  const canvas = canvasRef.value;
  if (!img || !canvas) return;

  const { w: targetW, h: targetH } = getTargetSizePx();
  const { data, w, h } = getGrayscaleData(img, targetW, targetH);
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

function calcOtsuThreshold(data: Uint8ClampedArray): number {
  const hist = new Array<number>(256).fill(0);
  for (let i = 0; i < data.length; i += 4) {
    hist[data[i]] += 1;
  }

  const total = data.length / 4;
  let sumAll = 0;
  for (let i = 0; i < 256; i++) {
    sumAll += i * hist[i];
  }

  let sumB = 0;
  let wB = 0;
  let maxVar = -1;
  let best = 128;

  for (let t = 0; t < 256; t++) {
    wB += hist[t];
    if (wB === 0) continue;
    const wF = total - wB;
    if (wF === 0) break;

    sumB += t * hist[t];
    const mB = sumB / wB;
    const mF = (sumAll - sumB) / wF;
    const between = wB * wF * (mB - mF) * (mB - mF);
    if (between > maxVar) {
      maxVar = between;
      best = t;
    }
  }

  return best;
}

function autoThreshold() {
  const img = originalImage.value;
  if (!img) {
    ElMessage.warning('请先上传图片');
    return;
  }
  const { w: targetW, h: targetH } = getTargetSizePx();
  const { data } = getGrayscaleData(img, targetW, targetH);
  threshold.value = calcOtsuThreshold(data);
  ElMessage.success(`已自动计算阈值：${threshold.value}`);
}

function applyPreset(type: 'soft' | 'balanced' | 'stamp') {
  preset.value = type;
  if (type === 'soft') {
    spacing.value = 12;
    dotScale.value = 0.75;
    lineWidth.value = 1.5;
    angle.value = 45;
    contrast.value = 130;
    brightness.value = 108;
    threshold.value = 120;
    fitScale.value = 82;
    return;
  }
  if (type === 'balanced') {
    spacing.value = 10;
    dotScale.value = 0.9;
    lineWidth.value = 2;
    angle.value = 45;
    contrast.value = 150;
    brightness.value = 100;
    threshold.value = 128;
    fitScale.value = 82;
    return;
  }
  spacing.value = 8;
  dotScale.value = 1;
  lineWidth.value = 2.5;
  angle.value = 40;
  contrast.value = 190;
  brightness.value = 92;
  threshold.value = 142;
  fitScale.value = 78;
}

function onPresetChange(value: string | number | boolean) {
  if (value === 'soft' || value === 'balanced' || value === 'stamp') {
    applyPreset(value);
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

watch([mode, spacing, dotScale, lineWidth, angle, contrast, brightness, invert, threshold], scheduleRender);

watch(widthMm, (value) => {
  if (syncingSize) return;
  widthMm.value = clampMm(value);
  if (lockAspect.value && sourceAspect > 0) {
    syncingSize = true;
    heightMm.value = Number((widthMm.value / sourceAspect).toFixed(1));
    syncingSize = false;
  }
  scheduleRender();
});

watch(heightMm, (value) => {
  if (syncingSize) return;
  heightMm.value = clampMm(value);
  if (lockAspect.value && sourceAspect > 0) {
    syncingSize = true;
    widthMm.value = Number((heightMm.value * sourceAspect).toFixed(1));
    syncingSize = false;
  }
  scheduleRender();
});

watch(lockAspect, (value) => {
  if (!value || sourceAspect <= 0) return;
  syncingSize = true;
  heightMm.value = Number((widthMm.value / sourceAspect).toFixed(1));
  syncingSize = false;
  scheduleRender();
});

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
            <span class="param-label">快速预设</span>
            <el-radio-group v-model="preset" size="small" @change="onPresetChange">
              <el-radio-button value="soft">柔和</el-radio-button>
              <el-radio-button value="balanced">均衡</el-radio-button>
              <el-radio-button value="stamp">印章</el-radio-button>
            </el-radio-group>
          </div>

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

          <div class="param-row">
            <span class="param-label">亮度 <em>{{ brightness }}%</em></span>
            <el-slider v-model="brightness" :min="60" :max="140" :step="1" />
          </div>

          <div class="param-row">
            <span class="param-label">输出尺寸（mm）</span>
            <div class="size-input-row">
              <el-input-number v-model="widthMm" :min="10" :max="300" :step="1" size="small" />
              <span class="size-sep">×</span>
              <el-input-number v-model="heightMm" :min="10" :max="300" :step="1" size="small" />
              <el-switch v-model="lockAspect" inline-prompt active-text="锁比" inactive-text="自由" />
            </div>
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
            <div class="param-row">
              <span class="param-label">主体阈值 <em>{{ threshold }}</em></span>
              <el-slider v-model="threshold" :min="0" :max="255" :step="1" />
            </div>
            <div class="param-row">
              <el-button size="small" @click="autoThreshold">自动阈值</el-button>
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
            <div class="preview-header-left">
              <span class="card-title">预览</span>
              <span v-if="hasImage" class="canvas-size-hint">
                {{ canvasRef?.width }} × {{ canvasRef?.height }} px ({{ formatMm(canvasRef?.width) }} × {{ formatMm(canvasRef?.height) }} mm)
              </span>
            </div>
            <div class="preview-actions" v-if="hasImage">
              <div class="preview-action-row">
                <el-radio-group v-model="previewMode" size="small">
                  <el-radio-button value="fit">适应</el-radio-button>
                  <el-radio-button value="actual">原始</el-radio-button>
                </el-radio-group>
                <div class="fit-scale-wrap" v-if="previewMode === 'fit'">
                  <span>缩放</span>
                  <el-slider v-model="fitScale" :min="60" :max="100" :step="1" style="width: 120px" />
                </div>
              </div>
            </div>
          </template>
          <div class="canvas-scroll">
            <div v-if="!hasImage" class="placeholder">
              <el-icon :size="64" color="#dcdfe6"><i-ep-picture /></el-icon>
              <p>上传图片后在此预览效果</p>
            </div>
            <canvas
              v-show="hasImage"
              ref="canvasRef"
              class="output-canvas"
              :class="previewMode"
              :style="previewMode === 'fit' ? { maxWidth: `${fitScale}%` } : undefined"
            />
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

.size-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.size-input-row :deep(.el-input-number) {
  width: 92px;
}

.size-sep {
  color: #909399;
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
  gap: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.preview-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.preview-actions {
  flex-shrink: 0;
}

.preview-action-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.fit-scale-wrap {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #606266;
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
  padding: 12px;
  box-sizing: border-box;
  display: flex;
  align-items: flex-start;
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
  width: auto;
  height: auto;
  display: block;
}

.output-canvas.fit {
  max-width: 82%;
  max-height: calc(58vh - 24px);
  margin: 0 auto;
}

.output-canvas.actual {
  max-width: none;
  max-height: none;
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

  .output-canvas.fit {
    max-width: 88%;
    max-height: calc(42vh - 24px);
  }

  .preview-card :deep(.el-card__header) {
    flex-direction: column;
    align-items: flex-start;
  }

  .preview-action-row {
    width: 100%;
    flex-wrap: wrap;
  }

  .size-input-row {
    flex-wrap: wrap;
  }
}
</style>
