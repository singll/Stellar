// 在开发环境中启动 mock 服务
// if (import.meta.env.DEV) {
//   try {
//     const { startMockServer } = await import('$lib/mocks/browser')
//     await startMockServer()
//   } catch (error) {
//     console.warn('MSW initialization failed:', error)
//   }
// }

export const prerender = true;
export const ssr = false;
