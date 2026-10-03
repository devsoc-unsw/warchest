import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute(
    '/purchase-requests/$purchaseRequestId/edit',
)({
    component: EditPurchaseRequestPage,
})

// Edit an existing draft purchase request
function EditPurchaseRequestPage() {
    const { purchaseRequestId } = Route.useParams()

    return (
        <main>
            <h1>Edit Purchase Request</h1>
            <p>Purchase Request ID: {purchaseRequestId}</p>
        </main>
    )
}