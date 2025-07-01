import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@connectrpc/connect-query";
import { getExpense } from "@/gen/expense/v1/expense-ExpenseService_connectquery";
import { ExpenseForm } from "@/components/ExpenseForm";
import { useAuth } from "@/contexts/AuthContext";

export const Route = createFileRoute("/expense/$id/edit")({
    component: EditExpense,
});

function EditExpense() {
    const { isAuthenticated, isLoading: authLoading } = useAuth();
    const { id } = Route.useParams();
    const { data, isLoading, error } = useQuery(
        getExpense,
        { id: BigInt(id) },
        { enabled: isAuthenticated },
    );

    if (authLoading || isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (!isAuthenticated) {
        window.location.href = "/login";
        return null;
    }

    if (error) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-lg text-red-600">
                    Error loading expense: {error.message}
                </div>
            </div>
        );
    }

    if (!data?.expense) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-lg text-gray-600">Expense not found</div>
            </div>
        );
    }

    return <ExpenseForm mode="edit" initialData={data.expense} />;
}
