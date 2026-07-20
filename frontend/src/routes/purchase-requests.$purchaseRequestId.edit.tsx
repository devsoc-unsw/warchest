import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute(
  '/purchase-requests/$purchaseRequestId/edit',
)({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/purchase-requests/$purchaseRequestId/edit"!</div>
}
