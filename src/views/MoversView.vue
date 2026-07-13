<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { TrendingDown, TrendingUp } from '@lucide/vue'

import { Card } from '@/components/ui/card'
import { apiChampions, type ChampionRow } from '../api'
import { percent } from '../format'
import { context } from '../state'
import ChampAvatar from '../components/ChampAvatar.vue'
import PageHeading from '../components/PageHeading.vue'
import StatusPane from '../components/StatusPane.vue'

const GAMES_MIN = 20
const MOVERS_MAX = 8

interface Mover {
    row: ChampionRow
    previous: number
    delta: number
}

let router = useRouter()
let loading = ref(true)
let error = ref('')
let movers = ref<Mover[]>([])

let previousPatch = computed(() => {
    let index = context.patches.indexOf(context.patch)

    if (index < 0 || index + 1 >= context.patches.length) return ''

    return context.patches[index + 1]
})

let risers = computed(() =>
    [...movers.value].sort((a, b) => b.delta - a.delta).slice(0, MOVERS_MAX),
)

let fallers = computed(() =>
    [...movers.value].sort((a, b) => a.delta - b.delta).slice(0, MOVERS_MAX),
)

async function load(): Promise<void> {
    if (!context.patch || !previousPatch.value) {
        movers.value = []
        loading.value = false

        return
    }

    loading.value = true
    error.value = ''

    try {
        let [current, previous] = await Promise.all([
            apiChampions(context.patch, context.queue),
            apiChampions(previousPatch.value, context.queue),
        ])

        let previousRates = new Map(
            previous
                .filter((row) => row.games >= GAMES_MIN)
                .map((row) => [row.champion_id, row.win_rate]),
        )

        movers.value = current
            .filter((row) => row.games >= GAMES_MIN && previousRates.has(row.champion_id))
            .map((row) => {
                let before = previousRates.get(row.champion_id) ?? 0

                return { row, previous: before, delta: row.win_rate - before }
            })
    } catch (raised) {
        error.value = String(raised)
    } finally {
        loading.value = false
    }
}

function deltaLabel(delta: number): string {
    let points = percent(delta)

    return delta >= 0 ? `+${points}` : points
}

watch(() => context.patch, load, { immediate: true })
</script>

<template>
    <section class="mx-auto max-w-5xl">
        <PageHeading eyebrow="Patch shifts" title="Movers">
            <span v-if="previousPatch" class="stat-mono text-xs text-muted-foreground">
                {{ previousPatch }} -> {{ context.patch }}
            </span>
        </PageHeading>

        <StatusPane v-if="error" variant="error" :message="error" />
        <StatusPane v-else-if="loading" variant="loading" message="Comparing patches…" />
        <StatusPane
            v-else-if="!previousPatch"
            variant="empty"
            message="Movers need two patches of data. They appear once a second patch has matches."
        />
        <StatusPane
            v-else-if="movers.length === 0"
            variant="empty"
            :message="`No champion has ${GAMES_MIN}+ games on both patches yet.`"
        />

        <div v-else class="rise-in grid items-start gap-4 min-[900px]:grid-cols-2">
            <Card class="panel">
                <h2 class="panel-title flex items-center gap-1.5">
                    <TrendingUp class="size-3.5 text-win" />
                    Biggest risers
                </h2>

                <div
                    v-for="mover in risers"
                    :key="mover.row.champion_id"
                    class="entry cursor-pointer"
                    @click="router.push(`/champions/${mover.row.champion_id}`)"
                >
                    <span class="flex flex-1 items-center gap-2.5">
                        <ChampAvatar
                            :champion-id="mover.row.champion_id"
                            :name="mover.row.champion_name"
                            size="sm"
                        />
                        <span class="font-medium">{{ mover.row.champion_name }}</span>
                    </span>

                    <span class="stat-mono text-xs text-muted-foreground">
                        {{ percent(mover.previous) }}% ->
                        {{ percent(mover.row.win_rate) }}%
                    </span>

                    <span class="stat-mono min-w-12 rounded-md bg-win/10 px-1.5 py-0.5 text-right text-xs font-medium text-win">
                        {{ deltaLabel(mover.delta) }}
                    </span>
                </div>
            </Card>

            <Card class="panel">
                <h2 class="panel-title flex items-center gap-1.5">
                    <TrendingDown class="size-3.5 text-loss" />
                    Biggest fallers
                </h2>

                <div
                    v-for="mover in fallers"
                    :key="mover.row.champion_id"
                    class="entry cursor-pointer"
                    @click="router.push(`/champions/${mover.row.champion_id}`)"
                >
                    <span class="flex flex-1 items-center gap-2.5">
                        <ChampAvatar
                            :champion-id="mover.row.champion_id"
                            :name="mover.row.champion_name"
                            size="sm"
                        />
                        <span class="font-medium">{{ mover.row.champion_name }}</span>
                    </span>

                    <span class="stat-mono text-xs text-muted-foreground">
                        {{ percent(mover.previous) }}% ->
                        {{ percent(mover.row.win_rate) }}%
                    </span>

                    <span class="stat-mono min-w-12 rounded-md bg-loss/10 px-1.5 py-0.5 text-right text-xs font-medium text-loss">
                        {{ deltaLabel(mover.delta) }}
                    </span>
                </div>
            </Card>
        </div>
    </section>
</template>
