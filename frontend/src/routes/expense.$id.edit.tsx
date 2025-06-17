import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@connectrpc/connect-query";
import { getExpense } from "@/gen/expense/v1/expense-ExpenseService_connectquery";
import { ExpenseForm } from "@/components/ExpenseForm";

export const Route = createFileRoute("/expense/$id/edit")({
    component: EditExpense,
});

function EditExpense() {
    const { id } = Route.useParams();
    const { data, isLoading, error } = useQuery(getExpense, { id });

    if (isLoading) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-lg text-gray-600">Loading expense...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-lg text-red-600">Error loading expense: {error.message}</div>
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