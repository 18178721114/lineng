import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import api from './api'
import Login from './views/Login.vue'
import Projects from './views/Projects.vue'
import Report from './views/Report.vue'
import Workers from './views/Workers.vue'
import Stats from './views/Stats.vue'
import ExportPage from './views/Export.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login', component: Login, meta: { title: '登录' } },
    { path: '/', component: Projects, meta: { title: '项目管理' } },
    { path: '/report/:id', component: Report, meta: { title: '日报填报' } },
    { path: '/workers', component: Workers, meta: { title: '人员管理' } },
    { path: '/stats', component: Stats, meta: { title: '数据统计' } },
    { path: '/export', component: ExportPage, meta: { title: '导出' } }
  ]
})

router.beforeEach((to) => {
  if (to.path !== '/login' && !localStorage.getItem('token')) return '/login'
})

const app = createApp(App)
for (const [key, comp] of Object.entries(ElementPlusIconsVue)) app.component(key, comp)
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.config.globalProperties.$api = api
app.mount('#app')
