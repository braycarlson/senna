<script setup lang="ts">
import { computed } from 'vue'

import { championSquare } from '../assets'

let props = defineProps<{
    championId: number
    name?: string | null
    size?: 'sm' | 'md' | 'lg'
}>()

const SIZE_CLASSES: Record<string, string> = {
    sm: 'size-7 rounded-md',
    md: 'size-9 rounded-lg',
    lg: 'size-20 rounded-xl',
}

let sizeClass = computed(() => SIZE_CLASSES[props.size ?? 'md'])
</script>

<template>
    <img
        class="bg-layer shrink-0 ring-1 ring-white/10"
        :class="sizeClass"
        :src="championSquare(championId)"
        :alt="name ?? String(championId)"
        loading="lazy"
    />
</template>
