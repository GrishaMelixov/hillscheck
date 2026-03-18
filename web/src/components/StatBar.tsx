interface StatBarProps {
  label: string
  value: number | undefined
  max: number
  color: string
  glow?: boolean
  thick?: boolean
}

export function StatBar({ label, value, max, color, glow, thick }: StatBarProps) {
  const v   = value ?? 0
  const pct = Math.min(100, Math.round((v / Math.max(max, 1)) * 100))
  const h   = thick ? '6px' : '4px'

  return (
    <div className="space-y-1.5">
      <div className="flex justify-between items-baseline">
        <span className="text-xs text-white/50 font-medium tracking-wide">{label}</span>
        <span className="text-xs font-mono text-white/30">{v}<span className="text-white/15">/{max}</span></span>
      </div>
      <div className="relative w-full rounded-full overflow-hidden" style={{ height: h, background: 'rgba(255,255,255,0.07)' }}>
        <div
          className="h-full rounded-full transition-all duration-700"
          style={{
            width: `${pct}%`,
            backgroundColor: color,
            boxShadow: glow && pct > 0 ? `0 0 12px ${color}70` : 'none',
          }}
        />
      </div>
    </div>
  )
}
