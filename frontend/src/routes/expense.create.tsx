import { createFileRoute } from "@tanstack/react-router";
import { ExpenseForm } from "@/components/ExpenseForm";
import { useAuth } from "@/contexts/AuthContext";

export const Route = createFileRoute("/expense/create")({
    component: CreateExpense,
});

function CreateExpense() {
    const { isAuthenticated, isLoading } = useAuth();

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (!isAuthenticated) {
        window.location.href = '/login';
        return null;
    }

    return <ExpenseForm mode="create" />;
}