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
  { icon: '⚔️', label: 'STR', key: 'strength'  as const, color: '#BF5AF2' },
  { icon: '📖', label: 'INT', key: 'intellect'  as const, color: '#30D158' },
  { icon: '🍀', label: 'LCK', key: 'luck'       as const, color: '#FFD60A' },
]

export function Avatar({ profile }: AvatarProps) {
  const level    = profile.level ?? 1
  const xpNeeded = XP_PER_LEVEL(level)
  const xp       = profile.xp   ?? 0
  const hp       = profile.hp   ?? 100
  const mana     = profile.mana ?? 100

  return (
    <div className="glass-gold p-5 space-y-5">

      {/* ── Character banner ── */}
      <div className="flex items-center gap-4">
        {/* Avatar ring */}
        <div className="relative shrink-0">
          <div
            className="w-[60px] h-[60px] rounded-[18px] flex items-center justify-center text-3xl"
            style={{
              background: 'linear-gradient(135deg, rgba(245,197,24,0.2), rgba(10,132,255,0.15))',
              border: '1px solid rgba(245,197,24,0.25)',
              boxShadow: '0 0 24px rgba(245,197,24,0.12)',
            }}
          >
            ⚔️
          </div>
          {/* Level badge */}
          <div
            className="absolute -bottom-1.5 -right-1.5 w-6 h-6 rounded-full flex items-center justify-center text-[11px] font-black text-black"
            style={{ background: '#F5C518', boxShadow: '0 0 8px rgba(245,197,24,0.6)' }}
          >
            {level}
          </div>
        </div>

        {/* Identity */}
        <div className="min-w-0 flex-1">
          <p className="text-[17px] font-semibold tracking-tight leading-none" style={{ color: '#F5C518' }}>
            Adventurer
          </p>
          <p className="text-[13px] text-white/40 mt-0.5">Level {level} · Hero</p>
          <div className="mt-2 flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-[#30D158]" style={{ boxShadow: '0 0 6px #30D158' }} />
            <span className="text-[11px] text-white/30">Online</span>
          </div>
        </div>
      </div>

      {/* ── XP bar ── */}
      <div>
        <StatBar label="✨ Experience" value={xp} max={xpNeeded} color="#F5C518" glow thick />
        <p className="text-right text-[11px] text-white/25 mt-1">
          {xpNeeded - xp > 0 ? `${xpNeeded - xp} XP to next level` : 'Max level!'}
        </p>
      </div>

      {/* ── Vitals ── */}
      <div className="grid grid-cols-2 gap-3">
        <StatBar label="❤️ HP"   value={hp}   max={1000} color="#FF453A" glow />
        <StatBar label="💎 Mana" value={mana} max={500}  color="#0A84FF" glow />
      </div>

      {/* ── Divider ── */}
      <div className="divider" />

      {/* ── Attributes ── */}
      <div className="grid grid-cols-3 gap-2">
        {ATTRS.map(attr => (
          <div
            key={attr.label}
            className="flex flex-col items-center gap-1.5 rounded-2xl p-3 text-center transition-colors"
            style={{
              background: 'rgba(255,255,255,0.04)',
              border: '1px solid rgba(255,255,255,0.07)',
            }}
          >
            <div className="text-xl">{attr.icon}</div>
            <div className="text-[10px] uppercase tracking-widest" style={{ color: 'rgba(255,255,255,0.3)' }}>{attr.label}</div>
            <div className="text-lg font-bold leading-none" style={{ color: attr.color }}>
              {profile[attr.key] ?? 0}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
