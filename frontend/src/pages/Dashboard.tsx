export function Dashboard() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome to Tamamo
        </p>
      </div>
      
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="mx-auto w-32 h-32 bg-orange-100 rounded-full flex items-center justify-center">
            <img 
              src="/logo.png" 
              alt="Tamamo Logo" 
              className="w-20 h-20 object-contain"
            />
          </div>
          <div className="space-y-2">
            <h2 className="text-2xl font-semibold">Tamamo Dashboard</h2>
            <p className="text-muted-foreground max-w-md">
              Your dashboard is ready. Additional features and content will be added here in future updates.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}