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

async function startScanner() {
  scannerError.value = null;
  try {
    active.value = true;

    // 优先使用后置摄像头约束，兼容移动设备
    const constraints: MediaStreamConstraints = { video: { facingMode: { ideal: 'environment' } } };

    codeReader.decodeFromConstraints(constraints, videoRef.value!, (result, err) => {
      if (err && !(err instanceof Error && (err as any).name === 'NotFoundError')) {
        // 非找不到设备的错误时记录日志
        console.warn('decode error', err);
      }

      if (result) {
        scanResult.value = result.getText();
        (async () => {
          const p = await store.findProductByBarcode(scanResult.value!);
          scannedProduct.value = p ?? null;
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
