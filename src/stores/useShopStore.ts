import dayjs from 'dayjs';
import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

import { db, defaultPricingRule, seedDemoData, seedLabData } from '@/db/appDb';
import type {
  InspirationInput,
  LabInspiration,
  LabProject,
  LabProjectInput,
  LabStage,
  NewOrderInput,
  Order,
  PricingRule,
  Product,
  ProductInput,
  ProofRecordInput,
} from '@/types';
import { calculateMinPrice, calculateSuggestedPrice } from '@/utils/pricing';

const stageProgressMap: Record<LabStage, number> = {
  灵感整理: 10,
  设计中: 35,
  打样中: 60,
  排产中: 82,
  待上新: 95,
  已上新: 100,
};

export const useShopStore = defineStore('shop', () => {
  const loading = ref(false);
  const initialized = ref(false);
  const products = ref<Product[]>([]);
  const orders = ref<Order[]>([]);
  const pricingRule = ref<PricingRule>({ ...defaultPricingRule });
  const labInspirations = ref<LabInspiration[]>([]);
  const labProjects = ref<LabProject[]>([]);

  async function loadAll() {
    products.value = await db.products.orderBy('createdAt').reverse().toArray();
    orders.value = await db.orders.orderBy('createdAt').reverse().toArray();
    pricingRule.value = (await db.pricingRules.get(1)) ?? { ...defaultPricingRule };
    labInspirations.value = await db.labInspirations.orderBy('updatedAt').reverse().toArray();
    labProjects.value = await db.labProjects.orderBy('updatedAt').reverse().toArray();
  }

  async function initialize() {
    if (initialized.value) return;

    loading.value = true;
    try {
      const productCount = await db.products.count();
      const rule = await db.pricingRules.get(1);
      const inspirationCount = await db.labInspirations.count();

      if (productCount === 0) {
        await seedDemoData();
      } else {
        if (!rule) {
          await db.pricingRules.put({ ...defaultPricingRule });
        }
        if (inspirationCount === 0) {
          await seedLabData();
        }
      }

      await loadAll();
      initialized.value = true;
    } finally {
      loading.value = false;
    }
  }

  async function addProduct(input: ProductInput) {
    const suggested = calculateSuggestedPrice(input.purchaseCost, input.packagingCost, pricingRule.value);

    await db.products.add({
      ...input,
      salePrice: input.salePrice > 0 ? input.salePrice : suggested,
      createdAt: new Date().toISOString(),
    });

    await loadAll();
  }

  async function addOrder(input: NewOrderInput) {
    const product = await db.products.get(input.productId);

    if (!product?.id) {
      throw new Error('未找到商品');
    }

    if (product.stock < input.quantity) {
      throw new Error('库存不足，请先补货');
    }

    const originalAmount = Number((product.salePrice * input.quantity).toFixed(2));
    const totalAmount = Number(Math.max(0, originalAmount - input.discountAmount).toFixed(2));
    const platformFee = Number((totalAmount * input.platformFeeRate).toFixed(2));
    const goodsCost = Number(((product.purchaseCost + product.packagingCost) * input.quantity).toFixed(2));
    const netProfit = Number((totalAmount - input.shippingCost - platformFee - goodsCost).toFixed(2));

    await db.transaction('rw', db.orders, db.products, async () => {
      await db.orders.add({
        orderNo: `ORD-${dayjs().format('YYYYMMDD-HHmmss')}`,
        channel: input.channel,
        customerName: input.customerName,
        items: [
          {
            productId: product.id!,
            productName: product.name,
            sku: product.sku,
            quantity: input.quantity,
            salePrice: product.salePrice,
            totalAmount,
            totalCost: goodsCost,
          },
        ],
        totalAmount,
        discountAmount: input.discountAmount,
        shippingCost: input.shippingCost,
        platformFee,
        goodsCost,
        netProfit,
        createdAt: new Date().toISOString(),
        status: 'paid',
      });

      await db.products.update(product.id!, {
        stock: product.stock - input.quantity,
      });
    });

    await loadAll();
  }

  async function savePricingRule(rule: PricingRule) {
    const payload = { ...rule, id: 1 };
    await db.pricingRules.put(payload);
    pricingRule.value = payload;
  }

  async function addInspiration(input: InspirationInput) {
    const now = new Date().toISOString();
    await db.labInspirations.add({
      ...input,
      createdAt: now,
      updatedAt: now,
    });
    await loadAll();
  }

  async function updateInspiration(id: number, input: InspirationInput) {
    await db.labInspirations.update(id, {
      ...input,
      updatedAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function deleteInspiration(id: number) {
    await db.labInspirations.delete(id);
    await loadAll();
  }

  async function addLabProject(input: LabProjectInput) {
    await db.labProjects.add({
      ...input,
      progress: Math.max(input.progress, stageProgressMap[input.stage]),
      sampleRecords: [],
      updatedAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function updateLabProject(id: number, input: LabProjectInput) {
    await db.labProjects.update(id, {
      ...input,
      progress: Math.max(input.progress, stageProgressMap[input.stage]),
      updatedAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function updateLabProjectStage(id: number, stage: LabStage) {
    const project = await db.labProjects.get(id);
    if (!project) {
      throw new Error('未找到研发产品');
    }

    await db.labProjects.update(id, {
      stage,
      progress: Math.max(project.progress, stageProgressMap[stage]),
      updatedAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function addProofRecord(projectId: number, input: ProofRecordInput) {
    const project = await db.labProjects.get(projectId);
    if (!project) {
      throw new Error('未找到研发产品');
    }

    await db.labProjects.update(projectId, {
      sampleRecords: [
        ...project.sampleRecords,
        {
          id: `proof-${Date.now()}`,
          ...input,
        },
      ],
      updatedAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function deleteLabProject(id: number) {
    await db.labProjects.delete(id);
    await loadAll();
  }

  async function convertLabProjectToProduct(id: number) {
    const project = await db.labProjects.get(id);
    if (!project) {
      throw new Error('未找到研发产品');
    }

    const existing = await db.products.where('sku').equals(`LAB-${id}`).first();
    if (existing) {
      throw new Error('该研发产品已转为正式商品');
    }

    await db.transaction('rw', db.products, db.labProjects, async () => {
      await db.products.add({
        name: project.name,
        sku: `LAB-${id}`,
        category: project.category,
        supplier: '待确认供应商',
        purchaseCost: 0,
        packagingCost: 0,
        salePrice: 0,
        stock: 0,
        safeStock: 5,
        note: `由研究所转入：${project.note}`,
        createdAt: new Date().toISOString(),
      });

      await db.labProjects.update(id, {
        stage: '已上新',
        progress: 100,
        updatedAt: new Date().toISOString(),
      });
    });

    await loadAll();
  }

  function getSuggestedPrice(product: Product) {
    return calculateSuggestedPrice(product.purchaseCost, product.packagingCost, pricingRule.value);
  }

  function getMinPrice(product: Product) {
    return calculateMinPrice(
      product.purchaseCost,
      product.packagingCost,
      pricingRule.value.extraCost,
    );
  }

  const stats = computed(() => {
    const sales = orders.value.reduce((sum, item) => sum + item.totalAmount, 0);
    const profit = orders.value.reduce((sum, item) => sum + item.netProfit, 0);
    const monthSales = orders.value
      .filter((item) => dayjs(item.createdAt).isSame(dayjs(), 'month'))
      .reduce((sum, item) => sum + item.totalAmount, 0);

    return {
      totalProducts: products.value.length,
      totalOrders: orders.value.length,
      totalSales: Number(sales.toFixed(2)),
      totalProfit: Number(profit.toFixed(2)),
      monthSales: Number(monthSales.toFixed(2)),
      lowStockCount: products.value.filter((item) => item.stock <= item.safeStock).length,
    };
  });

  const lowStockProducts = computed(() => products.value.filter((item) => item.stock <= item.safeStock));

  const recentOrders = computed(() => orders.value.slice(0, 5));

  const monthlyTrend = computed(() => {
    const labels = Array.from({ length: 6 }, (_, index) => dayjs().subtract(5 - index, 'month').format('MM月'));
    const map = new Map(labels.map((label) => [label, { sales: 0, profit: 0 }]));

    orders.value.forEach((order) => {
      const key = dayjs(order.createdAt).format('MM月');
      const current = map.get(key);
      if (current) {
        current.sales += order.totalAmount;
        current.profit += order.netProfit;
      }
    });

    return labels.map((label) => ({
      month: label,
      sales: Number(map.get(label)!.sales.toFixed(2)),
      profit: Number(map.get(label)!.profit.toFixed(2)),
    }));
  });

  const upcomingLaunches = computed(() =>
    [...labProjects.value]
      .filter((item) => item.launchDate)
      .sort((a, b) => dayjs(a.launchDate).valueOf() - dayjs(b.launchDate).valueOf())
      .slice(0, 6),
  );

  return {
    loading,
    initialized,
    products,
    orders,
    pricingRule,
    labInspirations,
    labProjects,
    stats,
    lowStockProducts,
    recentOrders,
    monthlyTrend,
    upcomingLaunches,
    initialize,
    addProduct,
    addOrder,
    savePricingRule,
    addInspiration,
    updateInspiration,
    deleteInspiration,
    addLabProject,
    updateLabProject,
    updateLabProjectStage,
    addProofRecord,
    deleteLabProject,
    convertLabProjectToProduct,
    getSuggestedPrice,
    getMinPrice,
  };
});