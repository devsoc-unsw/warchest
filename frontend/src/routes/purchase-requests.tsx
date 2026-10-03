import { createFileRoute } from "@tanstack/react-router";



//inspect users' all draft purchase requests
export const Route = createFileRoute('/purchase-requests')({
    component: PurchaseRequestsPage,
})
function PurchaseRequestsPage() {
    return (
        <main>
            <h1>Purchase Requests</h1>
            <p>Your draft purchase requests will appear here.</p>
        </main>
    )
}





