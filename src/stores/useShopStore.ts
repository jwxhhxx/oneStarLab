import dayjs from 'dayjs';
import { computed, ref } from 'vue';
import { acceptHMRUpdate, defineStore } from 'pinia';

import { db, defaultPricingRule } from '@/db/appDb';
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
  const categories = ref<import('@/types').Category[]>([]);
  const orders = ref<Order[]>([]);
  const pricingRule = ref<PricingRule>({ ...defaultPricingRule });
  const labInspirations = ref<LabInspiration[]>([]);
  const labProjects = ref<LabProject[]>([]);
  const drawingInspirations = ref<DrawingInspirationRecord[]>([]);
  const expenses = ref<ExpenseRecord[]>([]);
  const inventoryTransactions = ref<InventoryTransaction[]>([]);

  async function loadAll() {
    products.value = await db.products.orderBy('createdAt').reverse().toArray();
    categories.value = await db.categories.orderBy('name').toArray();
    orders.value = await db.orders.orderBy('createdAt').reverse().toArray();
    pricingRule.value = (await db.pricingRules.get(1)) ?? { ...defaultPricingRule };
    labInspirations.value = await db.labInspirations.orderBy('updatedAt').reverse().toArray();
    labProjects.value = await db.labProjects.orderBy('updatedAt').reverse().toArray();
    drawingInspirations.value = await db.drawingInspirations.orderBy('createdAt').reverse().toArray();
    expenses.value = await db.expenses.orderBy('createdAt').reverse().toArray();
    inventoryTransactions.value = await db.inventoryTransactions.orderBy('createdAt').reverse().toArray();
  }

  async function initialize() {
    if (initialized.value) return;

    loading.value = true;
    try {
      const rule = await db.pricingRules.get(1);

      if (!rule) {
        await db.pricingRules.put({ ...defaultPricingRule });
      }

      // Ensure categories table has sensible defaults and any existing product categories
      await ensureCategoriesSeeded();

      await loadAll();
      initialized.value = true;
    } finally {
      loading.value = false;
    }
  }

  async function ensureCategoriesSeeded() {
    const defaultCats = ['胶带', '贴纸', '便签', '套装'];

    const existing = await db.categories.toArray();
    const existingNames = new Set(existing.map((c) => c.name));

    const toAddSet = new Set<string>();

    // always ensure defaults exist
    for (const d of defaultCats) {
      if (!existingNames.has(d)) toAddSet.add(d);
    }

    // if there are no categories stored, extract from products for compatibility
    if (existing.length === 0) {
      const prods = await db.products.toArray();
      for (const p of prods) {
        if (p.category && p.category.trim() && !existingNames.has(p.category.trim())) {
          toAddSet.add(p.category.trim());
        }
      }
    }

    if (toAddSet.size > 0) {
      await db.transaction('rw', db.categories, async () => {
        for (const name of toAddSet) {
          await db.categories.add({ name, createdAt: new Date().toISOString() });
        }
      });
    }
  }

  async function addProduct(input: ProductInput) {
    // ensure category exists first
    const nm = (input.category || '').trim();
    if (nm) {
      await addCategory(nm);
    }

    // auto-generate SKU when not provided
    if (!(input.sku || '').trim()) {
      // try to get the category record to use its id as prefix
      let cat = null as (import('@/types').Category | undefined) | null;
      if (nm) cat = await db.categories.where('name').equals(nm).first();
      const categoryId = (cat && cat.id) ? cat.id : 0;

      // count existing products in this category to derive sequence
      const existingCount = nm ? await db.products.where('category').equals(nm).count() : await db.products.count();
      let seq = existingCount + 1;
      let sku = `${categoryId}-${String(seq).padStart(4, '0')}`;

      // ensure uniqueness (in case of race or manual SKUs)
      // increment seq until an unused SKU is found
      // note: this loop should be safe because sku space is large
      // and collisions are unlikely; it prevents accidental overwrite
      // if another product already has the same SKU.
      // eslint-disable-next-line no-constant-condition
      while (await db.products.where('sku').equals(sku).first()) {
        seq += 1;
        sku = `${categoryId}-${String(seq).padStart(4, '0')}`;
      }

      input.sku = sku;
    }

    const suggested = calculateSuggestedPrice(input.purchaseCost, input.packagingCost, pricingRule.value);

    await db.products.add({
      ...input,
      salePrice: input.salePrice > 0 ? input.salePrice : suggested,
      createdAt: new Date().toISOString(),
    });

    await loadAll();
  }

  async function updateProduct(id: number, input: ProductInput) {
    // ensure category exists
    if (input.category && input.category.trim()) {
      await addCategory(input.category.trim());
    }
    const suggested = calculateSuggestedPrice(input.purchaseCost, input.packagingCost, pricingRule.value);

    await db.products.update(id, {
      ...input,
      salePrice: input.salePrice > 0 ? input.salePrice : suggested,
    });

    await loadAll();
  }

  async function deleteProduct(id: number) {
    await db.products.delete(id);
    await loadAll();
  }

  async function addCategory(name: string) {
    const nm = (name || '').trim();
    if (!nm) return;
    const exists = await db.categories.where('name').equals(nm).first();
    if (!exists) {
      await db.categories.add({ name: nm, createdAt: new Date().toISOString() });
      await loadAll();
    }
  }

  async function generateSkuForCategory(categoryName: string, excludeProductId?: number) {
    const nm = (categoryName || '').trim();
    // ensure category exists
    let cat = null as (import('@/types').Category | undefined) | null;
    if (nm) {
      cat = await db.categories.where('name').equals(nm).first();
      if (!cat) {
        await db.categories.add({ name: nm, createdAt: new Date().toISOString() });
        cat = await db.categories.where('name').equals(nm).first();
      }
    }

    const categoryId = (cat && cat.id) ? cat.id : 0;

    const existingCount = nm ? await db.products.where('category').equals(nm).count() : await db.products.count();
    let seq = existingCount + 1;
    let sku = `${categoryId}-${String(seq).padStart(4, '0')}`;

    while (true) {
      const existing = await db.products.where('sku').equals(sku).first();
      if (!existing) break;
      if (excludeProductId && existing.id === excludeProductId) break;
      seq += 1;
      sku = `${categoryId}-${String(seq).padStart(4, '0')}`;
    }

    return sku;
  }

  async function deleteCategory(id: number, reassignToName?: string | null) {
    const cat = await db.categories.get(id);
    if (!cat) return;

    // find products using this category
    const productsUsing = await db.products.where('category').equals(cat.name).toArray();

    await db.transaction('rw', db.products, db.categories, async () => {
      if (productsUsing.length > 0) {
        if (reassignToName && reassignToName.trim()) {
          // ensure target category exists
          const targetName = reassignToName.trim();
          const exists = await db.categories.where('name').equals(targetName).first();
          if (!exists) {
            await db.categories.add({ name: targetName, createdAt: new Date().toISOString() });
          }

          // update products to use the new category name
          for (const p of productsUsing) {
            await db.products.update(p.id!, { category: targetName });
          }
        } else {
          // clear category field for affected products
          for (const p of productsUsing) {
            await db.products.update(p.id!, { category: '' });
          }
        }
      }

      await db.categories.delete(id);
    });

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

  function buildOrderPayload(
    product: Product,
    input: NewOrderInput,
    overrides: Partial<Pick<Order, 'orderNo' | 'createdAt' | 'status'>> = {},
  ) {
    const originalAmount = Number((product.salePrice * input.quantity).toFixed(2));
    const totalAmount = Number(Math.max(0, originalAmount - input.discountAmount).toFixed(2));
    const platformFee = Number((totalAmount * input.platformFeeRate).toFixed(2));
    const goodsCost = Number(((product.purchaseCost + product.packagingCost) * input.quantity).toFixed(2));
    const netProfit = Number((totalAmount - input.shippingCost - platformFee - goodsCost).toFixed(2));

    return {
      orderNo: overrides.orderNo ?? `ORD-${dayjs().format('YYYYMMDD-HHmmss')}`,
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
      createdAt: overrides.createdAt ?? new Date().toISOString(),
      status: overrides.status ?? 'paid',
    } satisfies Omit<Order, 'id'>;
  }

  async function addOrder(input: NewOrderInput) {
    const product = await db.products.get(input.productId);

    if (!product?.id) {
      throw new Error('未找到商品');
    }

    if (product.stock < input.quantity) {
      throw new Error('库存不足，请先补货');
    }

    const orderPayload = buildOrderPayload(product, input);

    await db.transaction('rw', db.orders, db.products, async () => {
      await db.orders.add(orderPayload);

      await db.products.update(product.id!, {
        stock: product.stock - input.quantity,
      });
    });

    await loadAll();
  }

  async function updateOrder(id: number, input: NewOrderInput) {
    const existingOrder = await db.orders.get(id);
    if (!existingOrder) {
      throw new Error('未找到订单');
    }

    const previousItem = existingOrder.items[0];
    if (!previousItem) {
      throw new Error('订单缺少商品信息');
    }

    await db.transaction('rw', db.orders, db.products, async () => {
      const previousProduct = await db.products.get(previousItem.productId);
      const nextProduct = await db.products.get(input.productId);

      if (!nextProduct?.id) {
        throw new Error('未找到商品');
      }

      if (previousProduct?.id && previousProduct.id !== nextProduct.id) {
        await db.products.update(previousProduct.id, {
          stock: previousProduct.stock + previousItem.quantity,
        });
      }

      const availableStock = nextProduct.stock + (nextProduct.id === previousItem.productId ? previousItem.quantity : 0);
      if (availableStock < input.quantity) {
        throw new Error('库存不足，请先补货');
      }

      const orderPayload = buildOrderPayload(nextProduct, input, {
        orderNo: existingOrder.orderNo,
        createdAt: existingOrder.createdAt,
        status: existingOrder.status,
      });

      await db.orders.update(id, orderPayload);
      await db.products.update(nextProduct.id, {
        stock: availableStock - input.quantity,
      });
    });

    await loadAll();
  }

  async function deleteOrder(id: number) {
    const existingOrder = await db.orders.get(id);
    if (!existingOrder) {
      throw new Error('未找到订单');
    }

    const previousItem = existingOrder.items[0];

    await db.transaction('rw', db.orders, db.products, async () => {
      if (previousItem) {
        const product = await db.products.get(previousItem.productId);
        if (product?.id) {
          await db.products.update(product.id, {
            stock: product.stock + previousItem.quantity,
          });
        }
      }

      await db.orders.delete(id);
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

  async function addDrawingInspiration(record: Omit<DrawingInspirationRecord, 'id'>) {
    await db.drawingInspirations.add(record);
    await loadAll();
  }

  async function updateDrawingInspiration(id: number, patch: Partial<DrawingInspirationRecord>) {
    await db.drawingInspirations.update(id, patch);
    await loadAll();
  }

  async function deleteDrawingInspiration(id: number) {
    await db.drawingInspirations.delete(id);
    await loadAll();
  }

  async function addExpense(input: ExpenseInput) {
    await db.expenses.add({
      ...input,
      amount: Number(input.amount),
      createdAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function addInventoryTransaction(input: InventoryTransactionInput) {
    await db.inventoryTransactions.add({
      ...input,
      createdAt: new Date().toISOString(),
    });
    await loadAll();
  }

  async function processInventoryTransaction(
    productId: number,
    type: 'in' | 'out',
    quantity: number,
    barcode?: string,
    note?: string,
  ) {
    const product = await db.products.get(productId);
    if (!product || !product.id) throw new Error('未找到商品');

    await db.transaction('rw', db.products, db.orders, db.inventoryTransactions, async () => {
      const newStock = type === 'in' ? product.stock + quantity : product.stock - quantity;
      if (type === 'out' && newStock < 0) throw new Error('库存不足');

      // 如果是出库，先创建订单记录（与库存变动放在同一事务中，保持原子性）
      if (type === 'out') {
        const orderInput = {
          productId,
          customerName: note ? `扫码:${note}` : '扫码出库',
          channel: '扫码',
          quantity,
          discountAmount: 0,
          shippingCost: 0,
          platformFeeRate: 0,
        } as NewOrderInput;

        const orderPayload = buildOrderPayload(product, orderInput);
        await db.orders.add(orderPayload);
      }

      await db.products.update(product.id!, { stock: newStock });

      await db.inventoryTransactions.add({
        productId,
        barcode: barcode ?? product.sku,
        type,
        quantity,
        note: note ?? '',
        createdAt: new Date().toISOString(),
      });
    });

    await loadAll();
  }

  async function createOrderFromScans(
    items: Array<{ productId: number; quantity: number; barcode?: string; note?: string }> ,
    type: 'in' | 'out',
    batchNote?: string,
  ) {
    if (!items || items.length === 0) throw new Error('没有扫描到任何商品');

    // 聚合同一商品
    const agg = new Map<number, { quantity: number; barcode?: string }>();
    for (const it of items) {
      const cur = agg.get(it.productId);
      if (cur) cur.quantity += it.quantity;
      else agg.set(it.productId, { quantity: it.quantity, barcode: it.barcode });
    }

    const productIds = Array.from(agg.keys());

    let newOrderNo = '';

    await db.transaction('rw', db.products, db.orders, db.inventoryTransactions, async () => {
      const productsList = await db.products.bulkGet(productIds);

      // validate
      for (const pid of productIds) {
        const p = productsList.find((x) => x && x.id === pid);
        if (!p) throw new Error(`未找到商品 id=${pid}`);
      }

      // check stock for out
      if (type === 'out') {
        for (const pid of productIds) {
          const p = productsList.find((x) => x && x.id === pid)!;
          const need = agg.get(pid)!.quantity;
          if (p.stock < need) throw new Error(`商品 ${p.name} 库存不足`);
        }
      }

      // build order items and totals
      const orderItems = [] as Array<import('@/types').OrderItem>;
      let totalAmount = 0;
      let goodsCost = 0;

      for (const pid of productIds) {
        const p = productsList.find((x) => x && x.id === pid)! as Product;
        const qty = agg.get(pid)!.quantity;
        const itemTotal = Number((p.salePrice * qty).toFixed(2));
        const itemCost = Number(((p.purchaseCost + p.packagingCost) * qty).toFixed(2));
        orderItems.push({
          productId: p.id!,
          productName: p.name,
          sku: p.sku,
          quantity: qty,
          salePrice: p.salePrice,
          totalAmount: itemTotal,
          totalCost: itemCost,
        });
        totalAmount += itemTotal;
        goodsCost += itemCost;
      }

      const platformFee = 0;
      const shippingCost = 0;
      const discountAmount = 0;
      const netProfit = Number((totalAmount - shippingCost - platformFee - goodsCost).toFixed(2));

      const orderPayload = {
        orderNo: `ORD-${dayjs().format('YYYYMMDD-HHmmss')}`,
        channel: '扫码',
        customerName: batchNote ? `扫码:${batchNote}` : '扫码出库批量',
        items: orderItems,
        totalAmount: Number(totalAmount.toFixed(2)),
        discountAmount,
        shippingCost,
        platformFee,
        goodsCost: Number(goodsCost.toFixed(2)),
        netProfit,
        createdAt: new Date().toISOString(),
        status: 'paid' as const,
      } as Omit<Order, 'id'>;

      const id = await db.orders.add(orderPayload as any);
      newOrderNo = orderPayload.orderNo;

      // update stocks and add inventory transactions
      for (const pid of productIds) {
        const p = productsList.find((x) => x && x.id === pid)! as Product;
        const qty = agg.get(pid)!.quantity;
        const newStock = type === 'in' ? p.stock + qty : p.stock - qty;
        await db.products.update(p.id!, { stock: newStock });

        await db.inventoryTransactions.add({
          productId: p.id!,
          barcode: agg.get(pid)!.barcode ?? p.sku,
          type,
          quantity: qty,
          note: batchNote ?? '',
          createdAt: new Date().toISOString(),
        });
      }
    });

    await loadAll();

    return newOrderNo;
  }

  async function findProductByBarcode(barcode: string) {
    return await db.products.where('sku').equals(barcode).first();
  }

  async function deleteInventoryTransaction(id: number) {
    await db.inventoryTransactions.delete(id);
    await loadAll();
  }

  async function updateExpense(id: number, input: ExpenseInput) {
    await db.expenses.update(id, {
      ...input,
      amount: Number(input.amount),
    });
    await loadAll();
  }

  async function deleteExpense(id: number) {
    await db.expenses.delete(id);
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