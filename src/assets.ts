import { reactive } from 'vue'

import { settingsGet } from './api'

const DDRAGON = 'https://ddragon.leagueoflegends.com'
const CDRAGON_RAW = 'https://raw.communitydragon.org'
const CDRAGON_CDN = 'https://cdn.communitydragon.org'
const CDRAGON_PLUGIN_PATH = 'latest/plugins/rcp-be-lol-game-data/global/default'
const PROBE_TIMEOUT_MS = 2500

const SHARD_PATHS: Record<number, string> = {
    5001: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodshealthplusicon.png`,
    5005: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodsattackspeedicon.png`,
    5007: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodscdrscalingicon.png`,
    5008: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodsadaptiveforceicon.png`,
    5010: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodsmovementspeedicon.png`,
    5011: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodshealthscalingicon.png`,
    5013: `${CDRAGON_PLUGIN_PATH}/v1/perk-images/statmods/statmodstenacityicon.png`,
}

export interface RuneEntry {
    id: number
    name: string
}

export interface RuneTree {
    id: number
    name: string
    slots: RuneEntry[][]
}

export interface ItemEntry {
    id: number
    name: string
    gold: number
    tags: string[]
}

interface AssetState {
    version: string
    dragonBase: string
    runeIcons: Record<number, string>
    runeTrees: RuneTree[]
    spellImages: Record<number, string>
    itemCatalog: ItemEntry[]
}

export let assets = reactive<AssetState>({
    version: '',
    dragonBase: '',
    runeIcons: {},
    runeTrees: [],
    spellImages: {},
    itemCatalog: [],
})

let dragonPromise: Promise<void> | null = null
let versionPromise: Promise<void> | null = null
let runesPromise: Promise<void> | null = null
let spellsPromise: Promise<void> | null = null
let itemsPromise: Promise<void> | null = null

function ensureDragonBase(): Promise<void> {
    if (!dragonPromise) {
        dragonPromise = (async () => {
            try {
                let settings = await settingsGet()
                let base = settings.base_url.trim().replace(/\/+$/, '')

                if (!base) return

                let controller = new AbortController()
                let timer = window.setTimeout(() => controller.abort(), PROBE_TIMEOUT_MS)

                let response = await fetch(`${base}/dragon/ddragon/api/versions.json`, {
                    signal: controller.signal,
                })

                window.clearTimeout(timer)

                if (response.ok) assets.dragonBase = `${base}/dragon`
            } catch {
                assets.dragonBase = ''
            }
        })()
    }

    return dragonPromise
}

function ddragonUrl(path: string): string {
    return assets.dragonBase ? `${assets.dragonBase}/ddragon/${path}` : `${DDRAGON}/${path}`
}

function cdragonUrl(path: string): string {
    return assets.dragonBase ? `${assets.dragonBase}/cdragon/${path}` : `${CDRAGON_RAW}/${path}`
}

interface RawRuneTree {
    id: number
    name: string
    icon: string
    slots: { runes: { id: number, name: string, icon: string }[] }[]
}

interface RawSummonerFile {
    data: Record<string, { key: string, image: { full: string } }>
}

interface RawItem {
    name: string
    gold?: { purchasable?: boolean, total: number }
    inStore?: boolean
    hideFromAll?: boolean
    requiredAlly?: string
    requiredChampion?: string
    maps?: Record<string, boolean>
    tags?: string[]
}

interface RawItemFile {
    data: Record<string, RawItem>
}

interface RawChampionSummary {
    id: number
    name: string
}

interface RawSkin {
    id: number
    name: string
    isBase?: boolean
    tilePath?: string
}

interface RawChampionDetail {
    skins?: RawSkin[]
}

async function fetchStaticJson<T>(path: string, direct: string): Promise<T> {
    await ensureDragonBase()

    let primary = path.startsWith('latest/') ? cdragonUrl(path) : ddragonUrl(path)

    if (primary !== direct) {
        try {
            let response = await fetch(primary)

            if (response.ok) return response.json()
        } catch (error) {
            console.error('primary asset fetch failed; falling back', error)
        }
    }

    let response = await fetch(direct)

    if (!response.ok) throw new Error(`${response.status}: static data fetch failed`)

    return response.json()
}

function ensureVersion(): Promise<void> {
    if (!versionPromise) {
        versionPromise = fetchStaticJson<string[]>('api/versions.json', `${DDRAGON}/api/versions.json`)
            .then((versions) => {
                assets.version = versions[0]
            })
            .catch((error) => {
                versionPromise = null

                console.error('version manifest load failed', error)
            })
    }

    return versionPromise
}

function ensureRunes(): Promise<void> {
    if (!runesPromise) {
        runesPromise = ensureVersion()
            .then(() =>
                fetchStaticJson<RawRuneTree[]>(
                    `cdn/${assets.version}/data/en_US/runesReforged.json`,
                    `${DDRAGON}/cdn/${assets.version}/data/en_US/runesReforged.json`,
                ),
            )
            .then((trees) => {
                let structured: RuneTree[] = []

                for (let tree of trees) {
                    assets.runeIcons[tree.id] = ddragonUrl(`cdn/img/${tree.icon}`)

                    let slots: RuneEntry[][] = []

                    for (let slot of tree.slots) {
                        let entries: RuneEntry[] = []

                        for (let rune of slot.runes) {
                            assets.runeIcons[rune.id] = ddragonUrl(`cdn/img/${rune.icon}`)

                            entries.push({ id: rune.id, name: rune.name })
                        }

                        slots.push(entries)
                    }

                    structured.push({ id: tree.id, name: tree.name, slots })
                }

                assets.runeTrees = structured
            })
            .catch((error) => {
                runesPromise = null

                console.error('rune manifest load failed', error)
            })
    }

    return runesPromise
}

function ensureSpells(): Promise<void> {
    if (!spellsPromise) {
        spellsPromise = ensureVersion()
            .then(() =>
                fetchStaticJson<RawSummonerFile>(
                    `cdn/${assets.version}/data/en_US/summoner.json`,
                    `${DDRAGON}/cdn/${assets.version}/data/en_US/summoner.json`,
                ),
            )
            .then((file) => {
                for (let key of Object.keys(file.data)) {
                    let spell = file.data[key]

                    assets.spellImages[Number(spell.key)] = spell.image.full
                }
            })
            .catch((error) => {
                spellsPromise = null

                console.error('spell manifest load failed', error)
            })
    }

    return spellsPromise
}

function ensureItems(): Promise<void> {
    if (!itemsPromise) {
        itemsPromise = ensureVersion()
            .then(() =>
                fetchStaticJson<RawItemFile>(
                    `cdn/${assets.version}/data/en_US/item.json`,
                    `${DDRAGON}/cdn/${assets.version}/data/en_US/item.json`,
                ),
            )
            .then((file) => {
                let byName = new Map<string, ItemEntry>()

                for (let key of Object.keys(file.data)) {
                    let item = file.data[key]
                    let gold = item.gold

                    if (!gold?.purchasable || gold.total <= 0) continue
                    if (item.inStore === false || item.hideFromAll) continue
                    if (item.requiredAlly || item.requiredChampion) continue
                    if (!item.maps?.['11'] && !item.maps?.['12']) continue

                    let id = Number(key)
                    let existing = byName.get(item.name)

                    if (!existing || id < existing.id) {
                        byName.set(item.name, {
                            id,
                            name: item.name,
                            gold: gold.total,
                            tags: item.tags ?? [],
                        })
                    }
                }

                let entries = [...byName.values()]

                entries.sort((a, b) => a.name.localeCompare(b.name))

                assets.itemCatalog = entries
            })
            .catch((error) => {
                itemsPromise = null

                console.error('item manifest load failed', error)
            })
    }

    return itemsPromise
}

export function assetsEnsureBuildIcons(): void {
    void ensureVersion()
    void ensureRunes()
    void ensureSpells()
}

export function assetsEnsureLoadoutEditors(): void {
    assetsEnsureBuildIcons()

    void ensureItems()
}

export interface ChampionEntry {
    id: number
    name: string
}

export async function fetchChampionCatalog(): Promise<ChampionEntry[]> {
    let path = `${CDRAGON_PLUGIN_PATH}/v1/champion-summary.json`
    let data = await fetchStaticJson<RawChampionSummary[]>(path, `${CDRAGON_RAW}/${path}`)
    let champions = Array.isArray(data) ? data : []

    return champions
        .filter((champion) => champion.id > 0)
        .map((champion) => ({ id: champion.id, name: champion.name }))
        .sort((a, b) => a.name.localeCompare(b.name))
}

export interface ChampionSkin {
    id: number
    name: string
    isBase: boolean
    tile: string
}

function cdragonAssetPath(path: string): string {
    return `${CDRAGON_PLUGIN_PATH}/${path.replace(/^\/lol-game-data\/assets\//i, '').toLowerCase()}`
}

export async function fetchChampionSkins(championId: number): Promise<ChampionSkin[]> {
    let path = `${CDRAGON_PLUGIN_PATH}/v1/champions/${championId}.json`
    let data = await fetchStaticJson<RawChampionDetail>(path, `${CDRAGON_RAW}/${path}`)
    let skins = Array.isArray(data.skins) ? data.skins : []

    return skins
        .filter((skin): skin is RawSkin & { tilePath: string } => Boolean(skin.tilePath))
        .map((skin) => ({
            id: skin.id,
            name: skin.isBase ? 'Default' : skin.name,
            isBase: Boolean(skin.isBase),
            tile: cdragonUrl(cdragonAssetPath(skin.tilePath)),
        }))
}

export function championSquare(championId: number): string {
    if (assets.dragonBase) return `${assets.dragonBase}/csquare/${championId}`

    return `${CDRAGON_CDN}/latest/champion/${championId}/square`
}

export function itemIcon(itemId: number): string {
    if (!assets.version) return ''

    return ddragonUrl(`cdn/${assets.version}/img/item/${itemId}.png`)
}

export function runeIcon(runeId: number): string {
    let stored = assets.runeIcons[runeId]

    if (stored) return stored

    let shard = SHARD_PATHS[runeId]

    return shard ? cdragonUrl(shard) : ''
}

export function spellIcon(spellId: number): string {
    let image = assets.spellImages[spellId]

    if (!image || !assets.version) return ''

    return ddragonUrl(`cdn/${assets.version}/img/spell/${image}`)
}
