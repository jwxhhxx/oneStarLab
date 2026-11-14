import type {
  BootstrapData,
  Category,
  DrawingInspirationRecord,
  ExpenseInput,
  ExpenseRecord,
  InspirationInput,
  InventoryTransaction,
  InventoryTransactionInput,
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

const API_BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1';
const API_TOKEN = import.meta.env.VITE_API_TOKEN as string | undefined;

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

interface RequestOptions {
  method?: Method;
  body?: unknown;
  skipAuthRedirect?: boolean;
}

interface AuthStatus {
  authenticated: boolean;
  usingDefault: boolean;
}

interface AuthResult {
  ok: boolean;
  usingDefault?: boolean;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  if (options.body !== undefined) headers['Content-Type'] = 'application/json';
  if (API_TOKEN) headers.Authorization = `Bearer ${API_TOKEN}`;

  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      method: options.method ?? 'GET',
      headers,
      credentials: 'include',
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    });
  } catch {
    throw new Error('无法连接后端服务，请确认 Go 服务已启动');
  }

  if (!response.ok) {
    if (response.status === 401 && !options.skipAuthRedirect) {
      window.dispatchEvent(new CustomEvent('onestarlab:unauthorized'));
    }
    let message = `请求失败（${response.status}）`;
    try {
      const data = (await response.json()) as { error?: string };
      if (data.error) message = data.error;
    } catch {
      // ignore malformed error body
    }
    throw new Error(message);
  }

  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

function query(params: Record<string, string | number | boolean | undefined>): string {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') search.set(key, String(value));
  });
  const text = search.toString();
  return text ? `?${text}` : '';
}

export const api = {
  auth: {
    status: () => request<AuthStatus>('/auth/status', { skipAuthRedirect: true }),
    login: (passcode: string) =>
      request<AuthResult>('/auth/login', {
        method: 'POST',
        body: { passcode },
        skipAuthRedirect: true,
      }),
    logout: () =>
      request<{ ok: boolean }>('/auth/logout', { method: 'POST', skipAuthRedirect: true }),
    changePasscode: (current: string, next: string) =>
      request<AuthResult>('/auth/passcode', { method: 'POST', body: { current, next } }),
  },

  bootstrap: () => request<BootstrapData>('/bootstrap'),

  products: {
    list: () => request<Product[]>('/products'),
    create: (input: ProductInput) => request<Product>('/products', { method: 'POST', body: input }),
    update: (id: number, input: ProductInput) =>
      request<Product>(`/products/${id}`, { method: 'PUT', body: input }),
    remove: (id: number) => request<{ ok: boolean }>(`/products/${id}`, { method: 'DELETE' }),
    findByBarcode: (sku: string) => request<Product>(`/products/sku/${encodeURIComponent(sku)}`),
    suggestSku: (category: string, excludeId?: number, createIfMissing = true) =>
      request<{ sku: string }>('/sku', {
        method: 'POST',
        body: { category, excludeProductId: excludeId ?? 0, createIfMissing },
      }),
  },

  categories: {
    list: () => request<Category[]>('/categories'),
    create: (name: string) => request<Category>('/categories', { method: 'POST', body: { name } }),
    remove: (id: number, reassignTo?: string | null) =>
      request<{ ok: boolean }>(`/categories/${id}${query({ reassignTo: reassignTo ?? undefined })}`, {
        method: 'DELETE',
      }),
  },

  orders: {
    list: () => request<Order[]>('/orders'),
    create: (input: NewOrderInput) => request<Order>('/orders', { method: 'POST', body: input }),
    update: (id: number, input: NewOrderInput) =>
      request<Order>(`/orders/${id}`, { method: 'PUT', body: input }),
    remove: (id: number) => request<{ ok: boolean }>(`/orders/${id}`, { method: 'DELETE' }),
  },

  pricing: {
    get: () => request<PricingRule>('/pricing-rule'),
    save: (rule: PricingRule) => request<PricingRule>('/pricing-rule', { method: 'PUT', body: rule }),
  },

  inventory: {
    list: () => request<InventoryTransaction[]>('/inventory/transactions'),
    create: (input: InventoryTransactionInput) =>
      request<InventoryTransaction>('/inventory/transactions', { method: 'POST', body: input }),
    remove: (id: number) =>
      request<{ ok: boolean }>(`/inventory/transactions/${id}`, { method: 'DELETE' }),
    process: (input: {
      productId: number;
      type: 'in' | 'out';
      quantity: number;
      barcode?: string;
      note?: string;
    }) => request<{ ok: boolean }>('/inventory/process', { method: 'POST', body: input }),
    scan: (input: {
      items: Array<{ productId: number; quantity: number; barcode?: string; note?: string }>;
      type: 'in' | 'out';
      note?: string;
    }) => request<{ orderNo: string }>('/inventory/scan', { method: 'POST', body: input }),
  },

  lab: {
    listInspirations: () => request<LabInspiration[]>('/lab/inspirations'),
    createInspiration: (input: InspirationInput) =>
      request<LabInspiration>('/lab/inspirations', { method: 'POST', body: input }),
    updateInspiration: (id: number, input: InspirationInput) =>
      request<LabInspiration>(`/lab/inspirations/${id}`, { method: 'PUT', body: input }),
    removeInspiration: (id: number) =>
      request<{ ok: boolean }>(`/lab/inspirations/${id}`, { method: 'DELETE' }),
    listProjects: () => request<LabProject[]>('/lab/projects'),
    createProject: (input: LabProjectInput) =>
      request<LabProject>('/lab/projects', { method: 'POST', body: input }),
    updateProject: (id: number, input: LabProjectInput) =>
      request<LabProject>(`/lab/projects/${id}`, { method: 'PUT', body: input }),
    removeProject: (id: number) =>
      request<{ ok: boolean }>(`/lab/projects/${id}`, { method: 'DELETE' }),
    updateStage: (id: number, stage: LabStage) =>
      request<LabProject>(`/lab/projects/${id}/stage`, { method: 'PATCH', body: { stage } }),
    addProof: (id: number, input: ProofRecordInput) =>
      request<LabProject>(`/lab/projects/${id}/proofs`, { method: 'POST', body: input }),
    convert: (id: number) =>
      request<Product>(`/lab/projects/${id}/convert`, { method: 'POST', body: {} }),
  },

  drawings: {
    list: () => request<DrawingInspirationRecord[]>('/drawings'),
    create: (input: Omit<DrawingInspirationRecord, 'id'>) =>
      request<DrawingInspirationRecord>('/drawings', { method: 'POST', body: input }),
    update: (id: number, patch: Partial<DrawingInspirationRecord>) =>
      request<DrawingInspirationRecord>(`/drawings/${id}`, { method: 'PUT', body: patch }),
    remove: (id: number) => request<{ ok: boolean }>(`/drawings/${id}`, { method: 'DELETE' }),
  },

  expenses: {
    list: () => request<ExpenseRecord[]>('/expenses'),
    create: (input: ExpenseInput) => request<ExpenseRecord>('/expenses', { method: 'POST', body: input }),
    update: (id: number, input: ExpenseInput) =>
      request<ExpenseRecord>(`/expenses/${id}`, { method: 'PUT', body: input }),
    remove: (id: number) => request<{ ok: boolean }>(`/expenses/${id}`, { method: 'DELETE' }),
  },

  markets: {
    list: () => request<MarketRecord[]>('/markets'),
    create: (input: MarketInput) => request<MarketRecord>('/markets', { method: 'POST', body: input }),
    update: (id: number, input: MarketInput) =>
      request<MarketRecord>(`/markets/${id}`, { method: 'PUT', body: input }),
    remove: (id: number) => request<{ ok: boolean }>(`/markets/${id}`, { method: 'DELETE' }),
  },

  settings: {
    getOnlineIncome: () =>
      request<{ onlineIncome: number; updatedAt: string }>('/settings/online-income'),
    saveOnlineIncome: (amount: number) =>
      request<{ onlineIncome: number; updatedAt: string }>('/settings/online-income', {
        method: 'PUT',
        body: { amount },
      }),
  },

  planner: {
    list: () => request<PlannerTemplate[]>('/planner-templates'),
    create: (input: PlannerTemplateInput) =>
      request<PlannerTemplate>('/planner-templates', { method: 'POST', body: input }),
    update: (id: number, input: PlannerTemplateInput) =>
      request<PlannerTemplate>(`/planner-templates/${id}`, { method: 'PUT', body: input }),
    remove: (id: number) => request<{ ok: boolean }>(`/planner-templates/${id}`, { method: 'DELETE' }),
  },

  exportSnapshot: async (): Promise<Blob> => {
    const headers: Record<string, string> = {};
    if (API_TOKEN) headers.Authorization = `Bearer ${API_TOKEN}`;
    const response = await fetch(`${API_BASE}/export`, { headers, credentials: 'include' });
    if (!response.ok) throw new Error(`导出失败（${response.status}）`);
    return await response.blob();
  },

  importSnapshot: async (snapshot: unknown): Promise<{ ok: boolean; source?: 'snapshot' | 'dexie' }> => {
    return await request<{ ok: boolean; source?: 'snapshot' | 'dexie' }>('/import', {
      method: 'POST',
      body: snapshot,
    });
  },
};
