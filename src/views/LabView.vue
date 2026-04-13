<script setup lang="ts">
import dayjs from 'dayjs';
import { ElMessage } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';

import { useShopStore } from '@/stores/useShopStore';
import type { InspirationInput, LabInspiration, LabProject, LabProjectInput, LabStage, ProofRecordInput } from '@/types';

const store = useShopStore();

const inspirationDialogVisible = ref(false);
const projectDialogVisible = ref(false);
const proofDialogVisible = ref(false);
const editingInspirationId = ref<number | null>(null);
const editingProjectId = ref<number | null>(null);
const proofTargetProjectId = ref<number | null>(null);

const stageOptions: LabStage[] = ['灵感整理', '设计中', '打样中', '排产中', '待上新', '已上新'];

const emptyInspiration = (): InspirationInput => ({
  title: '',
  tag: '',
  status: '灵感中',
  summary: '',
  keywords: '',
  image: '',
});

const emptyProject = (): LabProjectInput => ({
  name: '',
  category: '胶带',
  stage: '设计中',
  progress: 35,
  launchDate: dayjs().add(7, 'day').format('YYYY-MM-DD'),
  note: '',
  image: '',
});

const emptyProof = (): ProofRecordInput => ({
  title: '',
  date: dayjs().format('YYYY-MM-DD'),
  note: '',
  image: '',
});

const inspirationForm = reactive<InspirationInput>(emptyInspiration());
const projectForm = reactive<LabProjectInput>(emptyProject());
const proofForm = reactive<ProofRecordInput>(emptyProof());

const summaryCards = computed(() => [
  { label: '灵感池', value: `${store.labInspirations.length}` },
  { label: '设计 / 打样中', value: `${store.labProjects.filter((item) => ['设计中', '打样中'].includes(item.stage)).length}` },
  { label: '待上新 / 已上新', value: `${store.labProjects.filter((item) => ['待上新', '已上新'].includes(item.stage)).length}` },
  { label: '本月排期', value: `${store.upcomingLaunches.length}` },
]);

const notes = [
  '优先做低门槛高复购的贴纸与胶带，不要一次铺太多复杂套装。',
  '韩系风格重点不是“甜”，而是低饱和、安静、轻空气感。',
  '上新节奏建议按：灵感池 → 小样 → 试售 → 正式上架。',
];

function resetInspirationForm() {
  editingInspirationId.value = null;
  Object.assign(inspirationForm, emptyInspiration());
}

function resetProjectForm() {
  editingProjectId.value = null;
  Object.assign(projectForm, emptyProject());
}

function resetProofForm() {
  proofTargetProjectId.value = null;
  Object.assign(proofForm, emptyProof());
}

function openInspirationDialog(item?: LabInspiration) {
  if (item?.id) {
    editingInspirationId.value = item.id;
    Object.assign(inspirationForm, {
      title: item.title,
      tag: item.tag,
      status: item.status,
      summary: item.summary,
      keywords: item.keywords,
      image: item.image ?? '',
    });
  } else {
    resetInspirationForm();
  }

  inspirationDialogVisible.value = true;
}

function openProjectDialog(item?: LabProject) {
  if (item?.id) {
    editingProjectId.value = item.id;
    Object.assign(projectForm, {
      name: item.name,
      category: item.category,
      stage: item.stage,
      progress: item.progress,
      launchDate: item.launchDate,
      note: item.note,
      image: item.image ?? '',
    });
  } else {
    resetProjectForm();
  }

  projectDialogVisible.value = true;
}

function openProofDialog(projectId: number) {
  proofTargetProjectId.value = projectId;
  Object.assign(proofForm, emptyProof());
  proofDialogVisible.value = true;
}

async function submitInspiration() {
  try {
    if (editingInspirationId.value) {
      await store.updateInspiration(editingInspirationId.value, { ...inspirationForm });
      ElMessage.success('灵感条目已更新');
    } else {
      await store.addInspiration({ ...inspirationForm });
      ElMessage.success('已新增灵感条目');
    }

    inspirationDialogVisible.value = false;
    resetInspirationForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存灵感条目失败');
  }
}

async function submitProject() {
  try {
    if (editingProjectId.value) {
      await store.updateLabProject(editingProjectId.value, { ...projectForm });
      ElMessage.success('研发产品已更新');
    } else {
      await store.addLabProject({ ...projectForm });
      ElMessage.success('已新增研发产品');
    }

    projectDialogVisible.value = false;
    resetProjectForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存研发产品失败');
  }
}

async function submitProof() {
  try {
    if (!proofTargetProjectId.value) {
      throw new Error('请先选择研发产品');
    }

    await store.addProofRecord(proofTargetProjectId.value, { ...proofForm });
    ElMessage.success('已添加打样记录');
    proofDialogVisible.value = false;
    resetProofForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存打样记录失败');
  }
}

async function handleStageChange(projectId: number, stage: string) {
  try {
    await store.updateLabProjectStage(projectId, stage as LabStage);
    ElMessage.success('研发状态已切换');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '状态切换失败');
  }
}

function getProjectsForDate(day: string) {
  return store.labProjects.filter((item) => item.launchDate === day);
}

function readFileAsDataUrl(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ''));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}

async function handleImageUpload(event: Event, target: 'inspiration' | 'project' | 'proof') {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];

  if (!file) return;

  const dataUrl = await readFileAsDataUrl(file);

  if (target === 'inspiration') {
    inspirationForm.image = dataUrl;
  } else if (target === 'project') {
    projectForm.image = dataUrl;
  } else {
    proofForm.image = dataUrl;
  }

  input.value = '';
}

onMounted(() => {
  store.initialize();
});
</script>

<template>
  <div class="page-stack">
    <section class="page-hero page-hero--violet">
      <div>
        <div class="hero-kicker">研究所 · 灵感与研发区</div>
        <h2 class="hero-title">把灵感、打样和正在生产的产品集中整理</h2>
        <p class="hero-description">
          这里更像店铺的创作工作台，可以直接录入灵感、推进设计状态、记录打样，并安排上新日期。
        </p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">灵感 {{ store.labInspirations.length }} 条</span>
        <span class="hero-badge">研发中 {{ store.labProjects.length }} 款</span>
      </div>
    </section>

    <div class="summary-grid">
      <div v-for="item in summaryCards" :key="item.label" class="summary-tile">
        <strong>{{ item.label }}</strong>
        <span>{{ item.value }}</span>
      </div>
    </div>

    <div class="grid-two">
      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <div>
              <h3>灵感池</h3>
              <span class="inline-tip">可新增、编辑灵感条目，并附上参考图</span>
            </div>
            <el-button type="primary" @click="openInspirationDialog()">新增灵感</el-button>
          </div>
        </template>

        <div class="lab-grid">
          <article v-for="item in store.labInspirations" :key="item.id" class="lab-card">
            <img v-if="item.image" :src="item.image" class="lab-cover" alt="灵感图" />
            <div v-else class="lab-cover lab-cover--placeholder">暂无图片</div>

            <div class="lab-card__top">
              <div>
                <strong>{{ item.title }}</strong>
                <div class="inline-tip">{{ item.tag }}</div>
              </div>
              <el-tag round effect="plain">{{ item.status }}</el-tag>
            </div>

            <p class="lab-card__summary">{{ item.summary }}</p>
            <div class="lab-meta">{{ item.keywords }}</div>

            <div class="lab-card__footer">
              <span class="inline-tip">更新于 {{ dayjs(item.updatedAt).format('MM-DD') }}</span>
              <el-button link type="primary" @click="openInspirationDialog(item)">编辑</el-button>
            </div>
          </article>
        </div>
      </el-card>

      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <div>
              <h3>研发进度</h3>
              <span class="inline-tip">可切换阶段，记录打样和排产状态</span>
            </div>
            <el-button type="primary" plain @click="openProjectDialog()">新增研发产品</el-button>
          </div>
        </template>

        <div class="lab-progress-list">
          <div v-for="item in store.labProjects" :key="item.id" class="lab-progress-item">
            <div class="lab-progress-item__head">
              <div>
                <strong>{{ item.name }}</strong>
                <div class="inline-tip">{{ item.category }} · 预计 {{ item.launchDate }}</div>
              </div>
              <span class="insight-value">{{ item.progress }}%</span>
            </div>

            <img v-if="item.image" :src="item.image" class="lab-cover lab-cover--small" alt="研发图" />

            <div class="lab-progress-item__body">
              <el-select class="stage-select" :model-value="item.stage" @change="handleStageChange(item.id!, String($event))">
                <el-option v-for="stage in stageOptions" :key="stage" :label="stage" :value="stage" />
              </el-select>
              <div class="inline-tip">{{ item.note }}</div>
            </div>

            <el-progress :percentage="item.progress" :stroke-width="8" :show-text="false" />

            <div class="lab-card__actions">
              <el-button link type="primary" @click="openProjectDialog(item)">编辑</el-button>
              <el-button link @click="openProofDialog(item.id!)">添加打样记录</el-button>
            </div>

            <div class="lab-record-list">
              <div v-if="item.sampleRecords.length === 0" class="empty-note">还没有打样记录。</div>
              <div v-for="record in item.sampleRecords.slice().reverse()" :key="record.id" class="lab-record">
                <img v-if="record.image" :src="record.image" class="lab-record__image" alt="打样图" />
                <div>
                  <strong>{{ record.title }}</strong>
                  <div class="inline-tip">{{ record.date }}</div>
                  <div class="dialog-tip">{{ record.note }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <div class="grid-two">
      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>上新排期日历</h3>
            <span class="inline-tip">按日期查看准备上新的产品</span>
          </div>
        </template>

        <el-calendar class="lab-calendar">
          <template #date-cell="{ data }">
            <div class="lab-calendar-cell">
              <div class="calendar-date">{{ Number(data.day.slice(-2)) }}</div>
              <div class="calendar-list">
                <span
                  v-for="project in getProjectsForDate(data.day).slice(0, 2)"
                  :key="`${data.day}-${project.id}`"
                  class="calendar-badge"
                >
                  {{ project.name }}
                </span>
              </div>
            </div>
          </template>
        </el-calendar>
      </el-card>

      <div class="page-stack">
        <el-card class="section-card soft-table">
          <template #header>
            <div class="toolbar">
              <h3>近期上新安排</h3>
              <span class="inline-tip">优先关注最近一批</span>
            </div>
          </template>

          <el-table :data="store.upcomingLaunches">
            <el-table-column prop="name" label="产品" min-width="140" />
            <el-table-column prop="stage" label="阶段" width="110" />
            <el-table-column prop="launchDate" label="上新日期" width="120" />
            <el-table-column label="进度" width="90">
              <template #default="{ row }">
                <span class="insight-value">{{ row.progress }}%</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="section-card">
          <template #header>
            <div class="toolbar">
              <h3>研究备忘</h3>
              <span class="inline-tip">给自己留一些创作判断标准</span>
            </div>
          </template>

          <div class="insight-list">
            <div v-for="note in notes" :key="note" class="insight-item">
              <div class="inline-tip">{{ note }}</div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <el-dialog v-model="inspirationDialogVisible" :title="editingInspirationId ? '编辑灵感条目' : '新增灵感条目'" width="720px">
      <el-form label-width="92px">
        <div class="form-grid">
          <el-form-item label="标题"><el-input v-model="inspirationForm.title" /></el-form-item>
          <el-form-item label="分类标签"><el-input v-model="inspirationForm.tag" /></el-form-item>
          <el-form-item label="当前状态">
            <el-select v-model="inspirationForm.status">
              <el-option label="灵感中" value="灵感中" />
              <el-option label="待打样" value="待打样" />
              <el-option label="打样中" value="打样中" />
              <el-option label="待上新" value="待上新" />
            </el-select>
          </el-form-item>
          <el-form-item label="关键词"><el-input v-model="inspirationForm.keywords" /></el-form-item>
        </div>
        <el-form-item label="灵感说明"><el-input v-model="inspirationForm.summary" type="textarea" :rows="3" /></el-form-item>
        <div class="file-upload">
          <input type="file" accept="image/*" @change="(event) => handleImageUpload(event, 'inspiration')" />
          <img v-if="inspirationForm.image" :src="inspirationForm.image" class="file-upload__preview" alt="灵感预览" />
        </div>
      </el-form>
      <template #footer>
        <el-button @click="inspirationDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitInspiration">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="projectDialogVisible" :title="editingProjectId ? '编辑研发产品' : '新增研发产品'" width="760px">
      <el-form label-width="92px">
        <div class="form-grid">
          <el-form-item label="产品名"><el-input v-model="projectForm.name" /></el-form-item>
          <el-form-item label="品类"><el-input v-model="projectForm.category" /></el-form-item>
          <el-form-item label="阶段">
            <el-select v-model="projectForm.stage">
              <el-option v-for="stage in stageOptions" :key="stage" :label="stage" :value="stage" />
            </el-select>
          </el-form-item>
          <el-form-item label="进度">
            <el-input-number v-model="projectForm.progress" :min="0" :max="100" />
          </el-form-item>
          <el-form-item label="上新日期"><el-date-picker v-model="projectForm.launchDate" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        </div>
        <el-form-item label="研发备注"><el-input v-model="projectForm.note" type="textarea" :rows="3" /></el-form-item>
        <div class="file-upload">
          <input type="file" accept="image/*" @change="(event) => handleImageUpload(event, 'project')" />
          <img v-if="projectForm.image" :src="projectForm.image" class="file-upload__preview" alt="研发预览" />
        </div>
      </el-form>
      <template #footer>
        <el-button @click="projectDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitProject">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="proofDialogVisible" title="新增打样记录" width="720px">
      <el-form label-width="92px">
        <div class="form-grid">
          <el-form-item label="记录标题"><el-input v-model="proofForm.title" /></el-form-item>
          <el-form-item label="日期"><el-date-picker v-model="proofForm.date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        </div>
        <el-form-item label="记录内容"><el-input v-model="proofForm.note" type="textarea" :rows="3" /></el-form-item>
        <div class="file-upload">
          <input type="file" accept="image/*" @change="(event) => handleImageUpload(event, 'proof')" />
          <img v-if="proofForm.image" :src="proofForm.image" class="file-upload__preview" alt="打样预览" />
        </div>
      </el-form>
      <template #footer>
        <el-button @click="proofDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitProof">保存记录</el-button>
      </template>
    </el-dialog>
  </div>
</template>