<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
    apiBuild,
    apiChampions,
    apiMatchups,
    apiSynergies,
    type ChampionBuild,
    type ChampionRow,
    type PairRow,
} from '../api'
import { assetsEnsureBuildIcons } from '../assets'
import { percent, thousands } from '../format'
import { context } from '../state'
import BuildPanels from '../components/BuildPanels.vue'
import ChampAvatar from '../components/ChampAvatar.vue'
import LoadoutEditor from '../components/LoadoutEditor.vue'
import StatusPane from '../components/StatusPane.vue'
import WinMeter from '../components/WinMeter.vue'
import WinSparkline from '../components/WinSparkline.vue'

let props = defineProps<{ id: string }>()

const HISTORY_PATCHES_MAX = 6

let championId = computed(() => Number(props.id))
let build = ref<ChampionBuild | null>(null)
let summary = ref<ChampionRow | null>(null)
let matchups = ref<PairRow[]>([])
let synergies = ref<PairRow[]>([])
let history = ref<Array<{ label: string, value: number }>>([])
let loading = ref(false)
let error = ref('')

let historyDelta = computed(() => {
    if (history.value.length < 2) return null

    let latest = history.value[history.value.length - 1]
    let before = history.value[history.value.length - 2]

    return latest.value - before.value
})

let bestPage = computed(() => build.value?.rune_pages[0] ?? null)
let boots = computed(() => build.value?.boots.slice(0, 3) ?? [])
let coreBuilds = computed(() => build.value?.core_builds.slice(0, 4) ?? [])
let runeStyles = computed(() => build.value?.rune_styles.slice(0, 4) ?? [])
let skillOrders = computed(() => build.value?.skill_orders.slice(0, 3) ?? [])
let spellPairs = computed(() => build.value?.summoner_spells.slice(0, 3) ?? [])
let startingItems = computed(() => build.value?.starting_items.slice(0, 4) ?? [])

let hardestMatchups = computed(() =>
    [...matchups.value].sort((a, b) => a.win_rate - b.win_rate).slice(0, 8),
)

let bestAllies = computed(() =>
    [...synergies.value].sort((a, b) => b.win_rate - a.win_rate).slice(0, 8),
)

let serviceEmpty = computed(() => context.patchesLoaded && context.patches.length === 0)

let buildEmpty = computed(
    () =>
        startingItems.value.length === 0
        && coreBuilds.value.length === 0
        && !bestPage.value
        && spellPairs.value.length === 0
        && boots.value.length === 0
        && skillOrders.value.length === 0
        && hardestMatchups.value.length === 0
        && bestAllies.value.length === 0,
)

let hasCounters = computed(() => hardestMatchups.value.length > 0 || bestAllies.value.length > 0)

async function loadHistory(): Promise<void> {
    let patches = context.patches.slice(0, HISTORY_PATCHES_MAX)

    let lists = await Promise.all(
        patches.map((patch) => apiChampions(patch, context.queue).catch(() => [])),
    )

    let points = []

    for (let [index, list] of lists.entries()) {
        let row = list.find((entry) => entry.champion_id === championId.value)

        if (row) points.push({ label: patches[index], value: row.win_rate })
    }

    history.value = points.reverse()
}

async function load(): Promise<void> {
    if (!context.patch || Number.isNaN(championId.value)) return

    assetsEnsureBuildIcons()

    loading.value = true
    error.value = ''

    try {
        let [buildResult, championsResult, matchupResult, synergyResult] = await Promise.all([
            apiBuild(championId.value, context.patch, context.queue),
            apiChampions(context.patch, context.queue),
            apiMatchups(championId.value, context.patch, context.queue),
            apiSynergies(championId.value, context.patch, context.queue),
        ])

        build.value = buildResult
        summary.value = championsResult.find((row) => row.champion_id === championId.value) ?? null
        matchups.value = matchupResult
        synergies.value = synergyResult

        void loadHistory()
    } catch (raised) {
        error.value = String(raised)
    } finally {
        loading.value = false
    }
}

watch([() => context.patch, championId], load, { immediate: true })
</script>

<template>
    <section class="mx-auto max-w-6xl">
        <StatusPane v-if="error" variant="error" :message="error" />
        <StatusPane
            v-else-if="context.patchesFailed"
            variant="error"
            message="Could not reach the stats service. Check the base URL in Settings."
        />
        <StatusPane
            v-else-if="serviceEmpty"
            variant="empty"
            message="The service has no patch data yet, so there is no build to show for this champion."
        />
        <StatusPane v-else-if="loading || !build" variant="loading" message="Reading the mist…" />

        <template v-else>
            <Card class="rise-in mb-5 block overflow-hidden rounded-xl p-0">
                <div class="flex flex-wrap items-center gap-6 bg-gradient-to-r from-primary/[0.12] via-transparent to-transparent p-6">
                    <div class="flex items-center gap-5">
                        <ChampAvatar
                            class="ring-2 ring-primary/40"
                            :champion-id="build.champion_id"
                            :name="build.champion_name"
                            size="lg"
                        />

                        <div class="flex flex-col gap-1.5">
                            <h1 class="text-[30px] leading-none font-bold">
                                {{ build.champion_name ?? `Champion ${build.champion_id}` }}
                            </h1>

                            <div class="flex items-center gap-2">
                                <Badge class="bg-primary font-semibold text-primary-foreground">ARAM</Badge>
                                <Badge variant="secondary">Patch {{ build.patch }}</Badge>
                            </div>
                        </div>
                    </div>

                    <div v-if="summary" class="ml-auto flex items-center gap-6">
                        <div class="flex flex-col items-end gap-0.5">
                            <span class="flex items-center gap-1.5">
                                <span
                                    v-if="historyDelta !== null"
                                    class="stat-mono text-[11px]"
                                    :class="historyDelta >= 0 ? 'text-win' : 'text-loss'"
                                >
                                    {{ historyDelta >= 0 ? '+' : '' }}{{ percent(historyDelta) }}
                                </span>
                                <span
                                    class="stat-mono text-lg font-semibold"
                                    :class="summary.win_rate >= 0.5 ? 'text-win' : 'text-loss'"
                                >
                                    {{ percent(summary.win_rate) }}%
                                </span>
                            </span>
                            <span class="text-[11px] text-muted-foreground">win rate</span>
                        </div>

                        <div class="flex flex-col items-end gap-0.5">
                            <span class="stat-mono text-lg font-semibold">
                                {{ percent(summary.pick_rate) }}%
                            </span>
                            <span class="text-[11px] text-muted-foreground">pick rate</span>
                        </div>

                        <div class="flex flex-col items-end gap-0.5">
                            <span class="stat-mono text-lg font-semibold">{{ summary.kda.toFixed(2) }}</span>
                            <span class="text-[11px] text-muted-foreground">KDA</span>
                        </div>

                        <div class="flex flex-col items-end gap-0.5">
                            <span class="stat-mono text-lg font-semibold">
                                {{ thousands(summary.damage_average) }}k
                            </span>
                            <span class="text-[11px] text-muted-foreground">damage</span>
                        </div>

                        <div class="flex flex-col items-end gap-0.5">
                            <span class="stat-mono text-lg font-semibold">
                                {{ thousands(summary.gold_average) }}k
                            </span>
                            <span class="text-[11px] text-muted-foreground">gold</span>
                        </div>

                        <div class="flex flex-col items-end gap-0.5">
                            <span class="stat-mono text-lg font-semibold">
                                {{ summary.games.toLocaleString() }}
                            </span>
                            <span class="text-[11px] text-muted-foreground">games</span>
                        </div>

                        <WinSparkline v-if="history.length > 1" :points="history" />
                    </div>
                </div>

                <div
                    v-if="hardestMatchups.length > 0 || bestAllies.length > 0"
                    class="flex flex-wrap items-center gap-8 border-t border-white/[0.04] px-6 py-3"
                >
                    <span v-if="hardestMatchups.length > 0" class="flex items-center gap-2.5">
                        <span class="text-xs text-muted-foreground">Struggles vs</span>
                        <span
                            v-for="pair in hardestMatchups.slice(0, 3)"
                            :key="pair.champion_id"
                            class="flex items-center gap-1"
                            :title="`${pair.champion_name ?? pair.champion_id} · ${percent(pair.win_rate)}%`"
                        >
                            <ChampAvatar :champion-id="pair.champion_id" :name="pair.champion_name" size="sm" />
                            <span class="stat-mono text-xs text-loss">
                                {{ percent(pair.win_rate, 0) }}%
                            </span>
                        </span>
                    </span>

                    <span v-if="bestAllies.length > 0" class="flex items-center gap-2.5">
                        <span class="text-xs text-muted-foreground">Best with</span>
                        <span
                            v-for="pair in bestAllies.slice(0, 3)"
                            :key="pair.champion_id"
                            class="flex items-center gap-1"
                            :title="`${pair.champion_name ?? pair.champion_id} · ${percent(pair.win_rate)}%`"
                        >
                            <ChampAvatar :champion-id="pair.champion_id" :name="pair.champion_name" size="sm" />
                            <span class="stat-mono text-xs text-win">
                                {{ percent(pair.win_rate, 0) }}%
                            </span>
                        </span>
                    </span>
                </div>
            </Card>

            <Tabs class="rise-in" default-value="build">
                <TabsList class="mb-4 bg-white/[0.03]">
                    <TabsTrigger value="build">Build</TabsTrigger>
                    <TabsTrigger value="runes" :disabled="!bestPage && runeStyles.length === 0">
                        Runes
                    </TabsTrigger>
                    <TabsTrigger value="counters" :disabled="!hasCounters">Counters</TabsTrigger>
                    <TabsTrigger value="loadout">Loadout</TabsTrigger>
                </TabsList>

                <TabsContent v-if="buildEmpty" value="build">
                    <StatusPane
                        variant="empty"
                        message="No data for this champion on this patch yet. Stats appear once enough matches have been gathered."
                    />
                </TabsContent>

                <TabsContent v-else value="build">
                    <BuildPanels :build="build" :sections="['core', 'starting', 'spells', 'skills', 'path']" />
                </TabsContent>

                <TabsContent value="runes">
                    <BuildPanels :build="build" :sections="['page', 'styles']" />
                </TabsContent>

                <TabsContent value="counters" class="grid gap-4 lg:grid-cols-2">
                    <Card v-if="hardestMatchups.length > 0" class="panel">
                        <h2 class="panel-title">Hardest matchups</h2>

                        <div v-for="pair in hardestMatchups" :key="pair.champion_id" class="entry">
                            <span class="flex flex-1 items-center gap-2.5">
                                <ChampAvatar :champion-id="pair.champion_id" :name="pair.champion_name" size="sm" />
                                <span class="font-medium">{{ pair.champion_name ?? pair.champion_id }}</span>
                            </span>

                            <span class="games">{{ pair.games.toLocaleString() }}</span>
                            <WinMeter :win-rate="pair.win_rate" />
                        </div>
                    </Card>

                    <Card v-if="bestAllies.length > 0" class="panel">
                        <h2 class="panel-title">Best allies</h2>

                        <div v-for="pair in bestAllies" :key="pair.champion_id" class="entry">
                            <span class="flex flex-1 items-center gap-2.5">
                                <ChampAvatar :champion-id="pair.champion_id" :name="pair.champion_name" size="sm" />
                                <span class="font-medium">{{ pair.champion_name ?? pair.champion_id }}</span>
                            </span>

                            <span class="games">{{ pair.games.toLocaleString() }}</span>
                            <WinMeter :win-rate="pair.win_rate" />
                        </div>
                    </Card>
                </TabsContent>

                <TabsContent value="loadout">
                    <LoadoutEditor :champion-id="championId" />
                </TabsContent>
            </Tabs>
        </template>
    </section>
</template>

