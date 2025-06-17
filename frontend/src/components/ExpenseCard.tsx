import type { Expense } from "@/gen/expense/v1/expense_pb";

interface ExpenseCardProps {
    expense: Expense;
}

export function ExpenseCard(props: ExpenseCardProps) {
    const { expense } = props;

    return (
        <div className="divide-y divide-gray-100">
            <a
                href={`/expense/${expense.id}/edit`}
                className={
                    "block p-4 hover:bg-gray-50 transition-colors duration-150"
                }
            >
                <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-3">
                            <div className="flex-1">
                                <p className="text-sm font-medium text-gray-900 truncate">
                                    {expense.name}
                                </p>
                                {expense.isAutopay && (
                                    <p className="text-xs text-blue-600 mt-1 flex items-center">
                                        <svg
                                            className="w-3 h-3 mr-1"
                                            fill="currentColor"
                                            viewBox="0 0 20 20"
                                        >
                                            <title>test</title>
                                            <path
                                                fill-rule="evenodd"
                                                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                                clip-rule="evenodd"
                                            />
                                        </svg>
                                        Autopay
                                    </p>
                                )}
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center space-x-2">
                        <span className="text-lg font-semibold text-gray-900">
                            {expense.amount}
                        </span>
                        <svg
                            className="w-5 h-5 text-gray-400"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <title>test</title>
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M9 5l7 7-7 7"
                            />
                        </svg>
                    </div>
                </div>
            </a>
        </div>
    );
}