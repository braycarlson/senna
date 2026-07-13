<script setup lang="ts">
import { CircleAlert, LoaderCircle, Snowflake } from '@lucide/vue'

defineProps<{ variant: 'empty' | 'error' | 'loading', message: string }>()
</script>

<template>
    <div
        class="rise-in flex flex-col items-center gap-3 rounded-xl border border-dashed px-8 py-14 text-center"
        :class="variant === 'error' ? 'border-loss/40 bg-loss/[0.04]' : 'border-border/70 bg-white/[0.01]'"
    >
        <LoaderCircle v-if="variant === 'loading'" class="size-5 animate-spin text-primary" />
        <CircleAlert v-else-if="variant === 'error'" class="size-5 text-loss" />
        <Snowflake v-else class="size-5 text-muted-foreground/60" />

        <p
            class="max-w-md text-[13px] leading-relaxed"
            :class="variant === 'error' ? 'text-loss' : 'text-muted-foreground'"
        >
            {{ message }}
        </p>
    </div>
</template>
