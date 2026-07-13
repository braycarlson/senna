<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { Card } from '@/components/ui/card'
import { apiStats, type ServiceStats } from '../api'
import PageHeading from '../components/PageHeading.vue'
import StatTile from '../components/StatTile.vue'

const VERSION = '0.1.0'

let stats = ref<ServiceStats | null>(null)
let statsFailed = ref(false)

onMounted(async () => {
    try {
        stats.value = await apiStats()
    } catch {
        statsFailed.value = true
    }
})
</script>

<template>
    <section class="mx-auto max-w-2xl">
        <PageHeading eyebrow="senna" title="About" />

        <div class="rise-in flex flex-col gap-4">
            <Card class="block rounded-xl p-5">
                <h2 class="mb-3 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    The app
                </h2>

                <p class="text-[13px] leading-relaxed text-muted-foreground">
                    senna is an ARAM companion for the Howling Abyss. It ranks champions by real
                    match results, shows builds, runes, and matchups per champion, and applies
                    your runes, summoner spells, and skins automatically in champ select.
                </p>

                <p class="mt-3 text-[13px] text-muted-foreground">
                    Version <span class="stat-mono text-foreground">{{ VERSION }}</span>
                </p>
            </Card>

            <Card class="block rounded-xl p-5">
                <h2 class="mb-3 font-display text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                    The data
                </h2>

                <p v-if="statsFailed" class="text-[13px] text-muted-foreground">
                    Service stats are unavailable. The stats service could not be reached at the
                    configured base URL.
                </p>

                <template v-else-if="stats">
                    <p class="mb-4 text-[13px] leading-relaxed text-muted-foreground">
                        Stats are built from ARAM matches gathered continuously from the Riot API.
                        Numbers below cover everything on record so far.
                    </p>

                    <div class="grid grid-cols-2 gap-3">
                        <StatTile label="matches" :value="stats.counts.matches_stored.toLocaleString()" />
                        <StatTile label="player records" :value="stats.counts.participants_stored.toLocaleString()" />
                        <StatTile label="players indexed" :value="stats.counts.accounts_done.toLocaleString()" />
                        <StatTile label="players queued" :value="stats.counts.accounts_pending.toLocaleString()" />
                    </div>
                </template>
            </Card>
        </div>
    </section>
</template>
