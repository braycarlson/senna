<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Eraser } from '@lucide/vue'

import { Button } from '@/components/ui/button'
import { type RunePage } from '../api'
import { assets, runeIcon } from '../assets'

const SHARD_GROUPS = [
    { label: 'Offense', ids: [5008, 5005, 5007] },
    { label: 'Flex', ids: [5008, 5010, 5001] },
    { label: 'Defense', ids: [5011, 5013, 5001] },
]

const SHARD_LABELS: Record<number, string> = {
    5001: 'Health Scaling',
    5005: 'Attack Speed',
    5007: 'Ability Haste',
    5008: 'Adaptive Force',
    5010: 'Move Speed',
    5011: 'Health',
    5013: 'Tenacity',
}

let model = defineModel<RunePage | null>({ required: true })

let primaryStyle = ref(0)
let primaryPicks = ref<number[]>([0, 0, 0, 0])
let subStyle = ref(0)
let subPicks = ref<{ row: number, id: number }[]>([])
let shardPicks = ref<number[]>([0, 0, 0])

let primaryTree = computed(() => assets.runeTrees.find((tree) => tree.id === primaryStyle.value))
let subTree = computed(() => assets.runeTrees.find((tree) => tree.id === subStyle.value))

let primaryComplete = computed(
    () => primaryStyle.value > 0 && primaryPicks.value.every((id) => id > 0),
)
let subComplete = computed(() => subStyle.value > 0 && subPicks.value.length === 2)
let shardsComplete = computed(() => shardPicks.value.every((id) => id > 0))

let complete = computed(
    () => primaryComplete.value && subComplete.value && shardsComplete.value,
)

function pageBuild(): RunePage {
    let subIds = [...subPicks.value].sort((a, b) => a.row - b.row).map((pick) => pick.id)

    return {
        primary_style: primaryStyle.value,
        sub_style: subStyle.value,
        perks: [...primaryPicks.value, ...subIds, ...shardPicks.value],
    }
}

function reset(): void {
    primaryStyle.value = 0
    primaryPicks.value = [0, 0, 0, 0]
    subStyle.value = 0
    subPicks.value = []
    shardPicks.value = [0, 0, 0]
}

function hydrate(): void {
    let page = model.value

    if (!page || page.perks.length !== 9) {
        reset()

        return
    }

    primaryStyle.value = page.primary_style
    primaryPicks.value = page.perks.slice(0, 4)
    subStyle.value = page.sub_style
    shardPicks.value = page.perks.slice(6, 9)

    let tree = assets.runeTrees.find((entry) => entry.id === page.sub_style)
    let picks: { row: number, id: number }[] = []

    for (let id of page.perks.slice(4, 6)) {
        let row = tree?.slots.findIndex(
            (slot, index) => index > 0 && slot.some((rune) => rune.id === id),
        )

        if (row != null && row > 0) picks.push({ row, id })
    }

    subPicks.value = picks
}

function pickPrimaryStyle(styleId: number): void {
    if (primaryStyle.value === styleId) return

    primaryStyle.value = styleId
    primaryPicks.value = [0, 0, 0, 0]

    if (subStyle.value === styleId) {
        subStyle.value = 0
        subPicks.value = []
    }
}

function pickPrimaryRune(row: number, id: number): void {
    let picks = [...primaryPicks.value]

    picks[row] = picks[row] === id ? 0 : id

    primaryPicks.value = picks
}

function pickSubStyle(styleId: number): void {
    if (subStyle.value === styleId || styleId === primaryStyle.value) return

    subStyle.value = styleId
    subPicks.value = []
}

function pickSubRune(row: number, id: number): void {
    let picks = subPicks.value.filter((pick) => pick.row !== row)

    if (picks.length === subPicks.value.length || !subPicks.value.some((pick) => pick.id === id)) {
        picks.push({ row, id })
    }

    while (picks.length > 2) picks.shift()

    subPicks.value = picks
}

function pickShard(row: number, id: number): void {
    let picks = [...shardPicks.value]

    picks[row] = picks[row] === id ? 0 : id

    shardPicks.value = picks
}

function subPicked(row: number, id: number): boolean {
    return subPicks.value.some((pick) => pick.row === row && pick.id === id)
}

function clear(): void {
    reset()

    if (model.value !== null) model.value = null
}

watch([model, () => assets.runeTrees], hydrate, { immediate: true })

watch([primaryStyle, primaryPicks, subStyle, subPicks, shardPicks], () => {
    if (!complete.value) return

    let page = pageBuild()

    if (JSON.stringify(page) === JSON.stringify(model.value ?? null)) return

    model.value = page
})
</script>

<template>
    <div v-if="assets.runeTrees.length === 0" class="text-xs text-muted-foreground">
        Loading rune data…
    </div>

    <div v-else class="flex flex-col gap-4">
        <div class="flex items-center justify-between">
            <span
                class="text-xs"
                :class="complete ? 'text-win' : 'text-muted-foreground'"
            >
                {{ complete ? 'Page complete' : 'Pick a full page to save it' }}
            </span>

            <Button size="xs" variant="ghost" :disabled="!model && !primaryStyle" @click="clear">
                <Eraser class="size-3.5" />
                Clear
            </Button>
        </div>

        <div class="grid gap-4 min-[880px]:grid-cols-2">
            <div class="rune-column paths">
                <span class="rune-column-title">Primary path</span>

                <div class="rune-row">
                    <button
                        v-for="tree in assets.runeTrees"
                        :key="tree.id"
                        class="rune-cell w-16"
                        :class="{ selected: primaryStyle === tree.id }"
                        type="button"
                        @click="pickPrimaryStyle(tree.id)"
                    >
                        <img class="size-9" :src="runeIcon(tree.id)" :alt="tree.name" />
                        <span class="rune-name">{{ tree.name }}</span>
                    </button>
                </div>

                <div class="rune-divider"></div>

                <template v-if="primaryTree">
                    <div
                        v-for="(slot, row) in primaryTree.slots"
                        :key="row"
                        class="rune-row"
                    >
                        <button
                            v-for="rune in slot"
                            :key="rune.id"
                            class="rune-cell w-24"
                            :class="{ selected: primaryPicks[row] === rune.id }"
                            type="button"
                            @click="pickPrimaryRune(row, rune.id)"
                        >
                            <img class="size-11" :src="runeIcon(rune.id)" :alt="rune.name" />
                            <span class="rune-name">{{ rune.name }}</span>
                        </button>
                    </div>
                </template>

                <div v-else class="flex flex-1 items-center justify-center">
                    <p class="rune-hint">Pick a path to see its runes.</p>
                </div>
            </div>

            <div class="rune-column paths">
                <span class="rune-column-title">Secondary path</span>

                <div class="rune-row">
                    <button
                        v-for="tree in assets.runeTrees"
                        :key="tree.id"
                        class="rune-cell w-16"
                        :class="{ selected: subStyle === tree.id }"
                        type="button"
                        :disabled="tree.id === primaryStyle"
                        @click="pickSubStyle(tree.id)"
                    >
                        <img class="size-9" :src="runeIcon(tree.id)" :alt="tree.name" />
                        <span class="rune-name">{{ tree.name }}</span>
                    </button>
                </div>

                <div class="rune-divider"></div>

                <template v-if="subTree">
                    <div
                        v-for="(slot, index) in subTree.slots.slice(1)"
                        :key="index"
                        class="rune-row"
                    >
                        <button
                            v-for="rune in slot"
                            :key="rune.id"
                            class="rune-cell w-24"
                            :class="{ selected: subPicked(index + 1, rune.id) }"
                            type="button"
                            @click="pickSubRune(index + 1, rune.id)"
                        >
                            <img class="size-11" :src="runeIcon(rune.id)" :alt="rune.name" />
                            <span class="rune-name">{{ rune.name }}</span>
                        </button>
                    </div>

                </template>

                <div v-else class="flex flex-1 items-center justify-center">
                    <p class="rune-hint">Pick a second path.</p>
                </div>
            </div>
        </div>

        <div class="rune-column">
            <span class="rune-column-title">Shards</span>

            <div class="rune-divider"></div>

            <div class="grid grid-cols-3 gap-4">
                <div
                    v-for="(group, row) in SHARD_GROUPS"
                    :key="group.label"
                    class="flex flex-col items-center gap-2"
                >
                    <span class="text-[11px] tracking-[0.12em] text-muted-foreground/70 uppercase">
                        {{ group.label }}
                    </span>

                    <div class="rune-row">
                        <button
                            v-for="id in group.ids"
                            :key="`${row}-${id}`"
                            class="shard-cell"
                            :class="{ selected: shardPicks[row] === id }"
                            type="button"
                            :title="SHARD_LABELS[id]"
                            @click="pickShard(row, id)"
                        >
                            <img class="size-6" :src="runeIcon(id)" :alt="SHARD_LABELS[id]" />
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
@reference "../style.css";

.rune-column {
    @apply flex flex-col gap-3 rounded-xl bg-white/[0.015] p-4 ring-1 ring-white/[0.03] ring-inset;
}

.rune-column.paths {
    @apply min-h-[38rem];
}

.rune-column-title {
    @apply text-center text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase;
}

.rune-row {
    @apply flex items-start justify-center gap-2;
}

.rune-divider {
    @apply h-px w-full bg-white/[0.05];
}

.rune-hint {
    @apply text-center text-xs text-muted-foreground;
}

.rune-cell {
    @apply flex cursor-pointer flex-col items-center gap-1.5 rounded-lg p-2 transition-all;
}

.rune-cell img {
    @apply rounded-full bg-white/[0.03] opacity-55 ring-1 ring-white/[0.06] grayscale transition-all;
}

.rune-cell:hover img {
    @apply opacity-90 grayscale-0;
}

.rune-cell.selected {
    @apply bg-primary/10;
}

.rune-cell.selected img {
    @apply opacity-100 ring-2 ring-primary grayscale-0;
}

.rune-name {
    @apply line-clamp-2 h-8 text-center text-[11px] leading-snug text-muted-foreground;
}

.rune-cell.selected .rune-name {
    @apply text-primary;
}

.rune-cell:disabled {
    @apply cursor-default opacity-25;
}

.rune-cell:disabled:hover img {
    @apply opacity-55 grayscale;
}

.shard-cell {
    @apply flex cursor-pointer items-center justify-center rounded-full p-1.5 transition-all;
}

.shard-cell img {
    @apply rounded-full bg-white/[0.03] opacity-55 ring-1 ring-white/[0.06] grayscale transition-all;
}

.shard-cell:hover img {
    @apply opacity-90 grayscale-0;
}

.shard-cell.selected {
    @apply bg-primary/10;
}

.shard-cell.selected img {
    @apply opacity-100 ring-2 ring-primary grayscale-0;
}
</style>
