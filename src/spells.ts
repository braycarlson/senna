export const ARAM_SPELLS = [
    { id: 32, name: 'Mark' },
    { id: 4, name: 'Flash' },
    { id: 6, name: 'Ghost' },
    { id: 7, name: 'Heal' },
    { id: 3, name: 'Exhaust' },
    { id: 1, name: 'Cleanse' },
    { id: 14, name: 'Ignite' },
    { id: 21, name: 'Barrier' },
    { id: 13, name: 'Clarity' },
    { id: 12, name: 'Teleport' },
    { id: 11, name: 'Smite' },
]

export const SPELL_NAMES: Record<number, string> = Object.fromEntries(
    ARAM_SPELLS.map((spell) => [spell.id, spell.name]),
)

export const MODES = [
    { key: 'ARAM', label: 'ARAM' },
    { key: 'CLASSIC', label: "Summoner's Rift" },
    { key: 'URF', label: 'URF' },
    { key: 'ONEFORALL', label: 'One for All' },
]
