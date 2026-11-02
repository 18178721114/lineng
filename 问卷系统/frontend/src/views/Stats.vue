<template>
  <div>
    <div class="toolbar">
      <el-select v-model="projectId" placeholder="全部项目" clearable style="width: 280px" @change="load">
        <el-option label="全部项目" :value="0" />
        <el-option v-for="p in projects" :key="p.id" :label="p.unit + '·' + p.activityName" :value="p.id" />
      </el-select>
    </div>

    <el-row :gutter="14" class="cards">
      <el-col :span="6" v-for="c in cardItems" :key="c.key">
        <el-card shadow="hover" class="stat">
          <el-statistic :value="s.cards[c.key]" :title="c.label" />
        </el-card>
      </el-col>
    </el-row>

    <el-card class="box" shadow="never">
      <template #header><b>👤 人员出勤排行</b></template>
      <el-table :data="s.perWorker" border>
        <el-table-column type="index" width="50" label="#" />
        <el-table-column prop="name" label="姓名" width="160" />
        <el-table-column label="工数" width="120" prop="entries" sortable />
        <el-table-column label="天数" width="120" prop="days" sortable />
        <el-table-column label="占比">
          <template #default="{ row }">
            <el-progress :percentage="barPct(row.entries)" :stroke-width="14" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="box" shadow="never">
      <template #header><b>📁 项目汇总</b></template>
      <el-table :data="s.perProject" border>
        <el-table-column prop="name" label="项目" min-width="200" />
        <el-table-column prop="reports" label="日报数" width="110" sortable />
        <el-table-column prop="heads" label="用工人次" width="120" sortable />
        <el-table-column prop="entries" label="名单工数" width="120" sortable />
      </el-table>
    </el-card>

    <el-card class="box" shadow="never">
      <template #header><b>📋 日报明细（{{ s.details.length }}）</b></template>
      <el-table :data="s.details" border row-key="id">
        <el-table-column prop="reportDate" label="日期" width="120" />
        <el-table-column label="班次" min-width="200">
          <template #default="{ row }">白 {{ num(row.shifts.day) }} / 晚 {{ num(row.shifts.eve) }} / 凌 {{ num(row.shifts.dawn) }}</template>
        </el-table-column>
        <el-table-column label="餐费" min-width="200">
          <template #default="{ row }">早 {{ num(row.meals.breakfast) }} 中 {{ num(row.meals.lunch) }} 晚 {{ num(row.meals.dinner) }} 夜 {{ num(row.meals.supper) }} 水 {{ num(row.meals.water) }}</template>
        </el-table-column>
        <el-table-column label="设备" min-width="220">
          <template #default="{ row }">3吨{{ equipSum(row,'fork3') }} 5吨{{ equipSum(row,'fork5') }} 7吨{{ equipSum(row,'fork7') }} 吊{{ equipSum(row,'crane') }} 救{{ equipSum(row,'rescue') }}</template>
        </el-table-column>
        <el-table-column type="expand" label="明细" width="70">
          <template #default="{ row }">
            <div class="exp">
              <div v-if="row.workers.length"><b>名单：</b>{{ row.workers.map(w => w.name + (w.start ? '(' + w.start + '~' + w.end + ')' : '')).join('、') }}</div>
              <div v-if="row.outsourced"><b>外包：</b>{{ row.outsourcing.map(outText).join('；') }}</div>
              <div v-if="row.notes"><b>备注：</b>{{ row.notes }}</div>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import api from '../api'

const projects = ref([])
const projectId = ref(0)
const s = reactive({ cards: {}, perWorker: [], perProject: [], details: [] })
const cardItems = [
  { key: 'reports', label: '日报单数' },
  { key: 'heads', label: '用工总人次' },
  { key: 'entries', label: '名单总工数' },
  { key: 'distinctWorkers', label: '参与工人' }
]

const num = (v) => Number(v || 0)
const equipSum = (d, k) => { const e = (d.equipment || {})[k] || {}; return num(e.day) + num(e.eve) + num(e.dawn) }
// 外包明细：单位(白n/加班n 起~止 / 晚… / 凌…)
const outShifts = [{ key: 'day', label: '白' }, { key: 'eve', label: '晚' }, { key: 'dawn', label: '凌' }]
function outText(o) {
  const parts = outShifts.map(s => {
    const cnt = num(o[s.key])
    const otCnt = num(o[s.key + 'Ot'])
    const otStart = o[s.key + 'OtStart']
    const otEnd = o[s.key + 'OtEnd']
    const ot = otCnt || otStart
      ? `(加班${otCnt}${otStart ? ' ' + otStart + '~' + (otEnd || '') : ''})`
      : ''
    return `${s.label}${cnt}${ot}`
  })
  return `${o.name || '单位'} ${parts.join(' ')}`
}
const barPct = (v) => {
  const max = Math.max(...s.perWorker.map(w => w.entries), 1)
  return Math.round(v / max * 100)
}

async function load() {
  const q = projectId.value ? `?projectId=${projectId.value}` : ''
  Object.assign(s, await api.get('/stats' + q))
}
onMounted(async () => {
  projects.value = await api.get('/projects')
  load()
})
</script>

<style scoped>
.toolbar { margin-bottom: 14px; }
.cards { margin-bottom: 14px; }
.stat :deep(.el-statistic__head) { font-size: 13px; }
.box { margin-bottom: 14px; }
.exp { padding: 6px 14px; line-height: 1.9; }
</style>
