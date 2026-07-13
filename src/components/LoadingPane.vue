<script setup lang="ts">
import { LoaderCircle } from '@lucide/vue'

defineProps<{ message?: string }>()
</script>

<template>
    <div class="flex items-center justify-center">
        <span class="flex flex-col items-center gap-3">
            <LoaderCircle class="size-6 animate-spin text-primary" />
            <span class="text-xs text-muted-foreground">{{ message ?? 'Loading…' }}</span>
        </span>
    </div>
</template>
