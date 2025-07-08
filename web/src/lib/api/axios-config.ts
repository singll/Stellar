import axios from 'axios';
import { get } from 'svelte/store';
import { auth } from '$lib/stores/auth';
import { notifications } from '$lib/stores/notifications';

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const authState = get(auth);
    if (authState.token) {
      config.headers.Authorization = `Bearer ${authState.token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // 统一错误处理
    if (error.response) {
      const { status, data } = error.response;
      
             switch (status) {
         case 401:
           // 未授权，清除认证信息
           auth.logout();
           notifications.add({
             type: 'error',
             message: '登录已过期，请重新登录',
           });
           break;
         case 403:
           notifications.add({
             type: 'error',
             message: '权限不足',
           });
           break;
         case 404:
           notifications.add({
             type: 'error',
             message: '请求的资源不存在',
           });
           break;
         case 500:
           notifications.add({
             type: 'error',
             message: '服务器内部错误',
           });
           break;
         default:
           notifications.add({
             type: 'error',
             message: data?.message || '网络请求失败',
           });
       }
     } else if (error.request) {
       // 网络错误
       notifications.add({
         type: 'error',
         message: '网络连接失败，请检查网络设置',
       });
     } else {
       // 其他错误
       notifications.add({
         type: 'error',
         message: error.message || '请求处理失败',
       });
     }
    
    return Promise.reject(error);
  }
);

export default api; 