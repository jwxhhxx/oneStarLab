<script setup lang="ts">
import dayjs from 'dayjs';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

import PlannerSheet from '@/components/planner/PlannerSheet.vue';
import { useShopStore } from '@/stores/useShopStore';
import type {
  PlannerModuleConfig,
  PlannerModuleOptions,
  PlannerModuleType,
  PlannerSizeKey,
} from '@/types';

interface SizePreset {
  key: PlannerSizeKey;
  label: string;
  widthMm: number;
  heightMm: number;
  note?: string;
}

interface ModuleDefinition {
  type: PlannerModuleType;
  label: string;
  title: string;
  weight: number;
  options: PlannerModuleOptions;
}

interface ComboPreset {
  key: string;
  label: string;
  types: PlannerModuleType[];
}

interface Palette {
  label: string;
  accent: string;
  paper: string;
}

const store = useShopStore();

const PAGE_STYLE_ID = 'planner-page-size';
const MM_TO_PX = 96 / 25.4;

const sizePresets: SizePreset[] = [
  { key: 'm5', label: 'M5 手帐内页', widthMm: 62, heightMm: 105, note: '日系标准 62×105；部分品牌宽版为 74×105，可切到自定义微调。' },
  { key: 'm25', label: 'M2.5 手帐内页', widthMm: 42, heightMm: 105, note: 'M2.5 常见规格 42×105，适合极窄时间轴或打卡条。' },
  { key: 'a7', label: 'A7', widthMm: 74, heightMm: 105 },
  { key: 'a6per', label: 'A6per 个人', widthMm: 95, heightMm: 171, note: 'A6 personal 常见 95×171；Per Wide 约 120×171，可用自定义。' },
  { key: 'a6', label: 'A6', widthMm: 105, heightMm: 148 },
  { key: 'a5', label: 'A5', widthMm: 148, heightMm: 210 },
  { key: 'cm5', label: '5×5cm 方块', widthMm: 50, heightMm: 50 },
  { key: 'cm8', label: '8×8cm 方块', widthMm: 80, heightMm: 80 },
  { key: 'trifold', label: '三折页（A4 横向单页）', widthMm: 99, heightMm: 210, note: 'A4 横向三等分后的单页尺寸 99×210，打印后折叠。' },
  { key: 'custom', label: '自定义尺寸', widthMm: 105, heightMm: 148 },
];

const moduleLibrary: ModuleDefinition[] = [
  { type: 'calendar', label: '月历', title: '月历', weight: 3, options: { rows: 6 } },
  { type: 'weeklyPlan', label: '周计划', title: '本周计划', weight: 3, options: { rows: 2 } },
  { type: 'dailyPlan', label: '日计划', title: '今日计划', weight: 3, options: { startHour: 7, endHour: 22 } },
  { type: 'timeline', label: '时间轴', title: '时间轴', weight: 3, options: { startHour: 6, endHour: 23 } },
  { type: 'todo', label: '待办', title: '待办清单', weight: 2, options: { rows: 8 } },
  {
    type: 'quadrant',
    label: '四象限',
    title: '四象限',
    weight: 3,
    options: { labels: ['重要且紧急', '重要不紧急', '紧急不重要', '不重要不紧急'] },
  },
  { type: 'habitTracker', label: '月打卡', title: '月打卡', weight: 3, options: { rows: 8, days: 31 } },
  { type: 'meals', label: '三餐', title: '三餐记录', weight: 2, options: {} },
  { type: 'ledger', label: '账单', title: '账单', weight: 3, options: { rows: 10 } },
  { type: 'finance', label: '财务表', title: '财务表', weight: 3, options: { rows: 12 } },
  { type: 'frenchGrid', label: '法文格', title: '法文格', weight: 4, options: { cellMm: 8 } },
  { type: 'grid', label: '方格', title: '方格纸', weight: 4, options: { cellMm: 5 } },
  { type: 'dotGrid', label: '点阵', title: '点阵纸', weight: 4, options: { cellMm: 5 } },
  { type: 'lined', label: '横线', title: '横线纸', weight: 4, options: { lineMm: 7 } },
  { type: 'blank', label: '空白', title: '空白', weight: 4, options: {} },
];

const comboPresets: ComboPreset[] = [
  { key: 'daily', label: '日计划组合', types: ['dailyPlan', 'todo', 'meals'] },
  { key: 'weekly', label: '周计划组合', types: ['weeklyPlan', 'todo'] },
  { key: 'monthly', label: '月计划组合', types: ['calendar', 'habitTracker'] },
  { key: 'quadrant', label: '四象限组合', types: ['quadrant', 'todo'] },
  { key: 'finance', label: '记账组合', types: ['ledger', 'finance'] },
  { key: 'study', label: '学习组合', types: ['frenchGrid', 'todo'] },
  { key: 'grid', label: '空白网格', types: ['grid'] },
];

const palettes: Palette[] = [
  { label: '奶油', accent: '#b98a63', paper: '#fffaf2' },
  { label: '雾蓝', accent: '#6c7cff', paper: '#f7f9ff' },
  { label: '豆沙', accent: '#c97b8c', paper: '#fff7f8' },
  { label: '抹茶', accent: '#7a9b76', paper: '#f8fbf5' },
  { label: '灰紫', accent: '#8a7fb0', paper: '#f9f8ff' },
  { label: '墨黑', accent: '#4b4b4b', paper: '#ffffff' },
];

const moduleLabelMap = Object.fromEntries(moduleLibrary.map((item) => [item.type, item.label])) as Record<
  PlannerModuleType,
  string
>;

const sizeKey = ref<PlannerSizeKey>('a6');
const widthMm = ref(105);
const heightMm = ref(148);
const paddingMm = ref(4);
const accent = ref('#6c7cff');
const paper = ref('#f7f9ff');
const modules = ref<PlannerModuleConfig[]>([]);
const copies = ref(1);
const selectedTemplateId = ref<number>();

const stageRef = ref<HTMLElement>();
const previewScale = ref(1);
let resizeObserver: ResizeObserver | null = null;

function uid() {
  return `block-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

function makeModule(definition: ModuleDefinition): PlannerModuleConfig {
  return {
    id: uid(),
    type: definition.type,
    title: definition.title,
    weight: definition.weight,
    accent: accent.value,
    options: {
      ...definition.options,
      labels: definition.options.labels ? [...definition.options.labels] : undefined,
    },
  };
}

function buildCombo(types: PlannerModuleType[]) {
  return types.flatMap((type) => {
    const definition = moduleLibrary.find((item) => item.type === type);

    return definition ? [makeModule(definition)] : [];
  });
}

const currentPreset = computed(() => sizePresets.find((preset) => preset.key === sizeKey.value));

const templateOptions = computed(() =>
  store.plannerTemplates.flatMap((template) =>
    template.id == null
      ? []
      : [
          {
            value: template.id,
            label: `${template.name} · ${template.widthMm}×${template.heightMm}mm · ${dayjs(template.updatedAt).format('MM-DD HH:mm')}`,
          },
        ],
  ),
);

const previewWidthPx = computed(() => widthMm.value * MM_TO_PX * previewScale.value);
const previewHeightPx = computed(() => heightMm.value * MM_TO_PX * previewScale.value);

function applySize(key: PlannerSizeKey) {
  const preset = sizePresets.find((item) => item.key === key);

  if (!preset || key === 'custom') {
    sizeKey.value = key;
    return;
  }

  widthMm.value = preset.widthMm;
  heightMm.value = preset.heightMm;
  sizeKey.value = key;
}

function swapOrientation() {
  const nextWidth = heightMm.value;
  heightMm.value = widthMm.value;
  widthMm.value = nextWidth;
  sizeKey.value = 'custom';
}

function applyPalette(palette: Palette) {
  accent.value = palette.accent;
  paper.value = palette.paper;
}

function unifyAccent() {
  modules.value.forEach((block) => {
    block.accent = accent.value;
  });
}

function addModule(definition: ModuleDefinition) {
  if (modules.value.length >= 10) {
    ElMessage.warning('一页最多放 10 个模块');
    return;
  }

  modules.value.push(makeModule(definition));
}

function removeModule(id: string) {
  modules.value = modules.value.filter((block) => block.id !== id);
}

function moveModule(index: number, delta: number) {
  const target = index + delta;

  if (target < 0 || target >= modules.value.length) return;

  const next = [...modules.value];
  const [moved] = next.splice(index, 1);
  next.splice(target, 0, moved);
  modules.value = next;
}

async function applyCombo(combo: ComboPreset) {
  if (modules.value.length) {
    try {
      await ElMessageBox.confirm(`应用「${combo.label}」会替换当前 ${modules.value.length} 个模块。`, '应用组合', {
        confirmButtonText: '替换',
        cancelButtonText: '取消',
        type: 'warning',
      });
    } catch (error) {
      if (error === 'cancel' || error === 'close') return;
      return;
    }
  }

  modules.value = buildCombo(combo.types);
}

function maxRows(type: PlannerModuleType) {
  if (type === 'calendar') return 8;
  if (type === 'finance') return 12;
  if (type === 'habitTracker') return 20;
  if (type === 'weeklyPlan') return 4;

  return 30;
}

function moduleTypeLabel(type: PlannerModuleType) {
  return moduleLabelMap[type];
}

function labelsText(block: PlannerModuleConfig) {
  return (block.options.labels ?? []).join(' / ');
}

function setLabels(block: PlannerModuleConfig, value: string) {
  const labels = value
    .split(/[/、,，]/)
    .map((label) => label.trim())
    .filter(Boolean)
    .slice(0, 4);

  block.options = { ...block.options, labels };
}

function updatePreviewScale() {
  const stage = stageRef.value;

  if (!stage) return;

  const available = stage.clientWidth - 28;
  const natural = widthMm.value * MM_TO_PX;

  previewScale.value = natural > 0 ? Math.min(1, Math.max(0.15, available / natural)) : 1;
}

function syncPageStyle() {
  let styleElement = document.getElementById(PAGE_STYLE_ID) as HTMLStyleElement | null;

  if (!styleElement) {
    styleElement = document.createElement('style');
    styleElement.id = PAGE_STYLE_ID;
    document.head.appendChild(styleElement);
  }

  styleElement.textContent = `@page { size: ${widthMm.value}mm ${heightMm.value}mm; margin: 0; }`;
}

function handlePrint() {
  if (!modules.value.length) {
    ElMessage.warning('先添加至少一个模块再打印');
    return;
  }

  syncPageStyle();
  window.print();
}

async function handleSaveTemplate() {
  if (!modules.value.length) {
    ElMessage.warning('先添加至少一个模块再保存');
    return;
  }

  const suggested = `${currentPreset.value?.label ?? '自定义'} ${dayjs().format('MM-DD')}`;

  try {
    const { value } = await ElMessageBox.prompt('给这套内页起个名字，方便下次直接载入', '保存模板', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValue: suggested,
      inputValidator: (input) =>
        Boolean(input && input.trim().length > 0 && input.trim().length <= 30) || '请输入 1-30 个字符',
    });

    const id = await store.savePlannerTemplate({
      name: value.trim(),
      sizeKey: sizeKey.value,
      widthMm: widthMm.value,
      heightMm: heightMm.value,
      paddingMm: paddingMm.value,
      accent: accent.value,
      paper: paper.value,
      modules: modules.value,
    });

    selectedTemplateId.value = id;
    ElMessage.success(`模板已保存（#${id}）`);
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;

    ElMessage.error(error instanceof Error ? error.message : '保存模板失败');
  }
}

async function handleLoadTemplate() {
  if (selectedTemplateId.value == null) return;

  const template = store.plannerTemplates.find((item) => item.id === selectedTemplateId.value);

  if (!template) {
    ElMessage.warning('模板不存在，可能已被删除');
    return;
  }

  sizeKey.value = template.sizeKey;
  widthMm.value = template.widthMm;
  heightMm.value = template.heightMm;
  paddingMm.value = template.paddingMm;
  accent.value = template.accent;
  paper.value = template.paper;
  modules.value = template.modules.map((block) => ({
    ...block,
    id: uid(),
    options: {
      ...block.options,
      labels: block.options.labels ? [...block.options.labels] : undefined,
    },
  }));

  await nextTick();
  updatePreviewScale();
  ElMessage.success(`已载入模板「${template.name}」`);
}

async function handleDeleteTemplate() {
  if (selectedTemplateId.value == null) return;

  const template = store.plannerTemplates.find((item) => item.id === selectedTemplateId.value);

  try {
    await ElMessageBox.confirm(`删除模板「${template?.name ?? '未命名'}」后无法恢复，确认继续吗？`, '删除模板', {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'warning',
    });

    await store.deletePlannerTemplate(selectedTemplateId.value);
    selectedTemplateId.value = undefined;
    ElMessage.success('模板已删除');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;

    ElMessage.error(error instanceof Error ? error.message : '删除模板失败');
  }
}

async function resetPlanner() {
  try {
    await ElMessageBox.confirm('会恢复到 A6 默认配色和日计划组合，未保存的修改将丢失。', '重置内页', {
      confirmButtonText: '重置',
      cancelButtonText: '取消',
      type: 'warning',
    });
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    return;
  }

  sizeKey.value = 'a6';
  widthMm.value = 105;
  heightMm.value = 148;
  paddingMm.value = 4;
  accent.value = '#6c7cff';
  paper.value = '#f7f9ff';
  copies.value = 1;
  selectedTemplateId.value = undefined;
  modules.value = buildCombo(['dailyPlan', 'todo', 'meals']);
}

watch([widthMm, heightMm], () => {
  const preset = sizePresets.find((item) => item.key === sizeKey.value);

  if (preset && (preset.widthMm !== widthMm.value || preset.heightMm !== heightMm.value)) {
    sizeKey.value = 'custom';
  }
});

watch([widthMm, heightMm], syncPageStyle, { immediate: true });

watch([widthMm, heightMm], async () => {
  await nextTick();
  updatePreviewScale();
});

onMounted(async () => {
  modules.value = buildCombo(['dailyPlan', 'todo', 'meals']);
  syncPageStyle();
  await store.initialize();
  await nextTick();
  updatePreviewScale();

  window.addEventListener('resize', updatePreviewScale);

  if (stageRef.value) {
    resizeObserver = new ResizeObserver(updatePreviewScale);
    resizeObserver.observe(stageRef.value);
  }
});

onUnmounted(() => {
  window.removeEventListener('resize', updatePreviewScale);
  resizeObserver?.disconnect();
  document.getElementById(PAGE_STYLE_ID)?.remove();
});
</script>

<template>
  <div class="page-stack planner-page">
    <section class="page-hero page-hero--violet">
      <div>
        <p class="hero-kicker">内页生成器 · PRINTABLE PLANNER</p>
        <h1 class="hero-title">先定尺寸，再拼模块</h1>
        <p class="hero-description">
          从 M5、A7 到手帐三折页，按毫米真实排版；组合月历、时间轴、待办、四象限、账单与各种格纸，换好配色后直接打印或导出 1:1 PDF。
        </p>
        <div class="hero-badges">
          <span class="hero-badge">{{ widthMm }} × {{ heightMm }} mm</span>
          <span class="hero-badge">{{ modules.length }} 个模块</span>
          <span class="hero-badge">配色可调</span>
        </div>
      </div>
    </section>

    <div class="planner-layout">
      <div class="planner-controls">
        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">纸张与配色</span>
              <span class="muted">{{ currentPreset?.label ?? '自定义' }}</span>
            </div>
          </template>

          <div class="field-grid">
            <label class="field">
              <span>尺寸预设</span>
              <el-select :model-value="sizeKey" @update:model-value="applySize">
                <el-option v-for="preset in sizePresets" :key="preset.key" :label="preset.label" :value="preset.key" />
              </el-select>
            </label>
            <label class="field">
              <span>宽度 mm</span>
              <el-input-number v-model="widthMm" :min="20" :max="420" :step="1" />
            </label>
            <label class="field">
              <span>高度 mm</span>
              <el-input-number v-model="heightMm" :min="20" :max="420" :step="1" />
            </label>
            <label class="field">
              <span>留白 mm</span>
              <el-input-number v-model="paddingMm" :min="0" :max="20" :step="0.5" />
            </label>
          </div>

          <p v-if="currentPreset?.note" class="inline-tip">{{ currentPreset.note }}</p>

          <div class="inline-row">
            <el-button size="small" @click="swapOrientation">横竖切换</el-button>
            <span class="muted">实际 {{ widthMm }} × {{ heightMm }} mm</span>
          </div>

          <div class="inline-row">
            <span class="field-label">主题色</span>
            <el-color-picker v-model="accent" />
            <span class="field-label">纸色</span>
            <el-color-picker v-model="paper" />
            <el-button size="small" text @click="unifyAccent">模块统一主题色</el-button>
          </div>

          <div class="palette-row">
            <button
              v-for="palette in palettes"
              :key="palette.label"
              type="button"
              class="palette-chip"
              :style="{ background: palette.paper, borderColor: palette.accent }"
              @click="applyPalette(palette)"
            >
              <span :style="{ background: palette.accent }"></span>
              {{ palette.label }}
            </button>
          </div>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">模块库</span>
              <span class="muted">点一下加到页面</span>
            </div>
          </template>

          <div class="combo-row">
            <button
              v-for="combo in comboPresets"
              :key="combo.key"
              type="button"
              class="combo-chip"
              @click="applyCombo(combo)"
            >
              {{ combo.label }}
            </button>
          </div>

          <div class="module-grid">
            <button
              v-for="definition in moduleLibrary"
              :key="definition.type"
              type="button"
              class="module-chip"
              @click="addModule(definition)"
            >
              <strong>{{ definition.label }}</strong>
              <span>{{ definition.title }}</span>
            </button>
          </div>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">当前模块</span>
              <span class="muted">{{ modules.length }}/10</span>
            </div>
          </template>

          <p v-if="!modules.length" class="inline-tip">还没有模块，从上面的模块库点一个开始。</p>

          <div v-for="(block, index) in modules" :key="block.id" class="block-card">
            <div class="block-card-head">
              <span class="block-type">{{ moduleTypeLabel(block.type) }}</span>
              <el-input v-model="block.title" size="small" placeholder="标题（留空不显示）" class="block-title-input" />
              <el-input-number
                v-model="block.weight"
                :min="1"
                :max="8"
                size="small"
                controls-position="right"
                class="block-weight"
              />
              <el-color-picker v-model="block.accent" size="small" />
              <el-button size="small" text :disabled="index === 0" @click="moveModule(index, -1)">↑</el-button>
              <el-button size="small" text :disabled="index === modules.length - 1" @click="moveModule(index, 1)">↓</el-button>
              <el-button size="small" text type="danger" @click="removeModule(block.id)">删除</el-button>
            </div>

            <div v-if="block.type === 'calendar'" class="block-options">
              <label class="mini-field">
                周数
                <el-input-number v-model="block.options.rows" :min="4" :max="8" size="small" />
              </label>
            </div>

            <div v-else-if="block.type === 'habitTracker'" class="block-options">
              <label class="mini-field">
                习惯数
                <el-input-number v-model="block.options.rows" :min="1" :max="maxRows(block.type)" size="small" />
              </label>
              <label class="mini-field">
                天数
                <el-input-number v-model="block.options.days" :min="1" :max="31" size="small" />
              </label>
            </div>

            <div v-else-if="block.type === 'timeline' || block.type === 'dailyPlan'" class="block-options">
              <label class="mini-field">
                起始小时
                <el-input-number v-model="block.options.startHour" :min="0" :max="23" size="small" />
              </label>
              <label class="mini-field">
                结束小时
                <el-input-number v-model="block.options.endHour" :min="1" :max="24" size="small" />
              </label>
            </div>

            <div v-else-if="block.type === 'quadrant'" class="block-options block-options--wide">
              <label class="mini-field mini-field--wide">
                四个象限标题（用 / 分隔）
                <el-input
                  :model-value="labelsText(block)"
                  size="small"
                  placeholder="重要且紧急 / 重要不紧急 / 紧急不重要 / 不重要不紧急"
                  @update:model-value="setLabels(block, $event)"
                />
              </label>
            </div>

            <div v-else-if="block.type === 'grid' || block.type === 'dotGrid' || block.type === 'frenchGrid'" class="block-options">
              <label class="mini-field">
                格子 mm
                <el-input-number v-model="block.options.cellMm" :min="2" :max="20" :step="0.5" size="small" />
              </label>
            </div>

            <div v-else-if="block.type === 'lined'" class="block-options">
              <label class="mini-field">
                行距 mm
                <el-input-number v-model="block.options.lineMm" :min="4" :max="20" :step="0.5" size="small" />
              </label>
            </div>

            <div v-else-if="block.type === 'todo' || block.type === 'ledger' || block.type === 'finance' || block.type === 'weeklyPlan'" class="block-options">
              <label class="mini-field">
                行数
                <el-input-number v-model="block.options.rows" :min="1" :max="maxRows(block.type)" size="small" />
              </label>
            </div>

            <p v-else class="block-hint">该模块没有额外参数，用上方权重调整高度。</p>
          </div>
        </el-card>

        <el-card class="section-card" shadow="never">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">我的模板</span>
              <span class="muted">本地保存</span>
            </div>
          </template>

          <div class="inline-row">
            <el-select v-model="selectedTemplateId" placeholder="选择已保存模板" clearable class="template-select">
              <el-option v-for="option in templateOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </div>

          <div class="inline-row">
            <el-button :disabled="selectedTemplateId == null" @click="handleLoadTemplate">载入</el-button>
            <el-button :disabled="selectedTemplateId == null" @click="handleDeleteTemplate">删除</el-button>
            <el-button type="primary" plain @click="handleSaveTemplate">保存当前模板</el-button>
            <el-button text @click="resetPlanner">重置</el-button>
          </div>

          <p class="inline-tip">模板存在浏览器本地（IndexedDB），清理浏览器数据或换设备后会丢失。</p>
        </el-card>
      </div>

      <div class="planner-preview-panel">
        <el-card class="section-card preview-card" shadow="never">
          <template #header>
            <div class="preview-head">
              <span class="panel-title">实时预览</span>
              <span class="muted">{{ widthMm }}×{{ heightMm }}mm · {{ Math.round(previewScale * 100) }}%</span>
            </div>
          </template>

          <div ref="stageRef" class="preview-stage">
            <div class="preview-frame" :style="{ width: `${previewWidthPx}px`, height: `${previewHeightPx}px` }">
              <div class="preview-scaler" :style="{ transform: `scale(${previewScale})` }">
                <PlannerSheet
                  :width-mm="widthMm"
                  :height-mm="heightMm"
                  :padding-mm="paddingMm"
                  :accent="accent"
                  :paper="paper"
                  :modules="modules"
                />
              </div>
            </div>
          </div>

          <div class="print-actions">
            <span class="field-label">打印份数</span>
            <el-input-number v-model="copies" :min="1" :max="50" size="small" />
            <el-button type="primary" @click="handlePrint">打印 / 导出 PDF</el-button>
          </div>

          <p class="inline-tip">
            打印对话框里把「边距」设为无、关闭「页眉和页脚」，纸张选与内页一致的尺寸；选择「另存为 PDF」即可得到 1:1 文件。
          </p>
        </el-card>
      </div>
    </div>

    <Teleport to="body">
      <div id="planner-print-root">
        <div v-for="copy in copies" :key="copy" class="planner-print-page">
          <PlannerSheet
            :width-mm="widthMm"
            :height-mm="heightMm"
            :padding-mm="paddingMm"
            :accent="accent"
            :paper="paper"
            :modules="modules"
          />
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.planner-page {
  gap: 18px;
}

.planner-layout {
  display: grid;
  grid-template-columns: minmax(320px, 400px) minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.planner-controls {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.planner-preview-panel {
  position: sticky;
  top: 18px;
  min-width: 0;
}

.preview-card :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-head,
.preview-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.panel-title {
  color: #2f2b3a;
  font-weight: 600;
}

.muted {
  color: #8b8698;
  font-size: 12px;
}

.field-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.field > span,
.field-label {
  color: #6f6a80;
  font-size: 12px;
}

.inline-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}

.palette-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.palette-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px 3px 3px;
  border: 1px solid;
  border-radius: 999px;
  background: #fff;
  color: #555062;
  font-size: 12px;
  cursor: pointer;
}

.palette-chip span {
  width: 18px;
  height: 18px;
  border-radius: 50%;
}

.combo-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

.combo-chip {
  padding: 5px 10px;
  border: 1px solid #e4dff0;
  border-radius: 8px;
  background: #faf8ff;
  color: #5d5670;
  font-size: 12px;
  cursor: pointer;
}

.combo-chip:hover,
.module-chip:hover {
  border-color: #a99de0;
  color: #4d4480;
}

.module-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.module-chip {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 9px;
  border: 1px solid #e8e4f2;
  border-radius: 10px;
  background: #fff;
  color: #4f4a5e;
  text-align: left;
  cursor: pointer;
}

.module-chip strong {
  font-size: 13px;
  font-weight: 600;
}

.module-chip span {
  overflow: hidden;
  color: #9892a6;
  font-size: 11px;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.block-card {
  margin-bottom: 10px;
  padding: 10px;
  border: 1px solid #ece8f4;
  border-radius: 12px;
  background: #fff;
}

.block-card-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.block-type {
  padding: 2px 6px;
  border-radius: 6px;
  background: #f2effa;
  color: #6c63a8;
  font-size: 12px;
  white-space: nowrap;
}

.block-title-input {
  flex: 1;
  min-width: 120px;
}

.block-weight {
  width: 92px;
}

.block-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  margin-top: 8px;
}

.block-options--wide {
  display: block;
}

.mini-field {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6f6a80;
  font-size: 12px;
}

.mini-field :deep(.el-input-number) {
  width: 104px;
}

.mini-field--wide {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 4px;
}

.block-hint {
  margin: 8px 0 0;
  color: #9a94a8;
  font-size: 12px;
}

.template-select {
  width: 100%;
}

.preview-stage {
  display: flex;
  justify-content: center;
  max-height: 70vh;
  padding: 18px 10px;
  overflow: auto;
  border-radius: 14px;
  background:
    radial-gradient(circle at 1px 1px, #d9d4e8 1px, transparent 0) 0 0 / 14px 14px,
    #f5f3fa;
}

.preview-frame {
  position: relative;
  flex: none;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 16px 36px rgba(54, 45, 88, 0.18);
}

.preview-scaler {
  position: absolute;
  top: 0;
  left: 0;
  transform-origin: top left;
}

.print-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

@media (max-width: 1100px) {
  .planner-layout {
    grid-template-columns: 1fr;
  }

  .planner-preview-panel {
    position: static;
  }
}

@media (max-width: 640px) {
  .field-grid {
    grid-template-columns: 1fr;
  }

  .module-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .preview-stage {
    max-height: none;
  }
}
</style>

<style>
#planner-print-root {
  display: none;
}

@media print {
  html,
  body {
    margin: 0 !important;
    padding: 0 !important;
    background: #fff !important;
  }

  #app {
    display: none !important;
  }

  #planner-print-root {
    display: block !important;
  }

  #planner-print-root .planner-sheet {
    box-shadow: none;
  }

  .el-message,
  .el-message-box__wrapper,
  .el-notification,
  .el-overlay,
  .el-popper {
    display: none !important;
  }

  #planner-print-root .planner-print-page {
    break-after: page;
    page-break-after: always;
  }

  #planner-print-root .planner-print-page:last-child {
    break-after: auto;
    page-break-after: auto;
  }
}
</style>
