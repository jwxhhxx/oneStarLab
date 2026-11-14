import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

import { api } from '@/api/client';

export const useAuthStore = defineStore('auth', () => {
  const authenticated = ref(false);
  const usingDefault = ref(false);
  const checked = ref(false);
  const checking = ref(false);

  const loginHint = computed(() =>
    usingDefault.value ? '当前仍在使用初始口令 123456，建议登录后在「设置」中修改。' : '',
  );

  async function check(): Promise<void> {
    if (checking.value) return;
    checking.value = true;
    try {
      const status = await api.auth.status();
      authenticated.value = status.authenticated;
      usingDefault.value = status.usingDefault;
    } catch {
      authenticated.value = false;
    } finally {
      checked.value = true;
      checking.value = false;
    }
  }

  async function login(passcode: string): Promise<void> {
    const result = await api.auth.login(passcode);
    authenticated.value = true;
    usingDefault.value = Boolean(result.usingDefault);
    checked.value = true;
  }

  async function logout(): Promise<void> {
    try {
      await api.auth.logout();
    } finally {
      authenticated.value = false;
    }
  }

  async function changePasscode(current: string, next: string): Promise<void> {
    await api.auth.changePasscode(current, next);
    usingDefault.value = false;
  }

  function markLoggedOut(): void {
    authenticated.value = false;
    checked.value = true;
  }

  return {
    authenticated,
    usingDefault,
    checked,
    checking,
    loginHint,
    check,
    login,
    logout,
    changePasscode,
    markLoggedOut,
  };
});
