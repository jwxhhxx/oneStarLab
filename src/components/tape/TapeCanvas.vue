<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';

import type { TapeDesign } from '@/types';
import { drawTapeDesign } from '@/utils/tape';

const props = defineProps<{ design: TapeDesign }>();

const canvasRef = ref<HTMLCanvasElement | null>(null);
const PREVIEW_TARGET_WIDTH = 1400;

function render() {
  const canvas = canvasRef.value;
  if (!canvas) return;

  const pxPerMm = Math.min(
    16,
    Math.max(3.4, PREVIEW_TARGET_WIDTH / Math.max(props.design.lengthMm, 1)),
  );
  canvas.width = Math.max(1, Math.round(props.design.lengthMm * pxPerMm));
  canvas.height = Math.max(1, Math.round(props.design.widthMm * pxPerMm));

  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  drawTapeDesign(ctx, props.design, pxPerMm);
}

watch(() => props.design, render, { deep: true });
onMounted(render);
</script>

<template>
  <canvas ref="canvasRef" class="tape-canvas" />
</template>

<style scoped>
.tape-canvas {
  display: block;
  width: 100%;
  height: auto;
}
</style>
