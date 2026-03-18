
interface StatBarProps {
  label: string
  value: number
  max: number
  color: string
}

export function StatBar({ label, value, max, color }: StatBarProps) {
  const pct = Math.min(100, Math.round((value / max) * 100))
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs text-gray-400">
        <span>{label}</span>
        <span>{value} / {max}</span>
      </div>
      <div className="stat-bar">
        <div
          className="stat-bar-fill"
          style={{ width: `${pct}%`, backgroundColor: color }}
        />
      </div>
    </div>
  )
}
