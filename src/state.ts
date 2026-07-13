import { reactive } from 'vue'

import { apiPatches } from './api'

export const QUEUE_ARAM = 450

interface ContextState {
    patch: string
    patches: string[]
    patchesFailed: boolean
    patchesLoaded: boolean
    queue: number
}

export let context = reactive<ContextState>({
    patch: '',
    patches: [],
    patchesFailed: false,
    patchesLoaded: false,
    queue: QUEUE_ARAM,
})

const CONTEXT_RETRY_MS = 5000

let contextLoading = false
let retryTimer: number | undefined

export async function contextInit(): Promise<void> {
    if (contextLoading) return

    contextLoading = true

    if (retryTimer !== undefined) {
        clearTimeout(retryTimer)
        retryTimer = undefined
    }

    context.patchesFailed = false
    context.patchesLoaded = false

    try {
        let patches = await apiPatches()

        context.patches = patches
        context.patchesLoaded = true

        if (!context.patch && patches.length > 0) {
            context.patch = patches[0]
        }
    } catch (error) {
        context.patchesFailed = true

        console.error('patch list load failed', error)

        retryTimer = window.setTimeout(() => {
            void contextInit()
        }, CONTEXT_RETRY_MS)
    } finally {
        contextLoading = false
    }
}
