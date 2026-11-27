import { createFileRoute } from "@tanstack/react-router";
import LandingPage from "@/components/landing-page/LandingPage"

export const Route = createFileRoute("/")({ component: App });

function App() {
return (
  <LandingPage />
);
}