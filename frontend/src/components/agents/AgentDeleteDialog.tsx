import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Loader2, AlertTriangle } from 'lucide-react'
import { Agent, DELETE_AGENT, graphqlRequest } from '@/lib/graphql'

interface AgentDeleteDialogProps {
  agent: Agent
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AgentDeleteDialog({ agent, open, onOpenChange }: AgentDeleteDialogProps) {
  const navigate = useNavigate()
  const [confirmText, setConfirmText] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const expectedText = agent.agentId
  const isConfirmValid = confirmText === expectedText

  const handleDelete = async () => {
    if (!isConfirmValid) return

    try {
      setDeleting(true)
      setError(null)
      
      await graphqlRequest<{ deleteAgent: boolean }>(DELETE_AGENT, {
        id: agent.id
      })
      
      // Navigate back to agents list
      navigate('/agents')
    } catch (err) {
      console.error('Failed to delete agent:', err)
      setError(err instanceof Error ? err.message : 'Failed to delete agent')
      setDeleting(false)
    }
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!deleting) {
      setConfirmText('')
      setError(null)
      onOpenChange(newOpen)
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent className="max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center space-x-2">
            <AlertTriangle className="h-5 w-5 text-destructive" />
            <span>Delete Agent</span>
          </AlertDialogTitle>
          <AlertDialogDescription className="space-y-4">
            <p>
              This action <strong>cannot be undone</strong>. This will permanently delete the agent{' '}
              <strong>{agent.name}</strong> and all of its versions.
            </p>
            
            <p>
              Please type <strong>{expectedText}</strong> to confirm deletion.
            </p>

            <div className="space-y-2">
              <Label htmlFor="confirm-text">Agent ID</Label>
              <Input
                id="confirm-text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder={expectedText}
                disabled={deleting}
                className={
                  confirmText && !isConfirmValid
                    ? 'border-destructive focus-visible:ring-destructive'
                    : ''
                }
              />
            </div>

            {error && (
              <div className="text-sm text-destructive bg-destructive/10 p-3 rounded-md border border-destructive/20">
                {error}
              </div>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleting}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={!isConfirmValid || deleting}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {deleting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting...
              </>
            ) : (
              'Delete Agent'
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}