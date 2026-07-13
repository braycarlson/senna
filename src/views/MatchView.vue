<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowLeft } from '@lucide/vue'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { apiMatch, type MatchDetail, type MatchParticipant } from '../api'
import { assetsEnsureBuildIcons, itemIcon, runeIcon, spellIcon } from '../assets'
import { thousands } from '../format'
import ChampAvatar from '../components/ChampAvatar.vue'
import StatusPane from '../components/StatusPane.vue'

let props = defineProps<{ id: string }>()

let detail = ref<MatchDetail | null>(null)
let loading = ref(true)
let error = ref('')

let teams = computed(() => {
    if (!detail.value) return []

    let groups = new Map<number, MatchParticipant[]>()

    for (let participant of detail.value.participants) {
        let members = groups.get(participant.team_id) ?? []

        members.push(participant)
        groups.set(participant.team_id, members)
    }

    return [...groups.entries()]
        .sort((a, b) => a[0] - b[0])
        .map(([team, members]) => ({
            team,
            members,
            won: members[0]?.win ?? false,
        }))
})

let damageMax = computed(() => {
    let peak = 0

    for (let participant of detail.value?.participants ?? []) {
        peak = Math.max(peak, participant.damage_to_champions)
    }

    return Math.max(peak, 1)
})

function damageWidth(damage: number): string {
    return `${(damage / damageMax.value) * 100}%`
}

function items(participant: MatchParticipant): number[] {
    let ids = [
        participant.item0,
        participant.item1,
        participant.item2,
        participant.item3,
        participant.item4,
        participant.item5,
        participant.item6,
    ]

    return ids.filter((id) => id > 0)
}

function duration(seconds: number): string {
    let minutes = Math.floor(seconds / 60)
    let rest = String(Math.floor(seconds % 60)).padStart(2, '0')

    return `${minutes}:${rest}`
}

function playedAt(epochMs: number): string {
    return new Date(epochMs).toLocaleString(undefined, {
        month: 'short',
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit',
    })
}

async function load(): Promise<void> {
    if (!props.id) return

    assetsEnsureBuildIcons()

    loading.value = true
    error.value = ''

    try {
        detail.value = await apiMatch(props.id)
    } catch (raised) {
        error.value = String(raised)
        detail.value = null
    } finally {
        loading.value = false
    }
}

watch(() => props.id, load, { immediate: true })
</script>

<template>
    <section class="mx-auto max-w-5xl">
        <StatusPane v-if="error" variant="error" :message="error" />
        <StatusPane v-else-if="loading || !detail" variant="loading" message="Loading match…" />

        <template v-else>
            <div class="rise-in mb-6 flex items-end justify-between gap-4">
                <div>
                    <span class="font-display text-[11px] font-semibold tracking-[0.16em] text-primary uppercase">
                        Match detail
                    </span>

                    <div class="mt-1 flex items-center gap-3">
                        <h1 class="text-[26px] leading-tight font-bold">
                            {{ detail.game_mode }}
                        </h1>
                        <Badge variant="secondary">Patch {{ detail.patch }}</Badge>
                        <span class="stat-mono text-xs text-muted-foreground">
                            {{ duration(detail.game_duration) }} ·
                            {{ playedAt(detail.game_creation) }}
                        </span>
                    </div>
                </div>

                <Button size="sm" variant="secondary" @click="$router.back()">
                    <ArrowLeft class="size-4" />
                    Back
                </Button>
            </div>

            <div class="rise-in flex flex-col gap-4">
                <Card v-for="group in teams" :key="group.team" class="block gap-0 overflow-hidden rounded-xl py-0">
                    <h2
                        class="border-b px-4 py-3 font-display text-xs font-bold tracking-[0.14em] uppercase"
                        :class="group.won ? 'bg-win/[0.06] text-win' : 'bg-loss/[0.06] text-loss'"
                    >
                        {{ group.won ? 'Victory' : 'Defeat' }}
                    </h2>

                    <div class="flex flex-col">
                        <div
                            v-for="participant in group.members"
                            :key="participant.puuid"
                            class="flex items-center gap-4 border-t border-white/[0.04] px-4 py-2.5 first:border-t-0"
                        >
                            <span class="flex w-44 min-w-0 items-center gap-2.5">
                                <span class="relative shrink-0">
                                    <ChampAvatar
                                        :champion-id="participant.champion_id"
                                        :name="participant.champion_name"
                                    />
                                    <span class="stat-mono absolute -right-1 -bottom-1 rounded bg-black/80 px-1 text-[10px] leading-4">
                                        {{ participant.champion_level }}
                                    </span>
                                </span>
                                <span class="truncate text-[13px] font-medium">
                                    {{ participant.champion_name }}
                                </span>
                            </span>

                            <span class="flex shrink-0 items-center gap-1">
                                <img
                                    class="size-5 rounded-sm bg-layer"
                                    :src="spellIcon(participant.summoner1_id)"
                                    alt=""
                                />
                                <img
                                    class="size-5 rounded-sm bg-layer"
                                    :src="spellIcon(participant.summoner2_id)"
                                    alt=""
                                />
                                <img
                                    class="size-5"
                                    :src="runeIcon(participant.perk_keystone)"
                                    alt=""
                                />
                            </span>

                            <span class="stat-mono w-24 shrink-0 text-right text-[13px]">
                                {{ participant.kills }} /
                                <span class="text-loss">{{ participant.deaths }}</span> /
                                {{ participant.assists }}
                            </span>

                            <span class="flex w-32 shrink-0 flex-col items-end gap-1">
                                <span class="stat-mono text-xs text-muted-foreground">
                                    {{ participant.damage_to_champions.toLocaleString() }}
                                </span>
                                <span class="h-[3px] w-full overflow-hidden rounded-full bg-white/[0.07]">
                                    <span
                                        class="block h-full rounded-full bg-primary/80"
                                        :style="{ width: damageWidth(participant.damage_to_champions) }"
                                    ></span>
                                </span>
                            </span>

                            <span class="stat-mono w-16 shrink-0 text-right text-xs text-muted-foreground">
                                {{ thousands(participant.gold_earned) }}k
                            </span>

                            <span class="ml-auto flex shrink-0 items-center gap-1">
                                <img
                                    v-for="(item, index) in items(participant)"
                                    :key="index"
                                    class="size-6 rounded-sm bg-layer"
                                    :src="itemIcon(item)"
                                    alt=""
                                />
                            </span>
                        </div>
                    </div>
                </Card>
            </div>
        </template>
    </section>
</template>
