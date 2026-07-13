<script setup lang="ts">
defineProps<{ eyebrow: string, title: string }>()
</script>

<template>
    <div class="rise-in mb-6 flex items-end justify-between gap-4">
        <div>
            <span class="font-display text-[11px] font-semibold tracking-[0.16em] text-primary uppercase">
                {{ eyebrow }}
            </span>
            <h1 class="mt-1 text-[26px] leading-tight font-bold">{{ title }}</h1>
        </div>

        <slot />
    </div>
</template>
