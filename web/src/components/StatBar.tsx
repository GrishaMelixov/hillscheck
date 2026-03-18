interface StatBarProps {
  label: string
  value: number | undefined
  max: number
  color: string
  glow?: boolean
  thick?: boolean
}

export function StatBar({ label, value, max, color, glow, thick }: StatBarProps) {
  const v = value ?? 0
  const pct = Math.min(100, Math.round((v / Math.max(max, 1)) * 100))

  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs">
        <span className="text-gray-400 font-medium">{label}</span>
        <span className="text-gray-600">{v} / {max}</span>
      </div>
      <div className={`relative w-full rounded-full bg-gray-800/80 overflow-hidden ${thick ? 'h-3' : 'h-2'}`}>
        <div
          className="h-full rounded-full transition-all duration-700"
          style={{
            width: `${pct}%`,
            backgroundColor: color,
            boxShadow: glow && pct > 0 ? `0 0 10px ${color}60, 0 0 4px ${color}80` : 'none',
          }}
        />
      </div>
    </div>
  )
}
