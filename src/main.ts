import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';

import App from './App.vue';
import router from './router';
import { useAuthStore } from './stores/useAuthStore';
import './assets/main.css';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(ElementPlus);
app.mount('#app');

window.addEventListener('onestarlab:unauthorized', () => {
  const auth = useAuthStore(pinia);
  auth.markLoggedOut();
  const current = router.currentRoute.value;
  if (current.name !== 'login') {
    void router.replace({
      name: 'login',
      query: current.fullPath === '/' ? {} : { redirect: current.fullPath },
    });
  }
});
