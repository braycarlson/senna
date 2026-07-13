export function percent(rate: number, digits = 1): string {
    return (rate * 100).toFixed(digits)
}

export function thousands(value: number): string {
    return (value / 1000).toFixed(1)
}
