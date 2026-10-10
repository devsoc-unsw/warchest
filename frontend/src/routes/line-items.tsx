import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/line-items')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/line-items"!</div>
}
