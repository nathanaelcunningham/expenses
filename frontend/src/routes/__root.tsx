import { Outlet, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";

import Header from "../components/Header";
import { AuthProvider } from "../contexts/AuthContext";
import { authInterceptor } from "../lib/authInterceptor";

const url = import.meta.env.VITE_API_URL;
const finalTransport = createConnectTransport({
    baseUrl: url,
    interceptors: [authInterceptor],
});

const queryClient = new QueryClient();

function RootComponent() {
    return (
        <TransportProvider transport={finalTransport}>
            <QueryClientProvider client={queryClient}>
                <AuthProvider>
                    <RootLayout />
                </AuthProvider>
            </QueryClientProvider>
        </TransportProvider>
    );
}

function RootLayout() {
    return (
        <>
            <Header />
            <Outlet />
            <TanStackRouterDevtools />
        </>
    );
}

export const Route = createRootRoute({
    component: RootComponent,
});
