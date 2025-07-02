import { AccountsWrapper } from "@/components/AccountWrapper";
import {
    addAccount,
    getAccounts,
} from "@/gen/transaction/v1/transaction-TransactionService_connectquery";
import type { SimplefinAccount } from "@/gen/transaction/v1/transaction_pb";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/accounts")({
    component: () => (
        <AccountsWrapper>
            <TransactionAccounts />
        </AccountsWrapper>
    ),
});

function TransactionAccounts() {
    const { data, isLoading, error } = useQuery(getAccounts);

    const insertAccount = useMutation(addAccount);

    async function handleInsertAccount(account: SimplefinAccount) {
        try {
            await insertAccount.mutateAsync({
                accountId: account.id,
                name: account.name,
            });
        } catch (e) {
            window.alert(e);
        }
    }

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (error || data === undefined) {
        if (error) {
            return (
                <div>Error loading transaction accounts: {error.message}</div>
            );
        } else return <div>Error loading tranaction accounts</div>;
    }

    return (
        <div className="flex flex-col">
            {data.accounts.map((account) => (
                <>
                    <div>{account.name}</div>
                    <button onClick={() => handleInsertAccount(account)}>
                        Add Account
                    </button>
                </>
            ))}
        </div>
    );
}
