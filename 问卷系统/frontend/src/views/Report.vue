<template>
  <div>
    <el-page-header class="ph" @back="$router.back()">
      <template #content>
        <span class="ph-title">{{ project ? project.unit + ' · ' + project.activityName : '日报填报' }}</span>
      </template>
    </el-page-header>

    <el-card class="box" shadow="never">
      <el-form inline>
        <el-form-item label="日报日期" required>
          <el-date-picker v-model="reportDate" type="date" value-format="YYYY-MM-DD" @change="load" />
        </el-form-item>
        <el-tag v-if="report.updatedAt" type="success">已保存（{{ report.updatedAt }}）</el-tag>
        <el-tag v-else type="info">本日期暂无日报，填写后保存即新建</el-tag>
      </el-form>
    </el-card>

    <!-- ① 用工名单 -->
    <el-card class="box" shadow="never">
      <template #header><span class="sec">① 用工名单（先填工作时间，再选人）</span></template>
      <el-table :data="report.workers" border>
        <el-table-column type="index" width="50" label="#" />
        <el-table-column prop="name" label="姓名" width="160" />
        <el-table-column label="开始时间" width="180">
          <template #default="{ row }"><el-time-picker v-model="row.start" value-format="HH:mm" format="HH:mm" style="width:100%" /></template>
        </el-table-column>
        <el-table-column label="结束时间" width="180">
          <template #default="{ row }"><el-time-picker v-model="row.end" value-format="HH:mm" format="HH:mm" style="width:100%" /></template>
        </el-table-column>
        <el-table-column label="操作" align="center" min-width="120">
          <template #default="{ row, $index }">
            <el-button link type="danger" @click="report.workers.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button class="mt" type="primary" :icon="Plus" @click="openPicker">添加用工</el-button>
    </el-card>

    <!-- ② 用工人数 -->
    <el-card class="box" shadow="never">
      <template #header><span class="sec">② 用工人数</span></template>
      <el-row :gutter="16">
        <el-col :span="8" v-for="s in shiftDefs" :key="s.key">
          <div class="lbl">{{ s.label }}人数</div>
          <el-input-number v-model="report.shifts[s.key]" :min="0" style="width:100%" />
        </el-col>
      </el-row>
      <el-button class="mt" :icon="Tickets" @click="syncFromRoster">按名单带出（{{ report.workers.length }} 人）</el-button>
    </el-card>

    <!-- ③ 餐费 -->
    <el-card class="box" shadow="never">
      <template #header><span class="sec">③ 餐费（含水）</span></template>
      <el-row :gutter="14">
        <el-col :span="4.8" v-for="m in mealDefs" :key="m.key">
          <div class="lbl">{{ m.label }}</div>
          <el-input-number v-model="report.meals[m.key]" :min="0" style="width:100%" controls-position="right" />
        </el-col>
      </el-row>
    </el-card>

    <!-- ④ 设备台班 -->
    <el-card class="box" shadow="never">
      <template #header><span class="sec">④ 设备台班</span></template>
      <el-row :gutter="14">
        <el-col :span="12" v-for="eq in equipDefs" :key="eq.key" style="margin-bottom:14px">
          <el-card shadow="hover">
            <div class="equip-head"><b>{{ eq.label }}</b><span class="muted">合计 {{ equipTotal(eq.key) }} 台班</span></div>
            <div class="mt">
              <div class="lbl">台班数</div>
              <el-input-number v-model="report.equipment[eq.key].day" :min="0" style="width:100%" controls-position="right" />
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- ⑤ 外包 -->
    <el-card class="box" shadow="never">
      <template #header>
        <div class="hd">
          <span class="sec">⑤ 外包用工</span>
          <el-switch v-model="report.outsourced" />
        </div>
      </template>
      <template v-if="report.outsourced">
        <el-alert type="info" :closable="false" class="mb">每个外包单位按 白班 / 晚班 / 凌晨 三行填写：用工人数、加班人数、加班起止时间（无加班留空即可）。</el-alert>
        <el-table :data="outRows" border :span-method="outSpan">
          <el-table-column label="外包单位" min-width="170">
            <template #default="{ row }">
              <el-input v-if="row.si === 0" v-model="row.unit.name" placeholder="单位名称" />
            </template>
          </el-table-column>
          <el-table-column label="班次" width="86" align="center">
            <template #default="{ row }"><span class="shift-name">{{ row.label }}</span></template>
          </el-table-column>
          <el-table-column label="用工人数" width="120">
            <template #default="{ row }"><el-input-number v-model="row.unit[row.key]" :min="0" style="width:100%" /></template>
          </el-table-column>
          <el-table-column label="加班人数" width="120">
            <template #default="{ row }"><el-input-number v-model="row.unit[row.key + 'Ot']" :min="0" style="width:100%" /></template>
          </el-table-column>
          <el-table-column label="加班开始" width="140">
            <template #default="{ row }"><el-time-picker v-model="row.unit[row.key + 'OtStart']" value-format="HH:mm" format="HH:mm" style="width:100%" placeholder="无" /></template>
          </el-table-column>
          <el-table-column label="加班结束" width="140">
            <template #default="{ row }"><el-time-picker v-model="row.unit[row.key + 'OtEnd']" value-format="HH:mm" format="HH:mm" style="width:100%" placeholder="无" /></template>
          </el-table-column>
          <el-table-column label="操作" width="86">
            <template #default="{ row }">
              <el-button v-if="row.si === 0" link type="danger" @click="removeOutsource(row.ui)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-button class="mt" :icon="Plus" @click="addOutsource">添加外包单位</el-button>
      </template>
    </el-card>

    <!-- ⑥ 备注 -->
    <el-card class="box" shadow="never">
      <template #header><span class="sec">⑥ 备注</span></template>
      <el-input v-model="report.notes" type="textarea" :rows="3" placeholder="其他需要说明的情况" />
    </el-card>

    <div class="actions">
      <el-button v-if="report.id" type="danger" :icon="Delete" @click="del">删除本单</el-button>
      <el-button type="primary" :icon="Check" :loading="saving" @click="save">保存日报</el-button>
    </div>

    <!-- 选人弹窗 -->
    <el-dialog v-model="showPicker" title="添加用工" width="560px">
      <el-alert type="warning" :closable="false" class="mb">先选择本批工人的工作时间，再勾选人员，确认后整批加入名单。</el-alert>
      <el-form inline>
        <el-form-item label="开始"><el-time-picker v-model="pickStart" value-format="HH:mm" format="HH:mm" /></el-form-item>
        <el-form-item label="结束"><el-time-picker v-model="pickEnd" value-format="HH:mm" format="HH:mm" /></el-form-item>
      </el-form>
      <el-input v-model="kw" :prefix-icon="Search" placeholder="搜索 姓名/拼音/首字母" clearable class="mb" />
      <el-checkbox-group v-model="checked" class="picker-list">
        <el-checkbox v-for="w in filteredWorkers" :key="w.id" :value="w.name" border>
          {{ w.name }} <small class="muted">{{ w.jobType }} {{ w.phone }}</small>
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="showQuickAdd = true">新建人员</el-button>
        <el-button @click="showPicker = false">取消</el-button>
        <el-button type="primary" @click="confirmPick">加入名单（{{ checked.length }}）</el-button>
      </template>
    </el-dialog>

    <!-- 快捷新建人员 -->
    <el-dialog v-model="showQuickAdd" title="新建人员" width="420px" @closed="quick.name=''">
      <el-form label-width="60px">
        <el-form-item label="姓名" required><el-input v-model="quick.name" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="quick.phone" /></el-form-item>
        <el-form-item label="工种"><el-input v-model="quick.jobType" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuickAdd = false">取消</el-button>
        <el-button type="primary" :loading="qSaving" @click="quickAdd">确定创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Delete, Check, Tickets } from '@element-plus/icons-vue'
import api from '../api'

const route = useRoute()
const pid = Number(route.params.id)
const project = ref(null)
const reportDate = ref(new Date().toISOString().slice(0, 10))
const saving = ref(false)

const shiftDefs = [{ key: 'day', label: '白天' }, { key: 'eve', label: '晚上' }, { key: 'dawn', label: '凌晨' }]
const mealDefs = [{ key: 'breakfast', label: '早点' }, { key: 'lunch', label: '中餐' }, { key: 'dinner', label: '晚餐' }, { key: 'supper', label: '夜宵' }, { key: 'water', label: '水' }]
const equipDefs = [{ key: 'fork3', label: '🚜 叉车3吨' }, { key: 'fork5', label: '🚜 叉车5吨' }, { key: 'fork7', label: '🚜 叉车7吨' }, { key: 'crane', label: '🏗 吊车' }, { key: 'rescue', label: '🚑 救援车' }]

// 外包：一个单位展开成 白班/晚班/凌晨 三行，每行都有用工人数、加班人数、加班起止。
const outShiftDefs = [{ key: 'day', label: '白班' }, { key: 'eve', label: '晚班' }, { key: 'dawn', label: '凌晨' }]
const emptyOutsource = () => {
  const o = { name: '' }
  for (const s of outShiftDefs) {
    o[s.key] = ''
    o[s.key + 'Ot'] = ''
    o[s.key + 'OtStart'] = ''
    o[s.key + 'OtEnd'] = ''
  }
  return o
}
// 归一化后端返回的外包行，兼容旧数据（仅一组 otStart/otEnd 时归入白班）
function normalizeOutsourcing(list) {
  return (Array.isArray(list) ? list : []).map(u => {
    const o = emptyOutsource()
    o.name = u.name || ''
    for (const s of outShiftDefs) {
      o[s.key] = u[s.key] ?? ''
      o[s.key + 'Ot'] = u[s.key + 'Ot'] ?? ''
      o[s.key + 'OtStart'] = u[s.key + 'OtStart'] || ''
      o[s.key + 'OtEnd'] = u[s.key + 'OtEnd'] || ''
    }
    if (!o.dayOtStart && !o.dayOtEnd && (u.otStart || u.otEnd)) {
      o.dayOtStart = u.otStart || ''
      o.dayOtEnd = u.otEnd || ''
    }
    return o
  })
}
const emptyReport = () => ({
  id: 0, workers: [], shifts: { day: '', eve: '', dawn: '' },
  meals: { breakfast: '', lunch: '', dinner: '', supper: '', water: '' },
  equipment: { fork3: { day: '', eve: '', dawn: '' }, fork5: { day: '', eve: '', dawn: '' }, fork7: { day: '', eve: '', dawn: '' }, crane: { day: '', eve: '', dawn: '' }, rescue: { day: '', eve: '', dawn: '' } },
  outsourced: false, outsourcing: [], notes: '', updatedAt: ''
})
const report = reactive(emptyReport())

// 表格数据：每个外包单位展开成 白班/晚班/凌晨 三行
const outRows = computed(() => {
  const rows = []
  report.outsourcing.forEach((unit, ui) => {
    outShiftDefs.forEach((s, si) => rows.push({ ui, si, unit, key: s.key, label: s.label }))
  })
  return rows
})
// 单位列与操作列跨三行合并，其余列一行一格
function outSpan({ row, columnIndex }) {
  if (columnIndex === 0 || columnIndex === 6) {
    return row.si === 0 ? { rowspan: outShiftDefs.length, colspan: 1 } : { rowspan: 0, colspan: 0 }
  }
  return { rowspan: 1, colspan: 1 }
}

const equipTotal = (k) => {
  const e = report.equipment[k] || {}
  return Number(e.day || 0)
}

async function load() {
  const r = await api.get(`/report/${pid}/${reportDate.value}`)
  const base = emptyReport()
  const merged = { ...base, ...r }
  merged.shifts = { ...base.shifts, ...(r.shifts || {}) }
  merged.meals = { ...base.meals, ...(r.meals || {}) }
  merged.equipment = JSON.parse(JSON.stringify(base.equipment))
  for (const d of equipDefs) merged.equipment[d.key] = { ...base.equipment[d.key], ...((r.equipment || {})[d.key] || {}) }
  // 设备台班不再分班次：历史数据里晚班/凌晨的台班并入台班数
  for (const d of equipDefs) {
    const e = merged.equipment[d.key]
    e.day = Number(e.day || 0) + Number(e.eve || 0) + Number(e.dawn || 0)
    e.eve = 0
    e.dawn = 0
  }
  if (!Array.isArray(merged.workers)) merged.workers = []
  merged.outsourcing = normalizeOutsourcing(r.outsourcing)
  Object.assign(report, merged)
}

function syncFromRoster() {
  report.shifts.day = report.workers.length
  report.shifts.eve = ''
  report.shifts.dawn = ''
  ElMessage.info('已按名单带出白天人数')
}

// 选人弹窗
const showPicker = ref(false)
const kw = ref('')
const checked = ref([])
const pickStart = ref('08:00')
const pickEnd = ref('17:00')
const allWorkers = ref([])
const filteredWorkers = computed(() => {
  const q = kw.value.trim().toLowerCase()
  if (!q) return allWorkers.value
  return allWorkers.value.filter(w =>
    w.name.includes(kw.value.trim()) ||
    (w.pinyinFull || '').includes(q) ||
    (w.pinyinInitial || '').includes(q))
})
async function openPicker() {
  allWorkers.value = (await api.get('/workers')).filter(w => w.active)
  checked.value = []
  showPicker.value = true
}
function confirmPick() {
  if (!pickStart.value || !pickEnd.value) { ElMessage.warning('请先填写工作时间'); return }
  if (!checked.value.length) { ElMessage.warning('请先勾选工人'); return }
  const existed = new Set(report.workers.map(w => w.name))
  let added = 0
  for (const name of checked.value) {
    if (!existed.has(name)) {
      report.workers.push({ name, start: pickStart.value, end: pickEnd.value })
      added++
    }
  }
  showPicker.value = false
  ElMessage.success(`已加入 ${added} 人（${pickStart.value}~${pickEnd.value}）`)
}

// 快捷新建人员
const showQuickAdd = ref(false)
const quick = reactive({ name: '', phone: '', jobType: '' })
const qSaving = ref(false)
async function quickAdd() {
  if (!quick.name.trim()) { ElMessage.warning('姓名必填'); return }
  qSaving.value = true
  try {
    await api.post('/workers', { ...quick, active: true })
    allWorkers.value = (await api.get('/workers')).filter(w => w.active)
    if (!checked.value.includes(quick.name)) checked.value.push(quick.name)
    ElMessage.success('已创建并选中 ' + quick.name)
    showQuickAdd.value = false
  } finally { qSaving.value = false }
}

function addOutsource() {
  report.outsourcing.push(emptyOutsource())
}

function removeOutsource(ui) {
  report.outsourcing.splice(ui, 1)
}

async function save() {
  if (!report.workers.length && !report.shifts.day && !report.shifts.eve && !report.shifts.dawn) {
    ElMessage.warning('请先添加用工名单或填写用工人数'); return
  }
  for (let i = 0; i < report.workers.length; i++) {
    const w = report.workers[i]
    if (!w.name) { ElMessage.warning(`第 ${i + 1} 行未选人名`); return }
    if (!w.start || !w.end) { ElMessage.warning(`工人「${w.name}」的起止时间未填全`); return }
  }
  for (const [i, u] of report.outsourcing.entries()) {
    if (!u.name) { ElMessage.warning(`第 ${i + 1} 个外包单位未填名称`); return }
    for (const s of outShiftDefs) {
      const cnt = Number(u[s.key + 'Ot'] || 0)
      const st = u[s.key + 'OtStart']
      const en = u[s.key + 'OtEnd']
      if ((st || en) && !(st && en)) { ElMessage.warning(`外包单位「${u.name}」${s.label}的加班起止时间未填全`); return }
      if (cnt > 0 && !st) { ElMessage.warning(`外包单位「${u.name}」${s.label}填了加班人数，请补全加班起止时间`); return }
      const total = Number(u[s.key] || 0)
      if (total > 0 && cnt > total) { ElMessage.warning(`外包单位「${u.name}」${s.label}加班人数不能大于用工人数`); return }
    }
  }
  const payload = JSON.parse(JSON.stringify(report))
  delete payload.updatedAt
  if (!payload.shifts.day && !payload.shifts.eve && !payload.shifts.dawn && payload.workers.length) {
    payload.shifts.day = payload.workers.length
  }
  saving.value = true
  try {
    await api.post('/report', { ...payload, projectId: pid, reportDate: reportDate.value })
    ElMessage.success('日报已保存')
    load()
  } finally { saving.value = false }
}

async function del() {
  try {
    await ElMessageBox.confirm('确定删除本日期的日报？', '删除日报', { type: 'warning' })
    await api.delete('/report/' + report.id)
    ElMessage.success('已删除')
    Object.assign(report, emptyReport())
    load()
  } catch (e) { /* 取消 */ }
}

onMounted(async () => {
  project.value = await api.get('/projects/' + pid)
  load()
})
</script>

<style scoped>
.ph { margin-bottom: 12px; }
.ph-title { font-size: 16px; font-weight: 600; }
.box { margin-bottom: 14px; }
.sec { font-weight: 700; font-size: 15px; }
.hd { display: flex; align-items: center; justify-content: space-between; }
.lbl { color: #909399; font-size: 13px; margin-bottom: 6px; }
.mt { margin-top: 12px; }
.mb { margin-bottom: 12px; }
.muted { color: #909399; }
.equip-head { display: flex; justify-content: space-between; align-items: center; }
.shift-name { font-weight: 600; color: #606266; }
.picker-list { display: flex; flex-wrap: wrap; gap: 10px; max-height: 280px; overflow: auto; }
.actions { text-align: right; margin: 8px 0 20px; }
:deep(.el-col-4\.8) { width: 19.2%; }
</style>
