import React from 'react'
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

const XP_PER_LEVEL = (level: number) => level * level * 100

export function Avatar({ profile }: AvatarProps) {
  const xpNeeded = XP_PER_LEVEL(profile.level)

  return (
    <div className="card space-y-4">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="w-14 h-14 rounded-full bg-gradient-to-br from-rpg-purple to-rpg-blue flex items-center justify-center text-2xl select-none">
          ⚔️
        </div>
        <div>
          <p className="text-xs text-gray-500 uppercase tracking-wider">Level</p>
          <p className="text-3xl font-bold text-rpg-gold">{profile.level}</p>
        </div>
      </div>

      {/* XP bar */}
      <StatBar label="XP" value={profile.xp} max={xpNeeded} color="#f5c518" />

      {/* Vitals */}
      <div className="grid grid-cols-2 gap-2">
        <StatBar label="HP" value={profile.hp} max={1000} color="#dc2626" />
        <StatBar label="Mana" value={profile.mana} max={500} color="#2563eb" />
      </div>

      {/* Attributes */}
      <div className="border-t border-gray-800 pt-3 grid grid-cols-3 gap-2 text-center">
        {[
          { label: '⚔️ STR', value: profile.strength },
          { label: '📖 INT', value: profile.intellect },
          { label: '🍀 LCK', value: profile.luck },
        ].map((attr) => (
          <div key={attr.label} className="bg-gray-800 rounded-lg py-2">
            <p className="text-xs text-gray-400">{attr.label}</p>
            <p className="text-xl font-bold">{attr.value}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
