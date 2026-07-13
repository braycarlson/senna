<script setup lang="ts">
import { computed } from 'vue'

let props = defineProps<{ tier: string }>()

const TIER_STYLES: Record<string, string> = {
    'S+': 'bg-gold/15 text-gold ring-gold/40',
    'S': 'bg-primary/15 text-primary ring-primary/40',
    'A': 'bg-win/10 text-win ring-win/35',
    'B': 'bg-white/[0.06] text-foreground/80 ring-white/10',
    'C': 'bg-white/[0.04] text-muted-foreground ring-white/10',
    'D': 'bg-white/[0.03] text-muted-foreground/70 ring-white/5',
}

let classes = computed(() => TIER_STYLES[props.tier] ?? TIER_STYLES['D'])
</script>

<template>
    <span
        class="inline-flex min-w-9 items-center justify-center rounded-md px-1.5 py-0.5 font-display text-xs font-bold ring-1 ring-inset"
        :class="classes"
    >
        {{ tier }}
    </span>
</template>
