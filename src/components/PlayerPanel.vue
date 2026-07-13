<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronRight } from '@lucide/vue'

import { Card } from '@/components/ui/card'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import type { PlayerProfile } from '../api'
import { percent } from '../format'
import ChampAvatar from './ChampAvatar.vue'
import StatTile from './StatTile.vue'
import WinMeter from './WinMeter.vue'

let props = defineProps<{ profile: PlayerProfile }>()

let router = useRouter()

let totals = computed(() => {
    let games = 0
    let wins = 0

    for (let row of props.profile.champions) {
        games += row.games
        wins += row.wins
    }

    return { games, wins, rate: games > 0 ? wins / games : 0 }
})

let topChampion = computed(() => {
    let best = null

    for (let row of props.profile.champions) {
        if (!best || row.games > best.games) best = row
    }

    return best
})

function timeAgo(epochMs: number): string {
    let hours = Math.floor((Date.now() - epochMs) / 3_600_000)

    if (hours < 1) return 'just now'
    if (hours < 24) return `${hours}h ago`

    return `${Math.floor(hours / 24)}d ago`
}
</script>

<template>
    <div class="rise-in flex flex-col gap-4">
        <div v-if="totals.games > 0" class="grid grid-cols-2 gap-3 min-[720px]:grid-cols-4">
            <StatTile label="games recorded" :value="totals.games.toLocaleString()" />
            <StatTile label="wins" :value="totals.wins.toLocaleString()" />
            <StatTile label="win rate" :value="`${percent(totals.rate)}%`" />
            <StatTile label="most played" :value="topChampion?.champion_name ?? ''" />
        </div>

        <div class="grid items-start gap-4 min-[1100px]:grid-cols-2">
            <Card class="gap-0 overflow-hidden rounded-xl py-0">
                <h2 class="border-b bg-white/[0.02] px-4 py-3 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    Champions
                </h2>

                <div v-if="profile.champions.length === 0" class="p-12 text-center text-muted-foreground">
                    No games on record for this player yet. Use Update to fetch them.
                </div>

                <Table v-else>
                    <TableHeader>
                        <TableRow class="hover:bg-transparent">
                            <TableHead>Champion</TableHead>
                            <TableHead class="text-right">Games</TableHead>
                            <TableHead class="text-right">KDA</TableHead>
                            <TableHead class="text-right">Win rate</TableHead>
                        </TableRow>
                    </TableHeader>

                    <TableBody>
                        <TableRow
                            v-for="row in profile.champions"
                            :key="row.champion_id"
                            class="border-white/[0.04]"
                        >
                            <TableCell>
                                <span class="flex items-center gap-2.5">
                                    <ChampAvatar :champion-id="row.champion_id" :name="row.champion_name" size="sm" />
                                    <span class="font-medium">{{ row.champion_name }}</span>
                                </span>
                            </TableCell>
                            <TableCell class="stat-mono text-right text-muted-foreground">
                                {{ row.games }}
                            </TableCell>
                            <TableCell class="stat-mono text-right">
                                {{ row.kda.toFixed(2) }}
                            </TableCell>
                            <TableCell><WinMeter :win-rate="row.win_rate" /></TableCell>
                        </TableRow>
                    </TableBody>
                </Table>
            </Card>

            <Card class="gap-0 overflow-hidden rounded-xl py-0">
                <h2 class="border-b bg-white/[0.02] px-4 py-3 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    Recent matches
                </h2>

                <div v-if="profile.matches.length === 0" class="p-12 text-center text-muted-foreground">
                    No stored matches.
                </div>

                <div v-else class="flex flex-col gap-1.5 p-3">
                    <div
                        v-for="game in profile.matches"
                        :key="game.match_id"
                        class="group flex cursor-pointer items-center gap-3 rounded-lg border-l-2 py-2 pr-3 pl-3.5 transition-colors"
                        :class="game.win ? 'border-win/70 bg-win/[0.05] hover:bg-win/[0.09]' : 'border-loss/70 bg-loss/[0.05] hover:bg-loss/[0.09]'"
                        @click="router.push(`/matches/${game.match_id}`)"
                    >
                        <ChampAvatar :champion-id="game.champion_id" :name="game.champion_name" />

                        <div class="flex flex-1 flex-col">
                            <span class="font-semibold">{{ game.champion_name }}</span>
                            <span class="text-xs text-muted-foreground">
                                {{ game.patch }} · {{ timeAgo(game.game_creation) }}
                            </span>
                        </div>

                        <span class="stat-mono text-right text-[13px] font-medium">
                            {{ game.kills }} / <span class="text-loss">{{ game.deaths }}</span> /
                            {{ game.assists }}
                        </span>

                        <span
                            class="min-w-[60px] text-right font-display text-xs font-bold tracking-wider uppercase"
                            :class="game.win ? 'text-win' : 'text-loss'"
                        >
                            {{ game.win ? 'Victory' : 'Defeat' }}
                        </span>

                        <ChevronRight class="size-4 text-muted-foreground/40 transition-colors group-hover:text-foreground" />
                    </div>
                </div>
            </Card>
        </div>
    </div>
</template>
