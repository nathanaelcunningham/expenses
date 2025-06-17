import { createFileRoute } from "@tanstack/react-router";
import { ExpenseForm } from "@/components/ExpenseForm";

export const Route = createFileRoute("/expense/create")({
    component: CreateExpense,
});

function CreateExpense() {
    return <ExpenseForm mode="create" />;
}