<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowDown, ArrowUp, ArrowUpDown, Minus, Search } from '@lucide/vue'

import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import { apiTier, type ChampionRow } from '../api'
import { percent } from '../format'
import { useDelayedLoader } from '../loading'
import { context } from '../state'
import ChampAvatar from '../components/ChampAvatar.vue'
import LoadingPane from '../components/LoadingPane.vue'
import PageHeading from '../components/PageHeading.vue'
import StatusPane from '../components/StatusPane.vue'
import TierBadge from '../components/TierBadge.vue'
import WinMeter from '../components/WinMeter.vue'

type SortKey = 'champion_name' | 'tier' | 'win_rate' | 'pick_rate' | 'games' | 'kda'

const TIER_STEPS: Array<[number, string]> = [
    [0.54, 'S+'],
    [0.52, 'S'],
    [0.505, 'A'],
    [0.49, 'B'],
    [0.47, 'C'],
]

const TIER_RANK: Record<string, number> = { 'S+': 0, 'S': 1, 'A': 2, 'B': 3, 'C': 4, 'D': 5 }

const SORT_COLUMNS: Array<{ key: SortKey, label: string }> = [
    { key: 'win_rate', label: 'Win rate' },
    { key: 'pick_rate', label: 'Pick rate' },
    { key: 'kda', label: 'KDA' },
    { key: 'games', label: 'Games' },
]

const ASCENDING_FIRST = new Set<SortKey>(['champion_name', 'tier'])

let router = useRouter()
let rows = ref<ChampionRow[]>([])
let previousRanks = ref<Map<number, number>>(new Map())
let { loading, showLoader, begin, end } = useDelayedLoader()
let error = ref('')
let filter = ref('')
let sortKey = ref<SortKey>('win_rate')
let sortDescending = ref(true)

let currentRanks = computed(
    () => new Map(rows.value.map((row, index) => [row.champion_id, index + 1])),
)

function movement(championId: number): number | null {
    let current = currentRanks.value.get(championId)
    let previous = previousRanks.value.get(championId)

    if (current === undefined || previous === undefined) return null

    return previous - current
}

let serviceEmpty = computed(() => context.patchesLoaded && context.patches.length === 0)

let visible = computed(() => {
    let needle = filter.value.trim().toLowerCase()
    let matched = needle
        ? rows.value.filter((row) => row.champion_name.toLowerCase().includes(needle))
        : [...rows.value]

    matched.sort((a, b) => {
        let delta

        if (sortKey.value === 'champion_name') {
            delta = a.champion_name.localeCompare(b.champion_name)
        } else if (sortKey.value === 'tier') {
            delta =
                TIER_RANK[tierOf(a)] - TIER_RANK[tierOf(b)]
                || a.champion_name.localeCompare(b.champion_name)
        } else {
            delta = a[sortKey.value] - b[sortKey.value]
        }

        return sortDescending.value ? -delta : delta
    })

    return matched
})

function tierOf(row: ChampionRow): string {
    for (let [floor, label] of TIER_STEPS) {
        if (row.win_rate >= floor) return label
    }

    return 'D'
}

function kdaClass(kda: number): string {
    if (kda >= 4) return 'text-gold font-semibold'
    if (kda >= 3) return 'text-win font-semibold'

    return ''
}

function sortBy(key: SortKey): void {
    if (sortKey.value === key) {
        sortDescending.value = !sortDescending.value
    } else {
        sortKey.value = key
        sortDescending.value = !ASCENDING_FIRST.has(key)
    }
}

async function load(): Promise<void> {
    if (!context.patch) return

    begin()
    error.value = ''

    try {
        rows.value = await apiTier(context.patch, context.queue)
    } catch (raised) {
        error.value = String(raised)
    } finally {
        end()
    }

    previousRanks.value = new Map()

    let index = context.patches.indexOf(context.patch)

    if (index < 0 || index + 1 >= context.patches.length) return

    try {
        let previous = await apiTier(context.patches[index + 1], context.queue)

        previousRanks.value = new Map(previous.map((row, rank) => [row.champion_id, rank + 1]))
    } catch {
        previousRanks.value = new Map()
    }
}

watch(() => context.patch, load, { immediate: true })
</script>

<template>
    <section class="mx-auto max-w-5xl">
        <PageHeading eyebrow="Champion tier list" title="ARAM standings">
            <span v-if="visible.length > 0" class="stat-mono text-xs text-muted-foreground">
                {{ visible.length }} champions · patch {{ context.patch }}
            </span>
        </PageHeading>

        <transition name="fade" mode="out-in">
            <StatusPane v-if="error" key="error" variant="error" :message="error" />
            <StatusPane
                v-else-if="context.patchesFailed"
                key="unreachable"
                variant="error"
                message="Could not reach the stats service. Check the base URL in Settings."
            />
            <StatusPane
                v-else-if="serviceEmpty"
                key="no-patches"
                variant="empty"
                message="The service has no patch data yet. Stats appear once enough matches have been gathered."
            />

            <LoadingPane
                v-else-if="loading && showLoader"
                key="loading"
                class="min-h-[60vh]"
                message="Loading stats…"
            />

            <div v-else-if="loading" key="pending" class="min-h-[60vh]"></div>

            <div v-else key="content">
                <div class="relative mb-4 max-w-72">
                    <Search class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground/60" />
                    <Input
                        v-model="filter"
                        class="bg-white/[0.02] pl-9"
                        placeholder="Filter champions"
                        spellcheck="false"
                        autocomplete="off"
                    />
                </div>

                <StatusPane
                    v-if="rows.length === 0"
                    variant="empty"
                    message="No games for this patch yet. Stats appear once enough matches have been gathered."
                />
                <StatusPane
                    v-else-if="visible.length === 0"
                    variant="empty"
                    :message="`No champion matches “${filter.trim()}”.`"
                />

                <Card v-else class="gap-0 overflow-hidden rounded-xl py-0">
                    <Table>
                        <TableHeader>
                            <TableRow class="bg-white/[0.02] hover:bg-white/[0.02]">
                                <TableHead class="w-12 text-center">#</TableHead>
                                <TableHead>
                                    <button
                                        class="inline-flex cursor-pointer items-center gap-1 select-none hover:text-foreground"
                                        :class="{ 'text-primary': sortKey === 'champion_name' }"
                                        type="button"
                                        @click="sortBy('champion_name')"
                                    >
                                        Champion
                                        <template v-if="sortKey === 'champion_name'">
                                            <ArrowDown v-if="sortDescending" class="size-3" />
                                            <ArrowUp v-else class="size-3" />
                                        </template>
                                        <ArrowUpDown v-else class="size-3 opacity-40" />
                                    </button>
                                </TableHead>
                                <TableHead class="w-16 text-center">
                                    <button
                                        class="inline-flex cursor-pointer items-center gap-1 select-none hover:text-foreground"
                                        :class="{ 'text-primary': sortKey === 'tier' }"
                                        type="button"
                                        @click="sortBy('tier')"
                                    >
                                        Tier
                                        <template v-if="sortKey === 'tier'">
                                            <ArrowDown v-if="sortDescending" class="size-3" />
                                            <ArrowUp v-else class="size-3" />
                                        </template>
                                        <ArrowUpDown v-else class="size-3 opacity-40" />
                                    </button>
                                </TableHead>
                                <TableHead
                                    v-for="column in SORT_COLUMNS"
                                    :key="column.key"
                                    class="text-right"
                                >
                                    <button
                                        class="inline-flex cursor-pointer items-center gap-1 select-none hover:text-foreground"
                                        :class="{ 'text-primary': sortKey === column.key }"
                                        type="button"
                                        @click="sortBy(column.key)"
                                    >
                                        {{ column.label }}
                                        <template v-if="sortKey === column.key">
                                            <ArrowDown v-if="sortDescending" class="size-3" />
                                            <ArrowUp v-else class="size-3" />
                                        </template>
                                        <ArrowUpDown v-else class="size-3 opacity-40" />
                                    </button>
                                </TableHead>
                            </TableRow>
                        </TableHeader>

                        <TableBody>
                            <TableRow
                                v-for="(row, index) in visible"
                                :key="row.champion_id"
                                class="h-[52px] cursor-pointer border-white/[0.04] transition-colors hover:bg-primary/[0.04]"
                                @click="router.push(`/champions/${row.champion_id}`)"
                            >
                                <TableCell class="stat-mono text-center text-muted-foreground">
                                    <span class="inline-flex items-center gap-1">
                                        {{ index + 1 }}
                                        <template v-if="movement(row.champion_id) !== null">
                                            <span
                                                v-if="movement(row.champion_id)! > 0"
                                                class="flex items-center text-[10px] text-win"
                                            >
                                                <ArrowUp class="size-2.5" />{{ movement(row.champion_id) }}
                                            </span>
                                            <span
                                                v-else-if="movement(row.champion_id)! < 0"
                                                class="flex items-center text-[10px] text-loss"
                                            >
                                                <ArrowDown class="size-2.5" />{{ -movement(row.champion_id)! }}
                                            </span>
                                            <Minus v-else class="size-2.5 text-muted-foreground/30" />
                                        </template>
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <span class="flex items-center gap-3">
                                        <ChampAvatar :champion-id="row.champion_id" :name="row.champion_name" />
                                        <span class="font-semibold">{{ row.champion_name }}</span>
                                    </span>
                                </TableCell>
                                <TableCell class="text-center">
                                    <TierBadge :tier="tierOf(row)" />
                                </TableCell>
                                <TableCell><WinMeter :win-rate="row.win_rate" /></TableCell>
                                <TableCell class="stat-mono text-right text-muted-foreground">
                                    {{ percent(row.pick_rate) }}%
                                </TableCell>
                                <TableCell class="stat-mono text-right" :class="kdaClass(row.kda)">
                                    {{ row.kda.toFixed(2) }}
                                </TableCell>
                                <TableCell class="stat-mono text-right text-muted-foreground">
                                    {{ row.games.toLocaleString() }}
                                </TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </Card>
            </div>
        </transition>
    </section>
</template>
