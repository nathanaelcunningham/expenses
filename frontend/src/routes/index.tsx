import { listExpenses } from "@/gen/expense/v1/expense-ExpenseService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
    component: App,
});

function App() {
    const { data, isLoading, error } = useQuery(listExpenses);

    if (isLoading) {
        return <div>Loading expenses...</div>;
    }

    if (error) {
        return <div>Error loading expenses: {error.message}</div>;
    }

    if (!data?.expenses?.length) {
        return <div>No expenses found</div>;
    }

    return (
        <div className="text-center">
            {data.expenses.map((expense) => (
                <div key={expense.id}>{expense.name}</div>
            ))}
        </div>
    );
}
