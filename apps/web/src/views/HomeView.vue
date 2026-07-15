<script setup lang="ts">
import { RouterLink } from 'vue-router'
import AppIcon from '@/components/common/AppIcon.vue'
import MiniBoard from '@/components/board/MiniBoard.vue'
</script>

<template>
  <section class="page active">
    <div class="hero-panel">
      <div class="hero-copy">
        <span class="section-kicker">继续你的棋局</span>
        <h2>在每一次落子中，<br><em>看见自己的进步。</em></h2>
        <p>与可调节难度的 AI 对弈，导入经典棋谱，建立属于你的棋风学习库。</p>
        <div class="hero-actions"><RouterLink class="primary-button" to="/match"><AppIcon name="play" />继续上局</RouterLink><RouterLink class="text-button" to="/new-game">新建对局 <AppIcon name="arrow" /></RouterLink></div>
        <div class="continue-meta"><div><span>当前执色</span><strong class="red-text">红方</strong></div><div><span>对手</span><strong>AI · 进阶 6</strong></div><div><span>进行至</span><strong>第 18 回合</strong></div></div>
      </div>
      <div class="hero-board-wrap" aria-hidden="true"><MiniBoard /><div class="hero-stamp">残局<br>待续</div></div>
    </div>

    <div class="metric-strip" aria-label="近期统计">
      <article><span class="metric-icon cinnabar"><AppIcon name="board" /></span><div><span>近 30 日对局</span><strong>24</strong><small>较上月 +6</small></div></article>
      <article><span class="metric-icon jade"><AppIcon name="chart" /></span><div><span>当前胜率</span><strong>58%</strong><small>执红胜率 64%</small></div></article>
      <article><span class="metric-icon ochre"><AppIcon name="spark" /></span><div><span>学习库命中</span><strong>31%</strong><small>本周提升 4%</small></div></article>
      <article><span class="metric-icon ink"><AppIcon name="record" /></span><div><span>收录棋谱</span><strong>128</strong><small>3 个棋谱集合</small></div></article>
    </div>

    <div class="dashboard-grid">
      <article class="surface recent-match-panel">
        <div class="panel-header"><div><span class="section-kicker">最近对局</span><h3>复盘你的每一步</h3></div><RouterLink class="text-button small" to="/history">查看全部 <AppIcon name="arrow" /></RouterLink></div>
        <div class="match-list">
          <RouterLink v-for="match in [
            ['win','胜','AI · 休闲 4','中炮对屏风马 · 执红','42 回合','今日 09:42','38%'],
            ['loss','负','AI · 进阶 6','仙人指路 · 执黑','56 回合','昨日 21:18','26%'],
            ['draw','和','AI · 进阶 5','飞相局 · 执红','68 回合','7 月 13 日','44%']
          ]" :key="match[2]" class="match-row" to="/analysis">
            <span class="result-badge" :class="match[0]">{{ match[1] }}</span><span class="match-opponent"><strong>{{ match[2] }}</strong><small>{{ match[3] }}</small></span><span class="match-score"><strong>{{ match[4] }}</strong><small>{{ match[5] }}</small></span><span class="opening-hit"><span>学习库</span><strong>{{ match[6] }}</strong></span><AppIcon class="row-chevron" name="chevron" />
          </RouterLink>
        </div>
      </article>
      <article class="surface learning-summary">
        <div class="panel-header"><div><span class="section-kicker">学习库</span><h3>岭南名局精选</h3></div><span class="tag success">已启用</span></div>
        <div class="style-orbit"><svg viewBox="0 0 220 170" role="img" aria-label="棋风倾向雷达图"><polygon points="110,18 190,70 160,148 60,148 30,70" class="radar-grid"/><polygon points="110,44 166,78 146,128 76,132 48,76" class="radar-grid inner"/><polygon points="110,30 172,76 142,135 69,122 45,72" class="radar-value"/><g class="radar-labels"><text x="110" y="12">进攻</text><text x="201" y="73">活跃</text><text x="166" y="164">残局</text><text x="38" y="164">兑子</text><text x="5" y="73">稳健</text></g></svg></div>
        <div class="library-meta"><div><span>有效棋谱</span><strong>86 盘</strong></div><div><span>覆盖局面</span><strong>3,642</strong></div><div><span>画像置信</span><strong>较高</strong></div></div><RouterLink class="secondary-button full" to="/learning">查看学习报告</RouterLink>
      </article>
    </div>
    <div class="insight-banner"><span class="insight-mark"><AppIcon name="spark" /></span><div><span class="section-kicker">今日洞察</span><h3>你在“中炮对屏风马”的中局转换中，容易过早兑车。</h3><p>近 8 盘相似对局中出现 5 次，平均造成 0.7 分的局面损失。</p></div><RouterLink class="secondary-button" to="/analysis">查看专项复盘</RouterLink></div>
  </section>
</template>
