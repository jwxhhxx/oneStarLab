export interface Product {
  id?: number;
  name: string;
  sku: string;
  category: string;
  supplier: string;
  purchaseCost: number;
  packagingCost: number;
  salePrice: number;
  stock: number;
  safeStock: number;
  note?: string;
  createdAt: string;
}

export interface OrderItem {
  productId: number;
  productName: string;
  sku: string;
  quantity: number;
  salePrice: number;
  totalAmount: number;
  totalCost: number;
}

export interface Order {
  id?: number;
  orderNo: string;
  channel: string;
  customerName: string;
  items: OrderItem[];
  totalAmount: number;
  discountAmount: number;
  shippingCost: number;
  platformFee: number;
  goodsCost: number;
  netProfit: number;
  createdAt: string;
  status: 'paid' | 'refunded';
}

export interface PricingRule {
  id?: number;
  targetMargin: number;
  channelFeeRate: number;
  extraCost: number;
  roundingMode: 'none' | 'x.9' | 'whole';
}

export interface ProductInput {
  name: string;
  sku: string;
  category: string;
  supplier: string;
  purchaseCost: number;
  packagingCost: number;
  salePrice: number;
  stock: number;
  safeStock: number;
  note?: string;
}

export interface NewOrderInput {
  productId: number;
  customerName: string;
  channel: string;
  quantity: number;
  discountAmount: number;
  shippingCost: number;
  platformFeeRate: number;
}

export type InspirationStatus = '灵感中' | '待打样' | '打样中' | '待上新';
export type LabStage = '灵感整理' | '设计中' | '打样中' | '排产中' | '待上新' | '已上新';

export interface LabInspiration {
  id?: number;
  title: string;
  tag: string;
  status: InspirationStatus;
  summary: string;
  keywords: string;
  image?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProofRecord {
  id: string;
  title: string;
  date: string;
  note: string;
  image?: string;
}

export interface LabProject {
  id?: number;
  name: string;
  category: string;
  stage: LabStage;
  progress: number;
  launchDate: string;
  note: string;
  image?: string;
  sampleRecords: ProofRecord[];
  updatedAt: string;
}

export interface InspirationInput {
  title: string;
  tag: string;
  status: InspirationStatus;
  summary: string;
  keywords: string;
  image?: string;
}

export interface LabProjectInput {
  name: string;
  category: string;
  stage: LabStage;
  progress: number;
  launchDate: string;
  note: string;
  image?: string;
}

export interface ProofRecordInput {
  title: string;
  date: string;
  note: string;
  image?: string;
}
