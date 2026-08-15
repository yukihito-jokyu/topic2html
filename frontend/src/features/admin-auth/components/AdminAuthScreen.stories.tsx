import type { Meta, StoryObj } from "@storybook/react-vite";
import { useEffect } from "react";
import {
	AdminAuthScreen,
	AdminAuthView,
} from "@/features/admin-auth/components/AdminAuthScreen";
import {
	AdminAuthProvider,
	useAdminAuth,
} from "@/features/admin-auth/hooks/useAdminAuth";

const noop = () => {};

const meta = {
	title: "Features/AdminAuth/Screen",
	component: AdminAuthView,
	args: {
		onLoginStart: noop,
		onLogout: noop,
		onRetryBootstrap: noop,
		onRetryLogout: noop,
	},
} satisfies Meta<typeof AdminAuthView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Loading: Story = { args: { state: { kind: "loading" } } };
export const Login: Story = {
	args: { state: { kind: "login", failed: false } },
};
export const FailedLogin: Story = {
	args: { state: { kind: "login", failed: true } },
};
export const Authenticated: Story = {
	args: { state: { kind: "authenticated" } },
};
export const Unavailable: Story = {
	args: { state: { kind: "unavailable", retry: "bootstrap" } },
	parameters: { viewport: { defaultViewport: "mobile1" } },
};
export const LogoutUnavailable: Story = {
	args: { state: { kind: "unavailable", retry: "logout" } },
};

export const LogoutRequestFailed: Story = {
	render: () => (
		<AdminAuthProvider
			api={{
				bootstrap: async () => ({
					authenticated: true,
					csrfToken: crypto.randomUUID(),
				}),
				logout: async () => Promise.reject(new Error("unavailable")),
			}}
		>
			<AdminAuthScreen />
		</AdminAuthProvider>
	),
};

function ProtectedReadProbe() {
	const { fetchProtected } = useAdminAuth();
	useEffect(() => {
		void fetchProtected("/admin/future-feature/read");
	}, [fetchProtected]);
	return <p>保護された読取りを初期化しています。</p>;
}

export const ProtectedReadUnavailable: Story = {
	render: () => (
		<AdminAuthProvider
			api={{
				bootstrap: async () => ({
					authenticated: true,
					csrfToken: crypto.randomUUID(),
				}),
				logout: async () => new Response(null, { status: 200 }),
			}}
			fetcher={async () => new Response(null, { status: 503 })}
		>
			<AdminAuthScreen />
			<ProtectedReadProbe />
		</AdminAuthProvider>
	),
};
