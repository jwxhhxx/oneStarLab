<script setup lang="ts">
import { Lock } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useAuthStore } from '@/stores/useAuthStore';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const passcode = ref('');
const submitting = ref(false);

onMounted(() => {
  document.title = '登录 · OneStarLab';
});

async function submit() {
  const value = passcode.value.trim();
  if (!value) {
    ElMessage.warning('请输入口令');
    return;
  }
  submitting.value = true;
  try {
    await authStore.login(value);
    ElMessage.success('登录成功');
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/';
    await router.replace(redirect.startsWith('/') ? redirect : '/');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '登录失败');
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-orbit login-orbit--one" />
    <div class="login-orbit login-orbit--two" />

    <div class="login-card">
      <div class="brand-chip">OneStarLab</div>
      <h1>手帐店铺助手</h1>
      <p class="login-subtitle">输入访问口令后继续，数据保存在你部署的服务器上。</p>

      <el-input
        v-model="passcode"
        class="login-input"
        type="password"
        size="large"
        show-password
        placeholder="访问口令"
        :prefix-icon="Lock"
        autofocus
        @keyup.enter="submit"
      />

      <el-button
        type="primary"
        size="large"
        class="login-button"
        :loading="submitting"
        @click="submit"
      >
        进入后台
      </el-button>

      <p class="login-hint">初始口令为 123456，登录后请到「设置」中修改。</p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  position: relative;
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  overflow: hidden;
  background:
    radial-gradient(circle at 18% 18%, rgba(140, 124, 240, 0.28), transparent 46%),
    radial-gradient(circle at 82% 72%, rgba(245, 176, 132, 0.3), transparent 48%),
    linear-gradient(160deg, #fbf8f4 0%, #f4f1fb 52%, #eef4f8 100%);
}

.login-orbit {
  position: absolute;
  border-radius: 50%;
  border: 1px dashed rgba(124, 108, 240, 0.28);
  pointer-events: none;
}

.login-orbit--one {
  width: 420px;
  height: 420px;
  top: -160px;
  right: -100px;
}

.login-orbit--two {
  width: 280px;
  height: 280px;
  bottom: -120px;
  left: -80px;
  border-color: rgba(240, 168, 120, 0.32);
}

.login-card {
  position: relative;
  z-index: 1;
  width: min(420px, 100%);
  padding: 40px 36px 32px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.86);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 24px 60px rgba(49, 41, 75, 0.14);
  backdrop-filter: blur(14px);
  text-align: center;
}

.brand-chip {
  display: inline-flex;
  padding: 6px 14px;
  border-radius: 999px;
  background: linear-gradient(120deg, #7c6cf0, #a58cf2);
  color: #fff;
  font-size: 12px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.login-card h1 {
  margin: 20px 0 8px;
  font-size: 28px;
  color: #2f2a3d;
}

.login-subtitle {
  margin: 0 0 28px;
  color: #7b748a;
  font-size: 14px;
  line-height: 1.7;
}

.login-input :deep(.el-input__wrapper) {
  border-radius: 14px;
  padding: 6px 14px;
}

.login-button {
  width: 100%;
  margin-top: 18px;
  border-radius: 14px;
  font-weight: 600;
}

.login-hint {
  margin: 18px 0 0;
  color: #9a92aa;
  font-size: 12px;
  line-height: 1.6;
}
</style>
