import Dexie, { type Table } from 'dexie';

import type { DrawingInspirationRecord, ExpenseRecord, LabInspiration, LabProject, Order, Product, PricingRule } from '@/types';

class OneStarLabDb extends Dexie {
  products!: Table<Product, number>;
  orders!: Table<Order, number>;
  pricingRules!: Table<PricingRule, number>;
  labInspirations!: Table<LabInspiration, number>;
  labProjects!: Table<LabProject, number>;
  drawingInspirations!: Table<DrawingInspirationRecord, number>;
  expenses!: Table<ExpenseRecord, number>;

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

    this.version(3).stores({
      products: '++id,sku,category,name,stock,createdAt',
      orders: '++id,orderNo,channel,createdAt,status',
      pricingRules: '++id',
      labInspirations: '++id,status,tag,updatedAt',
      labProjects: '++id,stage,launchDate,updatedAt',
      drawingInspirations: '++id,createdAt,completed',
    });

    this.version(4).stores({
      products: '++id,sku,category,name,stock,createdAt',
      orders: '++id,orderNo,channel,createdAt,status',
      pricingRules: '++id',
      labInspirations: '++id,status,tag,updatedAt',
      labProjects: '++id,stage,launchDate,updatedAt',
      drawingInspirations: '++id,createdAt,completed',
      expenses: '++id,createdAt,purpose',
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