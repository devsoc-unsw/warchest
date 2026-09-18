import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/purchase-requests/new')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/purchase-requests/new"!</div>
}
