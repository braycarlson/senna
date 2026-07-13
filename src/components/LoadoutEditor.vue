<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { Check } from '@lucide/vue'

import { Card } from '@/components/ui/card'
import { settingsGet, settingsSet, type Settings } from '../api'
import SkinGridPicker from './SkinGridPicker.vue'
import SpellPairPicker from './SpellPairPicker.vue'

let props = defineProps<{ championId: number }>()

let settings = ref<Settings | null>(null)
let spellPair = ref<[number, number] | null>(null)
let error = ref('')
let saved = ref(false)
let savedTimer = 0

let selectedSkin = computed(() => settings.value?.loadouts[props.championId]?.skin_id ?? null)

async function load(): Promise<void> {
    error.value = ''

    try {
        let settingsResult = await settingsGet()

        settings.value = settingsResult
        spellPair.value = settingsResult.loadouts[props.championId]?.spells ?? null
    } catch (raised) {
        error.value = String(raised)
    }
}

function flashSaved(): void {
    saved.value = true

    window.clearTimeout(savedTimer)
    savedTimer = window.setTimeout(() => (saved.value = false), 2000)
}

async function persist(mutate: (loadout: { skin_id?: number | null, spells?: [number, number] | null }) => void): Promise<void> {
    if (!settings.value) return

    let loadout = { ...settings.value.loadouts[props.championId] }

    mutate(loadout)

    if (loadout.skin_id == null) delete loadout.skin_id
    if (loadout.spells == null) delete loadout.spells

    let loadouts = { ...settings.value.loadouts }

    if (Object.keys(loadout).length === 0) {
        delete loadouts[props.championId]
    } else {
        loadouts[props.championId] = loadout
    }

    settings.value = { ...settings.value, loadouts }
    error.value = ''

    try {
        await settingsSet(settings.value)

        flashSaved()
    } catch (raised) {
        error.value = String(raised)
    }
}

function pickSkin(skinId: number | null): void {
    if (skinId === selectedSkin.value) return

    void persist((loadout) => {
        loadout.skin_id = skinId
    })
}

function pickSpells(pair: [number, number] | null): void {
    let current = settings.value?.loadouts[props.championId]?.spells ?? null

    if (JSON.stringify(pair) === JSON.stringify(current)) return

    void persist((loadout) => {
        loadout.spells = pair
    })
}

watch(() => props.championId, load)

onMounted(load)

onUnmounted(() => window.clearTimeout(savedTimer))
</script>

<template>
    <div class="flex flex-col gap-4">
        <div class="flex items-center justify-between">
            <p class="text-[13px] text-muted-foreground">
                Preferences apply automatically in champ select. Unowned skins are skipped.
            </p>

            <span v-if="saved" class="flex items-center gap-1.5 text-[13px] text-win">
                <Check class="size-4" />
                Saved
            </span>
        </div>

        <p v-if="error" class="text-[13px] text-loss">{{ error }}</p>

        <Card class="block rounded-xl p-4">
            <h2 class="mb-3 text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                Summoner spells
            </h2>

            <SpellPairPicker
                id-prefix="loadout"
                v-model="spellPair"
                @update:model-value="pickSpells"
            />

            <p class="mt-2.5 text-xs text-muted-foreground">
                Both slots set a fixed pair for this champion. Either on Most played falls back to
                the app default from Settings.
            </p>
        </Card>

        <Card class="block rounded-xl p-4">
            <h2 class="mb-3 text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                Skin
            </h2>

            <SkinGridPicker
                :champion-id="props.championId"
                :model-value="selectedSkin"
                @update:model-value="pickSkin"
            />
        </Card>
    </div>
</template>
