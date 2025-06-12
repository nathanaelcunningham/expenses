import { Outlet, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";

import Header from "../components/Header";

const finalTransport = createConnectTransport({
    baseUrl: "http://localhost:8080",
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
