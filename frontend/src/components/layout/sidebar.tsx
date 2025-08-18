import { NavLink, useLocation } from "react-router-dom";
import { Home, Settings } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

const navigation = [
  {
    name: "Dashboard",
    href: "/",
    icon: Home,
  },
  {
    name: "Settings",
    href: "/settings",
    icon: Settings,
  },
];

export function Sidebar() {
  const location = useLocation();

  return (
    <div className="flex h-full w-52 flex-col fixed inset-y-0 z-50 bg-background border-r">
      <div className="flex h-14 items-center border-b px-4">
        <NavLink className="flex items-center space-x-2" to="/">
          <img 
            src="/logo.png" 
            alt="Tamamo Logo" 
            className="h-8 w-8 object-contain"
          />
          <span className="font-bold">Tamamo</span>
        </NavLink>
      </div>
      <nav className="flex-1 space-y-1 p-4">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href;
          return (
            <NavLink key={item.name} to={item.href}>
              <Button
                variant={isActive ? "secondary" : "ghost"}
                className={cn(
                  "w-full justify-start",
                  isActive && "bg-secondary text-secondary-foreground"
                )}>
                <item.icon className="mr-2 h-4 w-4" />
                {item.name}
              </Button>
            </NavLink>
          );
        })}
      </nav>
    </div>
  );
}