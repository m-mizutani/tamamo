import { useState } from 'react'
import { Users } from 'lucide-react'

interface AgentAvatarProps {
  imageUrl?: string
  name: string
  size?: number
  className?: string
  onClick?: () => void
}

export function AgentAvatar({ imageUrl, name, size = 32, className = '', onClick }: AgentAvatarProps) {
  const [imageLoadError, setImageLoadError] = useState(false)
  const baseClasses = `rounded-full flex items-center justify-center ${onClick ? 'cursor-pointer' : ''} ${className}`
  
  // If we have an image URL and it hasn't failed to load, show the image
  if (imageUrl && !imageLoadError) {
    return (
      <img
        src={imageUrl}
        alt={`${name} avatar`}
        className={`${baseClasses} object-cover`}
        style={{ width: size, height: size }}
        onClick={onClick}
        onError={() => setImageLoadError(true)}
      />
    )
  }

  // Default fallback with Users icon
  return (
    <div 
      className={`${baseClasses} bg-blue-100`}
      style={{ width: size, height: size }}
      onClick={onClick}
    >
      <Users className="text-blue-600" style={{ width: size * 0.5, height: size * 0.5 }} />
    </div>
  )
}