<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';

type CategoryKey = '植物' | '食物' | '动物' | '日常用品' | '水果' | '盐系小元素';

interface InspirationItem {
  category: CategoryKey;
  element: string;
}

interface CategoryConfig {
  label: CategoryKey;
  prefixes: string[];
  nouns: string[];
}

function buildPool(config: CategoryConfig) {
  const pool = new Set<string>();

  config.prefixes.forEach((prefix) => {
    config.nouns.forEach((noun) => {
      pool.add(`${prefix}${noun}`);
    });
  });

  return Array.from(pool);
}

function pickOne<T>(items: T[]): T {
  return items[Math.floor(Math.random() * items.length)];
}

function shuffle<T>(items: T[]) {
  const copied = [...items];
  for (let i = copied.length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1));
    [copied[i], copied[j]] = [copied[j], copied[i]];
  }
  return copied;
}

const categoryConfigs: CategoryConfig[] = [
  {
    label: '植物',
    prefixes: ['小', '迷你', '圆润', '细叶', '垂坠', '水培', '陶盆', '窗边', '晨露', '风干', '线稿', '简笔'],
    nouns: ['多肉', '仙人掌', '薄荷', '薰衣草', '雏菊', '向日葵', '银杏叶', '尤加利', '龟背竹', '小麦穗', '绣球花', '常春藤'],
  },
  {
    label: '食物',
    prefixes: ['热乎乎的', '刚出炉的', '奶香', '焦糖', '脆皮', '手作', '便当里的', '早餐风', '轻食风', '小份', '日常', '复古'],
    nouns: ['吐司', '可颂', '饭团', '蛋包饭', '布丁', '铜锣烧', '曲奇', '三明治', '拉面', '煎蛋', '松饼', '小蛋糕'],
  },
  {
    label: '动物',
    prefixes: ['圆脸', '眯眼', '打盹的', '抱枕般的', '软乎乎的', '趴着的', '探头的', '歪头', '背影', '小短腿', '手绘感', '涂鸦风'],
    nouns: ['橘猫', '柯基', '兔子', '仓鼠', '刺猬', '柴犬', '小熊', '小鹿', '企鹅', '海豹', '猫头鹰', '羊驼'],
  },
  {
    label: '日常用品',
    prefixes: ['桌面上的', '透明', '磨砂', '复古', '便携', '折叠', '迷你', '手账用', '奶油色', '木质', '软边', '圆角'],
    nouns: ['马克杯', '台灯', '胶带台', '剪刀', '订书机', '便签夹', '闹钟', '相机', '雨伞', '帆布包', '钥匙扣', '笔筒'],
  },
  {
    label: '水果',
    prefixes: ['新鲜', '切开的', '带叶子', '透明感', '冰镇', '夏日', '软萌', '手绘', '颗粒感', '甜香', '小清新', '果园里'],
    nouns: ['草莓', '蓝莓', '葡萄', '樱桃', '苹果', '梨', '橙子', '柠檬', '桃子', '无花果', '西瓜', '猕猴桃'],
  },
  {
    label: '盐系小元素',
    prefixes: ['雾灰', '奶白', '亚麻', '磨砂', '手写', '胶片感', '留白', '浅卡其', '旧纸感', '静物感', '低饱和', '柔光'],
    nouns: ['便签角', '纸胶带撕边', '回形针', '票据', '邮戳', '小星点', '网格纸', '折角纸片', '标签条', '线框', '印章', '半透明贴纸'],
  },
];

const categoryPools = categoryConfigs.map((config) => ({
  label: config.label,
  pool: buildPool(config),
}));

const selectedCategories = ref<CategoryKey[]>(categoryConfigs.map((item) => item.label));
const inspirationItems = ref<InspirationItem[]>([]);
const styleKeywords = ref<string[]>([]);

const stylePool = [
  '水彩',
  '马克笔',
  '扁平插画',
  '贴纸风',
  '蜡笔涂抹',
  '色铅笔',
  '轻复古',
  '盐系低饱和',
  '日杂封面感',
  '手账拼贴',
  '简笔线稿',
  '治愈手绘',
  '胶片颗粒',
  '纸感纹理',
  '留白构图',
  'Q版圆润',
  '暖调光影',
  '冷灰调',
  '奶油配色',
  '极简图标化',
];

const categoryStats = computed(() =>
  categoryPools.map((item) => ({
    label: item.label,
    count: item.pool.length,
  })),
);

const inspirationText = computed(() => {
  if (inspirationItems.value.length === 0) {
    return '点击“随机生成灵感”获得 3-4 个元素组合。';
  }

  return inspirationItems.value.map((item) => `${item.category}：${item.element}`).join(' / ');
});

const styleText = computed(() =>
  styleKeywords.value.length ? styleKeywords.value.join(' / ') : '暂未生成风格关键词',
);

function generateInspiration() {
  if (selectedCategories.value.length === 0) {
    ElMessage.warning('请至少选择一个元素类型');
    return;
  }

  const count = Math.random() > 0.5 ? 4 : 3;
  const categoryOrder = shuffle(selectedCategories.value);

  const finalCategories: CategoryKey[] = [];
  for (let i = 0; i < count; i += 1) {
    finalCategories.push(categoryOrder[i % categoryOrder.length]);
  }

  inspirationItems.value = finalCategories.map((category) => {
    const found = categoryPools.find((item) => item.label === category)!;
    return {
      category,
      element: pickOne(found.pool),
    };
  });

  const styleCount = Math.random() > 0.5 ? 3 : 2;
  styleKeywords.value = shuffle(stylePool).slice(0, styleCount);
}

function regenerate() {
  generateInspiration();
}

generateInspiration();
</script>

<template>
  <div class="page-stack">
    <section class="page-hero page-hero--violet">
      <div>
        <div class="hero-kicker">灵感菜单 · 随机绘画组合</div>
        <h2 class="hero-title">每次随机给你 3-4 个元素，快速开画</h2>
        <p class="hero-description">覆盖植物、食物、动物、日常用品、水果、盐系小元素，每类都提供 100+ 可画元素。</p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">类型 {{ categoryPools.length }} 类</span>
        <span class="hero-badge">单类元素 144 个起</span>
      </div>
    </section>

    <el-card class="section-card">
      <div class="toolbar">
        <div>
          <h3>类型筛选</h3>
          <span class="inline-tip">可多选，生成时会从已勾选类型中抽取</span>
        </div>
      </div>

      <el-checkbox-group v-model="selectedCategories" class="type-list">
        <el-checkbox v-for="item in categoryStats" :key="item.label" :label="item.label">
          {{ item.label }}（{{ item.count }}）
        </el-checkbox>
      </el-checkbox-group>

      <div class="actions-row">
        <el-button type="primary" @click="generateInspiration">随机生成灵感</el-button>
        <el-button @click="regenerate">再来一组</el-button>
      </div>
    </el-card>

    <el-card class="section-card">
      <template #header>
        <div class="toolbar">
          <h3>本次组合</h3>
          <span class="inline-tip">可以直接作为插画 prompt 或手帐元素练习题</span>
        </div>
      </template>

      <div class="result-pills">
        <el-tag v-for="(item, idx) in inspirationItems" :key="`${item.category}-${idx}`" size="large" effect="plain" class="result-pill">
          {{ item.category }} · {{ item.element }}
        </el-tag>
      </div>

      <div class="style-block">
        <div class="inline-tip">随机风格关键词</div>
        <div class="result-pills">
          <el-tag v-for="keyword in styleKeywords" :key="keyword" type="success" effect="light" class="result-pill">
            {{ keyword }}
          </el-tag>
        </div>
      </div>

      <div class="dialog-tip">{{ inspirationText }}</div>
      <div class="dialog-tip">风格：{{ styleText }}</div>
    </el-card>
  </div>
</template>

<style scoped>
.type-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 10px;
  margin-top: 10px;
}

.actions-row {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.result-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.result-pill {
  margin: 0;
}

.style-block {
  margin-top: 14px;
  display: grid;
  gap: 8px;
}
</style>
