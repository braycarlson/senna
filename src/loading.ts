import { ref, type Ref } from 'vue'

const LOADER_DELAY_MS = 150

export function useDelayedLoader(delayMs = LOADER_DELAY_MS): {
    loading: Ref<boolean>
    showLoader: Ref<boolean>
    begin: () => void
    end: () => void
} {
    let loading = ref(true)
    let showLoader = ref(false)
    let timer = 0

    function begin(): void {
        loading.value = true

        window.clearTimeout(timer)
        timer = window.setTimeout(() => (showLoader.value = loading.value), delayMs)
    }

    function end(): void {
        loading.value = false

        window.clearTimeout(timer)

        showLoader.value = false
    }

    return { loading, showLoader, begin, end }
}
