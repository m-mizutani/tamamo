import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Generate user initials from a name string
 * @param name - The name to generate initials from
 * @returns Uppercase initials (max 2 characters)
 */
export function getInitials(name: string): string {
  if (!name || name.trim() === "") {
    return "U";
  }

  const trimmedName = name.trim();
  const words = trimmedName.split(/\s+/);
  
  if (words.length >= 2) {
    // Take first character of first and second word
    return `${words[0][0]}${words[1][0]}`.toUpperCase();
  } else {
    // Take first two characters of the single word
    return trimmedName.substring(0, 2).toUpperCase();
  }
}