<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import JsBarcode from 'jsbarcode';
import { useShopStore } from '@/stores/useShopStore';

const store = useShopStore();
onMounted(() => {
  store.initialize();
});

const products = computed(() => store.products);
const selectedProductId = ref<number | null>(null);
const barcodeValue = ref('');
const svgRef = ref<SVGSVGElement | null>(null);

function generate() {
  const value = barcodeValue.value || products.value.find((p) => p.id === selectedProductId.value)?.sku || '';
  if (!value) {
    window.alert('请输入条码或先选择商品');
    return;
  }

  const svgEl = svgRef.value as any;
  try {
    JsBarcode(svgEl, value, {
      format: 'CODE128',
      displayValue: true,
      fontSize: 14,
      height: 60,
      width: 2,
    });
  } catch (e) {
    console.error(e);
    window.alert('生成失败：' + String(e));
  }
}

function downloadPNG() {
  if (!svgRef.value) return;

  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(svgRef.value);
  const blob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const img = new Image();
  img.onload = () => {
    const canvas = document.createElement('canvas');
    canvas.width = img.width;
    canvas.height = img.height;
    const ctx = canvas.getContext('2d');
    if (!ctx) {
      URL.revokeObjectURL(url);
      return;
    }

    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(img, 0, 0);

    canvas.toBlob(async (outBlob) => {
      if (!outBlob) {
        URL.revokeObjectURL(url);
        window.alert('导出失败');
        return;
      }

      const fileName = `barcode-${barcodeValue.value || selectedProductId.value || 'label'}-${new Date()
        .toISOString()
        .replace(/[:.]/g, '-')}.png`;

      // 尝试使用 Web Share API（支持直接分享到手机并保存到相册）
      const nav: any = navigator;
      try {
        const file = new File([outBlob], fileName, { type: 'image/png' });
        if (nav.canShare && nav.canShare({ files: [file] }) && nav.share) {
          try {
            await nav.share({ files: [file], title: fileName });
            URL.revokeObjectURL(url);
            return;
          } catch (e) {
            console.warn('share failed, fallback', e);
          }
        }
      } catch (e) {
        // ignore file/share errors and fallback
        console.warn('share api error', e);
      }

      // 回退：在新窗口打开图片（iOS Safari 支持长按保存），
      // 若浏览器允许，再尝试使用带 download 的 a 标签
      const blobUrl = URL.createObjectURL(outBlob);
      const win = window.open(blobUrl, '_blank');
      if (!win) {
        const a = document.createElement('a');
        a.href = blobUrl;
        a.download = fileName;
        document.body.appendChild(a);
        a.click();
        a.remove();
      }

      // 延迟释放，确保新窗口/下载有时间使用 URL
      setTimeout(() => URL.revokeObjectURL(blobUrl), 30000);
      URL.revokeObjectURL(url);
    }, 'image/png');
  };
  img.onerror = () => URL.revokeObjectURL(url);
  img.src = url;
}

function printLabel() {
  if (!svgRef.value) return;
  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(svgRef.value);
  const html = `
  <html>
    <head>
      <title>打印标签</title>
      <style>
        @page { size: 38mm 25mm; margin: 0; }
        body { margin: 0; display:flex; align-items:center; justify-content:center; height:100vh; }
        .label { width:38mm; height:25mm; display:flex; align-items:center; justify-content:center; }
      </style>
    </head>
    <body>
      <div class="label">${svgString}</div>
      <script>window.onload = () => { window.print(); };<\/script>
    </body>
  </html>`;

  const w = window.open('', '_blank');
  if (w) {
    w.document.write(html);
    w.document.close();
  } else {
    window.alert('浏览器阻止了弹窗，请允许弹窗以打印标签');
  }
}
</script>

<template>
  <div class="page-section">
    <el-card>
      <div style="display:flex;gap:12px;align-items:center;flex-wrap:wrap;">
        <el-select v-model="selectedProductId" placeholder="选择商品" style="min-width:220px">
          <el-option v-for="p in products" :key="p.id" :label="p.name + ' (' + p.sku + ')'" :value="p.id" />
        </el-select>

        <el-input v-model="barcodeValue" placeholder="手动输入条码（默认为商品 SKU）" style="min-width:260px" />

        <el-button type="primary" @click="generate">生成条形码</el-button>
        <el-button @click="downloadPNG">下载 PNG</el-button>
        <el-button type="success" @click="printLabel">打印标签</el-button>
      </div>

      <div style="margin-top:16px">
        <svg ref="svgRef"></svg>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page-section { padding: 12px; }
</style>
