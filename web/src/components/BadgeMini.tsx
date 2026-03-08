import type { UserBadge } from '../api/badges'

interface BadgeMiniProps {
  badges: UserBadge[]
  maxShow?: number
}

export default function BadgeMini({ badges, maxShow = 5 }: BadgeMiniProps) {
  if (!badges || badges.length === 0) return null

  const visible = badges.slice(0, maxShow)
  const remaining = badges.length - visible.length

  return (
    <div className="flex items-center gap-1">
      {visible.map((badge) => (
        <span
          key={badge.id}
          title={badge.name}
          className="text-lg leading-none"
          role="img"
          aria-label={badge.name}
        >
          {badge.icon_emoji}
        </span>
      ))}
      {remaining > 0 && (
        <span className="text-xs font-medium text-gray-400 ml-0.5">+{remaining}</span>
      )}
    </div>
  )
}
