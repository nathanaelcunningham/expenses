import type { Interceptor } from "@connectrpc/connect";

const SESSION_STORAGE_KEY = "auth_session_id";

/**
 * Authentication interceptor for Connect RPC calls
 * Adds session token to Authorization header for all requests
 */
export const authInterceptor: Interceptor = (next) => async (req) => {
    // Get session ID from localStorage
    let sessionId: string | null = null;
    try {
        sessionId = localStorage.getItem(SESSION_STORAGE_KEY);
    } catch (error) {
        // Handle localStorage not available (e.g., SSR)
        console.warn("localStorage not available:", error);
    }

    // Add Authorization header if session exists
    if (sessionId) {
        req.header.set("Authorization", `Bearer ${sessionId}`);
    }

    return await next(req);
};