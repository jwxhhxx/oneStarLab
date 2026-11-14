<script setup lang="ts">
import { Lock, Setting } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';

import { useAuthStore } from '@/stores/useAuthStore';

const authStore = useAuthStore();
const form = reactive({ current: '', next: '', confirm: '' });
const saving = ref(false);

onMounted(() => {
  void authStore.check();
});

async function submit() {
  const current = form.current.trim();
  const next = form.next.trim();
  const confirm = form.confirm.trim();

  if (!current) {
    ElMessage.warning('请输入当前口令');
    return;
  }
  if (next.length < 4 || next.length > 32) {
    ElMessage.warning('新口令长度需为 4-32 位');
    return;
  }
  if (next === current) {
    ElMessage.warning('新口令不能与当前口令相同');
    return;
  }
  if (next !== confirm) {
    ElMessage.warning('两次输入的新口令不一致');
    return;
  }

  saving.value = true;
  try {
    await authStore.changePasscode(current, next);
    ElMessage.success('口令已更新，下次登录请使用新口令');
    form.current = '';
    form.next = '';
    form.confirm = '';
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '修改失败');
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="page-stack">
    <section class="page-hero page-hero--violet">
      <div>
        <p class="hero-kicker">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </p>
        <h2 class="hero-title">账户安全</h2>
        <p class="hero-description">
          这个后台只有一道口令保护。修改后所有旧会话会失效，需要重新登录。
        </p>
        <div class="hero-badges">
          <span class="hero-badge">单口令登录</span>
          <span class="hero-badge" :class="{ 'hero-badge--danger': authStore.usingDefault }">
            {{ authStore.usingDefault ? '仍在使用初始口令' : '已自定义口令' }}
          </span>
        </div>
      </div>
    </section>

    <el-alert
      v-if="authStore.usingDefault"
      type="warning"
      show-icon
      :closable="false"
      title="当前口令是初始口令 123456"
      description="任何拿到地址的人都能用它登录。请立刻设置一个只有你知道的口令。"
    />

    <section class="section-card">
      <div class="section-heading">
        <div>
          <h3>修改访问口令</h3>
          <p>口令长度 4-32 位，支持中文、字母、数字和符号。</p>
        </div>
      </div>

      <el-form label-position="top" class="settings-form" @submit.prevent="submit">
        <div class="form-grid">
          <el-form-item label="当前口令">
            <el-input
              v-model="form.current"
              type="password"
              show-password
              :prefix-icon="Lock"
              placeholder="输入当前口令"
            />
          </el-form-item>
          <el-form-item label="新口令">
            <el-input
              v-model="form.next"
              type="password"
              show-password
              :prefix-icon="Lock"
              placeholder="4-32 位"
            />
          </el-form-item>
          <el-form-item label="确认新口令">
            <el-input
              v-model="form.confirm"
              type="password"
              show-password
              :prefix-icon="Lock"
              placeholder="再输入一次"
              @keyup.enter="submit"
            />
          </el-form-item>
        </div>

        <div class="dialog-actions">
          <el-button type="primary" :loading="saving" @click="submit">保存新口令</el-button>
        </div>
      </el-form>
    </section>

    <section class="section-card settings-tips">
      <h3>安全建议</h3>
      <ul>
        <li>不要把口令告诉别人，也不要在公开场合输入。</li>
        <li>通过 HTTPS 访问后台，避免口令在明文连接中传输。</li>
        <li>换设备或怀疑泄露时，立即修改口令，旧会话会全部失效。</li>
        <li>定期在「数据中心」导出备份，数据库文件损坏时可以恢复。</li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.section-heading h3 {
  margin: 0 0 6px;
  font-size: 18px;
  color: #2f2a3d;
}

.section-heading p {
  margin: 0;
  color: #7b748a;
  font-size: 13px;
}

.settings-form {
  max-width: 720px;
}

.settings-tips ul {
  margin: 12px 0 0;
  padding-left: 18px;
  color: #6f6880;
  font-size: 13px;
  line-height: 1.9;
}

@media (max-width: 768px) {
  .section-heading {
    flex-direction: column;
  }
}
</style>
