import { getFamilySettingByKey } from "@/gen/family/v1/family-FamilySettingsService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { Link } from "@tanstack/react-router";
import type { PropsWithChildren } from "react";

export function AccountsWrapper(props: PropsWithChildren) {
    const { data, isLoading, error } = useQuery(getFamilySettingByKey, {
        key: "simplefin_token",
    });

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (error || data === undefined) {
        if (error) {
            return <div>Error loading simplefin token{error.message}</div>;
        } else return <div>Error loading simplefin token</div>;
    }

    if (
        !data.familySetting ||
        !data.familySetting.settingValue ||
        data.familySetting.settingValue === ""
    ) {
        return (
            <div>
                <p>You have not added a simplefin token.</p>
                <button>
                    <Link to={`/settings`}>Click here to go to settings</Link>
                </button>
            </div>
        );
    }

    return <>{props.children}</>;
}
