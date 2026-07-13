<script setup lang="ts">
import { listen } from '@tauri-apps/api/event'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { Dices, Sparkles } from '@lucide/vue'

import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import {
    lcuApplyStatsRunes,
    lcuBenchSwap,
    lcuCurrentPage,
    lcuSelection,
    lcuPickableSkins,
    lcuReroll,
    lcuSetSkin,
    lcuSetSpells,
    type LiveSelection,
    type RunePageInfo,
} from '../api'
import {
    assetsEnsureBuildIcons,
    championSquare,
    fetchChampionSkins,
    spellIcon,
    type ChampionSkin,
} from '../assets'
import { ARAM_SPELLS } from '../spells'

let selection = ref<LiveSelection | null>(null)
let runePage = ref<RunePageInfo | null>(null)
let skins = ref<ChampionSkin[]>([])
let pickable = ref<Set<number>>(new Set())
let applyingRunes = ref(false)
let error = ref('')

let unlisteners: Array<() => void> = []

let active = computed(
    () => selection.value !== null && selection.value.in_select && selection.value.champion_id > 0,
)

let skinOptions = computed(() =>
    skins.value.filter((skin) => pickable.value.has(skin.id)),
)

async function loadSkinOptions(): Promise<void> {
    if (!active.value || !selection.value) {
        skins.value = []
        pickable.value = new Set()

        return
    }

    let championId = selection.value.champion_id

    let [catalog, pickableIds] = await Promise.all([
        fetchChampionSkins(championId).catch(() => []),
        lcuPickableSkins().catch(() => []),
    ])

    skins.value = catalog
    pickable.value = new Set(pickableIds)
}

async function run(action: () => Promise<unknown>): Promise<void> {
    error.value = ''

    try {
        await action()
    } catch (raised) {
        error.value = String(raised)
    }
}

function pickSlot(slot: 'd' | 'f', value: unknown): void {
    if (!selection.value) return

    let picked = Number(value)
    let other = slot === 'd' ? selection.value.spell_f : selection.value.spell_d

    if (!picked || picked === other) return

    let pair: [number, number] = slot === 'd' ? [picked, other] : [other, picked]

    if (!pair[0] || !pair[1]) return

    void run(() => lcuSetSpells(pair[0], pair[1]))
}

function pickSkin(skinId: number): void {
    void run(() => lcuSetSkin(skinId))
}

function swap(championId: number): void {
    void run(() => lcuBenchSwap(championId))
}

function requestReroll(): void {
    void run(() => lcuReroll())
}

async function reapplyRunes(): Promise<void> {
    if (!selection.value) return

    applyingRunes.value = true

    await run(() => lcuApplyStatsRunes(selection.value!.champion_id))

    runePage.value = await lcuCurrentPage().catch(() => null)
    applyingRunes.value = false
}

watch(
    () => [selection.value?.champion_id, selection.value?.in_select],
    () => void loadSkinOptions(),
)

onMounted(async () => {
    assetsEnsureBuildIcons()

    if (!('__TAURI_INTERNALS__' in window)) return

    unlisteners = await Promise.all([
        listen<LiveSelection>('lcu-selection', (event) => {
            selection.value = event.payload
        }),

        listen<RunePageInfo>('lcu-runepage', (event) => {
            runePage.value = event.payload
        }),
    ])

    selection.value = await lcuSelection().catch(() => null)
    runePage.value = await lcuCurrentPage().catch(() => null)
})

onUnmounted(() => {
    for (let unlisten of unlisteners) unlisten()
})
</script>

<template>
    <Card v-if="active && selection" class="block gap-0 overflow-hidden rounded-xl py-0">
        <div class="flex items-center justify-between border-b bg-white/[0.02] px-4 py-3">
            <h2 class="font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                Champ select controls
            </h2>

            <span v-if="runePage?.name" class="stat-mono max-w-64 truncate text-xs text-muted-foreground">
                Runes: {{ runePage.name }}
            </span>
        </div>

        <div class="flex flex-col gap-4 p-4">
            <p v-if="error" class="text-[13px] text-loss">{{ error }}</p>

            <div class="flex flex-wrap items-end gap-4">
                <div class="flex flex-col gap-1.5">
                    <span class="text-xs text-muted-foreground">Spell D</span>

                    <Select
                        :model-value="selection.spell_d ? String(selection.spell_d) : ''"
                        @update:model-value="pickSlot('d', $event)"
                    >
                        <SelectTrigger class="w-40 bg-white/[0.02]">
                            <SelectValue placeholder="Spell" />
                        </SelectTrigger>

                        <SelectContent>
                            <SelectItem
                                v-for="spell in ARAM_SPELLS"
                                :key="spell.id"
                                :value="String(spell.id)"
                                :disabled="spell.id === selection.spell_f"
                            >
                                <span class="flex items-center gap-2">
                                    <img class="size-5 rounded-sm bg-layer" :src="spellIcon(spell.id)" alt="" />
                                    {{ spell.name }}
                                </span>
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <div class="flex flex-col gap-1.5">
                    <span class="text-xs text-muted-foreground">Spell F</span>

                    <Select
                        :model-value="selection.spell_f ? String(selection.spell_f) : ''"
                        @update:model-value="pickSlot('f', $event)"
                    >
                        <SelectTrigger class="w-40 bg-white/[0.02]">
                            <SelectValue placeholder="Spell" />
                        </SelectTrigger>

                        <SelectContent>
                            <SelectItem
                                v-for="spell in ARAM_SPELLS"
                                :key="spell.id"
                                :value="String(spell.id)"
                                :disabled="spell.id === selection.spell_d"
                            >
                                <span class="flex items-center gap-2">
                                    <img class="size-5 rounded-sm bg-layer" :src="spellIcon(spell.id)" alt="" />
                                    {{ spell.name }}
                                </span>
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <Button
                    size="sm"
                    variant="secondary"
                    :disabled="applyingRunes"
                    @click="reapplyRunes"
                >
                    <Sparkles class="size-4" />
                    {{ applyingRunes ? 'Applying…' : 'Apply stats runes' }}
                </Button>

                <Button
                    v-if="selection.bench_enabled"
                    size="sm"
                    variant="secondary"
                    :disabled="selection.rerolls <= 0"
                    @click="requestReroll"
                >
                    <Dices class="size-4" />
                    Reroll ({{ selection.rerolls }})
                </Button>
            </div>

            <div v-if="skinOptions.length > 1" class="flex flex-col gap-2">
                <span class="text-xs text-muted-foreground">Skin</span>

                <div class="flex gap-2 overflow-x-auto pb-1.5">
                    <button
                        v-for="skin in skinOptions"
                        :key="skin.id"
                        class="flex w-20 shrink-0 cursor-pointer flex-col gap-1 rounded-lg p-1 text-left ring-1 ring-transparent transition-all hover:bg-white/[0.03] hover:ring-white/10"
                        :class="{ 'bg-primary/10 ring-2 ring-primary hover:ring-primary': selection.skin_id === skin.id }"
                        type="button"
                        @click="pickSkin(skin.id)"
                    >
                        <img
                            class="aspect-square w-full rounded-md bg-layer object-cover"
                            :src="skin.tile"
                            :alt="skin.name"
                            loading="lazy"
                        />
                        <span
                            class="truncate text-[11px]"
                            :class="selection.skin_id === skin.id ? 'text-primary' : 'text-muted-foreground'"
                        >
                            {{ skin.name }}
                        </span>
                    </button>
                </div>
            </div>

            <div v-if="selection.bench_enabled && selection.bench.length > 0" class="flex flex-col gap-2">
                <span class="text-xs text-muted-foreground">Bench · click to swap</span>

                <div class="flex items-center gap-2">
                    <button
                        v-for="championId in selection.bench"
                        :key="championId"
                        class="cursor-pointer rounded-lg ring-1 ring-white/10 transition-all hover:ring-2 hover:ring-primary"
                        type="button"
                        @click="swap(championId)"
                    >
                        <img
                            class="size-11 rounded-lg bg-layer"
                            :src="championSquare(championId)"
                            alt=""
                            loading="lazy"
                        />
                    </button>
                </div>
            </div>
        </div>
    </Card>
</template>
