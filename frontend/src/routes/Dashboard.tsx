import { createFileRoute } from '@tanstack/react-router'
import { DashboardComponent } from '../components/dashboard/Dashboard'

export const Route = createFileRoute('/Dashboard')({
  component: DashboardComponent,
})
