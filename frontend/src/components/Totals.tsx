import { getMonthlyIncome } from "@/gen/family/v1/family-FamilySettingsService_connectquery";
import { useNumberFormat } from "@/hooks/useNumberFormat";
import { useQuery } from "@connectrpc/connect-query";

interface TotalsProps {
    total_expenses: number;
}

export function Totals({ total_expenses }: TotalsProps) {
    const { formatFloat } = useNumberFormat();
    const { data, isLoading, error } = useQuery(getMonthlyIncome);

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (error || data === undefined || data.monthlyIncome === undefined) {
        if (error) {
            return <div>Error loading monthly income: {error.message}</div>;
        } else return <div>Error loading monthly income</div>;
    }

    const remaining = formatFloat(
        data.monthlyIncome.totalAmount - total_expenses,
    );

    return (
        <div className="space-y-4">
            <div className="bg-white rounded-lg shadow-sm border p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">
                    Monthly Summary
                </h2>
                <div className="space-y-3">
                    <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">
                            Total Monthly Income
                        </span>
                        <span className="text-2xl font-bold text-green-800">
                            ${formatFloat(data.monthlyIncome.totalAmount)}
                        </span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">
                            Total Monthly Expenses
                        </span>
                        <span className="text-2xl font-bold text-gray-900">
                            ${formatFloat(total_expenses)}
                        </span>
                    </div>
                    <div className="border-t pt-3">
                        <div className="flex justify-between items-center text-sm">
                            <span className="text-gray-600">
                                Remaining Balance
                            </span>
                            <span
                                className={`font-medium ${remaining < 0 ? "text-red-600" : "text-green-800"}`}
                            >
                                ${remaining}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

