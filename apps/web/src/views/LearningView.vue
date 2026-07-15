<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import AppModal from '@/components/common/AppModal.vue'
import { useLearningStore } from '@/stores/learning'
import { useUiStore } from '@/stores/ui'

const modalOpen = ref(false)
const versionName = ref('')
const learning = useLearningStore()
const ui = useUiStore()

onMounted(() => {
  if (!learning.loaded) learning.fetchVersions()
})

async function handleCreate() {
  await learning.createJob(versionName.value || undefined)
  modalOpen.value = false
  versionName.value = ''
  if (learning.completed) {
    ui.showToast('学习版本构建完成')
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  } catch {
    return dateStr.slice(0, 10)
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'active': return '当前启用'
    case 'ready': return '就绪'
    case 'superseded': return '已归档'
    case 'building': return '构建中'
    default: return status
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'active': return 'success'
    case 'ready': return 'neutral'
    case 'superseded': return 'neutral'
    default: return 'neutral'
  }
}
</script>

<template>
  <section class="page active">
    <!-- 当前版本 -->
    <div class="learning-hero">
      <div>
        <span class="section-kicker">当前启用版本</span>
        <h2>{{ learning.activeVersion?.name || '暂无学习版本' }}</h2>
        <p v-if="learning.activeVersion">
          基于 {{ learning.activeVersion.quality.validRecords }} 盘有效棋谱构建，包含 {{ learning.activeVersion.quality.coveredPositions }} 个可查询局面。
          AI 会在样本可信时参考学习库。
        </p>
        <p v-else>导入棋谱后，可在此创建学习版本。</p>
        <div v-if="learning.activeVersion" class="inline-meta">
          <span>发布于 {{ formatDate(learning.activeVersion.createdAt) }}</span>
          <span>算法 {{ learning.activeVersion.algorithm }}</span>
        </div>
      </div>
      <div class="learning-actions">
        <button class="secondary-button" @click="learning.fetchVersions()">刷新</button>
        <button class="primary-button" @click="modalOpen = true">
          <AppIcon name="spark" />创建新版本
        </button>
      </div>
    </div>

    <!-- 质量概览 -->
    <div v-if="learning.activeVersion" class="quality-grid">
      <article class="quality-card">
        <span>有效棋谱</span>
        <strong>{{ learning.activeVersion.quality.validRecords }}</strong>
        <small>通过校验的记录</small>
      </article>
      <article class="quality-card">
        <span>有效着数</span>
        <strong>{{ learning.activeVersion.quality.validMoves }}</strong>
        <small>全部合法着法</small>
      </article>
      <article class="quality-card">
        <span>局面覆盖</span>
        <strong>{{ learning.activeVersion.quality.coveredPositions }}</strong>
        <small>独立局面数</small>
      </article>
      <article class="quality-card">
        <span>低样本条目</span>
        <strong>{{ learning.activeVersion.quality.lowSampleEntries }}</strong>
        <small>样本 < 3 的局面</small>
      </article>
    </div>

    <!-- 版本列表 -->
    <article class="surface version-panel">
      <div class="panel-header">
        <div>
          <span class="section-kicker">版本记录</span>
          <h3>所有构建版本</h3>
        </div>
      </div>
      <div v-if="learning.versions.length === 0" class="empty-state">
        暂无学习版本。请先导入棋谱，然后创建新版本。
      </div>
      <div
        v-for="version in learning.versions"
        :key="version.id"
        class="version-row"
        :class="{ active: version.status === 'active' }"
      >
        <span class="version-line"><i /></span>
        <div>
          <strong>{{ version.name }}</strong>
          <small>{{ formatDate(version.createdAt) }} · {{ version.quality.validRecords }} 盘</small>
        </div>
        <span class="tag" :class="statusClass(version.status)">{{ statusLabel(version.status) }}</span>
        <button
          v-if="version.status === 'ready'"
          class="secondary-button small"
          @click="learning.activateVersion(version.id)"
        >启用</button>
        <button
          v-else-if="version.status === 'superseded'"
          class="text-button small"
          @click="learning.rollback(version.id)"
        >回滚至此</button>
      </div>
    </article>

    <!-- 创建版本弹窗 -->
    <AppModal
      :open="modalOpen"
      title="创建学习版本"
      description="将从已导入棋谱中构建新的学习版本。"
      @close="modalOpen = false"
    >
      <div class="form-field">
        <label>版本名称（可选）</label>
        <input v-model="versionName" type="text" placeholder="例如：岭南名局精选 v4" />
      </div>

      <div v-if="learning.running || learning.completed" class="fake-job">
        <div>
          <span>{{ learning.stage }}</span>
          <strong>{{ learning.progress }}%</strong>
        </div>
        <i><b :style="{ width: `${learning.progress}%` }" /></i>
      </div>

      <button
        class="primary-button full"
        :disabled="learning.running"
        @click="handleCreate"
      >
        {{ learning.completed ? '构建完成' : learning.running ? '构建中…' : '开始构建' }}
      </button>
    </AppModal>
  </section>
</template>
