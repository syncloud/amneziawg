import { ElMessage } from 'element-plus'

const opts = { showClose: true, duration: 3500 }

export const notify = {
  success: (msg: string) => ElMessage({ type: 'success', message: msg, ...opts }),
  error: (msg: string) => ElMessage({ type: 'error', message: msg, ...opts }),
  info: (msg: string) => ElMessage({ type: 'info', message: msg, ...opts }),
  warning: (msg: string) => ElMessage({ type: 'warning', message: msg, ...opts }),
}
