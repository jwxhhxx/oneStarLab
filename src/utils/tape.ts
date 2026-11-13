import type { Layer, Psd } from 'ag-psd';

import type { TapeDesign, TapePatternLayer, TapePatternType } from '@/types';

export const MM_PER_INCH = 25.4;

export const TAPE_PATTERNS: Array<{ type: TapePatternType; label: string }> = [
  { type: 'dots', label: '波点' },
  { type: 'raindrops', label: '雨滴' },
  { type: 'stars', label: '星星' },
  { type: 'hearts', label: '爱心' },
  { type: 'keys', label: '钥匙' },
  { type: 'wings', label: '翅膀' },
  { type: 'moons', label: '月亮' },
  { type: 'flowers', label: '小花' },
  { type: 'clouds', label: '云朵' },
  { type: 'bows', label: '蝴蝶结' },
];

export const TAPE_PALETTE = [
  '#e97a8a',
  '#f2b36b',
  '#f3dd7c',
  '#8fc9a4',
  '#7cc2d6',
  '#8f9bea',
  '#c49ce4',
  '#5f6672',
];

export const TAPE_WIDTH_PRESETS = [10, 15, 20, 25, 30, 40, 50];
export const TAPE_LENGTH_PRESETS = [5, 10, 15, 20, 30, 40, 60];

const TAU = Math.PI * 2;
const MAX_CANVAS_SIDE = 16000;

export function tapePatternLabel(type: TapePatternType): string {
  return TAPE_PATTERNS.find((pattern) => pattern.type === type)?.label ?? type;
}

export function createTapeLayer(
  type: TapePatternType,
  index = 0,
  palette: string[] = TAPE_PALETTE,
): TapePatternLayer {
  return {
    id: `tape-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    type,
    rows: 2,
    cols: 6,
    sizeMm: 5,
    color: palette[index % Math.max(palette.length, 1)] ?? '#e97a8a',
    multi: false,
    rotation: 0,
    stagger: false,
    opacity: 1,
  };
}

function drawDots(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.arc(0, 0, 0.3, 0, TAU);
  ctx.fill();
}

function drawRaindrops(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.moveTo(0, -0.5);
  ctx.bezierCurveTo(0.2, -0.26, 0.34, 0.02, 0.34, 0.18);
  ctx.arc(0, 0.18, 0.34, 0, Math.PI, false);
  ctx.bezierCurveTo(-0.34, 0.02, -0.2, -0.26, 0, -0.5);
  ctx.closePath();
  ctx.fill();
}

function drawStars(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  for (let point = 0; point < 10; point += 1) {
    const radius = point % 2 === 0 ? 0.46 : 0.19;
    const angle = -Math.PI / 2 + (point * Math.PI) / 5;
    const x = Math.cos(angle) * radius;
    const y = Math.sin(angle) * radius;
    if (point === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  }
  ctx.closePath();
  ctx.fill();
}

function drawHearts(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.moveTo(0, 0.42);
  ctx.bezierCurveTo(-0.5, 0.06, -0.44, -0.36, -0.17, -0.36);
  ctx.bezierCurveTo(-0.05, -0.36, 0, -0.2, 0, -0.09);
  ctx.bezierCurveTo(0, -0.2, 0.05, -0.36, 0.17, -0.36);
  ctx.bezierCurveTo(0.44, -0.36, 0.5, 0.06, 0, 0.42);
  ctx.closePath();
  ctx.fill();
}

function drawKeys(ctx: CanvasRenderingContext2D) {
  ctx.lineWidth = 0.08;
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.beginPath();
  ctx.arc(0, -0.24, 0.16, 0, TAU);
  ctx.stroke();
  ctx.beginPath();
  ctx.moveTo(0, -0.06);
  ctx.lineTo(0, 0.47);
  ctx.stroke();
  ctx.beginPath();
  ctx.moveTo(0, 0.31);
  ctx.lineTo(0.17, 0.31);
  ctx.stroke();
  ctx.beginPath();
  ctx.moveTo(0, 0.17);
  ctx.lineTo(0.12, 0.17);
  ctx.stroke();
}

function drawWings(ctx: CanvasRenderingContext2D) {
  for (let feather = 0; feather < 3; feather += 1) {
    ctx.save();
    ctx.rotate(-0.32 + feather * 0.3);
    ctx.beginPath();
    ctx.moveTo(0.04, 0.14);
    ctx.bezierCurveTo(-0.18, 0.02, -0.42, -0.16, -0.38, -0.44);
    ctx.bezierCurveTo(-0.16, -0.44, 0.06, -0.3, 0.12, -0.06);
    ctx.bezierCurveTo(0.13, 0.03, 0.09, 0.09, 0.04, 0.14);
    ctx.closePath();
    ctx.fill();
    ctx.restore();
  }
}

function drawMoons(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.arc(0, 0, 0.46, 0, TAU);
  ctx.arc(0.2, -0.04, 0.38, 0, TAU, true);
  ctx.fill('evenodd');
}

function drawFlowers(ctx: CanvasRenderingContext2D) {
  for (let petal = 0; petal < 5; petal += 1) {
    ctx.save();
    ctx.rotate((petal * TAU) / 5);
    ctx.beginPath();
    ctx.ellipse(0, -0.21, 0.15, 0.22, 0, 0, TAU);
    ctx.fill();
    ctx.restore();
  }
  ctx.save();
  ctx.fillStyle = '#fff8f0';
  ctx.globalAlpha = 0.9;
  ctx.beginPath();
  ctx.arc(0, 0, 0.1, 0, TAU);
  ctx.fill();
  ctx.restore();
}

function drawClouds(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.arc(-0.2, 0.07, 0.18, 0, TAU);
  ctx.arc(0.02, 0, 0.26, 0, TAU);
  ctx.arc(0.24, 0.1, 0.16, 0, TAU);
  ctx.rect(-0.38, 0.06, 0.76, 0.24);
  ctx.fill();
}

function drawBows(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.moveTo(0, 0.02);
  ctx.bezierCurveTo(-0.16, -0.28, -0.48, -0.34, -0.48, -0.04);
  ctx.bezierCurveTo(-0.48, 0.2, -0.18, 0.2, 0, 0.04);
  ctx.closePath();
  ctx.fill();

  ctx.beginPath();
  ctx.moveTo(0, 0.02);
  ctx.bezierCurveTo(0.16, -0.28, 0.48, -0.34, 0.48, -0.04);
  ctx.bezierCurveTo(0.48, 0.2, 0.18, 0.2, 0, 0.04);
  ctx.closePath();
  ctx.fill();

  ctx.beginPath();
  ctx.moveTo(-0.05, 0.08);
  ctx.lineTo(-0.2, 0.47);
  ctx.lineTo(-0.01, 0.36);
  ctx.closePath();
  ctx.fill();

  ctx.beginPath();
  ctx.moveTo(0.05, 0.08);
  ctx.lineTo(0.2, 0.47);
  ctx.lineTo(0.01, 0.36);
  ctx.closePath();
  ctx.fill();

  ctx.beginPath();
  ctx.arc(0, 0.03, 0.1, 0, TAU);
  ctx.fill();
}

function drawMotif(
  ctx: CanvasRenderingContext2D,
  type: TapePatternType,
  cx: number,
  cy: number,
  size: number,
  color: string,
  rotation: number,
  opacity: number,
) {
  ctx.save();
  ctx.translate(cx, cy);
  ctx.rotate((rotation * Math.PI) / 180);
  ctx.scale(size, size);
  ctx.globalAlpha = opacity;
  ctx.fillStyle = color;
  ctx.strokeStyle = color;

  switch (type) {
    case 'dots':
      drawDots(ctx);
      break;
    case 'raindrops':
      drawRaindrops(ctx);
      break;
    case 'stars':
      drawStars(ctx);
      break;
    case 'hearts':
      drawHearts(ctx);
      break;
    case 'keys':
      drawKeys(ctx);
      break;
    case 'wings':
      drawWings(ctx);
      break;
    case 'moons':
      drawMoons(ctx);
      break;
    case 'flowers':
      drawFlowers(ctx);
      break;
    case 'clouds':
      drawClouds(ctx);
      break;
    case 'bows':
      drawBows(ctx);
      break;
    default:
      drawDots(ctx);
  }

  ctx.restore();
}

function drawLayer(
  ctx: CanvasRenderingContext2D,
  layer: TapePatternLayer,
  design: TapeDesign,
  pxPerMm: number,
  width: number,
  height: number,
  layerIndex: number,
) {
  const margin = Math.min(2.2 * pxPerMm, Math.min(width, height) * 0.12);
  const innerWidth = Math.max(width - margin * 2, 1);
  const innerHeight = Math.max(height - margin * 2, 1);
  const cellWidth = innerWidth / Math.max(layer.cols, 1);
  const cellHeight = innerHeight / Math.max(layer.rows, 1);
  const size = Math.min(layer.sizeMm * pxPerMm, Math.max(height * 0.92, 1));
  const palette = design.palette.length > 0 ? design.palette : TAPE_PALETTE;

  ctx.save();
  ctx.beginPath();
  ctx.rect(0, 0, width, height);
  ctx.clip();

  for (let row = 0; row < layer.rows; row += 1) {
    for (let col = 0; col < layer.cols; col += 1) {
      const offset = layer.stagger && row % 2 === 1 ? cellWidth / 2 : 0;
      const cx = margin + (col + 0.5) * cellWidth + offset;
      const cy = margin + (row + 0.5) * cellHeight;
      const color = layer.multi
        ? palette[(row * layer.cols + col + layerIndex) % palette.length]
        : layer.color;
      drawMotif(ctx, layer.type, cx, cy, size, color, layer.rotation, layer.opacity);
    }
  }

  ctx.restore();
}

export function drawTapeDesign(
  ctx: CanvasRenderingContext2D,
  design: TapeDesign,
  pxPerMm: number,
): void {
  const width = ctx.canvas.width;
  const height = ctx.canvas.height;

  ctx.clearRect(0, 0, width, height);
  if (!design.transparent) {
    ctx.fillStyle = design.background;
    ctx.fillRect(0, 0, width, height);
  }

  design.layers.forEach((layer, index) => {
    drawLayer(ctx, layer, design, pxPerMm, width, height, index);
  });
}

export function createTapeCanvas(design: TapeDesign, pxPerMm: number): HTMLCanvasElement {
  const width = Math.max(1, Math.round(design.lengthMm * pxPerMm));
  const height = Math.max(1, Math.round(design.widthMm * pxPerMm));
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext('2d');

  if (!ctx) {
    throw new Error('当前浏览器无法创建画布');
  }

  drawTapeDesign(ctx, design, pxPerMm);
  return canvas;
}

function assertExportSize(design: TapeDesign, dpi: number) {
  const pxPerMm = dpi / MM_PER_INCH;
  const width = Math.round(design.lengthMm * pxPerMm);
  const height = Math.round(design.widthMm * pxPerMm);

  if (width > MAX_CANVAS_SIDE || height > MAX_CANVAS_SIDE) {
    throw new Error('导出尺寸过大，请缩短胶带长度或降低 DPI');
  }

  return { pxPerMm, width, height };
}

export async function exportTapePng(design: TapeDesign, dpi: number): Promise<Blob> {
  assertExportSize(design, dpi);
  const canvas = createTapeCanvas(design, dpi / MM_PER_INCH);

  return new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) resolve(blob);
      else reject(new Error('PNG 导出失败'));
    }, 'image/png');
  });
}

export async function exportTapePsd(design: TapeDesign, dpi: number): Promise<Blob> {
  const { pxPerMm, width, height } = assertExportSize(design, dpi);
  const { writePsd } = await import('ag-psd');
  const composite = createTapeCanvas(design, pxPerMm);
  const children: Layer[] = [];

  if (!design.transparent) {
    const background = document.createElement('canvas');
    background.width = width;
    background.height = height;
    const backgroundCtx = background.getContext('2d');
    if (backgroundCtx) {
      backgroundCtx.fillStyle = design.background;
      backgroundCtx.fillRect(0, 0, width, height);
      children.push({
        name: '底色',
        canvas: background,
        left: 0,
        top: 0,
        right: width,
        bottom: height,
      });
    }
  }

  design.layers.forEach((layer, index) => {
    const canvas = document.createElement('canvas');
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    drawLayer(ctx, layer, design, pxPerMm, width, height, index);
    children.push({
      name: `${tapePatternLabel(layer.type)} ${index + 1}`,
      canvas,
      left: 0,
      top: 0,
      right: width,
      bottom: height,
    });
  });

  const psd: Psd = { width, height, canvas: composite, children };
  const buffer = writePsd(psd);

  return new Blob([buffer], { type: 'image/vnd.adobe.photoshop' });
}
