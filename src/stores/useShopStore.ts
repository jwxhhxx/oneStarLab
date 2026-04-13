import dayjs from 'dayjs';
import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

import { db, defaultPricingRule, seedDemoData } from '@/db/appDb';
import type { NewOrderInput, Order, PricingRule, Product, ProductInput } from '@/types';
import { calculateMinPrice, calculateSuggestedPrice } from '@/utils/pricing';

export const useShopStore = defineStore('shop', () => {
  const loading = ref(false);
  const initialized = ref(false);
  const products = ref<Product[]>([]);
  const orders = ref<Order[]>([]);
  const pricingRule = ref<PricingRule>({ ...defaultPricingRule });

  async function loadAll() {
    products.value = await db.products.orderBy('createdAt').reverse().toArray();
    orders.value = await db.orders.orderBy('createdAt').reverse().toArray();
    pricingRule.value = (await db.pricingRules.get(1)) ?? { ...defaultPricingRule };
  }

  async function initialize() {
    if (initialized.value) return;

    loading.value = true;
    try {
      const productCount = await db.products.count();
      const rule = await db.pricingRules.get(1);

      if (productCount === 0) {
        await seedDemoData();
      } else if (!rule) {
        await db.pricingRules.put({ ...defaultPricingRule });
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

  return {
    loading,
    initialized,
    products,
    orders,
    pricingRule,
    stats,
    lowStockProducts,
    recentOrders,
    monthlyTrend,
    initialize,
    addProduct,
    addOrder,
    savePricingRule,
    getSuggestedPrice,
    getMinPrice,
  };
});