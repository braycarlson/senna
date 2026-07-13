import { createRouter, createWebHistory } from 'vue-router'

import { context, contextInit } from './state'
import AboutView from './views/AboutView.vue'
import ChampionView from './views/ChampionView.vue'
import EventLogView from './views/EventLogView.vue'
import HomeView from './views/HomeView.vue'
import MatchView from './views/MatchView.vue'
import MoversView from './views/MoversView.vue'
import PlayerView from './views/PlayerView.vue'
import PreferencesView from './views/PreferencesView.vue'
import SettingsView from './views/SettingsView.vue'
import TierView from './views/TierView.vue'

export let router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/', name: 'home', component: HomeView },
        { path: '/tier', name: 'tier', component: TierView },
        { path: '/movers', name: 'movers', component: MoversView },
        {
            path: '/champions/:id',
            name: 'champion',
            component: ChampionView,
            props: true,
        },
        {
            path: '/matches/:id',
            name: 'match',
            component: MatchView,
            props: true,
        },
        { path: '/player', name: 'player', component: PlayerView },
        { path: '/profile', redirect: '/' },
        {
            path: '/preferences',
            name: 'preferences',
            component: PreferencesView,
        },
        {
            path: '/settings',
            name: 'settings',
            component: SettingsView,
        },
        {
            path: '/events',
            name: 'events',
            component: EventLogView,
        },
        {
            path: '/about',
            name: 'about',
            component: AboutView,
        },
    ],
})

router.beforeEach(() => {
    if (context.patchesFailed) {
        void contextInit()
    }
})
