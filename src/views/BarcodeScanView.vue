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

async function startScanner() {
  try {
    const devices = await codeReader.listVideoInputDevices();
    const deviceId: string | null = devices.length ? devices[0].deviceId : null;
    active.value = true;
    codeReader.decodeFromVideoDevice(deviceId, videoRef.value!, (result, err) => {
      if (result) {
        scanResult.value = result.getText();
        (async () => {
          const p = await store.findProductByBarcode(scanResult.value!);
          scannedProduct.value = p ?? null;
        })();
      }
    });
  } catch (e) {
    console.error(e);
    window.alert('摄像头启动失败：' + String(e));
  }
}

function stopScanner() {
  try {
    codeReader.reset();
  } catch (e) {}
  active.value = false;
}

async function confirmProcess() {
  if (!scannedProduct.value || !scannedProduct.value.id) {
    window.alert('未识别到对应商品，请先创建或检查条码');
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

      <div v-if="scannedProduct" style="margin-top:12px">
        <strong>商品：</strong> {{ scannedProduct.name }} ({{ scannedProduct.sku }}) 当前库存：{{ scannedProduct.stock }}
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page-section { padding: 12px; }
</style>
