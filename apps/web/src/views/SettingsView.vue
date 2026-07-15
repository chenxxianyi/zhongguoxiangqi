<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { apiRequest } from '@/api/client'
import { useUiStore, type ThemeChoice } from '@/stores/ui'

const activeTab = ref<'preference' | 'engine' | 'data' | 'about'>('preference')
const ui = useUiStore()

// ── 引擎诊断 ──
const engineStatus = ref<string>('检查中…')
const engineName = ref<string>('')
const engineType = ref<string>('')
const engineHealthy = ref<boolean | null>(null)

async function checkEngineHealth() {
  engineStatus.value = '检查中…'
  engineHealthy.value = null
  try {
    const result = await apiRequest<{ name: string; status: string; type: string }>('/engines/health')
    engineName.value = result.name
    engineType.value = result.type
    engineHealthy.value = result.status === 'healthy'
    engineStatus.value = engineHealthy.value ? '运行正常' : '不可用'
  } catch {
    engineName.value = ''
    engineType.value = ''
    engineHealthy.value = false
    engineStatus.value = '无法连接'
  }
}

onMounted(checkEngineHealth)
const tabs = [
  { id: 'preference', label: '对弈偏好' }, { id: 'engine', label: '引擎诊断' }, { id: 'data', label: '数据与隐私' }, { id: 'about', label: '关于与许可证' },
] as const
</script>

<template>
  <section class="page active"><div class="settings-layout"><div class="settings-nav surface" role="tablist" aria-label="设置分类"><button v-for="tab in tabs" :key="tab.id" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id">{{ tab.label }}</button></div><div class="settings-content">
    <section v-if="activeTab === 'preference'" class="settings-tab active surface"><div class="panel-header"><div><span class="section-kicker">对弈偏好</span><h3>棋盘与交互</h3></div></div><div class="setting-row"><div><strong>默认棋盘方向</strong><small>创建新对局时使用的视角。</small></div><select aria-label="默认棋盘方向"><option>跟随执色</option><option>始终红方在下</option><option>始终黑方在下</option></select></div><div class="setting-row"><div><strong>显示棋盘坐标</strong><small>在棋盘边缘显示规范坐标。</small></div><label class="switch"><input type="checkbox" checked><span/></label></div><div class="setting-row"><div><strong>落子音效</strong><small>播放短促、克制的落子声音。</small></div><label class="switch"><input type="checkbox" checked><span/></label></div><div class="setting-row"><div><strong>减少动态效果</strong><small>关闭非必要过渡和动画。</small></div><label class="switch"><input type="checkbox"><span/></label></div><div class="setting-row"><div><strong>界面主题</strong><small>可随系统或手动切换。</small></div><select :value="ui.theme" aria-label="界面主题" @change="ui.setTheme(($event.target as HTMLSelectElement).value as ThemeChoice)"><option value="system">跟随系统</option><option value="light">浅色</option><option value="dark">深色</option></select></div></section>
    <section v-else-if="activeTab === 'engine'" class="settings-tab active surface"><div class="panel-header"><div><span class="section-kicker">引擎诊断</span><h3>搜索能力与运行状态</h3></div><span class="tag" :class="engineHealthy === true ? 'success' : engineHealthy === false ? 'neutral' : ''">{{ engineStatus }}</span></div><div class="engine-card"><span class="engine-logo">E</span><div><strong>{{ engineName || '内置引擎' }}</strong><small>类型：{{ engineType || 'Alpha-Beta' }}</small></div><span class="status-inline"><i :class="{ healthy: engineHealthy === true }" />{{ engineHealthy === true ? '正常' : engineHealthy === false ? '异常' : '检查中' }}</span></div><dl class="diagnostic-list"><div><dt>引擎名称</dt><dd>{{ engineName || '内置 Alpha-Beta 引擎' }}</dd></div><div><dt>状态</dt><dd>{{ engineHealthy === true ? '健康' : engineHealthy === false ? '不可用' : '检查中' }}</dd></div></dl><div class="notice"><AppIcon name="info" /><p>正式版本不会在页面展示可执行文件路径、原始命令行或敏感调试日志。</p></div><button class="secondary-button" @click="checkEngineHealth"><AppIcon name="refresh" />重新检查</button></section>
    <section v-else-if="activeTab === 'data'" class="settings-tab active surface"><div class="panel-header"><div><span class="section-kicker">数据与隐私</span><h3>管理你的棋局资料</h3></div></div><div class="data-action"><div><strong>导出个人数据</strong><small>包含对局、棋谱集合、学习版本和偏好设置。</small></div><button class="secondary-button"><AppIcon name="download" />创建导出</button></div><div class="data-action"><div><strong>原始棋谱保留策略</strong><small>删除原始文件前，会明确说明对已发布学习版本的影响。</small></div><button class="text-button">查看策略</button></div><div class="data-action danger-zone"><div><strong>删除账户数据</strong><small>此操作不可撤销，正式版本需要再次验证身份。</small></div><button class="danger-button">申请删除</button></div></section>
    <section v-else class="settings-tab active surface"><div class="panel-header"><div><span class="section-kicker">关于棋境</span><h3>Xiangqi Lab</h3></div><span class="tag success">已接入后端</span></div><p class="about-copy">棋境是一款面向中国象棋爱好者的人机对战、棋谱学习与复盘软件。前端已连接 Go 后端服务，所有着法均由服务端规则引擎验证，AI 以内置 Alpha-Beta 搜索进行对弈。</p><dl class="license-list"><div><dt>前端框架</dt><dd>Vue 3 + TypeScript + Vite</dd></div><div><dt>后端</dt><dd>Go + 标准库 net/http</dd></div><div><dt>状态与路由</dt><dd>Pinia + Vue Router</dd></div><div><dt>图标</dt><dd>项目内统一 SVG 图标</dd></div><div><dt>搜索引擎</dt><dd>内置 Alpha-Beta Negamax，支持 MultiPV</dd></div><div><dt>外部引擎</dt><dd>Pikafish 计划以独立进程接入，分发需遵循 GPLv3</dd></div></dl></section>
  </div></div></section>
</template>
