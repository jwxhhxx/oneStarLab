<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { BrowserMultiFormatReader } from '@zxing/library';
import { useShopStore } from '@/stores/useShopStore';

const store = useShopStore();
onMounted(() => {
  store.initialize();
});

const videoRef = ref<HTMLVideoElement | null>(null);
const codeReader = new BrowserMultiFormatReader();
const active = ref(false);
const scanResult = ref<string | null>(null);
const scannedProduct = ref<any | null>(null);
const qty = ref(1);
const actionType = ref<'in' | 'out'>('in');
const scannerError = ref<string | null>(null);
const manualBarcode = ref('');

const fileInputRef = ref<HTMLInputElement | null>(null);

function triggerFileInput() {
  fileInputRef.value?.click();
}

async function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement;
  if (!input.files || input.files.length === 0) return;
  const files = Array.from(input.files);
  const wasActive = active.value;
  if (wasActive) stopScanner();

  for (const f of files) {
    try {
      await decodeImageFile(f);
    } catch (err) {
      console.warn('decodeImageFile error', err);
    }
  }

  input.value = '';
  if (wasActive) startScanner();
}

async function decodeImageFile(file: File) {
  const url = URL.createObjectURL(file);
  const img = new Image();
  img.crossOrigin = 'anonymous';

  await new Promise<void>((resolve, reject) => {
    img.onload = () => resolve();
    img.onerror = (e) => reject(e);
    img.src = url;
  }).catch((e) => {
    URL.revokeObjectURL(url);
    throw e;
  });

  try {
    // 使用 ZXing 的图片解码接口
    const result = await (codeReader as any).decodeFromImageElement(img);
    const code = result?.getText?.();
    if (code) {
      await handleScan(code);
    } else {
      window.alert('未能从图片识别到条码');
    }
  } catch (e) {
    console.warn('decodeFromImageElement failed', e);
    window.alert('图片解析失败：' + String(e));
  } finally {
    URL.revokeObjectURL(url);
  }
}

const scannedItems = ref<Array<{ productId?: number; barcode: string; name?: string; sku?: string; quantity: number }>>([]);
const lastScanAt: Record<string, number> = {};

async function startScanner() {
  scannerError.value = null;
  try {
    active.value = true;

    // 优先使用后置摄像头约束，兼容移动设备
    const constraints: MediaStreamConstraints = { video: { facingMode: { ideal: 'environment' } } };

    codeReader.decodeFromConstraints(constraints, videoRef.value!, (result, err) => {
      if (err && !(err instanceof Error && (err as any).name === 'NotFoundError')) {
        console.warn('decode error', err);
      }

      if (result) {
        const code = result.getText();
        scanResult.value = code;
        // 异步处理扫描结果，加入扫描队列
        void (async () => {
          await handleScan(code);
        })();
      }
    });
  } catch (e: any) {
    console.error(e);
    // 更友好的错误提示，提示 HTTPS 与权限问题
    if (e && e.name === 'NotAllowedError') {
      scannerError.value = '未允许使用摄像头，请在浏览器中允许相机权限后重试。';
    } else if (e && e.name === 'NotFoundError') {
      scannerError.value = '未检测到摄像头设备。';
    } else {
      scannerError.value = '摄像头启动失败：' + String(e) + '\n请确保在 HTTPS 或 localhost 下使用，或使用下方手动输入条码。';
    }
    active.value = false;
  }
}

function stopScanner() {
  try {
    codeReader.reset();
  } catch (e) {}

  // 尝试停止 video 元素的媒体流
  try {
    const v = videoRef.value as HTMLVideoElement | null;
    const stream = v?.srcObject as MediaStream | null;
    if (stream) {
      stream.getTracks().forEach((t) => t.stop());
      v!.srcObject = null;
    }
  } catch (e) {
    console.warn('stop video tracks', e);
  }

  active.value = false;
}

async function confirmProcess() {
  if (!scannedProduct.value || !scannedProduct.value.id) {
    window.alert('未识别到对应商品，请先创建或检查条码，或使用下方手动输入');
    return;
  }
  try {
    await store.processInventoryTransaction(scannedProduct.value.id, actionType.value, Number(qty.value), scanResult.value ?? undefined, '扫码操作');
    window.alert('操作成功');
    scanResult.value = null;
    scannedProduct.value = null;
  } catch (e: any) {
    window.alert('操作失败：' + (e.message ?? e));
  }
}

async function handleScan(code: string) {
  const now = Date.now();
  if (lastScanAt[code] && now - lastScanAt[code] < 1000) return; // 去抖
  lastScanAt[code] = now;

  // 如果已在队列中，增加数量
  const exist = scannedItems.value.find((it) => it.barcode === code);
  if (exist) {
    exist.quantity += 1;
    return;
  }

  const p = await store.findProductByBarcode(code);
  if (!p || !p.id) {
    // 未识别的条码，提示并跳过
    console.warn('未识别条码：', code);
    window.alert(`未识别条码 ${code}，请先在商品中心添加该 SKU`);
    return;
  }

  scannedItems.value.push({ productId: p.id, barcode: code, name: p.name, sku: p.sku, quantity: 1 });
}

function removeScanned(index: number) {
  scannedItems.value.splice(index, 1);
}

function changeQuantity(index: number, delta: number) {
  const it = scannedItems.value[index];
  if (!it) return;
  it.quantity = Math.max(1, it.quantity + delta);
}

async function generateOrderFromScans() {
  if (!scannedItems.value.length) {
    window.alert('没有扫码项目');
    return;
  }

  try {
    const items = scannedItems.value.map((it) => ({ productId: it.productId!, quantity: it.quantity, barcode: it.barcode }));
    const orderNo = await store.createOrderFromScans(items, actionType.value, '扫码批量');
    window.alert(`订单已生成：${orderNo}`);
    scannedItems.value = [];
  } catch (e: any) {
    window.alert('生成订单失败：' + (e.message ?? e));
  }
}

async function confirmManualBarcode() {
  if (!manualBarcode.value) {
    window.alert('请输入条码');
    return;
  }

  const p = await store.findProductByBarcode(manualBarcode.value);
  if (!p || !p.id) {
    window.alert('未找到对应商品，请先在商品中心添加或检查条码');
    return;
  }

  try {
    await store.processInventoryTransaction(p.id, actionType.value, Number(qty.value), manualBarcode.value, '手动操作');
    window.alert('操作成功');
    manualBarcode.value = '';
  } catch (e: any) {
    window.alert('操作失败：' + (e.message ?? e));
  }
}

onMounted(() => {
  startScanner();
});

onUnmounted(() => {
  stopScanner();
});
</script>

<template>
  <div class="page-section">
    <el-card>
      <div>
        <video ref="videoRef" autoplay playsinline muted style="width:100%;max-height:60vh;background:#000"></video>
      </div>

      <div style="margin-top:12px;display:flex;gap:12px;align-items:center;flex-wrap:wrap;">
        <div>扫描结果：{{ scanResult ?? '-' }}</div>
        <el-select v-model="actionType" placeholder="操作类型" style="width:120px">
          <el-option label="入库" value="in" />
          <el-option label="出库" value="out" />
        </el-select>
        <el-input-number v-model="qty" :min="1" :max="9999" />
        <el-button type="primary" @click="confirmProcess" :disabled="!scannedProduct">确认</el-button>
        <el-button @click="triggerFileInput">上传图片</el-button>
        <input ref="fileInputRef" type="file" accept="image/*" style="display:none" @change="onFileSelected" multiple />
        <el-button @click="stopScanner">停止</el-button>
        <el-button @click="startScanner" v-if="!active">开始</el-button>
      </div>

      <div v-if="scannerError" style="margin-top:12px;color:#d9534f;">
        <strong>提示：</strong> {{ scannerError }}
        <div style="margin-top:6px;font-size:12px;color:#666">提示：扫码通常需要在 HTTPS 或 localhost 下使用，并允许浏览器访问相机权限。</div>
      </div>

      <div v-if="scannedProduct" style="margin-top:12px">
        <strong>商品：</strong> {{ scannedProduct.name }} ({{ scannedProduct.sku }}) 当前库存：{{ scannedProduct.stock }}
      </div>

      <div style="margin-top:12px">
        <div style="margin-bottom:8px">已扫码队列：</div>
        <div v-if="scannedItems.length">
          <el-table :data="scannedItems" style="width:100%" size="small">
            <el-table-column prop="sku" label="SKU" width="160">
              <template #default="{ row }">{{ row.sku }}</template>
            </el-table-column>
            <el-table-column prop="name" label="名称">
              <template #default="{ row }">{{ row.name }}</template>
            </el-table-column>
            <el-table-column label="数量" width="160">
              <template #default="{ row, $index }">
                <el-button size="mini" @click="changeQuantity($index, -1)">-</el-button>
                <span style="margin:0 8px">{{ row.quantity }}</span>
                <el-button size="mini" @click="changeQuantity($index, 1)">+</el-button>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ $index }">
                <el-button size="mini" type="danger" @click="removeScanned($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div style="margin-top:12px;display:flex;gap:8px;align-items:center;">
            <el-button type="primary" @click="generateOrderFromScans">生成订单</el-button>
            <el-button @click="scannedItems = []">清空队列</el-button>
            <div style="margin-left:auto;color:#666">SKU 数：{{ scannedItems.length }} &nbsp; 总数：{{ scannedItems.reduce((s, it) => s + it.quantity, 0) }}</div>
          </div>
        </div>
        <div v-else style="color:#999">队列为空：开始扫码会自动将商品加入队列。</div>
      </div>

      <el-divider />

      <div style="margin-top:12px">
        <div style="margin-bottom:8px">如果摄像头不可用，请手动输入条码进行入/出库：</div>
        <el-input v-model="manualBarcode" placeholder="输入条码或 SKU" style="max-width:320px;margin-bottom:8px" />
        <div style="display:flex;gap:8px;align-items:center">
          <el-select v-model="actionType" placeholder="操作类型" style="width:120px">
            <el-option label="入库" value="in" />
            <el-option label="出库" value="out" />
          </el-select>
          <el-input-number v-model="qty" :min="1" :max="9999" />
          <el-button type="primary" @click="confirmManualBarcode">手动执行</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page-section { padding: 12px; }
</style>
