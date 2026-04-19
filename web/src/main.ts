import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'

// Dev-only: start miragejs to mock all /api routes so the UI works
// in the browser without a running Go backend.
// Run with `npm run dev:stub` — sets VITE_STUB=1.
if (import.meta.env.VITE_STUB) {
  const { makeServer } = await import('./mirage/server')
  makeServer()
  // eslint-disable-next-line no-console
  console.info('[mirage] stub server active')
}

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.mount('#app')
