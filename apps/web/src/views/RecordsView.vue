<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import AppModal from '@/components/common/AppModal.vue'
import { useRecordsStore } from '@/stores/records'
import { useUiStore } from '@/stores/ui'

const recordsStore = useRecordsStore()
const ui = useUiStore()
const input = ref<HTMLInputElement | null>(null)
const searchQuery = ref('')
const dragging = ref(false)
const pendingDeleteId = ref<string | null>(null)
const pendingDeleteName = computed(() =>
  recordsStore.records.find((record) => record.id === pendingDeleteId.value)?.name ?? '这盘棋谱',
)

const filteredRecords = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase()
  if (!query) return recordsStore.records
  return recordsStore.records.filter((record) =>
    [record.name, record.format, record.result, record.outcome, ...(record.tags ?? [])]
      .some((value) => value?.toLocaleLowerCase().includes(query)),
  )
})

onMounted(() => {
  if (!recordsStore.loaded) void recordsStore.fetchRecords()
})

async function processFiles(files: File[]) {
  if (!files.length) return

  recordsStore.importing = true
  recordsStore.importProgress = 0
  recordsStore.importTitle = `正在处理 ${files.length} 个文件…`

  const count = await recordsStore.importRecords(files)
  ui.showToast(count > 0 ? `成功导入 ${count} 盘棋谱` : '未导入棋谱，请检查文件内容')
  if (input.value) input.value.value = ''
}

async function upload(event: Event) {
  await processFiles([...((event.target as HTMLInputElement).files ?? [])])
}

function handleDrop(event: DragEvent) {
  dragging.value = false
  const files = [...(event.dataTransfer?.files ?? [])]
    .filter((file) => /\.(txt|pgn|json)$/i.test(file.name))
  if (!files.length) {
    ui.showToast('请选择 TXT、PGN 或 JSON 棋谱文件')
    return
  }
  void processFiles(files)
}

async function removeRecord(id: string) {
  try {
    await recordsStore.deleteRecord(id)
    ui.showToast('棋谱已删除')
  } catch {
    ui.showToast('删除棋谱失败')
  } finally {
    pendingDeleteId.value = null
  }
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return '-'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function outcomeLabel(outcome: string): string {
  switch (outcome) {
    case 'red_win': return '红胜'
    case 'black_win': return '黑胜'
    case 'draw': return '和棋'
    case 'ongoing': return '未决'
    default: return outcome
  }
}

function outcomeClass(outcome: string): string {
  switch (outcome) {
    case 'red_win': return 'win-soft'
    case 'black_win': return 'loss-soft'
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
        <p>支持文本、PGN 坐标记谱和项目 JSON 格式，导入内容由后端逐着验证。</p>
      </div>
      <button class="primary-button" :disabled="recordsStore.importing" @click="input?.click()">
        <AppIcon name="upload" />{{ recordsStore.importing ? '导入中…' : '导入棋谱' }}
      </button>
      <input ref="input" type="file" accept=".txt,.pgn,.json" multiple hidden @change="upload">
    </div>

    <div class="records-grid">
      <div
        class="surface records-main record-dropzone"
        :class="{ dragging }"
        @dragenter.prevent="dragging = true"
        @dragover.prevent="dragging = true"
        @dragleave.self="dragging = false"
        @drop.prevent="handleDrop"
      >
        <div class="record-drop-hint"><AppIcon name="upload" />也可将 TXT、PGN、JSON 文件拖到这里导入</div>
        <div class="records-toolbar">
          <div class="search-field">
            <AppIcon name="search" />
            <input
              v-model="searchQuery"
              type="search"
              placeholder="搜索名称、格式或标签"
              aria-label="搜索棋谱"
            >
          </div>
          <span v-if="recordsStore.loaded" class="record-total">
            共 {{ recordsStore.totalRecords }} 盘
          </span>
        </div>
        <div v-if="recordsStore.loading" class="loading-state">正在读取后端棋谱…</div>
        <div v-else-if="recordsStore.error" class="error-state">
          <p>{{ recordsStore.error }}</p>
          <button class="secondary-button small" @click="recordsStore.fetchRecords">重新加载</button>
        </div>
        <div v-else class="record-table-wrap">
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
              <tr v-if="filteredRecords.length === 0">
                <td colspan="6" class="empty-state">
                  {{ searchQuery ? '没有匹配的后端棋谱' : '暂无棋谱，点击“导入棋谱”上传' }}
                </td>
              </tr>
              <tr v-for="record in filteredRecords" :key="record.id">
                <td><strong>{{ record.name }}</strong></td>
                <td>{{ record.format }}</td>
                <td>
                  <span class="tag" :class="outcomeClass(record.outcome)">
                    {{ outcomeLabel(record.outcome) }}
                  </span>
                </td>
                <td>{{ record.moveCount }} 手</td>
                <td>{{ formatDate(record.createdAt) }}</td>
                <td>
                  <button class="icon-button" :aria-label="`删除棋谱：${record.name}`" @click="pendingDeleteId = record.id">
                    <AppIcon name="trash" />
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
          <div v-if="recordsStore.importing || recordsStore.importProgress > 0" class="import-progress">
            <span :style="{ width: `${recordsStore.importProgress}%` }" />
          </div>
          <div class="import-stats">
            <div>
              <strong>{{ recordsStore.loaded ? recordsStore.totalRecords : '--' }}</strong>
              <span>后端棋谱</span>
            </div>
          </div>
        </article>

        <article class="surface collection-panel">
          <div class="panel-header">
            <div>
              <span class="section-kicker">棋谱集合</span>
              <h3 v-if="recordsStore.loaded">
                {{ recordsStore.totalRecords > 0 ? `${recordsStore.totalRecords} 盘已收录` : '暂无棋谱' }}
              </h3>
              <h3 v-else>等待后端数据</h3>
            </div>
          </div>
        </article>
      </aside>
    </div>

    <AppModal
      :open="pendingDeleteId !== null"
      title="删除这盘棋谱？"
      :description="`“${pendingDeleteName}”将从棋谱库中移除，此操作不可撤销。`"
      danger
      cancel-label="保留棋谱"
      confirm-label="确认删除"
      @close="pendingDeleteId = null"
      @confirm="pendingDeleteId && removeRecord(pendingDeleteId)"
    />
  </section>
</template>
