<script setup lang="ts">
import { computed } from 'vue'

let props = defineProps<{ winRate: number }>()

let percent = computed(() => (props.winRate * 100).toFixed(1))
let positive = computed(() => props.winRate >= 0.5)

let width = computed(() => `${Math.min(Math.max(props.winRate, 0), 1) * 100}%`)
</script>

<template>
    <div class="flex min-w-[132px] items-center justify-end gap-2.5">
        <span
            class="stat-mono min-w-[52px] text-right text-[13px] font-medium"
            :class="positive ? 'text-win' : 'text-loss'"
        >
            {{ percent }}%
        </span>

        <span class="relative h-[3px] w-16 overflow-hidden rounded-full bg-white/[0.07]">
            <span
                class="absolute inset-y-0 left-0 rounded-full transition-[width] duration-300"
                :class="positive ? 'bg-gradient-to-r from-win/40 to-win' : 'bg-gradient-to-r from-loss/40 to-loss'"
                :style="{ width }"
            ></span>
            <span class="absolute inset-y-0 left-1/2 w-px bg-white/20"></span>
        </span>
    </div>
</template>
