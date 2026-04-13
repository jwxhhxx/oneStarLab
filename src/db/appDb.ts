import dayjs from 'dayjs';
import Dexie, { type Table } from 'dexie';

import type { Order, Product, PricingRule } from '@/types';

class OneStarLabDb extends Dexie {
  products!: Table<Product, number>;
  orders!: Table<Order, number>;
  pricingRules!: Table<PricingRule, number>;

  constructor() {
    super('oneStarLabDb');

    this.version(1).stores({
      products: '++id,sku,category,name,stock,createdAt',
      orders: '++id,orderNo,channel,createdAt,status',
      pricingRules: '++id',
    });
  }
}

export const db = new OneStarLabDb();

export const defaultPricingRule: PricingRule = {
  id: 1,
  targetMargin: 0.45,
  channelFeeRate: 0.05,
  extraCost: 1.2,
  roundingMode: 'x.9',
};

const demoProducts: Omit<Product, 'id'>[] = [
  {
    name: '星夜和纸胶带',
    sku: 'TAPE-001',
    category: '胶带',
    supplier: '星纸工坊',
    purchaseCost: 7.8,
    packagingCost: 1.2,
    salePrice: 19.9,
    stock: 32,
    safeStock: 8,
    note: '适合拼贴与周记装饰',
    createdAt: dayjs().subtract(10, 'day').toISOString(),
  },
  {
    name: '春日碎花贴纸包',
    sku: 'STICKER-003',
    category: '贴纸',
    supplier: '纸间小铺',
    purchaseCost: 4.2,
    packagingCost: 0.8,
    salePrice: 12.9,
    stock: 15,
    safeStock: 10,
    note: '近期热卖款',
    createdAt: dayjs().subtract(8, 'day').toISOString(),
  },
  {
    name: '猫咪周计划便签',
    sku: 'NOTE-007',
    category: '便签',
    supplier: '晴空制纸',
    purchaseCost: 3.6,
    packagingCost: 0.6,
    salePrice: 9.9,
    stock: 48,
    safeStock: 12,
    note: '适合做赠品组合',
    createdAt: dayjs().subtract(6, 'day').toISOString(),
  },
];

export async function seedDemoData() {
  await db.products.bulkAdd(demoProducts);
  await db.pricingRules.put(defaultPricingRule);

  const products = await db.products.toArray();
  const tape = products.find((item) => item.sku === 'TAPE-001');
  const sticker = products.find((item) => item.sku === 'STICKER-003');

  if (tape?.id && sticker?.id) {
    const demoOrders: Omit<Order, 'id'>[] = [
      {
        orderNo: 'ORD-20260401-001',
        channel: '微信',
        customerName: '小满',
        items: [
          {
            productId: tape.id,
            productName: tape.name,
            sku: tape.sku,
            quantity: 2,
            salePrice: tape.salePrice,
            totalAmount: 39.8,
            totalCost: 18,
          },
        ],
        totalAmount: 39.8,
        discountAmount: 0,
        shippingCost: 5,
        platformFee: 0,
        goodsCost: 18,
        netProfit: 16.8,
        createdAt: dayjs().subtract(4, 'day').toISOString(),
        status: 'paid',
      },
      {
        orderNo: 'ORD-20260403-002',
        channel: '小红书',
        customerName: '阿桃',
        items: [
          {
            productId: sticker.id,
            productName: sticker.name,
            sku: sticker.sku,
            quantity: 3,
            salePrice: sticker.salePrice,
            totalAmount: 38.7,
            totalCost: 15,
          },
        ],
        totalAmount: 34.7,
        discountAmount: 4,
        shippingCost: 4,
        platformFee: 1.74,
        goodsCost: 15,
        netProfit: 13.96,
        createdAt: dayjs().subtract(2, 'day').toISOString(),
        status: 'paid',
      },
    ];

    await db.orders.bulkAdd(demoOrders);
  }
}