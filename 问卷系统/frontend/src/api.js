import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({ baseURL: '/api', timeout: 60000 })

api.interceptors.request.use((cfg) => {
  const t = localStorage.getItem('token')
  if (t) cfg.headers.Authorization = 'Bearer ' + t
  return cfg
})

api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const status = err.response?.status
    const msg = err.response?.data?.error || err.message || '请求失败'
    if (status === 401 && location.hash !== '#/login') {
      localStorage.removeItem('token')
      location.hash = '#/login'
    }
    ElMessage.error(msg)
    return Promise.reject(err)
  }
)

// 带鉴权的文件下载（导出 xlsx）
export async function download(path, filename) {
  const t = localStorage.getItem('token')
  const res = await fetch(path, { headers: { Authorization: 'Bearer ' + t } })
  if (!res.ok) {
    let msg = '导出失败'
    try { msg = (await res.json()).error || msg } catch (e) { /* ignore */ }
    ElMessage.error(msg)
    return
  }
  const blob = await res.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = decodeURIComponent(filename || '导出.xlsx')
  a.click()
  URL.revokeObjectURL(a.href)
}

export default api
