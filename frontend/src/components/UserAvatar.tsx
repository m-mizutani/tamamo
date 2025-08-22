import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { User } from '@/lib/graphql';
import { getInitials } from '@/lib/utils';

interface UserAvatarProps {
  user: User;
  size?: number;
  className?: string;
}

export function UserAvatar({ user, size = 48, className }: UserAvatarProps) {
  // Build avatar URL with size parameter
  const avatarUrl = `/api/users/${user.id}/avatar?size=${size}`;

  return (
    <Avatar className={className} style={{ width: size, height: size }}>
      <AvatarImage 
        src={avatarUrl} 
        alt={`${user.displayName || user.slackName}'s avatar`}
        onError={(e) => {
          // Hide broken image and show fallback
          e.currentTarget.style.display = 'none';
        }}
      />
      <AvatarFallback>
        {getInitials(user.displayName || user.slackName)}
      </AvatarFallback>
    </Avatar>
  );
}

// Simple variant that just takes user ID
interface UserAvatarByIdProps {
  userId: string;
  size?: number;
  className?: string;
  fallbackText?: string;
}

export function UserAvatarById({ userId, size = 48, className, fallbackText = '?' }: UserAvatarByIdProps) {
  const avatarUrl = `/api/users/${userId}/avatar?size=${size}`;

  return (
    <Avatar className={className} style={{ width: size, height: size }}>
      <AvatarImage 
        src={avatarUrl} 
        alt="User avatar"
        onError={(e) => {
          e.currentTarget.style.display = 'none';
        }}
      />
      <AvatarFallback>
        {fallbackText}
      </AvatarFallback>
    </Avatar>
  );
}