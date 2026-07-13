<script setup lang="ts">
import { ref } from 'vue'
import { RefreshCw, Search, Users } from '@lucide/vue'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { apiPlayer, apiPlayerRefresh, settingsGet, type PlayerProfile } from '../api'
import PageHeading from '../components/PageHeading.vue'
import PlayerPanel from '../components/PlayerPanel.vue'
import StatusPane from '../components/StatusPane.vue'

let riotId = ref('')
let profile = ref<PlayerProfile | null>(null)
let loading = ref(false)
let refreshing = ref(false)
let error = ref('')
let notice = ref('')

function parsed(): [string, string] | null {
    let raw = riotId.value.trim()
    let separator = raw.lastIndexOf('#')

    if (separator <= 0 || separator === raw.length - 1) return null

    return [raw.slice(0, separator), raw.slice(separator + 1)]
}

async function search(): Promise<void> {
    let identity = parsed()

    if (!identity) {
        error.value = 'Enter a Riot ID as name#tag.'

        return
    }

    loading.value = true
    error.value = ''
    notice.value = ''

    try {
        let settings = await settingsGet()

        profile.value = await apiPlayer(identity[0], identity[1], settings.region)
    } catch (raised) {
        error.value = String(raised)
        profile.value = null
    } finally {
        loading.value = false
    }
}

async function requestRefresh(): Promise<void> {
    let identity = parsed()

    if (!identity) return

    refreshing.value = true
    notice.value = ''

    try {
        let settings = await settingsGet()
        let result = await apiPlayerRefresh(identity[0], identity[1], settings.region)

        notice.value = result.queued
            ? 'Update queued. New matches appear after the next pass.'
            : 'An update is already in progress.'
    } catch (raised) {
        error.value = String(raised)
    } finally {
        refreshing.value = false
    }
}
</script>

<template>
    <section v-if="!profile && !loading && !error" class="mx-auto mt-20 max-w-md">
        <div class="rise-in flex flex-col items-center gap-5 text-center">
            <span class="flex size-14 items-center justify-center rounded-2xl bg-primary/12 ring-1 ring-primary/25 ring-inset">
                <Users class="size-6 text-primary" />
            </span>

            <div>
                <h1 class="text-2xl font-bold">Find a player</h1>
                <p class="mt-1.5 text-[13px] text-muted-foreground">
                    Look up ARAM match history by Riot ID.
                </p>
            </div>

            <form class="flex w-full gap-2" @submit.prevent="search">
                <Input
                    v-model="riotId"
                    class="h-10 flex-1 bg-white/[0.02]"
                    placeholder="Riot ID, e.g. Faker#KR1"
                    spellcheck="false"
                    autocomplete="off"
                />
                <Button class="h-10" type="submit" :disabled="loading">
                    <Search class="size-4" />
                    Search
                </Button>
            </form>
        </div>
    </section>

    <section v-else class="mx-auto max-w-6xl">
        <PageHeading eyebrow="Player lookup" title="Players">
            <form class="flex gap-2" @submit.prevent="search">
                <Input
                    v-model="riotId"
                    class="w-64 bg-white/[0.02]"
                    placeholder="Riot ID, e.g. Faker#KR1"
                    spellcheck="false"
                    autocomplete="off"
                />
                <Button type="submit" :disabled="loading">
                    <Search class="size-4" />
                    Search
                </Button>
                <Button
                    v-if="profile"
                    type="button"
                    variant="secondary"
                    :disabled="refreshing"
                    @click="requestRefresh"
                >
                    <RefreshCw class="size-4" :class="{ 'animate-spin': refreshing }" />
                    {{ refreshing ? 'Queueing…' : 'Update' }}
                </Button>
            </form>
        </PageHeading>

        <StatusPane v-if="error" variant="error" :message="error" />
        <p v-if="notice" class="rise-in mb-4 text-[13px] text-win">{{ notice }}</p>
        <StatusPane v-if="loading" variant="loading" message="Searching the mist…" />

        <PlayerPanel v-if="profile && !loading" :profile="profile" />
    </section>
</template>
