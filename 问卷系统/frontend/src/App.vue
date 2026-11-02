<template>
  <router-view v-if="isLogin" />
  <el-container v-else class="layout" style="height: 100vh">
    <el-aside width="220px" class="aside">
      <div class="logo">
        <el-icon size="22"><OfficeBuilding /></el-icon>
        <span>力能用工日报</span>
      </div>
      <el-menu :default-active="activeMenu" router class="menu" :default-openeds="['root']">
        <el-menu-item index="/"><el-icon><Files /></el-icon>项目管理</el-menu-item>
        <el-menu-item index="/workers"><el-icon><User /></el-icon>人员管理</el-menu-item>
        <el-menu-item index="/stats"><el-icon><Histogram /></el-icon>数据统计</el-menu-item>
        <el-menu-item index="/export"><el-icon><Download /></el-icon>导出</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span class="title">{{ pageTitle }}</span>
        <div class="user">
          <el-icon><UserFilled /></el-icon>
          <span>{{ username }}</span>
          <el-button text type="primary" @click="logout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isLogin = computed(() => route.path === '/login')
const username = computed(() => localStorage.getItem('username') || '')
const activeMenu = computed(() => route.path.startsWith('/report') ? '/' : route.path)
const pageTitle = computed(() => route.meta.title || '')

function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  router.replace('/login')
}
</script>

<style>
body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", sans-serif; }
.layout .aside { background: #1f2d3d; }
.layout .logo {
  height: 60px; display: flex; align-items: center; gap: 8px; padding: 0 18px;
  color: #fff; font-weight: 700; font-size: 16px; border-bottom: 1px solid #2c3e50;
}
.layout .menu { border-right: none; background: transparent; }
.layout .menu .el-menu-item { color: #c0c4cc; }
.layout .menu .el-menu-item.is-active { background: #3b6cff; color: #fff; }
.layout .header {
  display: flex; align-items: center; justify-content: space-between;
  background: #fff; border-bottom: 1px solid #ebeef5;
}
.layout .header .title { font-size: 17px; font-weight: 600; }
.layout .user { display: flex; align-items: center; gap: 8px; color: #606266; }
.layout .main { background: #f5f7fa; padding: 18px; }
</style>
