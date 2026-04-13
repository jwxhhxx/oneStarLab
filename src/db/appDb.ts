import dayjs from 'dayjs';
import Dexie, { type Table } from 'dexie';

import type { LabInspiration, LabProject, Order, Product, PricingRule } from '@/types';

class OneStarLabDb extends Dexie {
  products!: Table<Product, number>;
  orders!: Table<Order, number>;
  pricingRules!: Table<PricingRule, number>;
  labInspirations!: Table<LabInspiration, number>;
  labProjects!: Table<LabProject, number>;

  constructor() {
    super('oneStarLabDb');

    this.version(1).stores({
      products: '++id,sku,category,name,stock,createdAt',
      orders: '++id,orderNo,channel,createdAt,status',
      pricingRules: '++id',
    });

    this.version(2).stores({
      products: '++id,sku,category,name,stock,createdAt',
      orders: '++id,orderNo,channel,createdAt,status',
      pricingRules: '++id',
      labInspirations: '++id,status,tag,updatedAt',
      labProjects: '++id,stage,launchDate,updatedAt',
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

const demoInspirations: Omit<LabInspiration, 'id'>[] = [
  {
    title: '雾灰茶会系列',
    tag: '春夏灵感',
    status: '待打样',
    summary: '偏奶油灰与浅茶棕的拼贴调性，适合做和纸胶带、便签与套装页。',
    keywords: '留白 / 复古 / 低饱和',
    image: '',
    createdAt: dayjs().subtract(7, 'day').toISOString(),
    updatedAt: dayjs().subtract(1, 'day').toISOString(),
  },
  {
    title: '月光整理册',
    tag: '新本册方向',
    status: '灵感中',
    summary: '用极简封面和轻纹理纸张做一本偏记录感的周计划本。',
    keywords: '冷白 / 雾蓝 / 银灰',
    image: '',
    createdAt: dayjs().subtract(5, 'day').toISOString(),
    updatedAt: dayjs().toISOString(),
  },
];

const demoLabProjects: Omit<LabProject, 'id'>[] = [
  {
    name: '旧梦胶带套组',
    category: '胶带',
    stage: '打样中',
    progress: 60,
    launchDate: dayjs().add(8, 'day').format('YYYY-MM-DD'),
    note: '第一版颜色略深，准备调浅一点。',
    image: '',
    sampleRecords: [
      {
        id: 'proof-1',
        title: '第一轮打样',
        date: dayjs().subtract(2, 'day').format('YYYY-MM-DD'),
        note: '纸面显色偏深，需要降低蓝灰浓度。',
        image: '',
      },
    ],
    updatedAt: dayjs().toISOString(),
  },
  {
    name: '研究所限定套装',
    category: '套装',
    stage: '排产中',
    progress: 88,
    launchDate: dayjs().add(14, 'day').format('YYYY-MM-DD'),
    note: '清单已确认，等待供应商排期。',
    image: '',
    sampleRecords: [],
    updatedAt: dayjs().subtract(1, 'day').toISOString(),
  },
];

export async function seedLabData() {
  await db.labInspirations.bulkAdd(demoInspirations);
  await db.labProjects.bulkAdd(demoLabProjects);
}

export async function seedDemoData() {
  await db.products.bulkAdd(demoProducts);
  await db.pricingRules.put(defaultPricingRule);
  await seedLabData();

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