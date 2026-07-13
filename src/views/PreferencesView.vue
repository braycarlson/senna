<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Image, Plus, X } from '@lucide/vue'

import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
    settingsGet,
    settingsSet,
    type ItemBlock,
    type Loadout,
    type RunePage,
    type Settings,
} from '../api'
import {
    assetsEnsureLoadoutEditors,
    fetchChampionCatalog,
    itemIcon,
    runeIcon,
    spellIcon,
    type ChampionEntry,
} from '../assets'
import { useDelayedLoader } from '../loading'
import { MODES } from '../spells'
import ChampAvatar from '../components/ChampAvatar.vue'
import LoadingPane from '../components/LoadingPane.vue'
import ItemSetEditor from '../components/ItemSetEditor.vue'
import PageHeading from '../components/PageHeading.vue'
import RunePageEditor from '../components/RunePageEditor.vue'
import SkinGridPicker from '../components/SkinGridPicker.vue'
import SpellPairPicker from '../components/SpellPairPicker.vue'
import StatusPane from '../components/StatusPane.vue'
import { toastShow } from '../toast'

const ANY_MODE = 'ANY'

const MODE_OPTIONS = [{ key: ANY_MODE, label: 'Any mode' }, ...MODES]

const MODE_LABELS: Record<string, string> = Object.fromEntries(
    MODE_OPTIONS.map((mode) => [mode.key, mode.label]),
)

type Section = 'spells' | 'runes' | 'items' | 'skin'

const SECTION_TITLES: Record<Section, string> = {
    spells: 'Summoner spells',
    runes: 'Rune page',
    items: 'Item set',
    skin: 'Skin',
}

const SECTION_WIDTHS: Record<Section, string> = {
    spells: 'sm:max-w-lg',
    runes: 'sm:max-w-4xl',
    items: 'sm:max-w-5xl',
    skin: 'sm:max-w-4xl',
}

const ITEMS_PREVIEW_MAX = 3

let catalog = ref<ChampionEntry[]>([])
let settings = ref<Settings | null>(null)
let mode = ref(ANY_MODE)
let query = ref('')
let configuredOnly = ref(false)
let editing = ref<{ champion: ChampionEntry, section: Section } | null>(null)
let error = ref('')
let { loading, showLoader, begin, end } = useDelayedLoader()

function loadoutOf(source: Settings, modeKey: string, champion: number): Loadout {
    if (modeKey === ANY_MODE) return source.loadouts[champion] ?? {}

    return source.mode_loadouts[modeKey]?.[champion] ?? {}
}

function entryOf(champion: number): Loadout {
    if (!settings.value) return {}

    return loadoutOf(settings.value, mode.value, champion)
}

function itemsFlat(entry: Loadout): number[] {
    return entry.items?.flatMap((block) => block.items) ?? []
}

function itemsExtra(champion: number): number {
    return itemsFlat(entryOf(champion)).length - ITEMS_PREVIEW_MAX
}

function entrySet(entry: Loadout): boolean {
    return (
        entry.spells != null ||
        entry.runes != null ||
        (entry.items?.length ?? 0) > 0 ||
        entry.skin_id != null
    )
}

let rows = computed(() => {
    let needle = query.value.trim().toLowerCase()
    let filtered = catalog.value

    if (needle) {
        filtered = filtered.filter((champion) => champion.name.toLowerCase().includes(needle))
    }

    if (configuredOnly.value) {
        filtered = filtered.filter((champion) => entrySet(entryOf(champion.id)))
    }

    return filtered
})

let editingLoadout = computed<Loadout>(() => {
    if (!editing.value) return {}

    return entryOf(editing.value.champion.id)
})

function prune(entry: Loadout): void {
    if (entry.skin_id == null) delete entry.skin_id
    if (entry.spells == null) delete entry.spells
    if (entry.runes == null) delete entry.runes

    if (
        entry.items == null ||
        entry.items.every((block) => block.items.length === 0)
    ) {
        delete entry.items
    }
}

function withLoadout(
    source: Settings,
    modeKey: string,
    champion: number,
    entry: Loadout,
): Settings {
    let empty = Object.keys(entry).length === 0

    if (modeKey === ANY_MODE) {
        let loadouts = { ...source.loadouts }

        if (empty) {
            delete loadouts[champion]
        } else {
            loadouts[champion] = entry
        }

        return { ...source, loadouts }
    }

    let byChampion = { ...(source.mode_loadouts[modeKey] ?? {}) }

    if (empty) {
        delete byChampion[champion]
    } else {
        byChampion[champion] = entry
    }

    let modeLoadouts = { ...source.mode_loadouts }

    if (Object.keys(byChampion).length === 0) {
        delete modeLoadouts[modeKey]
    } else {
        modeLoadouts[modeKey] = byChampion
    }

    return { ...source, mode_loadouts: modeLoadouts }
}

async function write(champion: number, mutate: (entry: Loadout) => void): Promise<void> {
    error.value = ''

    try {
        let fresh = await settingsGet()
        let entry: Loadout = { ...loadoutOf(fresh, mode.value, champion) }

        mutate(entry)
        prune(entry)

        let next = withLoadout(fresh, mode.value, champion, entry)

        settings.value = next

        await settingsSet(next)

        toastShow('Preferences saved', 'ok', 2500)
    } catch (raised) {
        error.value = String(raised)
    }
}

function open(champion: ChampionEntry, section: Section): void {
    editing.value = { champion, section }
}

function close(openState: boolean): void {
    if (!openState) editing.value = null
}

function pickSpells(pair: [number, number] | null): void {
    if (!editing.value) return
    if (JSON.stringify(pair) === JSON.stringify(editingLoadout.value.spells ?? null)) return

    void write(editing.value.champion.id, (entry) => {
        entry.spells = pair
    })
}

function pickRunes(page: RunePage | null): void {
    if (!editing.value) return
    if (JSON.stringify(page) === JSON.stringify(editingLoadout.value.runes ?? null)) return

    void write(editing.value.champion.id, (entry) => {
        entry.runes = page
    })
}

function pickItems(items: ItemBlock[] | null): void {
    if (!editing.value) return
    if (JSON.stringify(items) === JSON.stringify(editingLoadout.value.items ?? null)) return

    void write(editing.value.champion.id, (entry) => {
        entry.items = items
    })
}

function pickSkin(skinId: number | null): void {
    if (!editing.value) return
    if (skinId === (editingLoadout.value.skin_id ?? null)) return

    void write(editing.value.champion.id, (entry) => {
        entry.skin_id = skinId
    })
}

function clearRow(champion: ChampionEntry): void {
    void write(champion.id, (entry) => {
        entry.skin_id = null
        entry.spells = null
        entry.runes = null
        entry.items = null
    })
}

onMounted(async () => {
    assetsEnsureLoadoutEditors()
    begin()

    try {
        let [catalogResult, settingsResult] = await Promise.all([
            fetchChampionCatalog(),
            settingsGet(),
        ])

        catalog.value = catalogResult
        settings.value = settingsResult
    } catch (raised) {
        error.value = String(raised)
    } finally {
        end()
    }
})
</script>

<template>
    <section class="mx-auto max-w-4xl">
        <PageHeading eyebrow="Champ select" title="Preferences" />

        <div class="rise-in flex flex-col gap-4">
            <Card class="block rounded-xl p-5">
                <h2 class="mb-1 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    Champion preferences
                </h2>

                <p class="mb-4 text-xs text-muted-foreground">
                    Preferences apply automatically in champ select. A mode entry overrides the
                    Any mode entry; anything unset falls back to the Settings spell pairs and the
                    ARAM stats build.
                </p>

                <div class="mb-4 flex flex-wrap items-center gap-3">
                    <Tabs v-model="mode">
                        <TabsList>
                            <TabsTrigger
                                v-for="option in MODE_OPTIONS"
                                :key="option.key"
                                :value="option.key"
                            >
                                {{ option.label }}
                            </TabsTrigger>
                        </TabsList>
                    </Tabs>

                    <Input
                        v-model="query"
                        class="max-w-52 bg-white/[0.02]"
                        placeholder="Search champions…"
                        spellcheck="false"
                    />

                    <label class="ml-auto flex items-center gap-2 text-xs text-muted-foreground">
                        Configured only
                        <Switch v-model="configuredOnly" />
                    </label>
                </div>

                <StatusPane v-if="error" class="mb-4" variant="error" :message="error" />

                <transition name="fade" mode="out-in" appear>
                    <LoadingPane
                        v-if="loading && showLoader"
                        key="loading"
                        class="min-h-[50vh]"
                        message="Loading champions…"
                    />

                    <div v-else-if="loading" key="pending" class="min-h-[50vh]"></div>

                    <div v-else key="table">
                        <Table class="table-fixed">
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Champion</TableHead>
                                    <TableHead class="w-28">Spells</TableHead>
                                    <TableHead class="w-28">Runes</TableHead>
                                    <TableHead class="w-36">Items</TableHead>
                                    <TableHead class="w-20">Skin</TableHead>
                                    <TableHead class="w-12"></TableHead>
                                </TableRow>
                            </TableHeader>

                            <TableBody>
                                <TableRow v-for="champion in rows" :key="champion.id">
                                    <TableCell>
                                        <span class="flex items-center gap-2.5 font-medium">
                                            <ChampAvatar :champion-id="champion.id" :name="champion.name" size="sm" />
                                            {{ champion.name }}
                                        </span>
                                    </TableCell>

                                    <TableCell>
                                        <button
                                            v-if="entryOf(champion.id).spells"
                                            class="cell-btn"
                                            type="button"
                                            @click="open(champion, 'spells')"
                                        >
                                            <img
                                                v-for="spell in entryOf(champion.id).spells"
                                                :key="spell"
                                                class="size-6 rounded-sm bg-layer"
                                                :src="spellIcon(spell)"
                                                alt=""
                                            />
                                        </button>
                                        <button
                                            v-else
                                            class="cell-add"
                                            type="button"
                                            @click="open(champion, 'spells')"
                                        >
                                            <Plus class="size-4" />
                                        </button>
                                    </TableCell>

                                    <TableCell>
                                        <button
                                            v-if="entryOf(champion.id).runes"
                                            class="cell-btn"
                                            type="button"
                                            @click="open(champion, 'runes')"
                                        >
                                            <img
                                                class="size-7 rounded-full bg-layer"
                                                :src="runeIcon(entryOf(champion.id).runes!.perks[0])"
                                                alt=""
                                            />
                                            <img
                                                class="size-4 opacity-80"
                                                :src="runeIcon(entryOf(champion.id).runes!.sub_style)"
                                                alt=""
                                            />
                                        </button>
                                        <button
                                            v-else
                                            class="cell-add"
                                            type="button"
                                            @click="open(champion, 'runes')"
                                        >
                                            <Plus class="size-4" />
                                        </button>
                                    </TableCell>

                                    <TableCell>
                                        <button
                                            v-if="itemsFlat(entryOf(champion.id)).length > 0"
                                            class="cell-btn"
                                            type="button"
                                            @click="open(champion, 'items')"
                                        >
                                            <img
                                                v-for="id in itemsFlat(entryOf(champion.id)).slice(0, ITEMS_PREVIEW_MAX)"
                                                :key="id"
                                                class="size-6 rounded-sm bg-layer"
                                                :src="itemIcon(id)"
                                                alt=""
                                            />
                                            <span
                                                v-if="itemsFlat(entryOf(champion.id)).length > ITEMS_PREVIEW_MAX"
                                                class="text-xs text-muted-foreground"
                                            >
                                                +{{ itemsExtra(champion.id) }}
                                            </span>
                                        </button>
                                        <button
                                            v-else
                                            class="cell-add"
                                            type="button"
                                            @click="open(champion, 'items')"
                                        >
                                            <Plus class="size-4" />
                                        </button>
                                    </TableCell>

                                    <TableCell>
                                        <button
                                            v-if="entryOf(champion.id).skin_id != null"
                                            class="cell-btn text-primary"
                                            type="button"
                                            @click="open(champion, 'skin')"
                                        >
                                            <Image class="size-5" />
                                        </button>
                                        <button
                                            v-else
                                            class="cell-add"
                                            type="button"
                                            @click="open(champion, 'skin')"
                                        >
                                            <Plus class="size-4" />
                                        </button>
                                    </TableCell>

                                    <TableCell>
                                        <Button
                                            v-if="entrySet(entryOf(champion.id))"
                                            size="icon-xs"
                                            variant="ghost"
                                            @click="clearRow(champion)"
                                        >
                                            <X class="size-3.5" />
                                        </Button>
                                    </TableCell>
                                </TableRow>

                                <TableRow v-if="rows.length === 0">
                                    <TableCell class="text-center text-muted-foreground" :colspan="6">
                                        No champions match.
                                    </TableCell>
                                </TableRow>
                            </TableBody>
                        </Table>
                    </div>
                </transition>
            </Card>
        </div>

        <Dialog :open="editing !== null" @update:open="close">
            <DialogContent v-if="editing" :class="SECTION_WIDTHS[editing.section]">
                <DialogHeader>
                    <DialogTitle>
                        {{ editing.champion.name }} · {{ SECTION_TITLES[editing.section] }}
                    </DialogTitle>

                    <DialogDescription>
                        {{ MODE_LABELS[mode] }} · changes save immediately.
                    </DialogDescription>
                </DialogHeader>

                <SpellPairPicker
                    v-if="editing.section === 'spells'"
                    id-prefix="pref"
                    :model-value="editingLoadout.spells ?? null"
                    @update:model-value="pickSpells"
                />

                <RunePageEditor
                    v-else-if="editing.section === 'runes'"
                    :model-value="editingLoadout.runes ?? null"
                    @update:model-value="pickRunes"
                />

                <ItemSetEditor
                    v-else-if="editing.section === 'items'"
                    :model-value="editingLoadout.items ?? null"
                    @update:model-value="pickItems"
                />

                <SkinGridPicker
                    v-else
                    :champion-id="editing.champion.id"
                    :model-value="editingLoadout.skin_id ?? null"
                    @update:model-value="pickSkin"
                />
            </DialogContent>
        </Dialog>
    </section>
</template>

<style scoped>
@reference "../style.css";

.cell-btn {
    @apply flex cursor-pointer items-center gap-1 rounded-md px-1.5 py-1 transition-colors;
}

.cell-btn:hover {
    @apply bg-white/[0.05];
}

.cell-add {
    @apply flex size-7 cursor-pointer items-center justify-center rounded-md text-muted-foreground/40 transition-colors;
}

.cell-add:hover {
    @apply bg-white/[0.05] text-muted-foreground;
}
</style>
