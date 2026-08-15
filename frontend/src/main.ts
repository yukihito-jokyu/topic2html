import { createElement } from "react";
import { createRoot } from "react-dom/client";

import "./index.css";
import { AdminAuthScreen } from "@/features/admin-auth/components";
import { AdminAuthProvider } from "@/features/admin-auth/hooks/useAdminAuth";

const root = document.getElementById("root");
if (!root) throw new Error("management UI root is missing");

createRoot(root).render(
	createElement(AdminAuthProvider, null, createElement(AdminAuthScreen)),
);
