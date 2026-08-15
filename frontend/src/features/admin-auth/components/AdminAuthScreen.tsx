import {
	AlertCircle,
	CircleCheck,
	LoaderCircle,
	LogIn,
	LogOut,
} from "lucide-react";
import { useCallback } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useAdminAuth } from "@/features/admin-auth/hooks/useAdminAuth";
import type { AdminAuthState } from "@/features/admin-auth/types";

type AdminAuthViewProps = {
	state: AdminAuthState;
	onRetryBootstrap(): void;
	onRetryLogout(): void;
	onLogout(): void;
	onLoginStart(): void;
};

export function AdminAuthView({
	state,
	onRetryBootstrap,
	onRetryLogout,
	onLogout,
	onLoginStart,
}: AdminAuthViewProps) {
	return (
		<main className="mx-auto flex min-h-screen w-full max-w-2xl items-center p-4 sm:p-8">
			<Card className="w-full">
				<CardHeader>
					<p className="text-primary text-sm font-semibold tracking-[0.16em] uppercase">
						Topic2html
					</p>
					<CardTitle className="text-2xl">管理画面</CardTitle>
					<CardDescription>
						管理機能を利用するには本人確認が必要です。
					</CardDescription>
				</CardHeader>
				<CardContent aria-atomic="true" aria-live="polite">
					{state.kind === "loading" && (
						<div
							className="flex items-center gap-3 text-muted-foreground"
							role="status"
						>
							<LoaderCircle
								aria-hidden="true"
								className="size-5 animate-spin"
							/>
							<span>認証状態を確認しています。</span>
						</div>
					)}
					{state.kind === "login" && (
						<div className="space-y-5">
							{state.failed && (
								<Alert variant="destructive">
									<AlertCircle aria-hidden="true" />
									<AlertTitle>本人確認に失敗しました</AlertTitle>
									<AlertDescription>
										管理操作は利用できません。もう一度お試しください。
									</AlertDescription>
								</Alert>
							)}
							<p className="text-muted-foreground leading-6">
								Google アカウントで本人確認を行います。
							</p>
							<form
								action="/admin/auth/google/start"
								method="post"
								onSubmit={onLoginStart}
							>
								<input name="return_path" type="hidden" value="/admin" />
								<Button className="w-full sm:w-auto" type="submit">
									<LogIn aria-hidden="true" />
									Googleでログイン
								</Button>
							</form>
						</div>
					)}
					{state.kind === "authenticated" && (
						<div className="space-y-4">
							<Alert variant="success">
								<CircleCheck aria-hidden="true" />
								<AlertTitle>認証済みです</AlertTitle>
								<AlertDescription>
									管理機能を利用できます。機能は順次追加されます。
								</AlertDescription>
							</Alert>
							{state.logoutProblem === "forbidden" && (
								<Alert variant="destructive">
									<AlertCircle aria-hidden="true" />
									<AlertTitle>ログアウトを完了できません</AlertTitle>
									<AlertDescription>
										認証状態を更新してから、もう一度お試しください。
									</AlertDescription>
								</Alert>
							)}
						</div>
					)}
					{state.kind === "unavailable" && (
						<Alert variant="destructive">
							<AlertCircle aria-hidden="true" />
							<AlertTitle>認証状態を確認・更新できません</AlertTitle>
							<AlertDescription>
								管理操作を続行できません。時間をおいて再試行してください。
							</AlertDescription>
						</Alert>
					)}
				</CardContent>
				<CardFooter className="flex flex-col items-stretch gap-3 sm:flex-row sm:justify-end">
					{state.kind === "unavailable" && state.retry === "bootstrap" && (
						<Button
							className="w-full sm:w-auto"
							onClick={onRetryBootstrap}
							type="button"
						>
							再試行
						</Button>
					)}
					{state.kind === "unavailable" && state.retry === "logout" && (
						<Button
							className="w-full sm:w-auto"
							onClick={onRetryLogout}
							type="button"
						>
							ログアウトを再試行
						</Button>
					)}
					{state.kind === "authenticated" &&
						state.logoutProblem === "forbidden" && (
							<Button
								className="w-full sm:w-auto"
								onClick={onRetryBootstrap}
								type="button"
							>
								認証状態を更新
							</Button>
						)}
					{state.kind === "authenticated" && (
						<Button
							className="w-full sm:w-auto"
							onClick={onLogout}
							type="button"
							variant="outline"
						>
							<LogOut aria-hidden="true" />
							ログアウト
						</Button>
					)}
				</CardFooter>
			</Card>
		</main>
	);
}

export function AdminAuthScreen() {
	const { bootstrap, logout, state } = useAdminAuth();
	const loginStart = useCallback(() => {
		const button = document.activeElement;
		if (button instanceof HTMLButtonElement) button.disabled = true;
	}, []);

	return (
		<AdminAuthView
			onLoginStart={loginStart}
			onLogout={() => void logout()}
			onRetryBootstrap={() => void bootstrap()}
			onRetryLogout={() => void logout()}
			state={state}
		/>
	);
}
