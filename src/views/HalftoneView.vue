<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

// ---- 类型定义 ----
type GeneratorMode = 'image' | 'text';
type StampShape = 'round' | 'rect';
type PatternMode = 'none' | 'dots' | 'grid' | 'sunburst';

const MM_PER_INCH = 25.4;
const EXPORT_DPI = 300;
const MAX_RENDER_SIDE = 3200;

// ---- 状态 ----
const generatorMode = ref<GeneratorMode>('text');
const spacing = ref(10);
const dotScale = ref(0.9);
const contrast = ref(150);
const brightness = ref(100);
const invert = ref(false);
const previewMode = ref<'fit' | 'actual'>('fit');
const fitScale = ref(82);
const preset = ref<'soft' | 'balanced' | 'stamp'>('balanced');
const widthMm = ref(60);
const heightMm = ref(60);
const lockAspect = ref(true);
const stampShape = ref<StampShape>('round');
const stampMainText = ref('好运常在');
const stampSubText = ref('ONESTARLAB');
const stampCenterText = ref('福');
const stampColor = ref('#c62828');
const stampBorderWidth = ref(10);
const stampDoubleRing = ref(true);
const stampOuterRingWidth = ref(4);
const stampOuterRingGap = ref(10);
const stampSubArc = ref(true);
const stampMainSize = ref(64);
const stampSubSize = ref(28);
const stampCenterSize = ref(96);
const stampPattern = ref<PatternMode>('dots');
const stampPatternDensity = ref(14);
const stampPatternOpacity = ref(20);
const stampDistress = ref(0);
const stampDistressScale = ref(14);

const isDragging = ref(false);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const uploadRef = ref<HTMLInputElement | null>(null);
const originalImage = ref<HTMLImageElement | null>(null);
const hasImage = ref(false);
let sourceAspect = 1;
let syncingSize = false;
const canRender = computed(() => generatorMode.value === 'text' || hasImage.value);
const route = useRoute();

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
        renderOutput();
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

function drawPatternLayer(ctx: CanvasRenderingContext2D, w: number, h: number) {
  if (stampPattern.value === 'none') return;
  const gap = Math.max(6, stampPatternDensity.value);

  ctx.save();
  ctx.globalAlpha = Math.max(0, Math.min(1, stampPatternOpacity.value / 100));
  ctx.strokeStyle = stampColor.value;
  ctx.fillStyle = stampColor.value;
  ctx.lineWidth = 1;

  if (stampPattern.value === 'dots') {
    for (let y = gap / 2; y < h; y += gap) {
      for (let x = gap / 2; x < w; x += gap) {
        ctx.beginPath();
        ctx.arc(x, y, 1.2, 0, Math.PI * 2);
        ctx.fill();
      }
    }
  }

  if (stampPattern.value === 'grid') {
    for (let y = 0; y <= h; y += gap) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(w, y);
      ctx.stroke();
    }
    for (let x = 0; x <= w; x += gap) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, h);
      ctx.stroke();
    }
  }

  if (stampPattern.value === 'sunburst') {
    const cx = w / 2;
    const cy = h / 2;
    const rayCount = Math.max(18, Math.round(360 / (gap * 0.8)));
    const radius = Math.sqrt(w * w + h * h);
    for (let i = 0; i < rayCount; i++) {
      const angle = (Math.PI * 2 * i) / rayCount;
      ctx.beginPath();
      ctx.moveTo(cx, cy);
      ctx.lineTo(cx + Math.cos(angle) * radius, cy + Math.sin(angle) * radius);
      ctx.stroke();
    }
  }

  ctx.restore();
}

function drawArcText(
  ctx: CanvasRenderingContext2D,
  text: string,
  cx: number,
  cy: number,
  radius: number,
  startAngle: number,
  endAngle: number,
  inward = false,
  reverse = false,
) {
  const chars = reverse ? [...text].reverse() : [...text];
  if (!chars.length) return;
  const step = chars.length === 1 ? 0 : (endAngle - startAngle) / (chars.length - 1);

  chars.forEach((ch, idx) => {
    const angle = startAngle + step * idx;
    ctx.save();
    ctx.translate(cx + Math.cos(angle) * radius, cy + Math.sin(angle) * radius);
    ctx.rotate(angle + (inward ? -Math.PI / 2 : Math.PI / 2));
    ctx.fillText(ch, 0, 0);
    ctx.restore();
  });
}

function noise2d(x: number, y: number, seed: number): number {
  let n = x * 374761393 + y * 668265263 + seed * 69069;
  n = (n ^ (n >> 13)) * 1274126177;
  return ((n ^ (n >> 16)) >>> 0) / 4294967295;
}

function applyDistress(ctx: CanvasRenderingContext2D, w: number, h: number) {
  if (stampDistress.value <= 0) return;

  const image = ctx.getImageData(0, 0, w, h);
  const data = image.data;
  const cutRatio = (stampDistress.value / 100) * 0.36;
  const cell = Math.max(4, stampDistressScale.value);

  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      const idx = (y * w + x) * 4;
      const r = data[idx];
      const g = data[idx + 1];
      const b = data[idx + 2];

      if (r > 242 && g > 242 && b > 242) continue;

      const nx = Math.floor(x / cell);
      const ny = Math.floor(y / cell);
      const grain = noise2d(nx, ny, 31);
      const edge = noise2d(x, y, 79);

      if (grain < cutRatio && edge < 0.6) {
        data[idx] = 255;
        data[idx + 1] = 255;
        data[idx + 2] = 255;
      }
    }
  }

  ctx.putImageData(image, 0, 0);
}

function drawRoundStamp(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const cx = w / 2;
  const cy = h / 2;
  const radius = Math.min(w, h) * 0.44;
  const ringWidth = stampBorderWidth.value;

  ctx.fillStyle = '#ffffff';
  ctx.fillRect(0, 0, w, h);

  ctx.save();
  ctx.beginPath();
  ctx.arc(cx, cy, radius - ringWidth, 0, Math.PI * 2);
  ctx.clip();
  drawPatternLayer(ctx, w, h);
  ctx.restore();

  ctx.strokeStyle = stampColor.value;
  ctx.lineWidth = ringWidth;
  ctx.beginPath();
  ctx.arc(cx, cy, radius, 0, Math.PI * 2);
  ctx.stroke();

  if (stampDoubleRing.value) {
    ctx.lineWidth = stampOuterRingWidth.value;
    ctx.beginPath();
    ctx.arc(cx, cy, radius + ringWidth / 2 + stampOuterRingGap.value, 0, Math.PI * 2);
    ctx.stroke();
  }

  ctx.fillStyle = stampColor.value;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';

  ctx.font = `${stampMainSize.value}px "KaiTi", "STKaiti", serif`;
  drawArcText(ctx, stampMainText.value, cx, cy, radius - ringWidth * 1.8, Math.PI * 0.83, Math.PI * 0.17);

  if (stampSubText.value.trim()) {
    ctx.font = `${stampSubSize.value}px "Helvetica", "Arial", sans-serif`;
    if (stampSubArc.value) {
      drawArcText(ctx, stampSubText.value, cx, cy, radius - ringWidth * 1.9, Math.PI * 0.17, Math.PI * 0.83, true, true);
    } else {
      ctx.fillText(stampSubText.value, cx, cy + radius * 0.54);
    }
  }

  if (stampCenterText.value.trim()) {
    ctx.font = `${stampCenterSize.value}px "KaiTi", "STKaiti", serif`;
    ctx.fillText(stampCenterText.value, cx, cy + radius * 0.02);
  }
}

function drawRoundedRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number) {
  const radius = Math.max(0, Math.min(r, Math.min(w, h) / 2));
  ctx.beginPath();
  ctx.moveTo(x + radius, y);
  ctx.lineTo(x + w - radius, y);
  ctx.quadraticCurveTo(x + w, y, x + w, y + radius);
  ctx.lineTo(x + w, y + h - radius);
  ctx.quadraticCurveTo(x + w, y + h, x + w - radius, y + h);
  ctx.lineTo(x + radius, y + h);
  ctx.quadraticCurveTo(x, y + h, x, y + h - radius);
  ctx.lineTo(x, y + radius);
  ctx.quadraticCurveTo(x, y, x + radius, y);
  ctx.closePath();
}

function drawRectStamp(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const pad = Math.min(w, h) * 0.12;
  const innerX = pad;
  const innerY = pad;
  const innerW = w - pad * 2;
  const innerH = h - pad * 2;
  const border = stampBorderWidth.value;

  ctx.fillStyle = '#ffffff';
  ctx.fillRect(0, 0, w, h);

  ctx.save();
  drawRoundedRect(ctx, innerX + border * 0.7, innerY + border * 0.7, innerW - border * 1.4, innerH - border * 1.4, Math.min(innerW, innerH) * 0.08);
  ctx.clip();
  drawPatternLayer(ctx, w, h);
  ctx.restore();

  ctx.strokeStyle = stampColor.value;
  ctx.lineWidth = border;
  drawRoundedRect(ctx, innerX, innerY, innerW, innerH, Math.min(innerW, innerH) * 0.1);
  ctx.stroke();

  ctx.fillStyle = stampColor.value;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.font = `${stampMainSize.value}px "KaiTi", "STKaiti", serif`;
  ctx.fillText(stampMainText.value, w / 2, h / 2 - h * 0.08);

  if (stampSubText.value.trim()) {
    ctx.font = `${stampSubSize.value}px "Helvetica", "Arial", sans-serif`;
    ctx.fillText(stampSubText.value, w / 2, h / 2 + h * 0.2);
  }

  if (stampDoubleRing.value) {
    ctx.lineWidth = Math.max(2, stampOuterRingWidth.value);
    drawRoundedRect(
      ctx,
      innerX - stampOuterRingGap.value,
      innerY - stampOuterRingGap.value,
      innerW + stampOuterRingGap.value * 2,
      innerH + stampOuterRingGap.value * 2,
      Math.min(innerW, innerH) * 0.12,
    );
    ctx.stroke();
  }
}

function renderTextStamp() {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const { w, h } = getTargetSizePx();
  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext('2d')!;
  ctx.clearRect(0, 0, w, h);

  if (stampShape.value === 'round') {
    drawRoundStamp(ctx, w, h);
  } else {
    drawRectStamp(ctx, w, h);
  }

  applyDistress(ctx, w, h);
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
  renderDot(ctx, data, w, h);
}

function renderOutput() {
  if (generatorMode.value === 'text') {
    renderTextStamp();
    return;
  }
  renderHalftone();
}

function applyPreset(type: 'soft' | 'balanced' | 'stamp') {
  preset.value = type;
  if (type === 'soft') {
    spacing.value = 12;
    dotScale.value = 0.75;
    contrast.value = 130;
    brightness.value = 108;
    fitScale.value = 82;
    return;
  }
  if (type === 'balanced') {
    spacing.value = 10;
    dotScale.value = 0.9;
    contrast.value = 150;
    brightness.value = 100;
    fitScale.value = 82;
    return;
  }
  spacing.value = 8;
  dotScale.value = 1;
  contrast.value = 190;
  brightness.value = 92;
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
  if (generatorMode.value === 'image' && !hasImage.value) return;
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    renderOutput();
  }, 250);
}

watch([spacing, dotScale, contrast, brightness, invert], scheduleRender);
watch([generatorMode, stampShape, stampMainText, stampSubText, stampCenterText, stampColor, stampBorderWidth, stampMainSize, stampSubSize, stampCenterSize, stampPattern, stampPatternDensity, stampPatternOpacity, stampDoubleRing, stampOuterRingWidth, stampOuterRingGap, stampSubArc, stampDistress, stampDistressScale], scheduleRender);

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

onMounted(() => {
  if (route.path === '/halftone') {
    generatorMode.value = 'image';
  }
  if (generatorMode.value === 'text') {
    renderOutput();
  }
});

watch(
  () => route.path,
  (path) => {
    if (path === '/halftone') {
      generatorMode.value = 'image';
      if (canRender.value) {
        renderOutput();
      }
    }
  },
);

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer);
});

// ---- 导出 PNG ----
function exportPng() {
  const canvas = canvasRef.value;
  if (!canvas || !canRender.value) {
    ElMessage.warning('请先准备内容');
    return;
  }
  if (canvas.width === 0 || canvas.height === 0) {
    renderOutput();
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
      <p class="page-desc">支持图片点阵半调与文字印章（含背景纹样）制作</p>
      <el-tabs v-model="generatorMode" class="generator-tabs" stretch>
        <el-tab-pane label="文字印章" name="text" />
        <el-tab-pane label="点阵印章" name="image" />
      </el-tabs>
    </div>

    <div class="halftone-layout">
      <!-- 左侧控制面板 -->
      <div class="control-panel">
        <!-- 上传区 -->
        <el-card class="panel-card">
          <template #header><span class="card-title">素材来源</span></template>
          <div v-if="generatorMode === 'image'">
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
          </div>
          <div v-else class="text-mode-hint">
            <p>文字印章模式无需上传图片，直接在右侧配置文字与背景纹样。</p>
          </div>
        </el-card>

        <!-- 效果参数 -->
        <el-card class="panel-card">
          <template #header><span class="card-title">效果参数</span></template>

          <div class="param-row" v-if="generatorMode === 'image'">
            <span class="param-label">快速预设</span>
            <el-radio-group v-model="preset" size="small" @change="onPresetChange">
              <el-radio-button value="soft">柔和</el-radio-button>
              <el-radio-button value="balanced">均衡</el-radio-button>
              <el-radio-button value="stamp">印章</el-radio-button>
            </el-radio-group>
          </div>

          <template v-if="generatorMode === 'image'">

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
            <span class="param-label">反色（黑底白纹）</span>
            <el-switch v-model="invert" />
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

          <div class="param-row">
            <span class="param-label">点大小系数 <em>{{ dotScale }}</em></span>
            <el-slider v-model="dotScale" :min="0.1" :max="1.0" :step="0.05" />
          </div>
          </template>

          <template v-else>
            <div class="param-row">
              <span class="param-label">章体样式</span>
              <el-radio-group v-model="stampShape" size="small">
                <el-radio-button value="round">圆章</el-radio-button>
                <el-radio-button value="rect">方章</el-radio-button>
              </el-radio-group>
            </div>

            <div class="param-row">
              <span class="param-label">主文字</span>
              <el-input v-model="stampMainText" placeholder="请输入主文字" maxlength="14" show-word-limit />
            </div>

            <div class="param-row">
              <span class="param-label">副文字</span>
              <el-input v-model="stampSubText" placeholder="如公司名/编号" maxlength="20" show-word-limit />
            </div>

            <div class="param-row" v-if="stampShape === 'round'">
              <span class="param-label">中心字</span>
              <el-input v-model="stampCenterText" placeholder="如 福 / ★" maxlength="2" />
            </div>

            <div class="param-row">
              <span class="param-label">印章颜色</span>
              <el-color-picker v-model="stampColor" />
            </div>

            <div class="param-row">
              <span class="param-label">边框粗细 <em>{{ stampBorderWidth }}px</em></span>
              <el-slider v-model="stampBorderWidth" :min="4" :max="22" :step="1" />
            </div>

            <div class="param-row">
              <span class="param-label">双层外圈</span>
              <el-switch v-model="stampDoubleRing" />
            </div>

            <template v-if="stampDoubleRing">
              <div class="param-row">
                <span class="param-label">外圈线宽 <em>{{ stampOuterRingWidth }}px</em></span>
                <el-slider v-model="stampOuterRingWidth" :min="2" :max="12" :step="1" />
              </div>
              <div class="param-row">
                <span class="param-label">外圈间距 <em>{{ stampOuterRingGap }}px</em></span>
                <el-slider v-model="stampOuterRingGap" :min="4" :max="24" :step="1" />
              </div>
            </template>

            <div class="param-row">
              <span class="param-label">主字大小 <em>{{ stampMainSize }}px</em></span>
              <el-slider v-model="stampMainSize" :min="32" :max="120" :step="1" />
            </div>

            <div class="param-row">
              <span class="param-label">副字大小 <em>{{ stampSubSize }}px</em></span>
              <el-slider v-model="stampSubSize" :min="16" :max="60" :step="1" />
            </div>

            <div class="param-row" v-if="stampShape === 'round' && stampSubText.trim()">
              <span class="param-label">副文弧排</span>
              <el-switch v-model="stampSubArc" inline-prompt active-text="弧形" inactive-text="直排" />
            </div>

            <div class="param-row" v-if="stampShape === 'round'">
              <span class="param-label">中心字大小 <em>{{ stampCenterSize }}px</em></span>
              <el-slider v-model="stampCenterSize" :min="40" :max="160" :step="1" />
            </div>

            <div class="param-row">
              <span class="param-label">背景纹样</span>
              <el-select v-model="stampPattern" style="width: 100%">
                <el-option label="无" value="none" />
                <el-option label="点阵" value="dots" />
                <el-option label="网格" value="grid" />
                <el-option label="放射纹" value="sunburst" />
              </el-select>
            </div>

            <div class="param-row" v-if="stampPattern !== 'none'">
              <span class="param-label">纹样密度 <em>{{ stampPatternDensity }}</em></span>
              <el-slider v-model="stampPatternDensity" :min="6" :max="32" :step="1" />
            </div>

            <div class="param-row" v-if="stampPattern !== 'none'">
              <span class="param-label">纹样透明度 <em>{{ stampPatternOpacity }}%</em></span>
              <el-slider v-model="stampPatternOpacity" :min="5" :max="60" :step="1" />
            </div>

            <div class="param-row">
              <span class="param-label">做旧强度 <em>{{ stampDistress }}%</em></span>
              <el-slider v-model="stampDistress" :min="0" :max="60" :step="1" />
            </div>

            <div class="param-row" v-if="stampDistress > 0">
              <span class="param-label">做旧颗粒 <em>{{ stampDistressScale }}</em></span>
              <el-slider v-model="stampDistressScale" :min="6" :max="32" :step="1" />
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
          </template>

          <div class="param-row">
            <el-button size="small" @click="renderOutput">立即渲染</el-button>
          </div>
        </el-card>

        <!-- 操作按钮 -->
        <div class="action-row">
          <el-button type="primary" :disabled="!canRender" @click="renderOutput" style="flex:1">
            重新生成
          </el-button>
          <el-button type="success" :disabled="!canRender" @click="exportPng" style="flex:1">
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
              <span v-if="canRender" class="canvas-size-hint">
                {{ canvasRef?.width }} × {{ canvasRef?.height }} px ({{ formatMm(canvasRef?.width) }} × {{ formatMm(canvasRef?.height) }} mm)
              </span>
            </div>
            <div class="preview-actions" v-if="canRender">
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
            <div v-if="!canRender" class="placeholder">
              <el-icon :size="64" color="#dcdfe6"><i-ep-picture /></el-icon>
              <p>{{ generatorMode === 'image' ? '上传图片后在此预览效果' : '请先输入文字参数后渲染' }}</p>
            </div>
            <canvas
              v-show="canRender"
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

.generator-tabs {
  margin-top: 12px;
  max-width: 420px;
}

.generator-tabs :deep(.el-tabs__header) {
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

.text-mode-hint {
  font-size: 12px;
  color: #606266;
  line-height: 1.6;
  background: #f8f9fb;
  border: 1px dashed #d8dbe2;
  border-radius: 8px;
  padding: 12px;
}

.text-mode-hint p {
  margin: 0;
}

.param-row {
  margin-bottom: 16px;
}

.param-row :deep(.el-input__wrapper) {
  border-radius: 8px;
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
