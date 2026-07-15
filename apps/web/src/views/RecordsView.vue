<script setup lang="ts">
import { ref } from 'vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { useUiStore } from '@/stores/ui'

interface RecordRow { players: string; event: string; opening: string; result: string; resultClass: string; collection: string; date: string }
const ui = useUiStore()
const input = ref<HTMLInputElement | null>(null)
const importing = ref(false)
const progress = ref(82)
const success = ref(42)
const importTitle = ref('7 月 14 日批次')
const records = ref<RecordRow[]>([
  { players:'许银川 对 吕钦',event:'第 12 届全国象棋个人赛',opening:'中炮对屏风马',result:'红胜',resultClass:'win-soft',collection:'岭南名局精选',date:'2024-10-18' },
  { players:'赵国荣 对 胡荣华',event:'全国象棋甲级联赛',opening:'飞相局',result:'和棋',resultClass:'neutral',collection:'大师残局集',date:'2023-06-09' },
  { players:'王天一 对 郑惟桐',event:'象棋冠军邀请赛',opening:'仙人指路',result:'黑胜',resultClass:'loss-soft',collection:'现代布局研究',date:'2025-01-22' },
  { players:'柳大华 对 李来群',event:'五羊杯冠军赛',opening:'顺炮直车',result:'红胜',resultClass:'win-soft',collection:'经典攻杀',date:'2022-11-03' },
])
function upload(event: Event) {
  const files = [...((event.target as HTMLInputElement).files ?? [])]
  if (!files.length) return
  importing.value = true; progress.value = 24; importTitle.value = '刚刚上传的批次'
  records.value.unshift({ players: files[0]?.name ?? '新棋谱', event:'刚刚导入的演示棋谱', opening:'待识别', result:'解析中', resultClass:'neutral', collection:'未分类', date:'刚刚' })
  ui.showToast(`已接收 ${files.length} 个文件，正在模拟安全校验`)
  window.setTimeout(() => { importing.value = false; progress.value = 100; success.value += files.length }, 1300)
}
</script>

<template>
  <section class="page active"><div class="section-intro split"><div><span class="section-kicker">棋谱管理</span><h2>让经典棋局成为你的学习资料</h2><p>支持文本、PGN、坐标记谱、常见中文记谱和项目 JSON 格式。</p></div><button class="primary-button" @click="input?.click()"><AppIcon name="upload" />导入棋谱</button><input ref="input" type="file" accept=".txt,.pgn,.json" multiple hidden @change="upload"></div>
  <div class="records-grid"><div class="surface records-main"><div class="records-toolbar"><div class="search-field"><AppIcon name="eye" /><input type="search" placeholder="搜索棋手、赛事或开局" aria-label="搜索棋谱"></div><button class="filter-button"><AppIcon name="filter" />筛选</button><span class="record-total">共 128 盘</span></div><div class="record-table-wrap"><table class="record-table"><thead><tr><th>对局</th><th>开局</th><th>结果</th><th>来源集合</th><th>日期</th><th></th></tr></thead><tbody><tr v-for="record in records" :key="`${record.players}-${record.date}`"><td><strong>{{ record.players }}</strong><small>{{ record.event }}</small></td><td>{{ record.opening }}</td><td><span class="tag" :class="record.resultClass">{{ record.result }}</span></td><td>{{ record.collection }}</td><td>{{ record.date }}</td><td><button class="icon-button" aria-label="查看棋谱"><AppIcon name="chevron" /></button></td></tr></tbody></table></div></div>
  <aside class="records-side"><article class="surface import-panel"><div class="panel-header"><div><span class="section-kicker">最近导入</span><h3>{{ importTitle }}</h3></div><span class="tag" :class="importing ? 'neutral' : 'success'">{{ importing ? '解析中' : '已完成' }}</span></div><div class="import-progress"><span :style="{ width: `${progress}%` }" /></div><div class="import-stats"><div><strong>{{ success }}</strong><span>成功</span></div><div><strong>3</strong><span>重复</span></div><div><strong>1</strong><span>失败</span></div><div><strong>2</strong><span>警告</span></div></div><button class="secondary-button full">查看导入报告</button></article>
  <article class="surface collection-panel"><div class="panel-header"><div><span class="section-kicker">棋谱集合</span><h3>按主题组织学习</h3></div><button class="icon-button" aria-label="棋谱集合选项"><AppIcon name="more" /></button></div><button v-for="collection in [['red','岭南名局精选','86 盘 · 已用于学习','86'],['green','大师残局集','24 盘','24'],['ochre','现代布局研究','18 盘','18']]" :key="collection[1]"><span class="collection-color" :class="collection[0]"/><span><strong>{{ collection[1] }}</strong><small>{{ collection[2] }}</small></span><b>{{ collection[3] }}</b></button></article></aside></div></section>
</template>
