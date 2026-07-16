<script setup lang="ts">
import AppIcon from '@/components/common/AppIcon.vue'
import { useGameSetupStore } from '@/stores/gameSetup'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'
import type { AiMode, Color, SideChoice } from '@/types/xiangqi'

const setup = useGameSetupStore()
const match = useMatchStore()
const ui = useUiStore()

const sideOptions: Array<{ value: SideChoice; title: string; note: string; piece: string }> = [
  { value: 'red', title: '执红', note: '先手行棋', piece: '帅' },
  { value: 'black', title: '执黑', note: '后手应对', piece: '将' },
  { value: 'random', title: '随机', note: '交给棋局', piece: '随机' },
]

const modes: Array<{ value: AiMode; title: string; note: string; icon: string }> = [
  { value: 'standard', title: '标准引擎', note: '根据当前难度稳定搜索，不参考个人学习库。', icon: 'board' },
  { value: 'library', title: '棋谱库优先', note: '命中可信局面时按频率与结果选择着法。', icon: 'record' },
  { value: 'style', title: '棋风模仿', note: '在评分损失允许范围内，模仿选定棋谱集合。', icon: 'spark' },
]

function resolveColor(side: SideChoice): Color {
  if (side === 'random') return Math.random() < 0.5 ? 'red' : 'black'
  return side
}

async function start() {
  const profile = setup.profile
  if (!profile) {
    ui.showToast('难度配置尚未从后端加载完成')
    return
  }
  const color = resolveColor(setup.side)
  ui.showToast(`正在创建对局：${color === 'red' ? '执红' : '执黑'}，${setup.modeLabel}，AI 难度 ${profile.level} 级`)
  try {
    await match.createMatch(color, profile.level, setup.mode, true)
  } catch {
    ui.showToast('创建对局失败，请确认后端服务已启动')
  }
}
</script>

<template>
  <section class="page active">
    <div class="setup-layout">
      <div class="setup-main">
        <div class="section-intro">
          <span class="section-kicker">对局设置</span>
          <h2>选择适合此刻的对手</h2>
          <p>每项难度都由真实的搜索资源与候选着策略构成，不以随机走棋伪装难度。</p>
        </div>

        <section class="setup-section" aria-labelledby="side-title">
          <div class="setup-title"><span>01</span><div><h3 id="side-title">选择执色</h3><p>红方先行，随机执色将在创建对局时确定。</p></div></div>
          <div class="option-grid three">
            <button
              v-for="option in sideOptions"
              :key="option.value"
              class="choice-card"
              :class="{ active: setup.side === option.value }"
              @click="setup.side = option.value"
            >
              <span v-if="option.value !== 'random'" class="piece-choice" :class="option.value === 'red' ? 'red-piece' : 'black-piece'">{{ option.piece }}</span>
              <span v-else class="split-piece"><i>帅</i><i>将</i></span>
              <strong>{{ option.title }}</strong>
              <small>{{ option.note }}</small>
              <span class="choice-check"><AppIcon name="check" /></span>
            </button>
          </div>
        </section>

        <section class="setup-section" aria-labelledby="difficulty-title">
          <div class="setup-title"><span>02</span><div><h3 id="difficulty-title">选择难度</h3><p>从入门到大师，搜索时间、搜索深度和节点预算逐级变化。</p></div></div>
          <div v-if="setup.loading" class="loading-state">正在读取后端难度配置…</div>
          <div v-else-if="setup.error" class="error-state">
            <p>{{ setup.error }}</p>
            <button class="secondary-button small" @click="setup.fetchProfiles">重新加载</button>
          </div>
          <div v-else-if="setup.profile" class="difficulty-selector">
            <input
              v-model.number="setup.difficulty"
              type="range"
              :min="setup.minDifficulty"
              :max="setup.maxDifficulty"
              :aria-label="`AI 难度，当前 ${setup.difficulty} 级`"
            >
            <div class="difficulty-numbers" aria-hidden="true">
              <span
                v-for="item in setup.profiles"
                :key="item.level"
                :class="{ active: setup.difficulty === item.level }"
              >
                {{ item.level }}
              </span>
            </div>
          </div>
          <div v-if="setup.profile" class="difficulty-detail">
            <div>
              <span class="difficulty-level">第 {{ setup.difficulty }} 级</span>
              <h4>{{ setup.profile.name }}</h4>
              <p>{{ setup.profile.description }}</p>
            </div>
            <dl>
              <div><dt>思考时间</dt><dd>{{ setup.profile.moveTimeMs }}ms</dd></div>
              <div><dt>搜索深度</dt><dd>{{ setup.profile.maxDepth }} 层</dd></div>
              <div><dt>候选着路数</dt><dd>{{ setup.profile.multiPV }} 路</dd></div>
            </dl>
          </div>
        </section>

        <section class="setup-section" aria-labelledby="mode-title">
          <div class="setup-title"><span>03</span><div><h3 id="mode-title">选择 AI 模式</h3><p>学习模式只在样本可信时参考棋谱，未命中会自动回退搜索引擎。</p></div></div>
          <div class="mode-list">
            <button
              v-for="mode in modes"
              :key="mode.value"
              class="mode-card"
              :class="{ active: setup.mode === mode.value }"
              @click="setup.mode = mode.value"
            >
              <span class="mode-icon"><AppIcon :name="mode.icon" /></span>
              <span><strong>{{ mode.title }}</strong><small>{{ mode.note }}</small></span>
              <span class="radio-dot" />
            </button>
          </div>
        </section>
      </div>

      <aside class="setup-summary surface">
        <span class="section-kicker">本局概要</span>
        <div class="summary-versus">
          <div><span class="piece-choice red-piece">帅</span><strong>你</strong><small>{{ setup.sideLabel }}</small></div>
          <span>对</span>
          <div><span class="piece-choice black-piece">将</span><strong>棋境 AI</strong><small>{{ setup.profile?.name || '等待后端配置' }}</small></div>
        </div>
        <div v-if="setup.profile" class="summary-lines">
          <div><span>AI 模式</span><strong>{{ setup.modeLabel }}</strong></div>
          <div><span>搜索深度</span><strong>{{ setup.profile.maxDepth }} 层</strong></div>
          <div><span>悔棋</span><strong>本局允许</strong></div>
          <div><span>思考时间</span><strong>{{ setup.profile.moveTimeMs }}ms</strong></div>
        </div>
        <div v-else class="notice"><AppIcon name="info" /><p>后端难度配置可用后才能创建对局。</p></div>
        <button class="primary-button full large" :disabled="!setup.profile || setup.loading" @click="start"><AppIcon name="play" />开始对局</button>
      </aside>
    </div>
  </section>
</template>
