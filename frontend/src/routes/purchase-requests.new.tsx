import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/purchase-requests/new')({
    component: RouteComponent,
})

function RouteComponent() {
    return (
        <main>
            <h1>Create Purchase Request</h1>
        </main>
    )
}

