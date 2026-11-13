import dayjs from 'dayjs';
import { computed, ref } from 'vue';
import { acceptHMRUpdate, defineStore } from 'pinia';

import { api } from '@/api/client';
import type {
  DrawingInspirationRecord,
  ExpenseInput,
  ExpenseRecord,
  InventoryTransaction,
  InventoryTransactionInput,
  InspirationInput,
  LabInspiration,
  LabProject,
  LabProjectInput,
  LabStage,
  MarketInput,
  MarketRecord,
  NewOrderInput,
  Order,
  PlannerTemplate,
  PlannerTemplateInput,
  PricingRule,
  Product,
  ProductInput,
  ProofRecordInput,
} from '@/types';
import { calculateMinPrice, calculateSuggestedPrice } from '@/utils/pricing';

const defaultPricingRule: PricingRule = {
  id: 1,
  targetMargin: 0.45,
  channelFeeRate: 0.05,
  extraCost: 1.2,
  roundingMode: 'x.9',
};

const MAX_CATEGORY_NAME_LENGTH = 10;

export const useShopStore = defineStore('shop', () => {
  const loading = ref(false);
  const initialized = ref(false);
  const products = ref<Product[]>([]);
  const categories = ref<import('@/types').Category[]>([]);
  const orders = ref<Order[]>([]);
  const pricingRule = ref<PricingRule>({ ...defaultPricingRule });
  const labInspirations = ref<LabInspiration[]>([]);
  const labProjects = ref<LabProject[]>([]);
  const drawingInspirations = ref<DrawingInspirationRecord[]>([]);
  const expenses = ref<ExpenseRecord[]>([]);
  const inventoryTransactions = ref<InventoryTransaction[]>([]);
  const markets = ref<MarketRecord[]>([]);
  const onlineIncome = ref(0);
  const onlineIncomeUpdatedAt = ref('');
  const plannerTemplates = ref<PlannerTemplate[]>([]);

  async function loadAll() {
    const data = await api.bootstrap();
    products.value = data.products;
    categories.value = data.categories;
    orders.value = data.orders;
    pricingRule.value = data.pricingRule;
    labInspirations.value = data.labInspirations;
    labProjects.value = data.labProjects;
    drawingInspirations.value = data.drawingInspirations;
    expenses.value = data.expenses;
    inventoryTransactions.value = data.inventoryTransactions;
    markets.value = data.markets;
    onlineIncome.value = data.onlineIncome;
    onlineIncomeUpdatedAt.value = data.onlineIncomeUpdatedAt;
    plannerTemplates.value = data.plannerTemplates;
  }

  async function initialize() {
    if (initialized.value) return;

    loading.value = true;
    try {
      await loadAll();
      initialized.value = true;
    } finally {
      loading.value = false;
    }
  }

  async function addCategory(name: string) {
    let nm = (name || '').trim();
    if (!nm) return;
    if (nm.length > MAX_CATEGORY_NAME_LENGTH) {
      nm = nm.slice(0, MAX_CATEGORY_NAME_LENGTH);
    }
    if (categories.value.some((item) => item.name === nm)) return;
    await api.categories.create(nm);
    categories.value = [...categories.value, { id: Date.now(), name: nm, createdAt: new Date().toISOString() }];
  }

  async function addProduct(input: ProductInput) {
    const nm = (input.category || '').trim();
    if (nm) {
      await addCategory(nm);
    }
    await api.products.create({ ...input, category: nm });
    await loadAll();
  }

  async function updateProduct(id: number, input: ProductInput) {
    if (input.category && input.category.trim()) {
      await addCategory(input.category.trim());
    }
    await api.products.update(id, input);
    await loadAll();
  }

  async function deleteProduct(id: number) {
    await api.products.remove(id);
    await loadAll();
  }

  async function generateSkuForCategory(
    categoryName: string,
    excludeProductId?: number,
    createIfMissing = true,
  ) {
    const result = await api.products.suggestSku(
      (categoryName || '').trim(),
      excludeProductId,
      createIfMissing,
    );
    return result.sku;
  }

  async function deleteCategory(id: number, reassignToName?: string | null) {
    await api.categories.remove(id, reassignToName ?? null);
    await loadAll();
  }

  function getCategoryUsageCounts() {
    const map: Record<string, number> = {};
    for (const p of products.value) {
      const key = p.category || '';
      map[key] = (map[key] || 0) + 1;
    }
    return map;
  }

  async function addOrder(input: NewOrderInput) {
    await api.orders.create(input);
    await loadAll();
  }

  async function updateOrder(id: number, input: NewOrderInput) {
    await api.orders.update(id, input);
    await loadAll();
  }

  async function deleteOrder(id: number) {
    await api.orders.remove(id);
    await loadAll();
  }

  async function savePricingRule(rule: PricingRule) {
    const payload = { ...rule, id: 1 };
    await api.pricing.save(payload);
    pricingRule.value = payload;
  }

  async function addInspiration(input: InspirationInput) {
    await api.lab.createInspiration(input);
    await loadAll();
  }

  async function updateInspiration(id: number, input: InspirationInput) {
    await api.lab.updateInspiration(id, input);
    await loadAll();
  }

  async function deleteInspiration(id: number) {
    await api.lab.removeInspiration(id);
    await loadAll();
  }

  async function addLabProject(input: LabProjectInput) {
    await api.lab.createProject(input);
    await loadAll();
  }

  async function updateLabProject(id: number, input: LabProjectInput) {
    await api.lab.updateProject(id, input);
    await loadAll();
  }

  async function updateLabProjectStage(id: number, stage: LabStage) {
    await api.lab.updateStage(id, stage);
    await loadAll();
  }

  async function addProofRecord(projectId: number, input: ProofRecordInput) {
    await api.lab.addProof(projectId, input);
    await loadAll();
  }

  async function deleteLabProject(id: number) {
    await api.lab.removeProject(id);
    await loadAll();
  }

  async function convertLabProjectToProduct(id: number) {
    await api.lab.convert(id);
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

  async function addDrawingInspiration(record: Omit<DrawingInspirationRecord, 'id'>) {
    await api.drawings.create(record);
    await loadAll();
  }

  async function updateDrawingInspiration(id: number, patch: Partial<DrawingInspirationRecord>) {
    await api.drawings.update(id, patch);
    await loadAll();
  }

  async function deleteDrawingInspiration(id: number) {
    await api.drawings.remove(id);
    await loadAll();
  }

  async function addExpense(input: ExpenseInput) {
    await api.expenses.create(input);
    await loadAll();
  }

  async function updateExpense(id: number, input: ExpenseInput) {
    await api.expenses.update(id, input);
    await loadAll();
  }

  async function deleteExpense(id: number) {
    await api.expenses.remove(id);
    await loadAll();
  }

  async function addInventoryTransaction(input: InventoryTransactionInput) {
    await api.inventory.create(input);
    await loadAll();
  }

  async function processInventoryTransaction(
    productId: number,
    type: 'in' | 'out',
    quantity: number,
    barcode?: string,
    note?: string,
  ) {
    await api.inventory.process({ productId, type, quantity, barcode, note });
    await loadAll();
  }

  async function createOrderFromScans(
    items: Array<{ productId: number; quantity: number; barcode?: string; note?: string }>,
    type: 'in' | 'out',
    batchNote?: string,
  ) {
    const result = await api.inventory.scan({ items, type, note: batchNote });
    await loadAll();
    return result.orderNo;
  }

  async function findProductByBarcode(barcode: string) {
    try {
      return await api.products.findByBarcode(barcode);
    } catch {
      return undefined;
    }
  }

  async function deleteInventoryTransaction(id: number) {
    await api.inventory.remove(id);
    await loadAll();
  }

  async function addMarket(input: MarketInput) {
    await api.markets.create(input);
    await loadAll();
  }

  async function updateMarket(id: number, input: MarketInput) {
    await api.markets.update(id, input);
    await loadAll();
  }

  async function deleteMarket(id: number) {
    await api.markets.remove(id);
    await loadAll();
  }

  async function saveOnlineIncome(amount: number) {
    const value = Number(amount);

    if (!Number.isFinite(value) || value < 0) {
      throw new Error('线上收入需为不小于 0 的数字');
    }

    const saved = await api.settings.saveOnlineIncome(value);
    onlineIncome.value = saved.onlineIncome;
    onlineIncomeUpdatedAt.value = saved.updatedAt;
  }

  async function savePlannerTemplate(input: PlannerTemplateInput, id?: number) {
    const template = id != null
      ? await api.planner.update(id, input)
      : await api.planner.create(input);
    await loadAll();
    return template.id ?? id ?? 0;
  }

  async function deletePlannerTemplate(id: number) {
    await api.planner.remove(id);
    await loadAll();
  }

  return {
    loading,
    initialized,
    products,
    orders,
    pricingRule,
    labInspirations,
    labProjects,
    drawingInspirations,
    expenses,
    inventoryTransactions,
    markets,
    onlineIncome,
    onlineIncomeUpdatedAt,
    plannerTemplates,
    stats,
    lowStockProducts,
    recentOrders,
    monthlyTrend,
    upcomingLaunches,
    loadAll,
    initialize,
    addProduct,
    updateProduct,
    deleteProduct,
    addOrder,
    updateOrder,
    deleteOrder,
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
    addDrawingInspiration,
    updateDrawingInspiration,
    deleteDrawingInspiration,
    addExpense,
    addInventoryTransaction,
    processInventoryTransaction,
    createOrderFromScans,
    findProductByBarcode,
    deleteInventoryTransaction,
    updateExpense,
    deleteExpense,
    addMarket,
    updateMarket,
    deleteMarket,
    saveOnlineIncome,
    savePlannerTemplate,
    deletePlannerTemplate,
    getSuggestedPrice,
    getMinPrice,
    getCategoryUsageCounts,
    generateSkuForCategory,
    categories,
    addCategory,
    deleteCategory,
  };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useShopStore, import.meta.hot));
}
