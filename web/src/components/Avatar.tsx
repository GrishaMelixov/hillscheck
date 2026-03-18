import { StatBar } from './StatBar'

interface GameProfile {
  level: number
  xp: number
  hp: number
  mana: number
  strength: number
  intellect: number
  luck: number
}

interface AvatarProps {
  profile: GameProfile
}

const XP_PER_LEVEL = (level: number) => Math.max(level, 1) * Math.max(level, 1) * 100

const ATTRS = [
  { icon: '⚔️', label: 'STR', key: 'strength' as const, color: '#a78bfa' },
  { icon: '📖', label: 'INT', key: 'intellect' as const, color: '#34d399' },
  { icon: '🍀', label: 'LCK', key: 'luck'      as const, color: '#fbbf24' },
]

export function Avatar({ profile }: AvatarProps) {
  const level    = profile.level    ?? 1
  const xpNeeded = XP_PER_LEVEL(level)
  const xp       = profile.xp      ?? 0
  const hp       = profile.hp      ?? 100
  const mana     = profile.mana    ?? 100

  return (
    <div className="card-glow space-y-4">

      {/* ── Character banner ── */}
      <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-gray-800/80 via-gray-900 to-gray-950 p-4 border border-gray-700/40">
        {/* subtle purple glow top-left */}
        <div className="absolute -top-6 -left-6 w-24 h-24 rounded-full bg-rpg-purple/20 blur-2xl pointer-events-none" />
        {/* subtle gold glow bottom-right */}
        <div className="absolute -bottom-4 -right-4 w-20 h-20 rounded-full bg-rpg-gold/10 blur-2xl pointer-events-none" />

        <div className="relative flex items-center gap-4">
          {/* Avatar circle */}
          <div className="relative shrink-0">
            <div className="w-16 h-16 rounded-full p-[2px]"
              style={{ background: 'linear-gradient(135deg, #7c3aed, #2563eb, #f5c518)' }}>
              <div className="w-full h-full rounded-full bg-gray-900 flex items-center justify-center text-3xl">
                ⚔️
              </div>
            </div>
            {/* Level badge */}
            <div className="absolute -bottom-1 -right-1 bg-rpg-gold text-gray-950 text-xs font-black w-6 h-6 rounded-full flex items-center justify-center shadow-lg">
              {level}
            </div>
          </div>

          {/* Name / level */}
          <div>
            <p className="text-rpg-gold font-bold tracking-wide text-base leading-tight">Adventurer</p>
            <p className="text-gray-500 text-xs mt-0.5">Level {level}</p>
            <div className="mt-1.5 flex items-center gap-1.5">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-green-400 shadow-[0_0_6px_#4ade80]" />
              <span className="text-gray-600 text-xs">Online</span>
            </div>
          </div>
        </div>
      </div>

      {/* ── XP ── */}
      <div>
        <StatBar label="✨ Experience" value={xp} max={xpNeeded} color="#f5c518" glow thick />
        <p className="text-right text-xs text-gray-600 mt-1">
          {xpNeeded - xp > 0 ? `${xpNeeded - xp} XP до следующего уровня` : 'Max level!'}
        </p>
      </div>

      {/* ── Vitals ── */}
      <div className="grid grid-cols-2 gap-3">
        <StatBar label="❤️ HP" value={hp} max={1000} color="#ef4444" glow />
        <StatBar label="💎 Mana" value={mana} max={500} color="#3b82f6" glow />
      </div>

      {/* ── Attributes ── */}
      <div className="divider" />
      <div className="grid grid-cols-3 gap-2">
        {ATTRS.map(attr => (
          <div
            key={attr.label}
            className="bg-gray-800/60 border border-gray-700/40 rounded-xl p-2.5 text-center hover:border-gray-600/60 transition-colors"
          >
            <div className="text-xl mb-1">{attr.icon}</div>
            <div className="text-gray-500 text-[10px] uppercase tracking-widest">{attr.label}</div>
            <div className="text-lg font-bold mt-0.5" style={{ color: attr.color }}>
              {profile[attr.key] ?? 0}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
