<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { getEngineHealth, getLicenses } from '@/api/system'
import { useUiStore, type ThemeChoice } from '@/stores/ui'
import type { EngineHealth, LicenseInfo } from '@/api/system'

const activeTab = ref<'preference' | 'engine' | 'about'>('preference')
const ui = useUiStore()
const engineHealth = ref<EngineHealth | null>(null)
const engineLoading = ref(false)
const engineError = ref<string | null>(null)
const licenseInfo = ref<LicenseInfo | null>(null)
const licensesLoading = ref(false)
const licensesError = ref<string | null>(null)

const tabs = [
  { id: 'preference', label: '界面偏好' },
  { id: 'engine', label: '引擎诊断' },
  { id: 'about', label: '关于与许可证' },
] as const

type TabId = typeof tabs[number]['id']

async function selectTab(tabId: TabId) {
  activeTab.value = tabId
  await nextTick()
  document.getElementById(`settings-tab-${tabId}`)?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number) {
  let nextIndex: number | undefined
  if (event.key === 'ArrowDown' || event.key === 'ArrowRight') nextIndex = (index + 1) % tabs.length
  else if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') nextIndex = (index - 1 + tabs.length) % tabs.length
  else if (event.key === 'Home') nextIndex = 0
  else if (event.key === 'End') nextIndex = tabs.length - 1
  else return
  event.preventDefault()
  void selectTab(tabs[nextIndex]!.id)
}

async function checkEngineHealth() {
  engineLoading.value = true
  engineError.value = null
  try {
    engineHealth.value = await getEngineHealth()
  } catch {
    engineHealth.value = null
    engineError.value = '无法从后端获取引擎状态'
  } finally {
    engineLoading.value = false
  }
}

async function loadLicenses() {
  licensesLoading.value = true
  licensesError.value = null
  try {
    licenseInfo.value = await getLicenses()
  } catch {
    licenseInfo.value = null
    licensesError.value = '无法从后端获取许可证信息'
  } finally {
    licensesLoading.value = false
  }
}

onMounted(() => {
  void checkEngineHealth()
  void loadLicenses()
})
</script>

<template>
  <section class="page active">
    <div class="settings-layout">
      <div class="settings-nav surface" role="tablist" aria-label="设置分类">
        <button
          v-for="(tab, index) in tabs"
          :key="tab.id"
          :class="{ active: activeTab === tab.id }"
          role="tab"
          :id="`settings-tab-${tab.id}`"
          :aria-controls="`settings-panel-${tab.id}`"
          :aria-selected="activeTab === tab.id"
          :tabindex="activeTab === tab.id ? 0 : -1"
          @click="selectTab(tab.id)"
          @keydown="handleTabKeydown($event, index)"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="settings-content">
        <section
          v-if="activeTab === 'preference'"
          id="settings-panel-preference"
          class="settings-tab active surface"
          role="tabpanel"
          aria-labelledby="settings-tab-preference"
          tabindex="0"
        >
          <div class="panel-header">
            <div>
              <span class="section-kicker">界面偏好</span>
              <h3>主题</h3>
            </div>
          </div>
          <div class="setting-row">
            <div>
              <strong>界面主题</strong>
              <small>主题保存在当前浏览器，不属于后端业务数据。</small>
            </div>
            <select
              :value="ui.theme"
              aria-label="界面主题"
              @change="ui.setTheme(($event.target as HTMLSelectElement).value as ThemeChoice)"
            >
              <option value="system">跟随系统</option>
              <option value="light">浅色</option>
              <option value="dark">深色</option>
            </select>
          </div>
        </section>

        <section
          v-else-if="activeTab === 'engine'"
          id="settings-panel-engine"
          class="settings-tab active surface"
          role="tabpanel"
          aria-labelledby="settings-tab-engine"
          tabindex="0"
        >
          <div class="panel-header">
            <div>
              <span class="section-kicker">引擎诊断</span>
              <h3>后端搜索引擎状态</h3>
            </div>
            <span
              class="tag"
              :class="engineHealth?.status === 'healthy' ? 'success' : 'neutral'"
            >
              {{ engineLoading ? '检查中' : engineHealth?.status || '不可用' }}
            </span>
          </div>

          <div v-if="engineLoading" class="loading-state">正在读取后端引擎状态…</div>
          <div v-else-if="engineError" class="error-state">
            <p>{{ engineError }}</p>
          </div>
          <template v-else-if="engineHealth">
            <div class="engine-card">
              <span class="engine-logo">E</span>
              <div>
                <strong>{{ engineHealth.name }}</strong>
                <small>类型：{{ engineHealth.type }}</small>
              </div>
              <span class="status-inline">
                <i class="healthy" />
                {{ engineHealth.status }}
              </span>
            </div>
            <dl class="diagnostic-list">
              <div><dt>引擎名称</dt><dd>{{ engineHealth.name }}</dd></div>
              <div><dt>引擎类型</dt><dd>{{ engineHealth.type }}</dd></div>
              <div><dt>后端状态</dt><dd>{{ engineHealth.status }}</dd></div>
            </dl>
          </template>

          <button class="secondary-button" :disabled="engineLoading" @click="checkEngineHealth">
            <AppIcon name="refresh" />重新检查
          </button>
        </section>

        <section
          v-else
          id="settings-panel-about"
          class="settings-tab active surface"
          role="tabpanel"
          aria-labelledby="settings-tab-about"
          tabindex="0"
        >
          <div class="panel-header">
            <div>
              <span class="section-kicker">关于棋境</span>
              <h3>Xiangqi Lab</h3>
            </div>
            <span class="tag" :class="licenseInfo ? 'success' : 'neutral'">
              {{ licensesLoading ? '加载中' : licenseInfo ? '后端数据' : '不可用' }}
            </span>
          </div>
          <p class="about-copy">
            棋境是一款中国象棋人机对战、棋谱学习与复盘软件。
          </p>

          <div v-if="licensesLoading" class="loading-state">正在读取后端许可证信息…</div>
          <div v-else-if="licensesError" class="error-state">
            <p>{{ licensesError }}</p>
            <button class="secondary-button small" @click="loadLicenses">重新加载</button>
          </div>
          <template v-else-if="licenseInfo">
            <dl class="license-list">
              <div>
                <dt>项目许可证</dt>
                <dd>{{ licenseInfo.application }}</dd>
              </div>
              <div v-for="engine in licenseInfo.externalEngines" :key="engine.name">
                <dt>{{ engine.name }} · {{ engine.status }}</dt>
                <dd>{{ engine.notice }}</dd>
              </div>
            </dl>
          </template>
        </section>
      </div>
    </div>
  </section>
</template>
