import { Outlet, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";

import Header from "../components/Header";
const url = import.meta.env.VITE_API_URL;
const finalTransport = createConnectTransport({
    baseUrl: url,
});

const queryClient = new QueryClient();

export const Route = createRootRoute({
    component: () => (
        <TransportProvider transport={finalTransport}>
            <QueryClientProvider client={queryClient}>
                <Header />

                <Outlet />
                <TanStackRouterDevtools />
            </QueryClientProvider>
        </TransportProvider>
    ),
});
