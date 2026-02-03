import request from './request'

// 认证API
export const authAPI = {
  login: (data) => request.post('/auth/login', data),
  logout: () => request.post('/auth/logout')
}

// 主机API
export const hostAPI = {
  list: () => request.get('/hosts'),
  get: (id) => request.get(`/hosts/${id}`),
  create: (data) => request.post('/hosts', data),
  update: (id, data) => request.put(`/hosts/${id}`, data),
  delete: (id) => request.delete(`/hosts/${id}`),
  test: (id) => request.post(`/hosts/${id}/test`)
}

// 任务API
export const taskAPI = {
  list: () => request.get('/tasks'),
  get: (id) => request.get(`/tasks/${id}`),
  create: (data) => request.post('/tasks', data),
  update: (id, data) => request.put(`/tasks/${id}`, data),
  delete: (id) => request.delete(`/tasks/${id}`),
  run: (id) => request.post(`/tasks/${id}/run`),
  logs: (id, params) => request.get(`/tasks/${id}/logs`, { params }),
  deleteBackup: (logId) => request.delete(`/backups/${logId}`)
}

// 存储API
export const storageAPI = {
  list: () => request.get('/storages'),
  get: (id) => request.get(`/storages/${id}`),
  create: (data) => request.post('/storages', data),
  update: (id, data) => request.put(`/storages/${id}`, data),
  delete: (id) => request.delete(`/storages/${id}`),
  test: (id) => request.post(`/storages/${id}/test`),
  getDiskSpace: (id) => request.get(`/storages/${id}/diskspace`)
}

// 通知API
export const notificationAPI = {
  list: () => request.get('/notifications'),
  get: (id) => request.get(`/notifications/${id}`),
  create: (data) => request.post('/notifications', data),
  update: (id, data) => request.put(`/notifications/${id}`, data),
  delete: (id) => request.delete(`/notifications/${id}`),
  test: (id) => request.post(`/notifications/${id}/test`)
}

// 日志API
export const logAPI = {
  list: (params) => request.get('/logs', { params })
}

// 备份API
export const backupAPI = {
  delete: (id) => request.delete(`/backups/${id}`),
  download: (id) => request.get(`/backups/${id}/download`, { responseType: 'blob' })
}

// 用户API
export const userAPI = {
  list: () => request.get('/users'),
  get: (id) => request.get(`/users/${id}`),
  create: (data) => request.post('/users', data),
  update: (id, data) => request.put(`/users/${id}`, data),
  delete: (id) => request.delete(`/users/${id}`)
}

// 仪表盘API
export const dashboardAPI = {
  stats: () => request.get('/dashboard/stats')
}
