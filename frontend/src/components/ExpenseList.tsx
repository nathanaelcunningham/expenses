import type { SortedExpense } from "@/gen/expense/v1/expense_pb";
import { ExpenseCard } from "./ExpenseCard";
import { toOrdinal } from "@/hooks/useOrdinal";

interface ExpenseListProps {
    expenseList: SortedExpense[];
}

export function ExpenseList({ expenseList }: ExpenseListProps) {
    return (
        <div className="space-y-6">
            {expenseList.length === 0 && (
                <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
                    <div className="text-gray-500 mb-2">No expenses found</div>
                    <p className="text-sm text-gray-400">
                        Add your first expense to get started
                    </p>
                </div>
            )}
            {expenseList.map((sortedExpense) => (
                <div key={sortedExpense.day}>
                    <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
                        <div className="bg-gray-50 px-4 py-3 border-b">
                            <h3 className="text-sm font-medium text-gray-700">
                                Due on the {toOrdinal(sortedExpense.day)}
                            </h3>
                        </div>
                    </div>
                    {sortedExpense.expenses.map((expense) => (
                        <ExpenseCard key={expense.id} expense={expense} />
                    ))}
                </div>
            ))}
        </div>
    );
}

