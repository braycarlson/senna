<script setup lang="ts">
import { listen } from '@tauri-apps/api/event'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
    Check,
    CircleAlert,
    Crown,
    House,
    Info,
    ScrollText,
    Settings,
    SlidersHorizontal,
    Snowflake,
    TrendingUp,
    Users,
} from '@lucide/vue'

import { lcuCurrentSummoner, lcuStatus, type CurrentSummoner } from './api'
import { context } from './state'
import { toast, toastShow } from './toast'

interface AppliedPayload {
    champion_name: string | null
    champion_id: number
    runes: boolean
    spells: boolean
    items: boolean
    skin: boolean
}

const NAV_ITEMS = [
    { icon: House, label: 'Home', to: '/', exact: true },
    { icon: Crown, label: 'Tier list', to: '/tier', exact: false },
    { icon: TrendingUp, label: 'Movers', to: '/movers', exact: false },
    { icon: Users, label: 'Players', to: '/player', exact: false },
    { icon: SlidersHorizontal, label: 'Preferences', to: '/preferences', exact: false },
    { icon: ScrollText, label: 'Event log', to: '/events', exact: false },
    { icon: Settings, label: 'Settings', to: '/settings', exact: false },
    { icon: Info, label: 'About', to: '/about', exact: false },
]

let route = useRoute()

function navActive(item: (typeof NAV_ITEMS)[number]): boolean {
    return item.exact ? route.path === item.to : route.path.startsWith(item.to)
}

let lcu = ref('disconnected')
let summoner = ref<CurrentSummoner | null>(null)

let lcuLabel = computed(() => {
    if (lcu.value === 'connected' && summoner.value) {
        return `${summoner.value.gameName}#${summoner.value.tagLine}`
    }

    return `League ${lcu.value}`
})

async function syncSummoner(): Promise<void> {
    if (lcu.value !== 'connected') {
        summoner.value = null

        return
    }

    try {
        summoner.value = await lcuCurrentSummoner()
    } catch {
        summoner.value = null
    }
}

onMounted(async () => {
    await Promise.all([
        listen<string>('lcu-status', (event) => {
            lcu.value = event.payload

            void syncSummoner()
        }),

        listen<AppliedPayload>('lcu-applied', (event) => {
            let target = event.payload.champion_name ?? `champion ${event.payload.champion_id}`
            let parts = []

            if (event.payload.runes) parts.push('runes')
            if (event.payload.spells) parts.push('spells')
            if (event.payload.items) parts.push('item set')
            if (event.payload.skin) parts.push('skin')
            if (parts.length === 0) return

            toastShow(`Applied ${parts.join(' + ')} for ${target}`, 'ok')
        }),

        listen<string>('lcu-error', (event) => {
            toastShow(`Champ select: ${event.payload}`, 'error')
        }),

        listen('lcu-accepted', () => {
            toastShow('Accepted the match.', 'ok')
        }),
    ])

    lcu.value = await lcuStatus()

    await syncSummoner()
})
</script>

<template>
    <div class="flex h-full">
        <aside class="flex w-56 shrink-0 flex-col border-r border-border/70 bg-[#0d0a17]">
            <div class="flex items-center gap-3 px-5 pt-6 pb-7">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/15 ring-1 ring-primary/30 ring-inset">
                    <Snowflake class="size-4 text-primary" />
                </span>

                <span class="font-display text-lg leading-none font-bold tracking-wide">senna</span>
            </div>

            <nav class="flex flex-col gap-0.5 px-3">
                <router-link
                    v-for="item in NAV_ITEMS"
                    :key="item.to"
                    :to="item.to"
                    class="nav-link"
                    :class="{ 'nav-active': navActive(item) }"
                >
                    <component :is="item.icon" class="size-4" />
                    {{ item.label }}
                </router-link>
            </nav>

            <div class="mt-auto flex flex-col gap-2.5 px-4 pb-5">
                <div
                    v-if="context.patch"
                    class="flex items-center justify-between rounded-lg bg-white/[0.03] px-3 py-2.5 ring-1 ring-white/[0.04] ring-inset"
                >
                    <span class="text-xs text-muted-foreground">Patch</span>
                    <span class="text-xs font-medium">{{ context.patch }}</span>
                </div>

                <div class="flex items-center gap-2 rounded-lg bg-white/[0.03] px-3 py-2.5 ring-1 ring-white/[0.04] ring-inset">
                    <span
                        class="size-1.5 shrink-0 rounded-full"
                        :class="{
                            'bg-win shadow-[0_0_8px_var(--jade)]': lcu === 'connected',
                            'bg-gold': lcu === 'searching',
                            'bg-muted-foreground/40': lcu !== 'connected' && lcu !== 'searching',
                        }"
                    ></span>
                    <span
                        class="truncate text-xs text-muted-foreground"
                        :class="{ capitalize: !summoner }"
                        :title="lcuLabel"
                    >
                        {{ lcuLabel }}
                    </span>
                </div>
            </div>
        </aside>

        <main class="min-w-0 flex-1 overflow-y-auto [scrollbar-gutter:stable]">
            <div class="px-8 py-8">
                <router-view v-slot="{ Component }">
                    <transition name="page" mode="out-in">
                        <component :is="Component" />
                    </transition>
                </router-view>
            </div>
        </main>

        <transition name="toast">
            <div
                v-if="toast.current"
                class="fixed right-6 bottom-6 z-10 flex max-w-md items-center gap-2.5 rounded-lg border bg-popover px-4 py-3 shadow-xl shadow-black/40"
                :class="toast.current.kind === 'error' ? 'border-loss/40' : 'border-primary/30'"
            >
                <CircleAlert v-if="toast.current.kind === 'error'" class="size-4 shrink-0 text-loss" />
                <Check v-else class="size-4 shrink-0 text-win" />
                <span class="text-[13px]">{{ toast.current.text }}</span>
            </div>
        </transition>
    </div>
</template>

<style scoped>
@reference "./style.css";

.nav-link {
    @apply relative flex items-center gap-2.5 rounded-lg px-3 py-2 font-medium text-muted-foreground transition-colors;
}

.nav-link:hover {
    @apply bg-white/[0.04] text-foreground;
}

.nav-link.nav-active {
    @apply bg-primary/10 text-primary;
}

.nav-link.nav-active::before {
    content: '';
    @apply absolute top-1/2 left-0 h-4 w-0.5 -translate-y-1/2 rounded-full bg-primary;
}

.toast-enter-active,
.toast-leave-active {
    transition: opacity 200ms ease, transform 200ms ease;
}

.toast-enter-from,
.toast-leave-to {
    opacity: 0;
    transform: translateY(10px);
}
</style>
