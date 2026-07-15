<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import AppModal from '@/components/common/AppModal.vue'
import XiangqiBoard from '@/components/board/XiangqiBoard.vue'
import { useMatchStore } from '@/stores/match'
import { useUiStore } from '@/stores/ui'

const tab = ref<'moves' | 'info'>('moves')
const resignOpen = ref(false)
const match = useMatchStore()
const ui = useUiStore()
const router = useRouter()
const moves = [['1.','炮二平五','马８进７'],['2.','马二进三','车９平８'],['3.','车一平二','炮８进４'],['4.','兵七进一','卒７进１'],['5.','车二进六','马２进３'],['6.','马八进七','象３进５'],['7.','车九进一','士４进５'],['8.','车九平六','车１平４'],['9.','车六进八','将５平４'],['10.','炮五平六','等待中']]
function undo() { ui.showToast(match.undo() ? '已撤销最近一次演示走子' : '当前没有可撤销的演示走子') }
async function confirmResign() { resignOpen.value = false; match.reset(); ui.showToast('本局已结束并保存到历史对局'); await router.push('/history') }
</script>

<template>
  <section class="page active match-page">
    <div class="match-shell">
      <div class="match-board-column">
        <div class="player-bar opponent"><div class="player-identity"><span class="avatar ai">AI</span><div><strong>棋境 AI</strong><small>进阶 6 · 外部引擎演示</small></div></div><div class="thinking-label"><span class="thinking-dot" />正在思考</div><time>08:42</time></div>
        <XiangqiBoard @undo="undo" @resign="resignOpen = true">
          <div class="player-bar self"><div class="player-identity"><span class="avatar">林</span><div><strong>林间棋客</strong><small>红方 · 轮到你走</small></div></div><div class="turn-label">你的回合</div><time>09:16</time></div>
        </XiangqiBoard>
      </div>
      <aside class="match-panel surface">
        <div class="match-panel-tabs" role="tablist"><button :class="{ active: tab === 'moves' }" role="tab" :aria-selected="tab === 'moves'" @click="tab = 'moves'">着法</button><button :class="{ active: tab === 'info' }" role="tab" :aria-selected="tab === 'info'" @click="tab = 'info'">局面</button></div>
        <div v-if="tab === 'moves'" class="match-tab active"><div class="opening-name"><span>当前开局</span><strong>中炮对屏风马 · 平炮兑车</strong><small>学习库命中至第 12 回合</small></div><ol class="move-list"><li v-for="(move,index) in moves" :key="move[0]" :class="{ current: index === moves.length - 1 }"><span>{{ move[0] }}</span><button>{{ move[1] }}</button><button>{{ move[2] }}</button></li></ol><div class="live-insight"><span><AppIcon name="spark" /></span><div><strong>本局学习库命中 3 次</strong><small>AI 的第 2、4、6 回合参考了“岭南名局精选”。</small></div></div></div>
        <div v-else class="match-tab active"><dl class="position-details"><div><dt>当前 FEN</dt><dd>2bakab2/4n4/4c1n2/...</dd></div><div><dt>对局版本</dt><dd>v18</dd></div><div><dt>当前行棋方</dt><dd>红方</dd></div><div><dt>规则集</dt><dd>休闲规则 1.0</dd></div><div><dt>重复局面</dt><dd>0 次</dd></div></dl></div>
      </aside>
    </div>
    <p class="demo-hint"><AppIcon name="info" />可点击棋子查看落点提示并体验走子；当前组件已按服务端权威规则接口预留数据边界。</p>
    <AppModal :open="resignOpen" title="确认认输？" description="本局将立即结束，并保存到历史对局中。" danger @close="resignOpen = false" @confirm="confirmResign" />
  </section>
</template>
