<script setup lang="ts">
import { computed } from 'vue'

import { percent } from '../format'

interface SparkPoint {
    label: string
    value: number
}

const WIDTH = 132
const HEIGHT = 36
const PAD = 5

let props = defineProps<{ points: SparkPoint[] }>()

let span = computed(() => {
    let values = props.points.map((point) => point.value)
    let low = Math.min(...values, 0.47)
    let high = Math.max(...values, 0.53)

    return { low, high }
})

function x(index: number): number {
    let count = props.points.length

    if (count === 1) return WIDTH / 2

    return PAD + (index * (WIDTH - PAD * 2)) / (count - 1)
}

function y(value: number): number {
    let { low, high } = span.value
    let unit = (value - low) / (high - low)

    return HEIGHT - PAD - unit * (HEIGHT - PAD * 2)
}

let line = computed(() =>
    props.points.map((point, index) => `${x(index)},${y(point.value)}`).join(' '),
)

let lastIndex = computed(() => props.points.length - 1)
</script>

<template>
    <svg
        v-if="points.length > 1"
        class="shrink-0"
        :width="WIDTH"
        :height="HEIGHT"
        :viewBox="`0 0 ${WIDTH} ${HEIGHT}`"
        role="img"
        aria-label="Win rate by patch"
    >
        <line
            class="stroke-white/10"
            :x1="PAD"
            :x2="WIDTH - PAD"
            :y1="y(0.5)"
            :y2="y(0.5)"
            stroke-dasharray="3 3"
        />

        <polyline
            class="stroke-primary"
            :points="line"
            fill="none"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
        />

        <circle
            class="fill-primary"
            :cx="x(lastIndex)"
            :cy="y(points[lastIndex].value)"
            r="3"
        />

        <g v-for="(point, index) in points" :key="point.label">
            <circle class="fill-transparent" :cx="x(index)" :cy="y(point.value)" r="7">
                <title>{{ point.label }} · {{ percent(point.value) }}%</title>
            </circle>
        </g>
    </svg>
</template>
