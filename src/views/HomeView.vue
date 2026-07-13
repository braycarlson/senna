<script setup lang="ts">
import { listen } from '@tauri-apps/api/event'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { ArrowRight, RefreshCw } from '@lucide/vue'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
    lcuCurrentChampion,
    lcuLobby,
    lcuCurrentSummoner,
    apiBuild,
    apiChampions,
    apiPlayer,
    apiPlayerRefresh,
    settingsGet,
    type ChampionBuild,
    type ChampionRow,
    type LobbyPlayer,
    type PlayerProfile,
} from '../api'
import { assetsEnsureBuildIcons, itemIcon, runeIcon, spellIcon } from '../assets'
import { percent } from '../format'
import { useDelayedLoader } from '../loading'
import { context } from '../state'
import BuildPanels from '../components/BuildPanels.vue'
import ChampAvatar from '../components/ChampAvatar.vue'
import LiveCockpit from '../components/LiveCockpit.vue'
import LoadingPane from '../components/LoadingPane.vue'
import LiveMatchPanel from '../components/LiveMatchPanel.vue'
import PageHeading from '../components/PageHeading.vue'
import PlayerPanel from '../components/PlayerPanel.vue'
import StatusPane from '../components/StatusPane.vue'

let riotId = ref<[string, string] | null>(null)
let profile = ref<PlayerProfile | null>(null)
let { loading, showLoader, begin, end } = useDelayedLoader()
let refreshing = ref(false)
let disconnected = ref(false)
let error = ref('')
let notice = ref('')

let championId = ref(0)
let summary = ref<ChampionRow | null>(null)
let build = ref<ChampionBuild | null>(null)
let roster = ref<LobbyPlayer[]>([])

let unlisteners: Array<() => void> = []

async function loadProfile(): Promise<void> {
    begin()

    disconnected.value = false
    error.value = ''

    try {
        let summoner = await lcuCurrentSummoner()

        if (!summoner.gameName || !summoner.tagLine) {
            disconnected.value = true

            return
        }

        riotId.value = [summoner.gameName, summoner.tagLine]

        let settings = await settingsGet()

        profile.value = await apiPlayer(summoner.gameName, summoner.tagLine, settings.region)
    } catch (raised) {
        let message = String(raised)

        if (message.includes('not available')) {
            disconnected.value = true
        } else {
            error.value = message
        }
    } finally {
        end()
    }
}

async function requestRefresh(): Promise<void> {
    if (!riotId.value) return

    refreshing.value = true
    notice.value = ''

    try {
        let settings = await settingsGet()
        let result = await apiPlayerRefresh(riotId.value[0], riotId.value[1], settings.region)

        notice.value = result.queued
            ? 'Update queued. New matches appear after the next pass.'
            : 'An update is already in progress.'
    } catch (raised) {
        error.value = String(raised)
    } finally {
        refreshing.value = false
    }
}

async function loadChampion(): Promise<void> {
    if (!championId.value || !context.patch) {
        summary.value = null
        build.value = null

        return
    }

    assetsEnsureBuildIcons()

    try {
        let [championsResult, buildResult] = await Promise.all([
            apiChampions(context.patch, context.queue),
            apiBuild(championId.value, context.patch, context.queue),
        ])

        summary.value =
            championsResult.find((row) => row.champion_id === championId.value) ?? null
        build.value = buildResult
    } catch {
        summary.value = null
        build.value = null
    }
}

watch([championId, () => context.patch], loadChampion)

onMounted(async () => {
    if ('__TAURI_INTERNALS__' in window) {
        unlisteners = await Promise.all([
            listen<number>('lcu-champion', (event) => {
                championId.value = event.payload
            }),

            listen<LobbyPlayer[]>('lcu-lobby', (event) => {
                roster.value = event.payload
            }),
        ])

        championId.value = await lcuCurrentChampion()
        roster.value = await lcuLobby()
    }

    await loadProfile()
})

onUnmounted(() => {
    for (let unlisten of unlisteners) unlisten()
})
</script>

<template>
    <section class="mx-auto max-w-6xl">
        <PageHeading eyebrow="Home" :title="riotId ? `${riotId[0]} #${riotId[1]}` : 'Welcome'">
            <Button
                v-if="riotId"
                type="button"
                variant="secondary"
                size="sm"
                :disabled="refreshing"
                @click="requestRefresh"
            >
                <RefreshCw class="size-4" :class="{ 'animate-spin': refreshing }" />
                {{ refreshing ? 'Queueing…' : 'Update' }}
            </Button>
        </PageHeading>

        <Card class="rise-in mb-4 block overflow-hidden rounded-xl p-0">
            <div
                v-if="championId && build"
                class="flex flex-wrap items-center gap-6 bg-gradient-to-r from-primary/[0.12] via-transparent to-transparent p-5"
            >
                <div class="flex items-center gap-4">
                    <ChampAvatar
                        class="ring-2 ring-primary/40"
                        :champion-id="championId"
                        :name="build.champion_name"
                        size="lg"
                    />

                    <div class="flex flex-col gap-1.5">
                        <h2 class="text-2xl leading-none font-bold">
                            {{ build.champion_name ?? `Champion ${championId}` }}
                        </h2>

                        <div class="flex items-center gap-2">
                            <Badge class="bg-primary font-semibold text-primary-foreground">
                                Live
                            </Badge>

                            <span v-if="summary" class="stat-mono text-xs text-muted-foreground">
                                {{ percent(summary.win_rate) }}% WR ·
                                {{ summary.kda.toFixed(2) }} KDA ·
                                {{ summary.games.toLocaleString() }} games
                            </span>
                        </div>
                    </div>
                </div>

                <div class="flex flex-1 flex-wrap items-center justify-end gap-5">
                    <span
                        v-if="build.summoner_spells.length > 0"
                        class="flex items-center gap-1.5"
                    >
                        <img
                            class="size-7 rounded-md bg-layer"
                            :src="spellIcon(build.summoner_spells[0].spell_a)"
                            :alt="build.summoner_spells[0].spell_a_name ?? ''"
                        />
                        <img
                            class="size-7 rounded-md bg-layer"
                            :src="spellIcon(build.summoner_spells[0].spell_b)"
                            :alt="build.summoner_spells[0].spell_b_name ?? ''"
                        />
                    </span>

                    <span
                        v-if="build.rune_pages.length > 0 && build.rune_pages[0].primary.length > 0"
                        class="flex items-center"
                    >
                        <img
                            class="size-8"
                            :src="runeIcon(build.rune_pages[0].primary[0].id)"
                            :alt="build.rune_pages[0].primary[0].name ?? ''"
                        />
                    </span>

                    <span v-if="build.core_builds.length > 0" class="flex items-center gap-1.5">
                        <img
                            v-for="item in build.core_builds[0].items"
                            :key="item.id"
                            class="size-7 rounded-md bg-layer"
                            :src="itemIcon(item.id)"
                            :alt="item.name ?? String(item.id)"
                            :title="item.name ?? String(item.id)"
                        />
                    </span>

                    <Button as-child size="sm" variant="secondary">
                        <router-link :to="`/champions/${championId}`">
                            Full build
                            <ArrowRight class="size-4" />
                        </router-link>
                    </Button>
                </div>
            </div>

            <div v-else class="flex items-center gap-3 p-5">
                <span class="relative flex size-2">
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary/60"></span>
                    <span class="relative inline-flex size-2 rounded-full bg-primary/70"></span>
                </span>

                <span class="text-[13px] text-muted-foreground">
                    Waiting for champ select. Your champion's build shows here the moment you hover
                    one.
                </span>
            </div>
        </Card>

        <LiveCockpit class="rise-in mb-4" />

        <LiveMatchPanel v-if="roster.length > 0" class="rise-in mb-4" :roster="roster" />

        <BuildPanels
            v-if="championId && build"
            class="rise-in mb-4"
            :build="build"
            :sections="['core', 'page', 'starting', 'spells', 'skills']"
        />

        <StatusPane v-if="error" variant="error" :message="error" />
        <p v-if="notice" class="rise-in mb-4 text-[13px] text-win">{{ notice }}</p>

        <transition name="fade" mode="out-in">
            <LoadingPane
                v-if="loading && showLoader"
                key="loading"
                class="min-h-[30vh]"
                message="Asking the League client who you are…"
            />

            <div v-else-if="loading" key="pending" class="min-h-[30vh]"></div>

            <StatusPane
                v-else-if="disconnected"
                key="disconnected"
                variant="empty"
                message="The League client is not running. Start it and reopen this page to see your profile."
            />

            <PlayerPanel v-else-if="profile" key="profile" :profile="profile" />

            <div v-else key="none"></div>
        </transition>
    </section>
</template>
