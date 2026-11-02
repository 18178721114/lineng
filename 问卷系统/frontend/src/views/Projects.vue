<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openCreate">新建项目</el-button>
      <el-button type="warning" :icon="Share" @click="openShare">分享</el-button>
    </div>

    <el-table :data="projects" v-loading="loading" border stripe>
      <el-table-column prop="no" label="编号" width="150" />
      <el-table-column label="用工单位 / 活动" min-width="200">
        <template #default="{ row }">
          <b>{{ row.unit }}</b>
          <div class="muted">{{ row.activityName }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="location" label="地点" min-width="120" />
      <el-table-column label="甲方负责人" min-width="150">
        <template #default="{ row }">{{ row.partyAOwner }} <span class="muted">{{ row.partyAPhone }}</span></template>
      </el-table-column>
      <el-table-column label="起止日期" width="210">
        <template #default="{ row }">{{ row.startDate }} ~ {{ row.endDate }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="statusOf(row).type">{{ statusOf(row).text }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="reportCount" label="日报数" width="90" align="center" />
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(row)">编辑</el-button>
          <el-button link type="success" :icon="Document" @click="$router.push('/report/' + row.id)">填报</el-button>
          <el-button link type="danger" :icon="Delete" @click="del(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showForm" :title="form.id ? '编辑项目' : '新建项目'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="用工单位" required><el-input v-model="form.unit" placeholder="如：天津某某公司" /></el-form-item>
        <el-form-item label="活动名称" required><el-input v-model="form.activityName" placeholder="如：李荣浩演唱会" /></el-form-item>
        <el-form-item label="地点" required><el-input v-model="form.location" /></el-form-item>
        <el-form-item label="甲方负责人" required>
          <el-select v-model="form.partyAOwner" filterable remote allow-create clearable default-first-option
            :remote-method="searchWorkers" :loading="workersLoading" :reserve-keyword="false"
            placeholder="搜索人员（支持汉字 / 拼音 / 首字母），也可直接输入"
            style="width: 100%" @change="onOwnerChange">
            <el-option v-for="w in workers" :key="w.id" :value="w.name"
              :label="w.name + (w.phone ? '（' + w.phone + '）' : '')" />
          </el-select>
        </el-form-item>
        <el-form-item label="甲方电话"><el-input v-model="form.partyAPhone" /></el-form-item>
        <el-form-item label="用工开始" required>
          <el-date-picker v-model="form.startDate" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" />
        </el-form-item>
        <el-form-item label="用工结束" required>
          <el-date-picker v-model="form.endDate" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分享 H5 填报链接（带防盗刷签名与有效期） -->
    <el-dialog v-model="showShare" title="分享 H5 填报链接" width="640px">
      <el-form label-width="110px">
        <el-form-item label="选择项目">
          <el-select v-model="shareProjectIds" multiple filterable placeholder="请选择要填报的项目（可多选）" style="width: 100%">
            <el-option v-for="p in projects" :key="p.id" :value="p.id" :label="p.no + ' ' + p.unit + '·' + p.activityName" />
          </el-select>
          <div class="muted" style="margin-top:4px">可选项目，手机端 H5 仅展示所选的项目。</div>
        </el-form-item>
        <el-form-item label="H5 页面地址">
          <el-input v-model="shareH5Url" placeholder="H5 部署地址，默认本站 /h5/index.html" />
          <div class="muted" style="margin-top:4px">手机需能访问该地址；本机/局域网默认不用改。</div>
        </el-form-item>
        <el-form-item label="有效期">
          <el-select v-model="shareHours" style="width: 100%">
            <el-option v-for="o in expireOptions" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
          <div class="muted" style="margin-top:4px">选「长期有效」生成的链接不会过期（签名仍校验）。</div>
        </el-form-item>
        <el-form-item v-if="shareLink" label="分享链接">
          <el-input :model-value="shareLink" readonly type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showShare = false">关闭</el-button>
        <el-button type="primary" :icon="Share" @click="genShareLink">生成并复制链接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus, EditPen, Document, Delete, Share } from '@element-plus/icons-vue'
import api from '../api'
import { buildShareLink, copyText, encodeIds } from '../share'

const projects = ref([])
const loading = ref(true)
const showForm = ref(false)
const saving = ref(false)
const emptyForm = { id: 0, unit: '', activityName: '', location: '', partyAOwner: '', partyAPhone: '', startDate: '', endDate: '' }
const form = reactive({ ...emptyForm })

// 人员管理列表：供「甲方负责人」下拉远程搜索（后端支持 汉字 / 拼音全拼 / 拼音首字母 / 手机号）
const workers = ref([])
const workersLoading = ref(false)
let searchTimer = null
let searchSeq = 0

async function loadWorkers(q = '') {
  const seq = ++searchSeq
  workersLoading.value = true
  try {
    const list = await api.get('/workers', q ? { params: { q } } : undefined)
    if (seq === searchSeq) workers.value = Array.isArray(list) ? list : []
  } catch (e) {
    if (seq === searchSeq) workers.value = []
  } finally {
    if (seq === searchSeq) workersLoading.value = false
  }
}
// 输入防抖：清空关键字时恢复完整列表
function searchWorkers(q) {
  clearTimeout(searchTimer)
  const kw = (q || '').trim()
  if (!kw) { loadWorkers(); return }
  searchTimer = setTimeout(() => loadWorkers(kw), 200)
}
function onOwnerChange(name) {
  const w = workers.value.find(x => x.name === name)
  if (w && w.phone) form.partyAPhone = w.phone
}

const statusOf = (p) => {
  const t = new Date().toISOString().slice(0, 10)
  if (p.endDate && t > p.endDate) return { text: '已结束', type: 'info' }
  if (p.startDate && t < p.startDate) return { text: '未开始', type: 'warning' }
  return { text: '进行中', type: 'success' }
}

async function load() {
  loading.value = true
  try { projects.value = await api.get('/projects') } finally { loading.value = false }
}

function openCreate() { Object.assign(form, emptyForm); loadWorkers(); showForm.value = true }
function openEdit(p) { Object.assign(form, JSON.parse(JSON.stringify(p))); loadWorkers(); showForm.value = true }

async function save() {
  if (!form.unit || !form.activityName || !form.location || !form.partyAOwner || !form.startDate || !form.endDate) {
    ElMessage.warning('带 * 项均为必填'); return
  }
  saving.value = true
  try {
    if (form.id) await api.put('/projects/' + form.id, form)
    else await api.post('/projects', form)
    showForm.value = false
    ElMessage.success('已保存')
    load()
  } finally { saving.value = false }
}

async function del(p) {
  try {
    await ElMessageBox.confirm(`确定删除「${p.unit}·${p.activityName}」？`, '删除项目', { type: 'warning' })
    await api.delete('/projects/' + p.id)
    ElMessage.success('已删除')
    load()
  } catch (e) { /* 取消 */ }
}

// ---------- 分享 H5 链接（防盗刷：HMAC 签名 + 有效期） ----------
const showShare = ref(false)
const shareProjectIds = ref([])
const shareH5Url = ref('')
const shareHours = ref(24)
const shareLink = ref('')
// 有效期下拉：value 为小时数，0 = 长期有效
const expireOptions = [
  { value: 12, label: '12 小时' },
  { value: 24, label: '1 天（24 小时）' },
  { value: 72, label: '3 天' },
  { value: 168, label: '7 天' },
  { value: 720, label: '30 天' },
  { value: 0, label: '长期有效' }
]

function defaultH5Url() {
  // 默认取本站同源下的 /h5/index.html（frontend/public/h5 由 vite 托管）
  const base = import.meta.env.BASE_URL || '/'
  return location.origin + (base.endsWith('/') ? base : base + '/') + 'h5/index.html'
}

function openShare() {
  shareProjectIds.value = projects.value.length ? [projects.value[0].id] : []
  shareH5Url.value = defaultH5Url()
  shareHours.value = 24
  shareLink.value = ''
  showShare.value = true
}

async function genShareLink() {
  if (!shareProjectIds.value.length) { ElMessage.warning('请先选择至少一个项目'); return }
  let h5Url = shareH5Url.value.trim()
  if (!h5Url) h5Url = defaultH5Url()
  try { new URL(h5Url) } catch (e) { ElMessage.error('H5 页面地址不是合法 URL'); return }
  const selected = projects.value.filter(p => shareProjectIds.value.includes(p.id))
  const ids = encodeIds(selected) // 项目 ID 列表参与签名，后端按 ID 返回真实项目
  const to = selected.map(p => p.no || (p.unit + '·' + p.activityName)).join('、')
  const hours = Number(shareHours.value) // 0 = 长期有效
  const { url, exp, forever } = buildShareLink(h5Url, hours, to, ids)
  shareLink.value = url
  const ok = await copyText(url)
  const tip = forever ? '长期有效（不会过期）' : '有效期至 ' + new Date(exp).toLocaleString('zh-CN')
  if (ok) ElMessage.success('链接已复制（' + selected.length + ' 个项目），' + tip)
  else ElMessage.info('复制失败，请手动复制下方链接')
}

onMounted(load)
</script>

<style scoped>
.toolbar { margin-bottom: 14px; }
.muted { color: #909399; font-size: 12px; }
</style>
