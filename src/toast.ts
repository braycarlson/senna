import { reactive } from 'vue'

interface ToastEntry {
    text: string
    kind: 'ok' | 'error'
}

export let toast = reactive<{ current: ToastEntry | null }>({ current: null })

let timer = 0

export function toastShow(text: string, kind: 'ok' | 'error' = 'ok', durationMs = 6000): void {
    toast.current = { text, kind }

    window.clearTimeout(timer)
    timer = window.setTimeout(() => (toast.current = null), durationMs)
}
