import { listFamilySettings } from "@/gen/family/v1/family-FamilySettingsService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/settings")({
    component: RouteComponent,
});

function RouteComponent() {
    const { data, isLoading, error } = useQuery(listFamilySettings);

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (error || data === undefined) {
        if (error) {
            return <div>Error loading settings: {error.message}</div>;
        } else return <div>Error loading settings</div>;
    }

    return <div></div>;
}
