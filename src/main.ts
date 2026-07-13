import '@fontsource-variable/bricolage-grotesque'
import '@fontsource-variable/geist'
import '@fontsource-variable/geist-mono'
import './style.css'

import { createApp } from 'vue'

import App from './App.vue'
import { router } from './router'
import { contextInit } from './state'

const FONT_WAIT_MS_MAX = 300

contextInit()

let fonts = Promise.all([
    document.fonts.load('700 14px "Bricolage Grotesque Variable"'),
    document.fonts.load('400 14px "Geist Variable"'),
    document.fonts.load('500 14px "Geist Mono Variable"'),
])

let deadline = new Promise((resolve) => window.setTimeout(resolve, FONT_WAIT_MS_MAX))

Promise.race([fonts, deadline]).then(() => {
    createApp(App).use(router).mount('#app')
})
