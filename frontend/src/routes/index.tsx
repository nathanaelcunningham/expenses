import { listExpenses } from "@/gen/expense/v1/expense-ExpenseService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { createFileRoute } from "@tanstack/react-router";
import { ExpenseList } from "@/components/ExpenseList";
import { Totals } from "@/components/Totals";
import { useAuth } from "@/contexts/AuthContext";

export const Route = createFileRoute("/")({
    component: App,
});

function App() {
    const { isAuthenticated, isLoading: authLoading } = useAuth();
    const { data, isLoading, error } = useQuery(listExpenses, {}, { enabled: isAuthenticated });

    if (authLoading) {
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

    if (isLoading) {
        return <div>Loading expenses...</div>;
    }

    if (error) {
        return <div>Error loading expenses: {error.message}</div>;
    }

    if (data === undefined) {
        return <div>Error loading expenses</div>;
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <div className="container mx-auto px-4 py-6 max-w-6xl">
                <div className="mb-8 flex justify-between items-start">
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900 mb-2">
                            Current Expenses
                        </h1>
                        <p className="text-gray-600">
                            Manage your monthly expenses and bills
                        </p>
                    </div>
                    <a
                        href="/expense/create"
                        className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 transition-colors duration-150"
                    >
                        <svg
                            className="w-4 h-4 mr-2"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <title>test</title>
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M12 4v16m8-8H4"
                            />
                        </svg>
                        Create New Expense
                    </a>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <div className="lg:col-span-2">
                        <ExpenseList expenseList={data.expenses} />
                    </div>
                    <div className="lg:col-span-1">
                        <Totals total={100} monthly_income={100.0} />
                    </div>
                </div>
            </div>
        </div>
    );
}

