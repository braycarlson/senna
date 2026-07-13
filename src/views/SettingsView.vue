<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Check, X } from '@lucide/vue'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { apiChampions, settingsGet, settingsSet, type Settings } from '../api'
import { assetsEnsureBuildIcons, spellIcon } from '../assets'
import { MODES } from '../spells'
import { context, contextInit } from '../state'
import ChampAvatar from '../components/ChampAvatar.vue'
import PageHeading from '../components/PageHeading.vue'
import SpellPairPicker from '../components/SpellPairPicker.vue'
import StatusPane from '../components/StatusPane.vue'

let settings = ref<Settings>({
    base_url: '',
    token: '',
    region: 'NA1',
    auto_accept: true,
    auto_runes: true,
    auto_spells: true,
    auto_items: true,
    default_spells: [4, 6],
    mode_spells: {},
    random_skin: false,
    loadouts: {},
    mode_loadouts: {},
})
let saving = ref(false)
let saved = ref(false)
let error = ref('')
let championNames = ref<Record<number, string>>({})

const REGIONS = ['NA1', 'EUW1', 'KR', 'EUN1', 'BR1', 'JP1', 'LA1', 'LA2', 'OC1', 'TR1']

let loadoutRows = computed(() =>
    Object.entries(settings.value.loadouts).map(([id, loadout]) => ({
        id: Number(id),
        loadout,
    })),
)

async function loadChampionNames(): Promise<void> {
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

function modePair(mode: string): [number, number] | null {
    return settings.value.mode_spells[mode] ?? null
}

function setModePair(mode: string, pair: [number, number] | null): void {
    let modeSpells = { ...settings.value.mode_spells }

    if (pair) {
        modeSpells[mode] = pair
    } else {
        delete modeSpells[mode]
    }

    settings.value = { ...settings.value, mode_spells: modeSpells }
}

async function removeLoadout(championId: number): Promise<void> {
    saving.value = true
    error.value = ''

    try {
        let fresh = await settingsGet()
        let loadouts = { ...fresh.loadouts }

        delete loadouts[championId]

        settings.value = { ...settings.value, loadouts, mode_loadouts: fresh.mode_loadouts }

        await settingsSet(settings.value)
    } catch (raised) {
        error.value = String(raised)
    } finally {
        saving.value = false
    }
}

async function save(): Promise<void> {
    saving.value = true
    saved.value = false
    error.value = ''

    try {
        let fresh = await settingsGet()

        settings.value = {
            ...settings.value,
            loadouts: fresh.loadouts,
            mode_loadouts: fresh.mode_loadouts,
        }

        await settingsSet(settings.value)

        saved.value = true

        await contextInit()
    } catch (raised) {
        error.value = String(raised)
    } finally {
        saving.value = false
    }
}

onMounted(async () => {
    assetsEnsureBuildIcons()

    settings.value = await settingsGet()

    await loadChampionNames()
})
</script>

<template>
    <section class="mx-auto max-w-2xl">
        <PageHeading eyebrow="Configuration" title="Settings" />

        <div class="rise-in flex flex-col gap-4">
            <Card class="block rounded-xl p-5">
                <h2 class="mb-4 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    Stats service
                </h2>

                <div class="flex flex-col gap-4">
                    <div class="flex flex-col gap-1.5">
                        <Label for="base-url">API base URL</Label>
                        <Input
                            id="base-url"
                            v-model="settings.base_url"
                            class="bg-white/[0.02]"
                            placeholder="http://127.0.0.1:8080"
                            spellcheck="false"
                        />
                    </div>

                    <div class="flex flex-col gap-1.5">
                        <Label for="token">API token (optional)</Label>
                        <Input
                            id="token"
                            v-model="settings.token"
                            class="bg-white/[0.02]"
                            type="password"
                            spellcheck="false"
                        />
                    </div>

                    <div class="flex flex-col gap-1.5">
                        <Label for="region">Default region for player lookup</Label>

                        <Select v-model="settings.region">
                            <SelectTrigger id="region" class="w-full bg-white/[0.02]">
                                <SelectValue placeholder="Region" />
                            </SelectTrigger>

                            <SelectContent>
                                <SelectItem v-for="region in REGIONS" :key="region" :value="region">
                                    {{ region }}
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
            </Card>

            <Card class="block rounded-xl p-5">
                <h2 class="mb-4 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    Live client
                </h2>

                <div class="flex flex-col gap-3">
                    <div class="flex items-center justify-between gap-4 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <div class="flex flex-col gap-0.5">
                            <Label for="auto-accept">Auto accept</Label>
                            <span class="text-xs text-muted-foreground">
                                Accept the ready check when a match is found.
                            </span>
                        </div>
                        <Switch id="auto-accept" v-model="settings.auto_accept" />
                    </div>

                    <div class="flex items-center justify-between gap-4 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <div class="flex flex-col gap-0.5">
                            <Label for="auto-runes">Auto runes</Label>
                            <span class="text-xs text-muted-foreground">
                                Apply the best rune page in champ select.
                            </span>
                        </div>
                        <Switch id="auto-runes" v-model="settings.auto_runes" />
                    </div>

                    <div class="flex items-center justify-between gap-4 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <div class="flex flex-col gap-0.5">
                            <Label for="auto-spells">Auto spells</Label>
                            <span class="text-xs text-muted-foreground">
                                Apply summoner spells in champ select.
                            </span>
                        </div>
                        <Switch id="auto-spells" v-model="settings.auto_spells" />
                    </div>

                    <div class="flex items-center justify-between gap-4 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <div class="flex flex-col gap-0.5">
                            <Label for="auto-items">Auto item set</Label>
                            <span class="text-xs text-muted-foreground">
                                Push the recommended build into the in-game shop.
                            </span>
                        </div>
                        <Switch id="auto-items" v-model="settings.auto_items" />
                    </div>

                    <div class="flex items-center justify-between gap-4 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <div class="flex flex-col gap-0.5">
                            <Label for="random-skin">Random skin</Label>
                            <span class="text-xs text-muted-foreground">
                                Pick a random owned skin when no loadout skin is set.
                            </span>
                        </div>
                        <Switch id="random-skin" v-model="settings.random_skin" />
                    </div>

                    <div class="flex flex-col gap-2 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <span class="text-sm font-medium">Default summoner spells</span>

                        <SpellPairPicker id-prefix="default" v-model="settings.default_spells" />

                        <span class="text-xs text-muted-foreground">
                            Fallback for every mode. A champion loadout pair or a mode pair below
                            takes priority. Most played uses the ARAM stats pick.
                        </span>
                    </div>

                    <div class="flex flex-col gap-4 rounded-lg bg-white/[0.02] px-4 py-3 ring-1 ring-white/[0.03] ring-inset">
                        <span class="text-sm font-medium">Spells per game mode</span>

                        <div v-for="mode in MODES" :key="mode.key" class="flex flex-col gap-1.5">
                            <span class="text-xs text-muted-foreground">{{ mode.label }}</span>

                            <SpellPairPicker
                                :id-prefix="`mode-${mode.key}`"
                                :model-value="modePair(mode.key)"
                                @update:model-value="setModePair(mode.key, $event)"
                            />
                        </div>

                        <span class="text-xs text-muted-foreground">
                            Most played on both slots falls back to the default pair above.
                        </span>
                    </div>
                </div>
            </Card>

            <div class="flex items-center gap-3">
                <Button :disabled="saving" @click="save">Save settings</Button>

                <span v-if="saved" class="flex items-center gap-1.5 text-[13px] text-win">
                    <Check class="size-4" />
                    Saved
                </span>
            </div>

            <StatusPane v-if="error" variant="error" :message="error" />

            <Card v-if="loadoutRows.length > 0" class="block rounded-xl p-5">
                <h2 class="mb-4 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    Champion loadouts
                </h2>

                <div class="flex flex-col gap-2">
                    <div
                        v-for="row in loadoutRows"
                        :key="row.id"
                        class="flex items-center gap-3 rounded-lg bg-white/[0.02] px-3 py-2 ring-1 ring-white/[0.03] ring-inset"
                    >
                        <ChampAvatar :champion-id="row.id" :name="championNames[row.id]" size="sm" />

                        <router-link
                            class="flex-1 font-medium transition-colors hover:text-primary"
                            :to="`/champions/${row.id}`"
                        >
                            {{ championNames[row.id] ?? `Champion ${row.id}` }}
                        </router-link>

                        <span v-if="row.loadout.spells" class="flex items-center gap-1">
                            <img
                                v-for="spell in row.loadout.spells"
                                :key="spell"
                                class="size-5 rounded-sm bg-layer"
                                :src="spellIcon(spell)"
                                alt=""
                            />
                        </span>

                        <Badge v-if="row.loadout.skin_id != null" variant="secondary">Skin</Badge>

                        <Button
                            size="icon-xs"
                            variant="ghost"
                            :disabled="saving"
                            @click="removeLoadout(row.id)"
                        >
                            <X class="size-3.5" />
                        </Button>
                    </div>
                </div>
            </Card>

        </div>
    </section>
</template>
