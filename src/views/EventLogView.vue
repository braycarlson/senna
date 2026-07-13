<script setup lang="ts">
import { listen } from '@tauri-apps/api/event'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ArrowLeftRight, Crosshair, Flag, Sparkles, Swords, Zap } from '@lucide/vue'

import { Card } from '@/components/ui/card'
import { apiChampions, lcuEvents, type LcuEvent } from '../api'
import { assetsEnsureBuildIcons, championSquare, spellIcon } from '../assets'
import { SPELL_NAMES } from '../spells'
import { context } from '../state'
import PageHeading from '../components/PageHeading.vue'
import StatusPane from '../components/StatusPane.vue'

const KIND_ICONS = {
    phase: Flag,
    assign: Swords,
    champion: Swords,
    hover: Crosshair,
    trade: ArrowLeftRight,
    spells: Zap,
    apply: Sparkles,
}

let events = ref<LcuEvent[]>([])
let championNames = ref<Record<number, string>>({})
let unlisten: (() => void) | null = null

let eventsReversed = computed(() => [...events.value].reverse())

function championName(id?: number): string {
    if (!id) return 'a champion'

    return championNames.value[id] ?? `Champion ${id}`
}

function timeOf(ts: number): string {
    return new Date(ts).toLocaleTimeString(undefined, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    })
}

function appliedParts(event: LcuEvent): string {
    let parts = []

    if (event.runes) parts.push('runes')
    if (event.spells) parts.push('spells')
    if (event.items) parts.push('item set')
    if (event.skin) parts.push('skin')

    return parts.join(' + ')
}

async function loadNames(): Promise<void> {
    if (!context.patch) return

    try {
        let rows = await apiChampions(context.patch, context.queue)
        let names: Record<number, string> = {}

        for (let row of rows) names[row.champion_id] = row.champion_name

        championNames.value = names
    } catch {
        championNames.value = {}
    }
}

onMounted(async () => {
    assetsEnsureBuildIcons()

    void loadNames()

    if ('__TAURI_INTERNALS__' in window) {
        unlisten = await listen<LcuEvent>('lcu-event', (event) => {
            events.value.push(event.payload)
        })

        events.value = await lcuEvents().catch(() => [])
    }
})

onUnmounted(() => {
    unlisten?.()
})
</script>

<template>
    <section class="mx-auto max-w-3xl">
        <PageHeading eyebrow="Champ select activity" title="Event log">
            <span v-if="events.length > 0" class="stat-mono text-xs text-muted-foreground">
                {{ events.length }} events
            </span>
        </PageHeading>

        <StatusPane
            v-if="events.length === 0"
            variant="empty"
            message="Nothing yet. Champ select activity shows up here: picks, hovers, trades, spell swaps, and everything senna applies."
        />

        <Card v-else class="rise-in block gap-0 overflow-hidden rounded-xl py-0">
            <div class="flex flex-col">
                <div
                    v-for="(event, index) in eventsReversed"
                    :key="`${event.ts}-${index}`"
                    class="flex items-center gap-3 border-t border-white/[0.04] px-4 py-2.5 first:border-t-0"
                    :class="{ 'bg-primary/[0.04]': event.self }"
                >
                    <span class="stat-mono w-20 shrink-0 text-xs text-muted-foreground/70">
                        {{ timeOf(event.ts) }}
                    </span>

                    <component
                        :is="KIND_ICONS[event.kind] ?? Flag"
                        class="size-4 shrink-0"
                        :class="event.kind === 'apply' ? 'text-primary' : 'text-muted-foreground/60'"
                    />

                    <span v-if="event.kind === 'phase'" class="text-[13px] font-medium">
                        {{ event.note }}
                    </span>

                    <span v-else-if="event.kind === 'apply'" class="flex items-center gap-2 text-[13px]">
                        <img
                            v-if="event.champion_id"
                            class="size-6 rounded-md bg-layer"
                            :src="championSquare(event.champion_id)"
                            alt=""
                        />
                        senna applied {{ appliedParts(event) }} for
                        {{ championName(event.champion_id) }}
                    </span>

                    <span v-else-if="event.kind === 'trade'" class="flex items-center gap-2 text-[13px]">
                        <span class="font-medium">{{ event.player }}</span>
                        <img
                            v-if="event.champion_id"
                            class="size-6 rounded-md bg-layer"
                            :src="championSquare(event.champion_id)"
                            alt=""
                        />
                        traded with
                        <span class="font-medium">{{ event.other }}</span>
                        <img
                            v-if="event.other_champion_id"
                            class="size-6 rounded-md bg-layer"
                            :src="championSquare(event.other_champion_id)"
                            alt=""
                        />
                    </span>

                    <span v-else-if="event.kind === 'spells'" class="flex items-center gap-2 text-[13px]">
                        <span class="font-medium">{{ event.player }}</span>
                        set
                        <template v-if="event.spells">
                            <img class="size-5 rounded-sm bg-layer" :src="spellIcon(event.spells[0])" alt="" />
                            {{ SPELL_NAMES[event.spells[0]] ?? event.spells[0] }} +
                            <img class="size-5 rounded-sm bg-layer" :src="spellIcon(event.spells[1])" alt="" />
                            {{ SPELL_NAMES[event.spells[1]] ?? event.spells[1] }}
                        </template>
                    </span>

                    <span v-else class="flex items-center gap-2 text-[13px]">
                        <span class="font-medium">{{ event.player }}</span>
                        {{ event.kind === 'hover' ? 'is hovering' : event.kind === 'assign' ? 'is on' : 'switched to' }}
                        <img
                            v-if="event.champion_id"
                            class="size-6 rounded-md bg-layer"
                            :src="championSquare(event.champion_id)"
                            alt=""
                        />
                        {{ championName(event.champion_id) }}
                    </span>
                </div>
            </div>
        </Card>
    </section>
</template>
