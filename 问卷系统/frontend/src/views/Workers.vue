<template>
  <div>
    <div class="toolbar">
      <el-input v-model="kw" placeholder="按 姓名/拼音/首字母 搜索，如 zhangsan / zs" clearable :prefix-icon="Search"
        style="width: 320px" />
      <el-button type="primary" :icon="Plus" @click="openCreate">新建人员</el-button>
    </div>

    <el-table :data="filtered" border stripe>
      <el-table-column label="姓名" width="120">
        <template #default="{ row }"><b>{{ row.name }}</b></template>
      </el-table-column>
      <el-table-column prop="pinyinFull" label="拼音全拼" min-width="140" />
      <el-table-column prop="pinyinInitial" label="首字母" width="90" />
      <el-table-column prop="jobType" label="工种" min-width="120" />
      <el-table-column prop="phone" label="电话" min-width="130" />
      <el-table-column prop="note" label="备注" min-width="120" show-overflow-tooltip />
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" :icon="Delete" @click="del(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showForm" :title="form.id ? '编辑人员' : '新建人员'" width="520px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="姓名" required><el-input v-model="form.name" placeholder="保存后自动生成拼音" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="工种"><el-input v-model="form.jobType" placeholder="如：电工/叉车司机/木工" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.note" /></el-form-item>
        <el-alert v-if="form.id" :title="'拼音：' + form.pinyinFull + '（' + form.pinyinInitial + '）'" type="info" :closable="false" />
      </el-form>
      <template #footer>
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus, Search, EditPen, Delete } from '@element-plus/icons-vue'
import api from '../api'

const workers = ref([])
const kw = ref('')
const showForm = ref(false)
const saving = ref(false)
const emptyForm = { id: 0, name: '', phone: '', jobType: '', note: '', pinyinFull: '', pinyinInitial: '' }
const form = reactive({ ...emptyForm })

const filtered = computed(() => {
  const q = kw.value.trim().toLowerCase()
  if (!q) return workers.value
  return workers.value.filter(w =>
    w.name.includes(kw.value.trim()) ||
    (w.pinyinFull || '').includes(q) ||
    (w.pinyinInitial || '').includes(q))
})

async function load() { workers.value = await api.get('/workers') }
function openCreate() { Object.assign(form, emptyForm); showForm.value = true }
function openEdit(w) { Object.assign(form, JSON.parse(JSON.stringify(w))); showForm.value = true }

async function save() {
  if (!form.name.trim()) { ElMessage.warning('姓名必填'); return }
  saving.value = true
  try {
    if (form.id) await api.put('/workers/' + form.id, form)
    else await api.post('/workers', { ...form, active: true })
    showForm.value = false
    ElMessage.success('已保存')
    load()
  } finally { saving.value = false }
}

async function del(w) {
  try {
    await ElMessageBox.confirm(`确定删除「${w.name}」？`, '删除人员', { type: 'warning' })
    await api.delete('/workers/' + w.id)
    ElMessage.success('已删除')
    load()
  } catch (e) { /* 取消 */ }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 12px; align-items: center; margin-bottom: 14px; }
</style>
