import { useMutation } from "@connectrpc/connect-query";
import { useNavigate } from "@tanstack/react-router";
import {
    createExpense,
    updateExpense,
} from "@/gen/expense/v1/expense-ExpenseService_connectquery";
import type { Expense } from "@/gen/expense/v1/expense_pb";
import { useAppForm } from "@/hooks/form";

interface ExpenseFormProps {
    mode: "create" | "edit";
    initialData?: Expense;
}

export function ExpenseForm({ mode, initialData }: ExpenseFormProps) {
    const navigate = useNavigate();
    const createExpenseMutation = useMutation(createExpense);
    const updateExpenseMutation = useMutation(updateExpense);

    const form = useAppForm({
        defaultValues: {
            name: initialData?.name || "",
            amount: initialData?.amount || 0,
            dayOfMonthDue: initialData?.dayOfMonthDue || 1,
            isAutopay: initialData?.isAutopay || false,
        },
        validators: {
            onBlur: ({ value }) => {
                const errors = {
                    fields: {},
                } as {
                    fields: Record<string, string>;
                };

                if (value.name.trim().length === 0) {
                    errors.fields.name = "Name is required";
                }

                if (value.amount <= 0) {
                    errors.fields.amount = "Amount must be greater than 0";
                }

                if (value.dayOfMonthDue < 1 || value.dayOfMonthDue > 31) {
                    errors.fields.dayOfMonthDue =
                        "Day must be between 1 and 31";
                }

                return errors;
            },
        },
        onSubmit: async ({ value }) => {
            try {
                if (mode === "create") {
                    await createExpenseMutation.mutateAsync({
                        name: value.name,
                        amount: value.amount,
                        dayOfMonthDue: value.dayOfMonthDue,
                        isAutopay: value.isAutopay,
                    });
                } else {
                    await updateExpenseMutation.mutateAsync({
                        id: initialData!.id,
                        name: value.name,
                        amount: value.amount,
                        dayOfMonthDue: value.dayOfMonthDue,
                        isAutopay: value.isAutopay,
                    });
                }
                navigate({ to: "/" });
            } catch (error) {
                console.error("Error saving expense:", error);
            }
        },
    });

    const isSubmitting =
        createExpenseMutation.isPending || updateExpenseMutation.isPending;

    return (
        <div className="min-h-screen bg-gray-50">
            <div className="container mx-auto px-4 py-6 max-w-2xl">
                <div className="bg-white rounded-lg shadow-md p-6">
                    <h1 className="text-2xl font-bold text-gray-900 mb-6">
                        {mode === "create"
                            ? "Create New Expense"
                            : "Edit Expense"}
                    </h1>

                    <form
                        onSubmit={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            form.handleSubmit();
                        }}
                        className="space-y-6"
                    >
                        <form.AppField
                            name="name"
                            validators={{
                                onBlur: ({ value }) => {
                                    if (!value || value.trim().length === 0) {
                                        return "Name is required";
                                    }
                                    return undefined;
                                },
                            }}
                        >
                            {(field) => (
                                <field.TextField
                                    label="Expense Name"
                                    placeholder="e.g., Netflix, Rent, Groceries"
                                />
                            )}
                        </form.AppField>

                        <form.AppField
                            name="amount"
                            validators={{
                                onBlur: ({ value }) => {
                                    if (value <= 0) {
                                        return "Amount must be greater than 0";
                                    }
                                    return undefined;
                                },
                            }}
                        >
                            {(field) => (
                                <field.NumberField
                                    label="Amount ($)"
                                    placeholder="0.00"
                                    step="0.01"
                                />
                            )}
                        </form.AppField>

                        <form.AppField
                            name="dayOfMonthDue"
                            validators={{
                                onBlur: ({ value }) => {
                                    if (value < 1 || value > 31) {
                                        return "Day must be between 1 and 31";
                                    }
                                    return undefined;
                                },
                            }}
                        >
                            {(field) => (
                                <field.NumberField
                                    label="Day of Month Due"
                                    placeholder="1"
                                    min="1"
                                    max="31"
                                />
                            )}
                        </form.AppField>

                        <form.AppField name="isAutopay">
                            {(field) => (
                                <field.CheckboxField label="Autopay Enabled" />
                            )}
                        </form.AppField>

                        <div className="flex justify-end space-x-4 pt-6">
                            <button
                                type="button"
                                onClick={() => navigate({ to: "/" })}
                                className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                            >
                                Cancel
                            </button>
                            <button
                                type="submit"
                                disabled={isSubmitting}
                                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {isSubmitting
                                    ? "Saving..."
                                    : mode === "create"
                                      ? "Create Expense"
                                      : "Update Expense"}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
}
