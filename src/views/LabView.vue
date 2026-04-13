<script setup lang="ts">
const inspirationCards = [
  {
    title: '雾灰茶会系列',
    tag: '春夏灵感',
    status: '待打样',
    summary: '偏奶油灰与浅茶棕的拼贴调性，适合做和纸胶带、便签与套装页。',
    points: ['关键词：留白、复古、低饱和', '适配品类：胶带 / 贴纸 / 小卡'],
  },
  {
    title: '月光整理册',
    tag: '新本册方向',
    status: '结构讨论中',
    summary: '用极简封面和轻纹理纸张做一本偏记录感的周计划本。',
    points: ['关键词：冷白、雾蓝、银灰', '适配用户：日常记录 / 学习计划'],
  },
  {
    title: '窗边花影贴纸包',
    tag: '插画主题',
    status: '画面修稿中',
    summary: '保留韩系静物插画感，适合做一整组少女但不甜腻的花影贴纸。',
    points: ['关键词：静物、窗帘、柔焦', '适配品类：PET贴纸 / 包装卡'],
  },
];

const pipelineItems = [
  {
    name: '月光便签册',
    stage: '结构设计',
    progress: 35,
    eta: '4月16日',
    note: '内页排版已定，封面材质待确认。',
  },
  {
    name: '旧梦胶带套组',
    stage: '打样中',
    progress: 60,
    eta: '4月19日',
    note: '第一版颜色略深，准备调浅一点。',
  },
  {
    name: '窗边花影贴纸',
    stage: '插画修稿',
    progress: 75,
    eta: '4月14日',
    note: '正在统一边框和文字排版。',
  },
  {
    name: '研究所限定套装',
    stage: '排产准备',
    progress: 90,
    eta: '4月22日',
    note: '清单已确认，等待供应商排期。',
  },
];

const notes = [
  '优先做低门槛高复购的贴纸与胶带，不要一次铺太多复杂套装。',
  '韩系风格重点不是“甜”，而是低饱和、安静、轻空气感。',
  '上新节奏建议按：灵感池 → 小样 → 试售 → 正式上架。',
];
</script>

<template>
  <div class="page-stack">
    <section class="page-hero page-hero--violet">
      <div>
        <div class="hero-kicker">研究所 · 灵感与研发区</div>
        <h2 class="hero-title">把灵感、打样和正在生产的产品集中整理</h2>
        <p class="hero-description">
          这里更像店铺的创作工作台，可以先记录想法，再跟进每个产品目前做到哪一步。
        </p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">灵感 {{ inspirationCards.length }} 条</span>
        <span class="hero-badge">研发中 {{ pipelineItems.length }} 款</span>
      </div>
    </section>

    <div class="summary-grid">
      <div class="summary-tile">
        <strong>灵感池</strong>
        <span>{{ inspirationCards.length }}</span>
      </div>
      <div class="summary-tile">
        <strong>打样 / 修稿中</strong>
        <span>{{ pipelineItems.filter((item) => item.progress < 80).length }}</span>
      </div>
      <div class="summary-tile">
        <strong>接近排产</strong>
        <span>{{ pipelineItems.filter((item) => item.progress >= 80).length }}</span>
      </div>
      <div class="summary-tile">
        <strong>本周重点</strong>
        <span>先推胶带</span>
      </div>
    </div>

    <div class="grid-two">
      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>灵感池</h3>
            <span class="inline-tip">记录正在酝酿的系列与创作方向</span>
          </div>
        </template>

        <div class="lab-grid">
          <article v-for="item in inspirationCards" :key="item.title" class="lab-card">
            <div class="lab-card__top">
              <strong>{{ item.title }}</strong>
              <el-tag round effect="plain">{{ item.tag }}</el-tag>
            </div>
            <p class="lab-card__summary">{{ item.summary }}</p>
            <ul class="lab-points">
              <li v-for="point in item.points" :key="point">{{ point }}</li>
            </ul>
            <div class="lab-card__footer">
              <span>{{ item.status }}</span>
            </div>
          </article>
        </div>
      </el-card>

      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>研发进度</h3>
            <span class="inline-tip">正在设计、打样和排产中的产品</span>
          </div>
        </template>

        <div class="lab-progress-list">
          <div v-for="item in pipelineItems" :key="item.name" class="lab-progress-item">
            <div class="lab-progress-item__head">
              <div>
                <strong>{{ item.name }}</strong>
                <div class="inline-tip">{{ item.stage }} · 预计 {{ item.eta }}</div>
              </div>
              <span class="insight-value">{{ item.progress }}%</span>
            </div>
            <el-progress :percentage="item.progress" :stroke-width="8" :show-text="false" />
            <div class="dialog-tip">{{ item.note }}</div>
          </div>
        </div>
      </el-card>
    </div>

    <div class="grid-two">
      <el-card class="section-card soft-table">
        <template #header>
          <div class="toolbar">
            <h3>研究所排期表</h3>
            <span class="inline-tip">按阶段追踪设计与生产节奏</span>
          </div>
        </template>

        <el-table :data="pipelineItems">
          <el-table-column prop="name" label="产品" min-width="160" />
          <el-table-column prop="stage" label="阶段" width="120" />
          <el-table-column prop="eta" label="预计时间" width="110" />
          <el-table-column label="完成度" width="110">
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
</template>