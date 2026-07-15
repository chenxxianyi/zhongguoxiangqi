import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: () => import('@/views/HomeView.vue'), meta: { title: '从一盘好棋开始', eyebrow: '今日棋境' } },
    { path: '/new-game', component: () => import('@/views/NewGameView.vue'), meta: { title: '新建一盘对局', eyebrow: '对局设置' } },
    { path: '/match/:id', component: () => import('@/views/MatchView.vue'), meta: { title: '对弈进行中', eyebrow: '人机对战' } },
    { path: '/records', component: () => import('@/views/RecordsView.vue'), meta: { title: '棋谱库', eyebrow: '导入、整理与研读' } },
    { path: '/learning', component: () => import('@/views/LearningView.vue'), meta: { title: '学习中心', eyebrow: '棋谱学习与棋风画像' } },
    { path: '/analysis/:matchId?', component: () => import('@/views/AnalysisView.vue'), meta: { title: '复盘分析', eyebrow: '对局复盘' } },
    { path: '/history', component: () => import('@/views/HistoryView.vue'), meta: { title: '历史对局', eyebrow: '记录每一次进步' } },
    { path: '/settings', component: () => import('@/views/SettingsView.vue'), meta: { title: '设置与诊断', eyebrow: '偏好、引擎与数据' } },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

export default router
