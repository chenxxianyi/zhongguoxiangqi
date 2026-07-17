import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/main.css'
import './styles/tokens.css'
import './styles/motion.css'
import './styles/polish.css'
import './styles/board.css'

createApp(App).use(createPinia()).use(router).mount('#app')
