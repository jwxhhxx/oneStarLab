<script setup lang="ts">
import { computed } from 'vue';

import type { PlannerModuleConfig } from '@/types';

const props = defineProps<{
  block: PlannerModuleConfig;
  accent: string;
}>();

const weekdays = ['一', '二', '三', '四', '五', '六', '日'];
const weekdayNames = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
const mealNames = ['早餐', '午餐', '晚餐', '加餐 / 饮水'];

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

const rows = computed(() => clamp(Math.round(props.block.options.rows ?? 6), 1, 40));
const calRows = computed(() => clamp(Math.round(props.block.options.rows ?? 6), 4, 8));
const days = computed(() => clamp(Math.round(props.block.options.days ?? 31), 1, 31));
const cellMm = computed(() => clamp(props.block.options.cellMm ?? 5, 2, 20));
const lineMm = computed(() => clamp(props.block.options.lineMm ?? 7, 4, 20));

const hours = computed(() => {
  const start = clamp(Math.round(props.block.options.startHour ?? 6), 0, 23);
  const end = clamp(Math.round(props.block.options.endHour ?? 22), start + 1, 24);

  return Array.from({ length: end - start }, (_, index) => start + index);
});

const quadrantLabels = computed(() => {
  const labels = props.block.options.labels;

  return labels && labels.length === 4 ? labels : ['重要且紧急', '重要不紧急', '紧急不重要', '不重要不紧急'];
});

function padHour(hour: number) {
  return String(hour).padStart(2, '0');
}

function hexToRgb(hex: string) {
  const normalized = hex.replace('#', '');
  const full = normalized.length === 3 ? normalized.split('').map((char) => char + char).join('') : normalized;
  const value = Number.parseInt(full, 16);

  if (Number.isNaN(value)) {
    return { r: 108, g: 124, b: 255 };
  }

  return { r: (value >> 16) & 255, g: (value >> 8) & 255, b: value & 255 };
}

function rgba(hex: string, alpha: number) {
  const { r, g, b } = hexToRgb(hex);

  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

const accentColor = computed(() => props.block.accent || props.accent);

const blockStyle = computed(() => ({
  flex: `${clamp(props.block.weight, 1, 8)} 1 0`,
  '--accent': accentColor.value,
  '--line': rgba(accentColor.value, 0.55),
  '--soft': rgba(accentColor.value, 0.24),
  '--fill': rgba(accentColor.value, 0.08),
  '--ink-soft': rgba(accentColor.value, 0.72),
}));

const gridStyle = computed(() => {
  const size = `${cellMm.value}mm`;

  return {
    backgroundImage:
      `repeating-linear-gradient(to right, var(--soft) 0 0.12mm, transparent 0.12mm ${size}),` +
      `repeating-linear-gradient(to bottom, var(--soft) 0 0.12mm, transparent 0.12mm ${size})`,
  };
});

const dotStyle = computed(() => ({
  backgroundImage: 'radial-gradient(circle, var(--line) 0 0.14mm, transparent 0.16mm)',
  backgroundSize: `${cellMm.value}mm ${cellMm.value}mm`,
}));

const linedStyle = computed(() => ({
  backgroundImage: `repeating-linear-gradient(to bottom, var(--soft) 0 0.12mm, transparent 0.12mm ${lineMm.value}mm)`,
}));

const frenchStyle = computed(() => {
  const size = `${cellMm.value}mm`;
  const guide = `${Math.max(2, cellMm.value / 4)}mm`;

  return {
    backgroundImage:
      `repeating-linear-gradient(to bottom, var(--line) 0 0.16mm, transparent 0.16mm ${size}),` +
      `repeating-linear-gradient(to bottom, var(--soft) 0 0.08mm, transparent 0.08mm ${guide}),` +
      `repeating-linear-gradient(to right, var(--soft) 0 0.08mm, transparent 0.08mm ${size})`,
  };
});
</script>

<template>
  <section class="planner-block" :style="blockStyle">
    <header v-if="block.title" class="pb-head">
      <span class="pb-title">{{ block.title }}</span>
      <span class="pb-head-rule"></span>
    </header>

    <div class="pb-body">
      <div v-if="block.type === 'calendar'" class="pb-calendar">
        <div class="pb-calendar-week">
          <span v-for="day in weekdays" :key="day">{{ day }}</span>
        </div>
        <div class="pb-calendar-grid" :style="{ gridTemplateRows: `repeat(${calRows}, minmax(0, 1fr))` }">
          <span v-for="cell in calRows * 7" :key="cell"></span>
        </div>
      </div>

      <div v-else-if="block.type === 'timeline'" class="pb-timeline">
        <div v-for="hour in hours" :key="hour" class="pb-time-row">
          <span class="pb-time">{{ padHour(hour) }}:00</span>
          <span class="pb-rule"></span>
        </div>
      </div>

      <div v-else-if="block.type === 'todo'" class="pb-todo">
        <div v-for="row in rows" :key="row" class="pb-todo-row">
          <span class="pb-check"></span>
          <span class="pb-rule"></span>
        </div>
      </div>

      <div v-else-if="block.type === 'quadrant'" class="pb-quadrant">
        <div v-for="(label, index) in quadrantLabels" :key="index" class="pb-quadrant-cell">
          <span class="pb-quadrant-label">{{ label }}</span>
          <span v-for="line in 3" :key="line" class="pb-quadrant-line"></span>
        </div>
      </div>

      <div v-else-if="block.type === 'ledger'" class="pb-table pb-table--ledger">
        <div class="pb-table-head">
          <span>日期</span>
          <span>项目</span>
          <span>收入</span>
          <span>支出</span>
        </div>
        <div v-for="row in rows" :key="row" class="pb-table-row">
          <span v-for="cell in 4" :key="cell"></span>
        </div>
      </div>

      <div v-else-if="block.type === 'finance'" class="pb-table pb-table--finance">
        <div class="pb-table-head">
          <span>月份</span>
          <span>收入</span>
          <span>支出</span>
          <span>结余</span>
        </div>
        <div v-for="row in rows" :key="row" class="pb-table-row">
          <span class="pb-table-month">{{ row }}月</span>
          <span></span>
          <span></span>
          <span></span>
        </div>
      </div>

      <div v-else-if="block.type === 'frenchGrid'" class="pb-pattern" :style="frenchStyle"></div>

      <div v-else-if="block.type === 'dailyPlan'" class="pb-daily">
        <div class="pb-daily-priority">
          <div v-for="index in 3" :key="index" class="pb-daily-priority-row">
            <span class="pb-priority-index">{{ index }}</span>
            <span class="pb-rule"></span>
          </div>
        </div>
        <div class="pb-daily-hours">
          <div v-for="hour in hours" :key="hour" class="pb-daily-hour">
            <span class="pb-time">{{ padHour(hour) }}</span>
            <span class="pb-rule"></span>
          </div>
        </div>
      </div>

      <div v-else-if="block.type === 'meals'" class="pb-meals">
        <div v-for="meal in mealNames" :key="meal" class="pb-meal">
          <span class="pb-meal-name">{{ meal }}</span>
          <span class="pb-meal-lines">
            <span v-for="line in 2" :key="line" class="pb-meal-line"></span>
          </span>
        </div>
      </div>

      <div
        v-else-if="block.type === 'habitTracker'"
        class="pb-habit"
        :style="{
          gridTemplateColumns: `minmax(6mm, 1fr) repeat(${days}, minmax(0, 1fr))`,
          gridTemplateRows: `auto repeat(${rows}, minmax(0, 1fr))`,
        }"
      >
        <span class="pb-habit-name">习惯</span>
        <span v-for="day in days" :key="`day-${day}`" class="pb-habit-day">{{ day }}</span>
        <template v-for="row in rows" :key="`row-${row}`">
          <span class="pb-habit-name"></span>
          <span v-for="day in days" :key="`cell-${row}-${day}`" class="pb-habit-cell"></span>
        </template>
      </div>

      <div v-else-if="block.type === 'weeklyPlan'" class="pb-weekly">
        <div v-for="day in weekdayNames" :key="day" class="pb-weekly-row">
          <span class="pb-weekly-day">{{ day }}</span>
          <span class="pb-weekly-lines">
            <span v-for="line in rows" :key="line" class="pb-weekly-line"></span>
          </span>
        </div>
      </div>

      <div v-else-if="block.type === 'grid'" class="pb-pattern" :style="gridStyle"></div>

      <div v-else-if="block.type === 'dotGrid'" class="pb-pattern" :style="dotStyle"></div>

      <div v-else-if="block.type === 'lined'" class="pb-pattern" :style="linedStyle"></div>

      <div v-else class="pb-blank"></div>
    </div>
  </section>
</template>

<style scoped>
.planner-block {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  break-inside: avoid;
}

.pb-head {
  display: flex;
  align-items: center;
  gap: 1.2mm;
  margin-bottom: 0.5mm;
  color: var(--accent);
  flex: none;
}

.pb-title {
  max-width: 70%;
  overflow: hidden;
  font-size: 1.05em;
  font-weight: 600;
  letter-spacing: 0.04em;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.pb-head-rule {
  flex: 1;
  height: 0.18mm;
  background: var(--line);
}

.pb-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-rule {
  flex: 1;
  border-bottom: 0.1mm solid var(--soft);
}

.pb-time {
  flex: none;
  width: 7.2mm;
  color: var(--ink-soft);
  font-size: 0.78em;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.pb-calendar {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-calendar-week {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  color: var(--ink-soft);
  font-size: 0.72em;
  text-align: center;
}

.pb-calendar-week span {
  padding-bottom: 0.4mm;
}

.pb-calendar-grid {
  display: grid;
  flex: 1;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  min-height: 0;
  border-top: 0.12mm solid var(--soft);
  border-left: 0.12mm solid var(--soft);
}

.pb-calendar-grid span {
  border-right: 0.1mm solid var(--soft);
  border-bottom: 0.1mm solid var(--soft);
}

.pb-timeline,
.pb-todo,
.pb-daily-hours {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-time-row,
.pb-todo-row,
.pb-daily-hour {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 1.4mm;
  min-height: 0;
}

.pb-check {
  flex: none;
  width: 2.4mm;
  height: 2.4mm;
  border: 0.14mm solid var(--line);
  border-radius: 0.35mm;
}

.pb-quadrant {
  display: grid;
  flex: 1;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  min-height: 0;
  border-top: 0.12mm solid var(--soft);
  border-left: 0.12mm solid var(--soft);
}

.pb-quadrant-cell {
  display: flex;
  flex-direction: column;
  gap: 0.5mm;
  min-height: 0;
  padding: 0.8mm;
  border-right: 0.12mm solid var(--soft);
  border-bottom: 0.12mm solid var(--soft);
}

.pb-quadrant-label {
  overflow: hidden;
  color: var(--accent);
  font-size: 0.8em;
  font-weight: 600;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.pb-quadrant-line {
  flex: 1;
  border-bottom: 0.1mm solid var(--soft);
}

.pb-table {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  border: 0.12mm solid var(--soft);
}

.pb-table-head,
.pb-table-row {
  display: grid;
  min-height: 0;
}

.pb-table--ledger .pb-table-head,
.pb-table--ledger .pb-table-row {
  grid-template-columns: 1.1fr 2.1fr 1fr 1fr;
}

.pb-table--finance .pb-table-head,
.pb-table--finance .pb-table-row {
  grid-template-columns: 0.9fr 1.3fr 1.3fr 1.2fr;
}

.pb-table-head {
  flex: none;
  background: var(--fill);
  color: var(--accent);
  font-size: 0.76em;
  font-weight: 600;
}

.pb-table-row {
  flex: 1;
  font-size: 0.76em;
}

.pb-table-head span,
.pb-table-row span {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 0 0.4mm;
  border-right: 0.1mm solid var(--soft);
}

.pb-table-head span:last-child,
.pb-table-row span:last-child {
  border-right: none;
}

.pb-table-row + .pb-table-row {
  border-top: 0.1mm solid var(--soft);
}

.pb-table-row span:first-child {
  border-top: none;
}

.pb-table-month {
  color: var(--ink-soft);
}

.pb-pattern {
  flex: 1;
  min-height: 0;
}

.pb-daily {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-daily-priority {
  display: flex;
  flex: none;
  flex-direction: column;
  gap: 0.4mm;
  margin-bottom: 0.8mm;
}

.pb-daily-priority-row {
  display: flex;
  align-items: center;
  gap: 1.2mm;
  min-height: 0;
}

.pb-priority-index {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 2.6mm;
  height: 2.6mm;
  border: 0.14mm solid var(--line);
  border-radius: 50%;
  color: var(--accent);
  font-size: 0.66em;
  line-height: 1;
}

.pb-meals {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-meal {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.4mm;
  min-height: 0;
  padding: 0.4mm 0;
}

.pb-meal-name {
  color: var(--accent);
  font-size: 0.82em;
  font-weight: 600;
}

.pb-meal-lines {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-meal-line {
  flex: 1;
  border-bottom: 0.1mm solid var(--soft);
}

.pb-habit {
  display: grid;
  flex: 1;
  min-height: 0;
  border-top: 0.1mm solid var(--soft);
  border-left: 0.1mm solid var(--soft);
}

.pb-habit-name,
.pb-habit-day,
.pb-habit-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-right: 0.1mm solid var(--soft);
  border-bottom: 0.1mm solid var(--soft);
}

.pb-habit-name {
  justify-content: flex-start;
  padding-left: 0.6mm;
  color: var(--ink-soft);
  font-size: 0.66em;
  white-space: nowrap;
}

.pb-habit-day {
  color: var(--accent);
  font-size: 0.58em;
}

.pb-weekly {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-weekly-row {
  display: flex;
  flex: 1;
  align-items: stretch;
  gap: 1.2mm;
  min-height: 0;
}

.pb-weekly-day {
  flex: none;
  width: 6.4mm;
  color: var(--accent);
  font-size: 0.76em;
  line-height: 1.2;
}

.pb-weekly-lines {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.pb-weekly-line {
  flex: 1;
  border-bottom: 0.1mm solid var(--soft);
}

.pb-blank {
  flex: 1;
}
</style>
