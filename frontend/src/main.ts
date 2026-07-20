import { createPinia } from 'pinia'
import { createApp } from 'vue'

import '@fortawesome/fontawesome-free/css/all.min.css'
import 'vant/es/loading/style'
import 'vant/es/popup/style'
import 'vant/es/switch/style'
import 'vant/es/toast/style'
import 'vant/es/uploader/style'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import 'vue3-video-play/dist/style.css'
import '@/assets/main.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.mount('#app')
