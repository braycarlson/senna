<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { Card } from '@/components/ui/card'
import {
    apiChampions,
    apiPlayerChampions,
    type ChampionRow,
    type LobbyPlayer,
    type PlayerChampionRow,
} from '../api'
import { championSquare } from '../assets'
import { percent } from '../format'
import { context } from '../state'
import ChampAvatar from './ChampAvatar.vue'
import WinMeter from './WinMeter.vue'

let props = defineProps<{ roster: LobbyPlayer[] }>()

let championRows = ref<Map<number, ChampionRow>>(new Map())
let playerRows = ref<Map<string, PlayerChampionRow[]>>(new Map())

let teams = computed(() => {
    let groups = new Map<number, LobbyPlayer[]>()

    for (let player of props.roster) {
        let members = groups.get(player.team) ?? []

        members.push(player)
        groups.set(player.team, members)
    }

    let ordered = [...groups.entries()].sort((a, b) => {
        let aOwn = a[1].some((player) => player.self) ? 0 : 1
        let bOwn = b[1].some((player) => player.self) ? 0 : 1

        return aOwn - bOwn || a[0] - b[0]
    })

    return ordered.map(([team, members], index) => ({
        team,
        members,
        label:
            ordered.length === 1
                ? 'Your team'
                : index === 0
                    ? 'Your team'
                    : 'Enemy team',
    }))
})

function championOf(player: LobbyPlayer): ChampionRow | null {
    return championRows.value.get(player.champion_id) ?? null
}

function onChampion(player: LobbyPlayer): PlayerChampionRow | null {
    let rows = playerRows.value.get(player.puuid)

    if (!rows) return null

    return rows.find((row) => row.champion_id === player.champion_id) ?? null
}

function hasHistory(player: LobbyPlayer): boolean {
    return (playerRows.value.get(player.puuid) ?? []).length > 0
}

function mains(player: LobbyPlayer): PlayerChampionRow[] {
    return (playerRows.value.get(player.puuid) ?? []).slice(0, 3)
}

let teamRates = computed(() => {
    if (teams.value.length !== 2) return null

    let rates = teams.value.map((group) => {
        let known = group.members
            .map((player) => championOf(player))
            .filter((row): row is ChampionRow => row !== null)

        if (known.length === 0) return 0

        let sum = known.reduce((total, row) => total + row.win_rate, 0)

        return sum / known.length
    })

    if (rates[0] === 0 || rates[1] === 0) return null

    return { own: rates[0], enemy: rates[1], share: rates[0] / (rates[0] + rates[1]) }
})

let ownBarWidth = computed(() => `${(teamRates.value?.share ?? 0) * 100}%`)

let enemyBarLeft = computed(() => `${(teamRates.value?.share ?? 0) * 100 + 0.5}%`)

async function load(): Promise<void> {
    if (!context.patch) return

    apiChampions(context.patch, context.queue)
        .then((rows) => {
            championRows.value = new Map(rows.map((row) => [row.champion_id, row]))
        })
        .catch(() => {})

    let lookups = props.roster
        .filter((player) => player.puuid)
        .map(async (player) => {
            let rows = await apiPlayerChampions(player.puuid).catch(() => [])

            playerRows.value.set(player.puuid, rows)
        })

    await Promise.all(lookups)

    playerRows.value = new Map(playerRows.value)
}

watch([() => props.roster, () => context.patch], load, { immediate: true })
</script>

<template>
    <Card class="block gap-0 overflow-hidden rounded-xl py-0">
        <h2 class="border-b bg-white/[0.02] px-4 py-3 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
            Live match
        </h2>

        <div v-if="teamRates" class="flex items-center gap-3 border-b border-white/[0.04] px-4 py-3">
            <span class="stat-mono text-xs font-medium text-primary">
                {{ percent(teamRates.own) }}%
            </span>

            <span class="relative h-[5px] flex-1 overflow-hidden rounded-full bg-white/[0.06]">
                <span
                    class="absolute inset-y-0 left-0 rounded-l-full bg-primary/80"
                    :style="{ width: ownBarWidth }"
                ></span>
                <span
                    class="absolute inset-y-0 rounded-r-full bg-loss/70"
                    :style="{ left: enemyBarLeft, right: 0 }"
                ></span>
            </span>

            <span class="stat-mono text-xs font-medium text-loss">
                {{ percent(teamRates.enemy) }}%
            </span>

            <span class="text-[11px] text-muted-foreground">avg champion WR</span>
        </div>

        <div class="grid gap-0 min-[900px]:grid-cols-2 min-[900px]:divide-x min-[900px]:divide-border/60">
            <div v-for="group in teams" :key="group.team" class="p-3">
                <span
                    class="mb-2 block px-1 font-display text-[10px] font-semibold tracking-[0.16em] uppercase"
                    :class="group.label === 'Your team' ? 'text-primary' : 'text-loss/80'"
                >
                    {{ group.label }}
                </span>

                <div class="flex flex-col gap-1.5">
                    <div
                        v-for="player in group.members"
                        :key="player.puuid || `${player.team}-${player.name}-${player.champion_id}`"
                        class="flex items-center gap-3 rounded-lg bg-white/[0.02] px-3 py-2 ring-1 ring-white/[0.03] ring-inset"
                        :class="{ 'ring-primary/40 bg-primary/[0.05]': player.self }"
                    >
                        <ChampAvatar
                            v-if="player.champion_id > 0"
                            :champion-id="player.champion_id"
                            :name="championOf(player)?.champion_name"
                            size="sm"
                        />
                        <span
                            v-else
                            class="flex size-7 shrink-0 items-center justify-center rounded-md bg-white/[0.04] text-xs text-muted-foreground/50"
                        >
                            ?
                        </span>

                        <div class="flex min-w-0 flex-1 flex-col">
                            <span class="truncate text-[13px] font-medium">
                                {{ player.name || 'Hidden' }}
                            </span>

                            <span class="truncate text-xs text-muted-foreground">
                                <template v-if="onChampion(player)">
                                    {{ onChampion(player)!.games }} games ·
                                    {{ percent(onChampion(player)!.win_rate, 0) }}% on
                                    {{ championOf(player)?.champion_name ?? 'champion' }}
                                </template>
                                <template v-else-if="hasHistory(player)">
                                    First game on
                                    {{ championOf(player)?.champion_name ?? 'this champion' }}
                                </template>
                                <template v-else>No match history yet</template>
                            </span>
                        </div>

                        <span v-if="mains(player).length > 0" class="flex shrink-0 items-center gap-1">
                            <img
                                v-for="main in mains(player)"
                                :key="main.champion_id"
                                class="size-5 rounded bg-layer ring-1 ring-white/10"
                                :src="championSquare(main.champion_id)"
                                :alt="main.champion_name"
                                :title="`${main.champion_name} · ${main.games} games · ${percent(main.win_rate, 0)}%`"
                                loading="lazy"
                            />
                        </span>

                        <WinMeter
                            v-if="championOf(player)"
                            :win-rate="championOf(player)!.win_rate"
                        />
                    </div>
                </div>
            </div>
        </div>
    </Card>
</template>
