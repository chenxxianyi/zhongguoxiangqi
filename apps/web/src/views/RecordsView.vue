<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { useRecordsStore } from '@/stores/records'
import { useUiStore } from '@/stores/ui'

const recordsStore = useRecordsStore()
const ui = useUiStore()
const input = ref<HTMLInputElement | null>(null)

onMounted(() => {
  if (!recordsStore.loaded) recordsStore.fetchRecords()
})

async function upload(event: Event) {
  const files = [...((event.target as HTMLInputElement).files ?? [])]
  if (!files.length) return

  recordsStore.importing = true
  recordsStore.importProgress = 0
  recordsStore.importTitle = `正在处理 ${files.length} 个文件…`

  const count = await recordsStore.importRecords(files)
  ui.showToast(`成功导入 ${count} 盘棋谱`)
  // 重置 input 以允许重复选择相同文件
  if (input.value) input.value.value = ''
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  } catch {
    return dateStr.slice(0, 10)
  }
}

function outcomeLabel(outcome: string): string {
  switch (outcome) {
    case 'red_win': return '红胜'
    case 'black_win': return '黑胜'
    case 'draw': return '和棋'
    default: return outcome
  }
}

function outcomeClass(outcome: string): string {
  switch (outcome) {
    case 'red_win': return 'win-soft'
    case 'black_win': return 'loss-soft'
    case 'draw': return 'neutral'
    default: return 'neutral'
  }
}
</script>

<template>
  <section class="page active">
    <div class="section-intro split">
      <div>
        <span class="section-kicker">棋谱管理</span>
        <h2>让经典棋局成为你的学习资料</h2>
        <p>支持文本、PGN、坐标记谱、常见中文记谱和项目 JSON 格式。</p>
      </div>
      <button class="primary-button" :disabled="recordsStore.importing" @click="input?.click()">
        <AppIcon name="upload" />{{ recordsStore.importing ? '导入中…' : '导入棋谱' }}
      </button>
      <input ref="input" type="file" accept=".txt,.pgn,.json" multiple hidden @change="upload">
    </div>

    <div class="records-grid">
      <div class="surface records-main">
        <div class="records-toolbar">
          <div class="search-field"><AppIcon name="eye" /><input type="search" placeholder="搜索棋手、赛事或开局" aria-label="搜索棋谱"></div>
          <button class="filter-button"><AppIcon name="filter" />筛选</button>
          <span class="record-total">共 {{ recordsStore.totalRecords }} 盘</span>
        </div>
        <div class="record-table-wrap">
          <table class="record-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>格式</th>
                <th>结果</th>
                <th>着数</th>
                <th>日期</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="recordsStore.records.length === 0">
                <td colspan="6" class="empty-state">暂无棋谱，点击"导入棋谱"按钮上传</td>
              </tr>
              <tr v-for="record in recordsStore.records" :key="record.id">
                <td><strong>{{ record.name }}</strong></td>
                <td>{{ record.format }}</td>
                <td><span class="tag" :class="outcomeClass(record.outcome)">{{ outcomeLabel(record.outcome) }}</span></td>
                <td>{{ record.moveCount }} 步</td>
                <td>{{ formatDate(record.createdAt) }}</td>
                <td>
                  <button class="icon-button" aria-label="删除" @click="recordsStore.deleteRecord(record.id)">
                    <AppIcon name="eye" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <aside class="records-side">
        <article class="surface import-panel">
          <div class="panel-header">
            <div>
              <span class="section-kicker">最近导入</span>
              <h3>{{ recordsStore.importTitle || '暂无导入记录' }}</h3>
            </div>
            <span class="tag" :class="recordsStore.importing ? 'neutral' : 'success'">
              {{ recordsStore.importing ? '导入中' : '就绪' }}
            </span>
          </div>
          <div class="import-progress" v-if="recordsStore.importing || recordsStore.importProgress > 0">
            <span :style="{ width: `${recordsStore.importProgress}%` }" />
          </div>
          <div class="import-stats">
            <div><strong>{{ recordsStore.totalRecords }}</strong><span>总棋谱</span></div>
          </div>
        </article>

        <article class="surface collection-panel">
          <div class="panel-header">
            <div>
              <span class="section-kicker">棋谱集合</span>
              <h3>{{ recordsStore.totalRecords > 0 ? `${recordsStore.totalRecords} 盘已收录` : '暂无棋谱' }}</h3>
            </div>
          </div>
        </article>
      </aside>
    </div>
  </section>
</template>
