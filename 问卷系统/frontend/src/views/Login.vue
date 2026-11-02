<template>
  <div class="login-wrap">
    <el-card class="login-card" shadow="always">
      <div class="brand">
        <el-icon size="28" color="#3b6cff"><OfficeBuilding /></el-icon>
        <h2>力能用工日报管理系统</h2>
      </div>
      <p class="sub">北京力能艺术文化服务有限公司</p>
      <el-form :model="form" @submit.prevent="onSubmit" label-width="64px">
        <el-form-item label="账号" required>
          <el-input v-model="form.username" placeholder="请输入账号" clearable :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="密码" required>
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password :prefix-icon="Lock" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%" size="large">
          登 录
        </el-button>
      </el-form>
      <div class="tip">默认管理员账号：admin / admin123</div>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import api from '../api'

const router = useRouter()
const form = reactive({ username: '', password: '' })
const loading = ref(false)

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    const res = await api.post('/login', form)
    localStorage.setItem('token', res.token)
    localStorage.setItem('username', res.username)
    router.replace('/')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #3b6cff 0%, #6f9bff 100%);
}
.login-card { width: 380px; border-radius: 14px; padding: 14px 24px 28px; }
.brand { display: flex; align-items: center; gap: 10px; justify-content: center; margin-bottom: 4px; }
.brand h2 { margin: 0; font-size: 19px; }
.sub { text-align: center; color: #909399; margin: 0 0 18px; font-size: 13px; }
.tip { text-align: center; color: #c0c4cc; font-size: 12px; margin-top: 14px; }
</style>
