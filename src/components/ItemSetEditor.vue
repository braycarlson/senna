<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowDownAZ, ArrowDown01, ArrowUp01, ChevronDown, ChevronUp, PackagePlus, X } from '@lucide/vue'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { assets, itemIcon } from '../assets'
import { type ItemBlock } from '../api'

const BLOCKS_MAX = 8
const BLOCK_ITEMS_MAX = 12

const FILTERS = [
    { label: 'Attack Damage', tags: ['Damage'] },
    { label: 'Critical Strike', tags: ['CriticalStrike'] },
    { label: 'Attack Speed', tags: ['AttackSpeed'] },
    { label: 'On-Hit', tags: ['OnHit'] },
    { label: 'Armor Pen', tags: ['ArmorPenetration'] },
    { label: 'Ability Power', tags: ['SpellDamage'] },
    { label: 'Mana', tags: ['Mana', 'ManaRegen'] },
    { label: 'Magic Pen', tags: ['MagicPenetration'] },
    { label: 'Health', tags: ['Health', 'HealthRegen'] },
    { label: 'Armor', tags: ['Armor'] },
    { label: 'Magic Resist', tags: ['SpellBlock'] },
    { label: 'Ability Haste', tags: ['CooldownReduction'] },
    { label: 'Movement', tags: ['Boots', 'NonbootsMovement'] },
    { label: 'Sustain', tags: ['LifeSteal', 'SpellVamp'] },
]

type SortKey = 'gold-asc' | 'gold-desc' | 'name'

type DragSource =
    | { kind: 'catalog', id: number }
    | { kind: 'chip', block: number, index: number }

let model = defineModel<ItemBlock[] | null>({ required: true })

let blocks = ref<ItemBlock[]>([])
let active = ref(0)
let query = ref('')
let filters = ref<Set<number>>(new Set())
let sort = ref<SortKey>('gold-asc')
let drag = ref<DragSource | null>(null)
let dragOver = ref<number | 'new' | 'remove' | null>(null)
let lastCommitted = ''

let itemNames = computed(() => {
    let names: Record<number, string> = {}

    for (let item of assets.itemCatalog) names[item.id] = item.name

    return names
})

let results = computed(() => {
    let needle = query.value.trim().toLowerCase()
    let groups = [...filters.value].map((index) => FILTERS[index].tags)

    let matched = assets.itemCatalog.filter((item) => {
        if (needle && !item.name.toLowerCase().includes(needle)) return false

        return groups.every((tags) => tags.some((tag) => item.tags.includes(tag)))
    })

    matched.sort((a, b) => {
        if (sort.value === 'name') return a.name.localeCompare(b.name)
        if (sort.value === 'gold-desc') return b.gold - a.gold || a.name.localeCompare(b.name)

        return a.gold - b.gold || a.name.localeCompare(b.name)
    })

    return matched
})

function sortCycle(): void {
    sort.value =
        sort.value === 'gold-asc' ? 'gold-desc' : sort.value === 'gold-desc' ? 'name' : 'gold-asc'
}

function filterToggle(index: number): void {
    let next = new Set(filters.value)

    if (next.has(index)) {
        next.delete(index)
    } else {
        next.add(index)
    }

    filters.value = next
}

function cleaned(): ItemBlock[] {
    return blocks.value
        .map((block) => ({ name: block.name.trim(), items: [...block.items] }))
        .filter((block) => block.name.length > 0 || block.items.length > 0)
}

function commit(): void {
    let next = cleaned()
    let value = next.length > 0 ? next : null

    lastCommitted = JSON.stringify(value)

    model.value = value
}

function hydrate(): void {
    let value = model.value ?? null

    if (JSON.stringify(value) === lastCommitted) return

    lastCommitted = JSON.stringify(value)

    blocks.value = value
        ? value.map((block) => ({ name: block.name, items: [...block.items] }))
        : [{ name: 'Core build', items: [] }]

    active.value = 0
}

function blockAdd(): void {
    if (blocks.value.length >= BLOCKS_MAX) return

    blocks.value.push({ name: `Section ${blocks.value.length + 1}`, items: [] })

    active.value = blocks.value.length - 1

    commit()
}

function blockRemove(index: number): void {
    blocks.value.splice(index, 1)

    if (blocks.value.length === 0) blocks.value.push({ name: 'Core build', items: [] })
    if (active.value >= blocks.value.length) active.value = blocks.value.length - 1

    commit()
}

function blockMove(index: number, delta: number): void {
    let target = index + delta

    if (target < 0 || target >= blocks.value.length) return

    let [moved] = blocks.value.splice(index, 1)

    blocks.value.splice(target, 0, moved)

    if (active.value === index) active.value = target

    commit()
}

function itemAdd(blockIndex: number, id: number): void {
    let block = blocks.value[blockIndex]

    if (!block) return
    if (block.items.includes(id)) return
    if (block.items.length >= BLOCK_ITEMS_MAX) return

    block.items.push(id)

    commit()
}

function itemRemove(blockIndex: number, itemIndex: number): void {
    blocks.value[blockIndex].items.splice(itemIndex, 1)

    commit()
}

function dragPayload(event: DragEvent): void {
    if (!event.dataTransfer) return

    event.dataTransfer.setData('text/plain', 'senna-item')
    event.dataTransfer.effectAllowed = 'move'
}

function dragCatalogStart(id: number, event: DragEvent): void {
    dragPayload(event)

    drag.value = { kind: 'catalog', id }
}

function dragChipStart(blockIndex: number, itemIndex: number, event: DragEvent): void {
    dragPayload(event)

    drag.value = { kind: 'chip', block: blockIndex, index: itemIndex }
}

function dragEnd(): void {
    drag.value = null
    dragOver.value = null
}

function dropOnBlock(targetBlock: number, targetIndex: number | null): void {
    let source = drag.value

    dragEnd()

    if (!source) return

    let items = blocks.value[targetBlock].items

    if (source.kind === 'catalog') {
        if (items.includes(source.id)) return
        if (items.length >= BLOCK_ITEMS_MAX) return

        items.splice(targetIndex ?? items.length, 0, source.id)

        active.value = targetBlock

        commit()

        return
    }

    let [moved] = blocks.value[source.block].items.splice(source.index, 1)

    if (targetBlock !== source.block) {
        if (items.length >= BLOCK_ITEMS_MAX || items.includes(moved)) {
            blocks.value[source.block].items.splice(source.index, 0, moved)

            return
        }
    }

    let index = targetIndex ?? items.length

    if (targetBlock === source.block && targetIndex != null && source.index < targetIndex) {
        index = targetIndex - 1
    }

    items.splice(index, 0, moved)

    commit()
}

function dropOnNew(): void {
    let source = drag.value

    dragEnd()

    if (!source) return
    if (blocks.value.length >= BLOCKS_MAX) return

    let id

    if (source.kind === 'catalog') {
        id = source.id
    } else {
        ;[id] = blocks.value[source.block].items.splice(source.index, 1)
    }

    blocks.value.push({ name: `Section ${blocks.value.length + 1}`, items: [id] })

    active.value = blocks.value.length - 1

    commit()
}

function dropOnCatalog(): void {
    let source = drag.value

    dragEnd()

    if (source?.kind !== 'chip') return

    itemRemove(source.block, source.index)
}

watch(model, hydrate, { immediate: true })
</script>

<template>
    <div v-if="assets.itemCatalog.length === 0" class="text-xs text-muted-foreground">
        Loading item data…
    </div>

    <div v-else class="grid gap-4 min-[860px]:grid-cols-[310px_1fr]">
        <div class="flex h-[62vh] min-h-0 flex-col gap-2.5">
            <div class="flex items-center gap-2">
                <Input
                    v-model="query"
                    class="h-8 flex-1 bg-white/[0.02]"
                    placeholder="Search items…"
                    spellcheck="false"
                />

                <Button size="xs" variant="outline" @click="sortCycle">
                    <ArrowDown01 v-if="sort === 'gold-asc'" class="size-3.5" />
                    <ArrowUp01 v-else-if="sort === 'gold-desc'" class="size-3.5" />
                    <ArrowDownAZ v-else class="size-3.5" />
                    {{ sort === 'name' ? 'Name' : 'Gold' }}
                </Button>
            </div>

            <div class="flex flex-wrap gap-1">
                <button
                    v-for="(group, index) in FILTERS"
                    :key="group.label"
                    class="filter-pill"
                    :class="{ active: filters.has(index) }"
                    type="button"
                    @click="filterToggle(index)"
                >
                    {{ group.label }}
                </button>

                <button
                    v-if="filters.size > 0"
                    class="filter-pill clear"
                    type="button"
                    @click="filters = new Set()"
                >
                    <X class="size-3" />
                    Clear
                </button>
            </div>

            <div
                class="grid min-h-0 flex-1 auto-rows-min grid-cols-4 gap-1.5 overflow-y-auto rounded-lg bg-white/[0.015] p-2 ring-1 ring-white/[0.03] ring-inset transition-all"
                :class="{ 'ring-loss/50': drag?.kind === 'chip' }"
                @dragover.prevent="dragOver = 'remove'"
                @drop.prevent="dropOnCatalog"
            >
                <button
                    v-for="item in results"
                    :key="item.id"
                    class="catalog-item"
                    type="button"
                    draggable="true"
                    :title="`${item.name} · ${item.gold}g`"
                    @dragstart="dragCatalogStart(item.id, $event)"
                    @dragend="dragEnd"
                    @click="itemAdd(active, item.id)"
                >
                    <img class="size-10 rounded-md bg-layer" :src="itemIcon(item.id)" alt="" />
                    <span class="stat-mono text-[10px] text-muted-foreground">{{ item.gold }}</span>
                </button>

                <p v-if="results.length === 0" class="col-span-4 py-6 text-center text-xs text-muted-foreground">
                    No items match.
                </p>
            </div>

            <p class="text-[11px] text-muted-foreground/70">
                Click an item to add it to the highlighted section, or drag it into any section.
                Drag a section item back here to remove it.
            </p>
        </div>

        <div class="flex h-[62vh] min-h-0 flex-col gap-2.5 overflow-y-auto pr-1">
            <div
                v-for="(block, blockIndex) in blocks"
                :key="blockIndex"
                class="block-card"
                :class="{
                    active: active === blockIndex,
                    over: dragOver === blockIndex && drag !== null,
                }"
                @click="active = blockIndex"
                @dragover.prevent="dragOver = blockIndex"
                @drop.prevent="dropOnBlock(blockIndex, null)"
            >
                <div class="flex items-center gap-2">
                    <Input
                        :model-value="block.name"
                        class="h-7 max-w-44 bg-white/[0.02] text-[13px]"
                        placeholder="Section name"
                        spellcheck="false"
                        @update:model-value="block.name = String($event ?? '')"
                        @blur="commit"
                        @keydown.enter="commit"
                    />

                    <span class="text-xs text-muted-foreground">
                        {{ block.items.length }}/{{ BLOCK_ITEMS_MAX }}
                    </span>

                    <span class="ml-auto flex items-center gap-0.5">
                        <Button
                            size="icon-xs"
                            variant="ghost"
                            :disabled="blockIndex === 0"
                            @click.stop="blockMove(blockIndex, -1)"
                        >
                            <ChevronUp class="size-3.5" />
                        </Button>
                        <Button
                            size="icon-xs"
                            variant="ghost"
                            :disabled="blockIndex === blocks.length - 1"
                            @click.stop="blockMove(blockIndex, 1)"
                        >
                            <ChevronDown class="size-3.5" />
                        </Button>
                        <Button size="icon-xs" variant="ghost" @click.stop="blockRemove(blockIndex)">
                            <X class="size-3.5" />
                        </Button>
                    </span>
                </div>

                <div v-if="block.items.length > 0" class="flex flex-wrap items-center gap-2">
                    <span
                        v-for="(id, itemIndex) in block.items"
                        :key="`${id}-${itemIndex}`"
                        class="item-chip"
                        :class="{ dragging: drag?.kind === 'chip' && drag.block === blockIndex && drag.index === itemIndex }"
                        draggable="true"
                        :title="itemNames[id] ?? String(id)"
                        @dragstart="dragChipStart(blockIndex, itemIndex, $event)"
                        @dragend="dragEnd"
                        @dragover.prevent.stop="dragOver = blockIndex"
                        @drop.prevent.stop="dropOnBlock(blockIndex, itemIndex)"
                    >
                        <img class="size-10 rounded-md bg-layer" :src="itemIcon(id)" alt="" />
                        <button
                            class="item-chip-remove"
                            type="button"
                            @click.stop="itemRemove(blockIndex, itemIndex)"
                        >
                            <X class="size-3" />
                        </button>
                    </span>
                </div>

                <p v-else class="text-xs text-muted-foreground/70">
                    Drop or click items to fill this section.
                </p>
            </div>

            <button
                v-if="blocks.length < BLOCKS_MAX"
                class="new-block"
                :class="{ over: dragOver === 'new' && drag !== null }"
                type="button"
                @click="blockAdd"
                @dragover.prevent="dragOver = 'new'"
                @drop.prevent="dropOnNew"
            >
                <PackagePlus class="size-4" />
                Drop an item here to start a new section, or click to add one.
            </button>
        </div>
    </div>
</template>

<style scoped>
@reference "../style.css";

.filter-pill {
    @apply flex cursor-pointer items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] text-muted-foreground ring-1 ring-white/[0.07] transition-colors;
}

.filter-pill:hover {
    @apply text-foreground ring-white/20;
}

.filter-pill.active {
    @apply bg-primary/15 text-primary ring-primary/50;
}

.filter-pill.clear {
    @apply text-loss ring-loss/40;
}

.catalog-item {
    @apply flex cursor-grab flex-col items-center gap-0.5 rounded-lg p-1.5 transition-colors;
}

.catalog-item:hover {
    @apply bg-white/[0.05];
}

.block-card {
    @apply flex shrink-0 cursor-pointer flex-col gap-2.5 rounded-lg bg-white/[0.02] p-3 ring-1 ring-white/[0.04] ring-inset transition-all;
}

.block-card.active {
    @apply bg-primary/[0.06] ring-primary/50;
}

.block-card.over {
    @apply ring-2 ring-primary;
}

.item-chip {
    @apply relative cursor-grab rounded-md ring-1 ring-white/10 transition-all;
}

.item-chip.dragging {
    @apply opacity-40;
}

.item-chip-remove {
    @apply absolute -top-1.5 -right-1.5 hidden size-4 cursor-pointer items-center justify-center rounded-full bg-loss text-white;
}

.item-chip:hover .item-chip-remove {
    @apply flex;
}

.new-block {
    @apply flex shrink-0 cursor-pointer items-center justify-center gap-2 rounded-lg border-2 border-dashed border-white/10 px-4 py-5 text-xs text-muted-foreground/70 transition-colors;
}

.new-block:hover {
    @apply border-white/25 text-muted-foreground;
}

.new-block.over {
    @apply border-primary text-primary;
}
</style>
