<script setup lang="ts">
import { computed } from 'vue'

import { Card } from '@/components/ui/card'
import type { ChampionBuild } from '../api'
import { itemIcon, runeIcon, spellIcon } from '../assets'
import WinMeter from './WinMeter.vue'

type Section = 'core' | 'starting' | 'spells' | 'skills' | 'path' | 'page' | 'styles'

const SKILL_LETTERS: Record<string, string> = { '1': 'Q', '2': 'W', '3': 'E', '4': 'R' }

let props = defineProps<{ build: ChampionBuild, sections: Section[] }>()

let bestPage = computed(() => props.build.rune_pages[0] ?? null)
let boots = computed(() => props.build.boots.slice(0, 3))
let coreBuilds = computed(() => props.build.core_builds.slice(0, 4))
let runeStyles = computed(() => props.build.rune_styles.slice(0, 4))
let skillOrders = computed(() => props.build.skill_orders.slice(0, 3))
let skillPaths = computed(() => props.build.skill_sequences.slice(0, 3))
let spellPairs = computed(() => props.build.summoner_spells.slice(0, 3))
let startingItems = computed(() => props.build.starting_items.slice(0, 4))

function lettersOf(sequence: string): string[] {
    return sequence.split(',').map((id) => SKILL_LETTERS[id] ?? '?')
}

function show(section: Section): boolean {
    return props.sections.includes(section)
}

let empty = computed(
    () =>
        !(show('core') && coreBuilds.value.length > 0)
        && !(show('starting') && startingItems.value.length > 0)
        && !(show('spells') && (spellPairs.value.length > 0 || boots.value.length > 0))
        && !(show('skills') && skillOrders.value.length > 0)
        && !(show('path') && skillPaths.value.length > 0)
        && !(show('page') && bestPage.value)
        && !(show('styles') && runeStyles.value.length > 0),
)
</script>

<template>
    <div v-if="!empty" class="grid gap-4 lg:grid-cols-2">
        <Card v-if="show('core') && coreBuilds.length > 0" class="panel">
            <h2 class="panel-title">Core build</h2>

            <div v-for="(entry, index) in coreBuilds" :key="index" class="entry">
                <span class="flex flex-1 items-center gap-1.5">
                    <template v-for="(item, position) in entry.items" :key="item.id">
                        <span v-if="position > 0" class="text-muted-foreground/50">›</span>
                        <img
                            class="item-icon"
                            :src="itemIcon(item.id)"
                            :alt="item.name ?? String(item.id)"
                            :title="item.name ?? String(item.id)"
                        />
                    </template>
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>
        </Card>

        <Card v-if="show('starting') && startingItems.length > 0" class="panel">
            <h2 class="panel-title">Starting items</h2>

            <div v-for="(entry, index) in startingItems" :key="index" class="entry">
                <span class="flex flex-1 items-center gap-1.5">
                    <img
                        v-for="item in entry.items"
                        :key="item.id"
                        class="item-icon"
                        :src="itemIcon(item.id)"
                        :alt="item.name ?? String(item.id)"
                        :title="item.name ?? String(item.id)"
                    />
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>
        </Card>

        <Card v-if="show('spells') && (spellPairs.length > 0 || boots.length > 0)" class="panel">
            <h2 class="panel-title">Spells &amp; boots</h2>

            <div v-for="(entry, index) in spellPairs" :key="`spell-${index}`" class="entry">
                <span class="flex flex-1 items-center gap-1.5">
                    <img
                        class="spell-icon"
                        :src="spellIcon(entry.spell_a)"
                        :alt="entry.spell_a_name ?? String(entry.spell_a)"
                        :title="entry.spell_a_name ?? String(entry.spell_a)"
                    />
                    <img
                        class="spell-icon"
                        :src="spellIcon(entry.spell_b)"
                        :alt="entry.spell_b_name ?? String(entry.spell_b)"
                        :title="entry.spell_b_name ?? String(entry.spell_b)"
                    />
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>

            <div v-for="entry in boots" :key="`boot-${entry.id}`" class="entry">
                <span class="flex flex-1 items-center gap-1.5">
                    <img
                        class="item-icon"
                        :src="itemIcon(entry.id)"
                        :alt="entry.name ?? String(entry.id)"
                        :title="entry.name ?? String(entry.id)"
                    />
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>
        </Card>

        <Card v-if="show('skills') && skillOrders.length > 0" class="panel">
            <h2 class="panel-title">Skill priority</h2>

            <div v-for="entry in skillOrders" :key="entry.skill_max_order" class="entry">
                <span class="flex flex-1 items-center gap-2">
                    <template
                        v-for="(letter, index) in entry.skill_max_order.split('>')"
                        :key="index"
                    >
                        <span v-if="index > 0" class="text-muted-foreground/50">›</span>
                        <span class="flex size-7 items-center justify-center rounded-md bg-primary/12 font-display font-bold text-primary ring-1 ring-primary/20 ring-inset">
                            {{ letter }}
                        </span>
                    </template>
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>
        </Card>

        <Card v-if="show('path') && skillPaths.length > 0" class="panel">
            <h2 class="panel-title">First levels</h2>

            <div v-for="entry in skillPaths" :key="entry.sequence" class="entry">
                <span class="flex flex-1 items-center gap-1">
                    <span
                        v-for="(letter, index) in lettersOf(entry.sequence)"
                        :key="index"
                        class="flex size-6 items-center justify-center rounded-md bg-white/[0.04] font-display text-xs font-bold text-foreground/80 ring-1 ring-white/[0.06] ring-inset"
                    >
                        {{ letter }}
                    </span>
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>
        </Card>

        <Card v-if="show('page') && bestPage" class="panel">
            <h2 class="panel-title">Best rune page</h2>

            <div class="entry flex-col items-start gap-3.5">
                <div class="flex items-center gap-2.5">
                    <img
                        v-for="(rune, index) in bestPage.primary"
                        :key="rune.id"
                        class="rune-icon"
                        :class="index === 0 ? 'size-11 rounded-full bg-white/[0.04] ring-1 ring-gold/30' : 'size-7'"
                        :src="runeIcon(rune.id)"
                        :alt="rune.name ?? String(rune.id)"
                        :title="rune.name ?? String(rune.id)"
                    />
                </div>

                <div class="flex items-center gap-2.5">
                    <img
                        v-for="rune in bestPage.sub"
                        :key="rune.id"
                        class="rune-icon size-7"
                        :src="runeIcon(rune.id)"
                        :alt="rune.name ?? String(rune.id)"
                        :title="rune.name ?? String(rune.id)"
                    />

                    <span class="h-5 w-px bg-border"></span>

                    <img
                        v-for="(shard, index) in bestPage.shards"
                        :key="index"
                        class="rune-icon size-[22px]"
                        :src="runeIcon(shard.id)"
                        :alt="shard.name ?? String(shard.id)"
                        :title="shard.name ?? String(shard.id)"
                    />
                </div>

                <div class="flex w-full items-center justify-end gap-3">
                    <span class="games">{{ bestPage.games.toLocaleString() }}</span>
                    <WinMeter :win-rate="bestPage.win_rate" />
                </div>
            </div>
        </Card>

        <Card v-if="show('styles') && runeStyles.length > 0" class="panel">
            <h2 class="panel-title">Style pairs</h2>

            <div
                v-for="entry in runeStyles"
                :key="`${entry.primary_style}-${entry.sub_style}`"
                class="entry"
            >
                <span class="flex flex-1 items-center gap-2">
                    <img
                        class="rune-icon size-6"
                        :src="runeIcon(entry.primary_style)"
                        :alt="entry.primary_name ?? String(entry.primary_style)"
                    />
                    <span class="font-medium">{{ entry.primary_name ?? entry.primary_style }}</span>
                    <span class="text-muted-foreground/50">+</span>
                    <img
                        class="rune-icon size-6"
                        :src="runeIcon(entry.sub_style)"
                        :alt="entry.sub_name ?? String(entry.sub_style)"
                    />
                    <span class="text-muted-foreground">{{ entry.sub_name ?? entry.sub_style }}</span>
                </span>

                <span class="games">{{ entry.games.toLocaleString() }}</span>
                <WinMeter :win-rate="entry.win_rate" />
            </div>
        </Card>
    </div>
</template>
