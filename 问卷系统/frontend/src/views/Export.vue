<template>
  <div>
    <el-card class="box" shadow="never">
      <template #header><b>💰 结算单</b></template>
      <el-alert type="info" :closable="false" class="mb">
        结构对齐《本地（北京）结算单》：人工 → 餐费(含水) → 叉车3/5/7吨 → 吊车 → 救援车 → 税费6% → 劳务总合计；单价可参数化，默认取项目配置。
      </el-alert>
      <el-form label-width="100px">
        <el-form-item label="项目" required>
          <el-select v-model="s.projectId" placeholder="选择项目" style="width: 320px" @change="applyDefaultPrices">
            <el-option v-for="p in projects" :key="p.id" :label="p.unit + '·' + p.activityName" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始日期" required>
          <el-date-picker v-model="s.start" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="结束日期" required>
          <el-date-picker v-model="s.end" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="单价设置">
          <el-collapse class="price-box">
            <el-collapse-item title="点开设置单价（默认取项目配置）">
              <el-row :gutter="14">
                <el-col :span="8" v-for="f in priceFields" :key="f.key">
                  <div class="lbl">{{ f.label }}</div>
                  <el-input-number v-model="s[f.key]" :min="0" controls-position="right" style="width:100%" />
                </el-col>
              </el-row>
            </el-collapse-item>
          </el-collapse>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Download" :loading="busy" @click="exportSettlement">导出结算单 xlsx</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="box" shadow="never">
      <template #header><b>📅 考勤表</b></template>
      <el-alert type="info" :closable="false" class="mb">
        结构对齐《考勤表》：Sheet「m月份」，行=人员（按拼音排序），列=1~31 日，格=当日工数。
      </el-alert>
      <el-form label-width="100px">
        <el-form-item label="项目" required>
          <el-select v-model="a.projectId" placeholder="选择项目" style="width: 320px">
            <el-option v-for="p in projects" :key="p.id" :label="p.unit + '·' + p.activityName" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="月份" required>
          <el-date-picker v-model="a.month" type="month" value-format="YYYY-MM" placeholder="选择月份" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Download" :loading="busy" @click="exportAttendance">导出考勤表 xlsx</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import api, { download } from '../api'

const projects = ref([])
const busy = ref(false)
const priceFields = [
  { key: 'laborPrice', label: '人工 元/人' },
  { key: 'mealPrice', label: '餐费 元/份' },
  { key: 'waterPrice', label: '水 元/份' },
  { key: 'fork3Price', label: '叉车3吨 元/台班' },
  { key: 'fork5Price', label: '叉车5吨 元/台班' },
  { key: 'fork7Price', label: '叉车7吨 元/台班' },
  { key: 'cranePrice', label: '吊车 元/台班' },
  { key: 'rescuePrice', label: '救援车 元/台班' },
  { key: 'taxRate', label: '税率 %' }
]
const s = reactive({ projectId: 0, start: '', end: '', laborPrice: 380, mealPrice: 20, waterPrice: 5, fork3Price: 1200, fork5Price: 1600, fork7Price: 2000, cranePrice: 1200, rescuePrice: 800, taxRate: 6 })
const a = reactive({ projectId: 0, month: new Date().toISOString().slice(0, 7) })

const labelOf = (id) => projects.value.find(p => p.id === id)?.unit + '·' + projects.value.find(p => p.id === id)?.activityName || ''
function applyDefaultPrices(id) {
  const p = projects.value.find(x => x.id === id)
  if (!p?.prices) return
  try {
    const m = JSON.parse(p.prices)
    s.laborPrice = m.labor ?? s.laborPrice
    s.mealPrice = m.meal ?? s.mealPrice
    s.waterPrice = m.water ?? s.waterPrice
    s.fork3Price = m.fork3 ?? s.fork3Price
    s.fork5Price = m.fork5 ?? s.fork5Price
    s.fork7Price = m.fork7 ?? s.fork7Price
    s.cranePrice = m.crane ?? s.cranePrice
    s.rescuePrice = m.rescue ?? s.rescuePrice
    s.taxRate = m.taxRate ?? s.taxRate
    if (p.startDate) s.start = p.startDate
    if (p.endDate) s.end = p.endDate
  } catch (e) { /* ignore */ }
}

async function exportSettlement() {
  if (!s.projectId) { ElMessage.warning('请选择项目'); return }
  if (!s.start || !s.end) { ElMessage.warning('请选择起止日期'); return }
  busy.value = true
  try {
    const qs = new URLSearchParams({ projectId: s.projectId, start: s.start, end: s.end })
    for (const f of priceFields) qs.set(f.key, s[f.key])
    await download('/api/export/settlement?' + qs.toString(), `结算单_${labelOf(s.projectId)}_${s.start}至${s.end}.xlsx`)
  } finally { busy.value = false }
}
async function exportAttendance() {
  if (!a.projectId) { ElMessage.warning('请选择项目'); return }
  if (!a.month) { ElMessage.warning('请选择月份'); return }
  busy.value = true
  try {
    const qs = new URLSearchParams({ projectId: a.projectId, month: a.month })
    await download('/api/export/attendance?' + qs.toString(), `考勤表_${labelOf(a.projectId)}_${a.month}.xlsx`)
  } finally { busy.value = false }
}

onMounted(async () => {
  projects.value = await api.get('/projects')
  if (projects.value.length) {
    s.projectId = projects.value[0].id
    a.projectId = projects.value[0].id
    applyDefaultPrices(projects.value[0].id)
  }
})
</script>

<style scoped>
.box { margin-bottom: 14px; }
.mb { margin-bottom: 14px; }
.lbl { color: #909399; font-size: 13px; margin-bottom: 6px; }
.price-box { width: 100%; }
</style>
