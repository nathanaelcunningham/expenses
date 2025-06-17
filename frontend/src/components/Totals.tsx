interface TotalsProps {
    total: number;
    monthly_income: number;
}

export function Totals({ total, monthly_income }: TotalsProps) {
    return (
        <div className="space-y-4">
            <div className="bg-white rounded-lg shadow-sm border p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">
                    Monthly Summary
                </h2>
                <div className="space-y-3">
                    <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">
                            Total Monthly Expenses
                        </span>
                        <span className="text-2xl font-bold text-gray-900">
                            ${total}
                        </span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">
                            Total Monthly Income
                        </span>
                        <span className="text-2xl font-bold text-green-800">
                            ${monthly_income}
                        </span>
                    </div>
                    <div className="border-t pt-3">
                        <div className="flex justify-between items-center text-sm">
                            <span className="text-gray-600">Remaining</span>
                            <span className="font-medium text-gray-900">
                                ${monthly_income - total}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}