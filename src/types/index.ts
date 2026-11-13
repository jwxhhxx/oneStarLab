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

export interface Category {
  id?: number;
  name: string;
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

export type DrawingCategoryKey = '植物' | '食物' | '动物' | '日常用品' | '水果' | '盐系小元素';

export interface DrawingInspirationItem {
  category: DrawingCategoryKey;
  element: string;
}

export interface DrawingInspirationRecord {
  id?: number;
  items: DrawingInspirationItem[];
  styles: string[];
  createdAt: string;
  completed: boolean;
}

export interface ExpenseRecord {
  id?: number;
  purpose: string;
  amount: number;
  createdAt: string;
}

export interface ExpenseInput {
  purpose: string;
  amount: number;
}

export interface MarketRecord {
  id?: number;
  name: string;
  date: string;
  location: string;
  income: number;
  note?: string;
  createdAt: string;
}

export interface MarketInput {
  name: string;
  date: string;
  location: string;
  income: number;
  note?: string;
}

export interface AppSetting {
  key: string;
  value: string;
  updatedAt: string;
}

export type PlannerSizeKey = 'm5' | 'm25' | 'a7' | 'a6per' | 'a6' | 'a5' | 'cm5' | 'cm8' | 'trifold' | 'custom';

export type PlannerModuleType =
  | 'calendar'
  | 'timeline'
  | 'todo'
  | 'quadrant'
  | 'ledger'
  | 'finance'
  | 'frenchGrid'
  | 'dailyPlan'
  | 'meals'
  | 'habitTracker'
  | 'weeklyPlan'
  | 'grid'
  | 'dotGrid'
  | 'lined'
  | 'blank';

export interface PlannerModuleOptions {
  rows?: number;
  days?: number;
  startHour?: number;
  endHour?: number;
  cellMm?: number;
  lineMm?: number;
  labels?: string[];
}

export interface PlannerModuleConfig {
  id: string;
  type: PlannerModuleType;
  title: string;
  weight: number;
  accent: string;
  options: PlannerModuleOptions;
}

export interface PlannerTemplate {
  id?: number;
  name: string;
  sizeKey: PlannerSizeKey;
  widthMm: number;
  heightMm: number;
  paddingMm: number;
  accent: string;
  paper: string;
  modules: PlannerModuleConfig[];
  createdAt: string;
  updatedAt: string;
}

export interface PlannerTemplateInput {
  name: string;
  sizeKey: PlannerSizeKey;
  widthMm: number;
  heightMm: number;
  paddingMm: number;
  accent: string;
  paper: string;
  modules: PlannerModuleConfig[];
}

export interface InventoryTransaction {
  id?: number;
  productId: number;
  barcode?: string;
  type: 'in' | 'out';
  quantity: number;
  note?: string;
  createdAt: string;
}

export interface InventoryTransactionInput {
  productId: number;
  barcode?: string;
  type: 'in' | 'out';
  quantity: number;
  note?: string;
}
