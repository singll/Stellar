import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

// 创建 worker 实例
export const worker = setupWorker(...handlers)

// 导出启动函数
export async function startMockServer() {
  if (import.meta.env.DEV) {
    return worker.start({
      onUnhandledRequest: 'bypass',
      serviceWorker: {
        url: '/mockServiceWorker.js',
        options: {
          scope: '/'
        }
      }
    })
  }
} 