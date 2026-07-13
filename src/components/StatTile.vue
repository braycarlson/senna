<script setup lang="ts">
defineProps<{ label: string, value: string }>()
</script>

<template>
    <div class="flex flex-col gap-1 rounded-lg bg-white/[0.03] px-4 py-3 ring-1 ring-white/[0.04] ring-inset">
        <span class="stat-mono text-xl font-semibold text-primary">{{ value }}</span>
        <span class="text-xs text-muted-foreground">{{ label }}</span>
    </div>
</template>
