import { UserAvatar } from './UserAvatar';
import { User } from '@/lib/graphql';
import { cn } from '@/lib/utils';

interface UserDisplayProps {
  user: User;
  size?: number;
  showEmail?: boolean;
  layout?: 'horizontal' | 'vertical';
  className?: string;
}

export function UserDisplay({ 
  user, 
  size = 32, 
  showEmail = false, 
  layout = 'horizontal',
  className 
}: UserDisplayProps) {
  const isHorizontal = layout === 'horizontal';

  return (
    <div 
      className={cn(
        'flex items-center gap-2',
        isHorizontal ? 'flex-row' : 'flex-col text-center',
        className
      )}
    >
      <UserAvatar user={user} size={size} />
      <div className={cn('flex flex-col', isHorizontal ? 'text-left' : 'text-center')}>
        <span className="text-sm font-medium text-gray-900">
          {user.displayName || user.slackName}
        </span>
        {showEmail && user.email && (
          <span className="text-xs text-gray-500">
            {user.email}
          </span>
        )}
      </div>
    </div>
  );
}

// Compact variant for list items
interface UserDisplayCompactProps {
  user: User;
  size?: number;
  className?: string;
}

export function UserDisplayCompact({ user, size = 24, className }: UserDisplayCompactProps) {
  return (
    <div className={cn('flex items-center gap-1.5', className)}>
      <UserAvatar user={user} size={size} />
      <span className="text-sm text-gray-700 truncate">
        {user.displayName || user.slackName}
      </span>
    </div>
  );
}