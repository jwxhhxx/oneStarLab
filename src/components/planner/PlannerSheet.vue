<script setup lang="ts">
import { computed } from 'vue';

import type { PlannerModuleConfig } from '@/types';

import PlannerBlock from './PlannerBlock.vue';

const props = defineProps<{
  widthMm: number;
  heightMm: number;
  paddingMm: number;
  accent: string;
  paper: string;
  modules: PlannerModuleConfig[];
}>();

const fontPt = computed(() => Math.min(9, Math.max(5.2, props.widthMm * 0.07)));
const gapMm = computed(() => Math.min(2.6, Math.max(0.8, props.paddingMm * 0.45)));

const sheetStyle = computed(() => ({
  width: `${props.widthMm}mm`,
  height: `${props.heightMm}mm`,
  padding: `${props.paddingMm}mm`,
  gap: `${gapMm.value}mm`,
  fontSize: `${fontPt.value}pt`,
  background: props.paper,
  '--planner-accent': props.accent,
  '--planner-paper': props.paper,
  '--planner-ink': '#393634',
}));
</script>

<template>
  <div class="planner-sheet" :style="sheetStyle">
    <PlannerBlock v-for="block in modules" :key="block.id" :block="block" :accent="accent" />
  </div>
</template>

<style scoped>
.planner-sheet {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--planner-ink);
  font-family: 'Noto Sans SC', 'PingFang SC', 'Microsoft YaHei', 'Hiragino Sans GB', sans-serif;
  line-height: 1.28;
  print-color-adjust: exact;
  -webkit-print-color-adjust: exact;
}
</style>
