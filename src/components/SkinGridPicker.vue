<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Ban } from '@lucide/vue'

import { Switch } from '@/components/ui/switch'
import { lcuOwnedSkins } from '../api'
import { fetchChampionSkins, type ChampionSkin } from '../assets'

let props = defineProps<{ championId: number }>()
let model = defineModel<number | null>({ required: true })

let skins = ref<ChampionSkin[]>([])
let owned = ref<Set<number> | null>(null)
let showOwnedOnly = ref(false)
let error = ref('')

function skinOwned(skin: ChampionSkin): boolean {
    return !owned.value || skin.isBase || owned.value.has(skin.id)
}

let visibleSkins = computed(() => {
    if (!showOwnedOnly.value || !owned.value) return skins.value

    return skins.value.filter(skinOwned)
})

async function load(): Promise<void> {
    error.value = ''
    skins.value = []
    owned.value = null

    if (props.championId <= 0) return

    try {
        let [skinsResult, ownedResult] = await Promise.all([
            fetchChampionSkins(props.championId),
            lcuOwnedSkins(props.championId).catch(() => null),
        ])

        skins.value = skinsResult
        owned.value = ownedResult ? new Set(ownedResult) : null
    } catch (raised) {
        error.value = String(raised)
    }
}

function pick(skinId: number | null): void {
    if (skinId === model.value) return

    model.value = skinId
}

watch(() => props.championId, load)

onMounted(load)
</script>

<template>
    <div class="flex flex-col gap-3">
        <div class="flex items-center justify-between">
            <p v-if="error" class="text-[13px] text-loss">{{ error }}</p>
            <p v-else class="text-xs text-muted-foreground">Unowned skins are skipped.</p>

            <label v-if="owned" class="flex items-center gap-2 text-xs text-muted-foreground">
                Show owned
                <Switch v-model="showOwnedOnly" />
            </label>
            <span v-else class="text-xs text-muted-foreground/70">
                League offline, ownership unknown
            </span>
        </div>

        <div class="grid max-h-[62vh] grid-cols-3 gap-2.5 overflow-y-auto pr-1 min-[900px]:grid-cols-5">
            <button
                class="skin-tile"
                :class="{ selected: model === null }"
                type="button"
                @click="pick(null)"
            >
                <span class="flex aspect-square w-full items-center justify-center rounded-lg bg-white/[0.03]">
                    <Ban class="size-6 text-muted-foreground/50" />
                </span>
                <span class="skin-name">No preference</span>
            </button>

            <button
                v-for="skin in visibleSkins"
                :key="skin.id"
                class="skin-tile"
                :class="{ selected: model === skin.id, unowned: !skinOwned(skin) }"
                type="button"
                :disabled="!skinOwned(skin)"
                @click="pick(skin.id)"
            >
                <img
                    class="aspect-square w-full rounded-lg bg-layer object-cover"
                    :src="skin.tile"
                    :alt="skin.name"
                    loading="lazy"
                />
                <span class="skin-name">{{ skin.name }}</span>
            </button>
        </div>
    </div>
</template>

<style scoped>
@reference "../style.css";

.skin-tile {
    @apply flex cursor-pointer flex-col gap-1.5 rounded-xl p-1.5 text-left ring-1 ring-transparent transition-all;
}

.skin-tile:hover {
    @apply bg-white/[0.03] ring-white/10;
}

.skin-tile.selected {
    @apply bg-primary/10 ring-2 ring-primary;
}

.skin-name {
    @apply truncate px-0.5 text-xs text-muted-foreground;
}

.skin-tile.selected .skin-name {
    @apply text-primary;
}

.skin-tile.unowned {
    @apply cursor-default opacity-35 grayscale;
}

.skin-tile.unowned:hover {
    @apply bg-transparent ring-transparent;
}
</style>
