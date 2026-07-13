<script setup lang="ts">
import { ref, watch } from 'vue'

import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { spellIcon } from '../assets'
import { ARAM_SPELLS } from '../spells'

const AUTO = 'auto'

let props = defineProps<{ idPrefix: string }>()
let model = defineModel<[number, number] | null>({ required: true })

let spellD = ref(AUTO)
let spellF = ref(AUTO)

watch(
    model,
    (pair) => {
        spellD.value = pair ? String(pair[0]) : AUTO
        spellF.value = pair ? String(pair[1]) : AUTO
    },
    { immediate: true },
)

function pickSlot(slot: 'd' | 'f', value: unknown): void {
    let picked = value == null ? AUTO : String(value)

    if (slot === 'd') {
        spellD.value = picked
    } else {
        spellF.value = picked
    }

    let pair: [number, number] | null =
        spellD.value !== AUTO && spellF.value !== AUTO && spellD.value !== spellF.value
            ? [Number(spellD.value), Number(spellF.value)]
            : null

    if (JSON.stringify(pair) === JSON.stringify(model.value ?? null)) return

    model.value = pair
}
</script>

<template>
    <div class="grid max-w-md grid-cols-2 gap-3">
        <div class="flex flex-col gap-1.5">
            <Label :for="`${props.idPrefix}-spell-d`">Slot D</Label>

            <Select :model-value="spellD" @update:model-value="pickSlot('d', $event)">
                <SelectTrigger :id="`${props.idPrefix}-spell-d`" class="w-full bg-white/[0.02]">
                    <SelectValue />
                </SelectTrigger>

                <SelectContent>
                    <SelectItem :value="AUTO">Most played</SelectItem>
                    <SelectItem
                        v-for="spell in ARAM_SPELLS"
                        :key="spell.id"
                        :value="String(spell.id)"
                        :disabled="String(spell.id) === spellF"
                    >
                        <span class="flex items-center gap-2">
                            <img class="size-5 rounded-sm bg-layer" :src="spellIcon(spell.id)" alt="" />
                            {{ spell.name }}
                        </span>
                    </SelectItem>
                </SelectContent>
            </Select>
        </div>

        <div class="flex flex-col gap-1.5">
            <Label :for="`${props.idPrefix}-spell-f`">Slot F</Label>

            <Select :model-value="spellF" @update:model-value="pickSlot('f', $event)">
                <SelectTrigger :id="`${props.idPrefix}-spell-f`" class="w-full bg-white/[0.02]">
                    <SelectValue />
                </SelectTrigger>

                <SelectContent>
                    <SelectItem :value="AUTO">Most played</SelectItem>
                    <SelectItem
                        v-for="spell in ARAM_SPELLS"
                        :key="spell.id"
                        :value="String(spell.id)"
                        :disabled="String(spell.id) === spellD"
                    >
                        <span class="flex items-center gap-2">
                            <img class="size-5 rounded-sm bg-layer" :src="spellIcon(spell.id)" alt="" />
                            {{ spell.name }}
                        </span>
                    </SelectItem>
                </SelectContent>
            </Select>
        </div>
    </div>
</template>
