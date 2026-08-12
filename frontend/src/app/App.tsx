import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Toaster } from "@/components/ui/sonner";

export function App() {
  return (
    <main className="mx-auto max-w-3xl p-6">
      <Card>
        <CardHeader>
          <CardTitle role="heading" aria-level={1}>
            topic2html
          </CardTitle>
          <CardDescription>管理画面</CardDescription>
        </CardHeader>
        <CardContent className="text-muted-foreground">
          管理画面の最小起動確認用の表示です。
        </CardContent>
      </Card>
      <Toaster />
    </main>
  );
}
