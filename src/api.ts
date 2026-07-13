import { invoke as invokeTauri } from '@tauri-apps/api/core'

export interface RunePage {
    primary_style: number
    sub_style: number
    perks: number[]
}

export interface ItemBlock {
    name: string
    items: number[]
}

export interface Loadout {
    skin_id?: number | null
    spells?: [number, number] | null
    runes?: RunePage | null
    items?: ItemBlock[] | null
}

export interface Settings {
    base_url: string
    token: string
    region: string
    auto_accept: boolean
    auto_runes: boolean
    auto_spells: boolean
    auto_items: boolean
    default_spells: [number, number] | null
    mode_spells: Record<string, [number, number]>
    random_skin: boolean
    loadouts: Record<number, Loadout>
    mode_loadouts: Record<string, Record<number, Loadout>>
}

export interface ChampionRow {
    champion_id: number
    champion_name: string
    games: number
    wins: number
    win_rate: number
    kda: number
    pick_rate: number
    kills_average: number
    deaths_average: number
    assists_average: number
    damage_average: number
    gold_average: number
}

export interface BuildItem {
    id: number
    name: string | null
}

export interface BuildEntry {
    id: number
    name: string | null
    games: number
    wins: number
    win_rate: number
}

export interface ItemSetEntry {
    items: BuildItem[]
    games: number
    wins: number
    win_rate: number
}

export interface SpellPairEntry {
    spell_a: number
    spell_a_name: string | null
    spell_b: number
    spell_b_name: string | null
    games: number
    wins: number
    win_rate: number
}

export interface RuneStyleEntry {
    primary_style: number
    primary_name: string | null
    sub_style: number
    sub_name: string | null
    games: number
    wins: number
    win_rate: number
}

export interface RunePageEntry {
    primary_style: number
    sub_style: number
    primary: BuildItem[]
    sub: BuildItem[]
    shards: BuildItem[]
    games: number
    wins: number
    win_rate: number
}

export interface SkillOrderEntry {
    skill_max_order: string
    games: number
    wins: number
    win_rate: number
}

export interface SkillSequenceEntry {
    sequence: string
    games: number
    wins: number
    win_rate: number
}

export interface ChampionBuild {
    champion_id: number
    champion_name: string | null
    patch: string
    queue_id: number
    starting_items: ItemSetEntry[]
    core_builds: ItemSetEntry[]
    boots: BuildEntry[]
    items: BuildEntry[]
    keystones: BuildEntry[]
    summoner_spells: SpellPairEntry[]
    rune_styles: RuneStyleEntry[]
    rune_pages: RunePageEntry[]
    skill_orders: SkillOrderEntry[]
    skill_sequences: SkillSequenceEntry[]
}

export interface PairRow {
    champion_id: number
    champion_name: string | null
    games: number
    wins: number
    win_rate: number
}

export interface PlayerChampionRow {
    champion_id: number
    champion_name: string
    games: number
    wins: number
    win_rate: number
    kda: number
}

export interface PlayerMatch {
    match_id: string
    patch: string
    queue_id: number
    champion_id: number
    champion_name: string
    win: boolean
    kills: number
    deaths: number
    assists: number
    game_creation: number
}

export interface PlayerProfile {
    puuid: string
    champions: PlayerChampionRow[]
    matches: PlayerMatch[]
}

export interface MatchParticipant {
    puuid: string
    champion_id: number
    champion_name: string
    team_id: number
    win: boolean
    kills: number
    deaths: number
    assists: number
    champion_level: number
    gold_earned: number
    damage_to_champions: number
    item0: number
    item1: number
    item2: number
    item3: number
    item4: number
    item5: number
    item6: number
    summoner1_id: number
    summoner2_id: number
    perk_keystone: number
    perk_primary_style: number
    perk_sub_style: number
}

export interface MatchDetail {
    match_id: string
    patch: string
    queue_id: number
    game_mode: string
    game_creation: number
    game_duration: number
    participants: MatchParticipant[]
}

export interface CrawlCounts {
    accounts_pending: number
    accounts_done: number
    matches_stored: number
    participants_stored: number
}

export interface ServiceStats {
    counts: CrawlCounts
    patches: string[]
}

export interface RefreshResponse {
    puuid: string
    queued: boolean
}

export interface CurrentSummoner {
    gameName: string
    tagLine: string
}

export interface LobbyPlayer {
    puuid: string
    name: string
    champion_id: number
    team: number
    self: boolean
}

export interface LiveSelection {
    in_select: boolean
    champion_id: number
    skin_id: number
    spell_d: number
    spell_f: number
    bench_enabled: boolean
    bench: number[]
    rerolls: number
    game_mode: string
}

export interface LcuEvent {
    ts: number
    kind: 'phase' | 'assign' | 'champion' | 'hover' | 'trade' | 'spells' | 'apply'
    note?: string
    player?: string
    self?: boolean
    other?: string
    champion_id?: number
    other_champion_id?: number
    spells?: [number, number]
    runes?: boolean
    items?: boolean
    skin?: boolean
}

export interface RunePageInfo {
    id: number
    name: string
}

const BROWSER_BASE_URL = 'http://127.0.0.1:8080'

const BROWSER_SETTINGS: Settings = {
    base_url: BROWSER_BASE_URL,
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
}

function browserQuery(pairs: Record<string, unknown>): string {
    let params = new URLSearchParams()

    for (let [name, value] of Object.entries(pairs)) {
        if (value !== undefined && value !== null) params.set(name, String(value))
    }

    let raw = params.toString()

    return raw ? `?${raw}` : ''
}

async function browserFetch<T>(path: string, method: string = 'GET'): Promise<T> {
    let response = await fetch(`${BROWSER_BASE_URL}${path}`, { method })

    if (!response.ok) throw new Error(`${response.status}: ${await response.text()}`)

    return response.json()
}

async function invokeBrowser<T>(command: string, args: Record<string, unknown> = {}): Promise<T> {
    switch (command) {
        case 'settings_get':
            return BROWSER_SETTINGS as T
        case 'settings_set':
            return undefined as T
        case 'lcu_status':
            return 'disconnected' as T
        case 'lcu_current_champion':
            return 0 as T
        case 'lcu_lobby':
            return [] as T
        case 'lcu_selection':
            return {
                in_select: false,
                champion_id: 0,
                skin_id: 0,
                spell_d: 0,
                spell_f: 0,
                bench_enabled: false,
                bench: [],
                rerolls: 0,
                game_mode: '',
            } as T
        case 'lcu_events':
            return [] as T
        case 'api_player_champions':
            return browserFetch(`/players/${String(args.puuid)}/champions${browserQuery({ limit: args.limit })}`)
        case 'api_match':
            return browserFetch(`/matches/${String(args.matchId)}`)
        case 'api_patches':
            return browserFetch('/patches')
        case 'api_stats':
            return browserFetch('/stats')
        case 'api_champions':
            return browserFetch(`/aram/champions${browserQuery({ patch: args.patch, queue: args.queue })}`)
        case 'api_tier':
            return browserFetch(`/aram/tier${browserQuery({ patch: args.patch, queue: args.queue, games_min: args.gamesMin })}`)
        case 'api_build':
            return browserFetch(`/aram/champions/${String(args.championId)}/build${browserQuery({ patch: args.patch, queue: args.queue })}`)
        case 'api_matchups':
            return browserFetch(`/aram/champions/${String(args.championId)}/matchups${browserQuery({ patch: args.patch, queue: args.queue, games_min: args.gamesMin })}`)
        case 'api_synergies':
            return browserFetch(`/aram/champions/${String(args.championId)}/synergies${browserQuery({ patch: args.patch, queue: args.queue, games_min: args.gamesMin })}`)
        case 'api_player':
            return browserFetch(`/players/by-riot-id/${String(args.name)}/${String(args.tag)}${browserQuery({ region: args.region, limit: args.limit })}`)
        case 'api_player_refresh':
            return browserFetch(`/players/by-riot-id/${String(args.name)}/${String(args.tag)}/refresh${browserQuery({ region: args.region })}`, 'POST')
        default:
            throw new Error(`${command} is not available in the browser`)
    }
}

let invoke: typeof invokeTauri =
    '__TAURI_INTERNALS__' in window ? invokeTauri : (invokeBrowser as typeof invokeTauri)

const CACHE_ENTRIES_MAX = 128

let cache = new Map<string, Promise<unknown>>()

function cached<T>(key: string, factory: () => Promise<T>): Promise<T> {
    let existing = cache.get(key)

    if (existing) return existing as Promise<T>

    if (cache.size >= CACHE_ENTRIES_MAX) {
        let oldest = cache.keys().next().value

        if (oldest !== undefined) cache.delete(oldest)
    }

    let promise = factory().catch((error) => {
        cache.delete(key)

        throw error
    })

    cache.set(key, promise)

    return promise
}

export let settingsGet = () => invoke<Settings>('settings_get')
export let settingsSet = (settings: Settings) => {
    if (!settings.base_url.trim()) throw new Error('base url must not be empty')

    return invoke<void>('settings_set', { settings })
}

export let apiPatches = () => cached('patches', () => invoke<string[]>('api_patches'))
export let apiStats = () => invoke<ServiceStats>('api_stats')

export let apiChampions = (patch?: string, queue?: number) =>
    cached(`champions:${patch}:${queue}`, () =>
        invoke<ChampionRow[]>('api_champions', { patch, queue }),
    )

export let apiTier = (patch?: string, queue?: number, gamesMin?: number) =>
    cached(`tier:${patch}:${queue}:${gamesMin}`, () =>
        invoke<ChampionRow[]>('api_tier', { patch, queue, gamesMin }),
    )

export let apiBuild = (championId: number, patch?: string, queue?: number) =>
    cached(`build:${championId}:${patch}:${queue}`, () =>
        invoke<ChampionBuild>('api_build', { championId, patch, queue }),
    )

export let apiMatchups = (championId: number, patch?: string, queue?: number, gamesMin?: number) =>
    cached(`matchups:${championId}:${patch}:${queue}:${gamesMin}`, () =>
        invoke<PairRow[]>('api_matchups', { championId, patch, queue, gamesMin }),
    )

export let apiSynergies = (championId: number, patch?: string, queue?: number, gamesMin?: number) =>
    cached(`synergies:${championId}:${patch}:${queue}:${gamesMin}`, () =>
        invoke<PairRow[]>('api_synergies', { championId, patch, queue, gamesMin }),
    )

export let apiPlayer = (name: string, tag: string, region?: string, limit?: number) =>
    invoke<PlayerProfile>('api_player', { name, tag, region, limit })

export let apiPlayerChampions = (puuid: string, limit?: number) =>
    cached(`playerchamps:${puuid}:${limit}`, () =>
        invoke<PlayerChampionRow[]>('api_player_champions', { puuid, limit }),
    )

export let apiMatch = (matchId: string) =>
    cached(`match:${matchId}`, () => invoke<MatchDetail>('api_match', { matchId }))

export let apiPlayerRefresh = (name: string, tag: string, region?: string) =>
    invoke<RefreshResponse>('api_player_refresh', { name, tag, region })

export let lcuStatus = () => invoke<string>('lcu_status')
export let lcuCurrentChampion = () => invoke<number>('lcu_current_champion')
export let lcuLobby = () => invoke<LobbyPlayer[]>('lcu_lobby')
export let lcuSelection = () => invoke<LiveSelection>('lcu_selection')
export let lcuEvents = () => invoke<LcuEvent[]>('lcu_events')
export let lcuCurrentPage = () => invoke<RunePageInfo>('lcu_current_page')
export let lcuPickableSkins = () => invoke<number[]>('lcu_pickable_skins')

export let lcuSetSpells = (spellD: number, spellF: number) => {
    if (spellD <= 0 || spellF <= 0 || spellD === spellF) {
        throw new Error('two distinct spells are required')
    }

    return invoke<void>('lcu_set_spells', { spellD, spellF })
}

export let lcuSetSkin = (skinId: number) => {
    if (skinId <= 0) throw new Error('skin id must be positive')

    return invoke<void>('lcu_set_skin', { skinId })
}

export let lcuBenchSwap = (championId: number) => {
    if (championId <= 0) throw new Error('champion id must be positive')

    return invoke<void>('lcu_bench_swap', { championId })
}

export let lcuReroll = () => invoke<void>('lcu_reroll')

export let lcuApplyStatsRunes = (championId: number) => {
    if (championId <= 0) throw new Error('champion id must be positive')

    return invoke<boolean>('lcu_apply_stats_runes', { championId })
}
export let lcuCurrentSummoner = () => invoke<CurrentSummoner>('lcu_current_summoner')
export let lcuOwnedSkins = (championId: number) => invoke<number[]>('lcu_owned_skins', { championId })
