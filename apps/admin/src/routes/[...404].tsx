import { A } from "@solidjs/router";
import { Button } from "@qp/ui";

export default function NotFound() {
  return (
    <main class="flex min-h-screen flex-col items-center justify-center gap-4 bg-background p-4 text-foreground">
      <p class="text-6xl font-bold text-primary">404</p>
      <p class="text-muted-foreground">This page doesn't exist.</p>
      <Button as={A} href="/">
        Back to dashboard
      </Button>
    </main>
  );
}
