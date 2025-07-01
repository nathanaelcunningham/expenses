import { 
    listFamilySettings,
    createFamilySetting,
    updateFamilySetting,
    deleteFamilySetting
} from "@/gen/family/v1/family-FamilySettingsService_connectquery";
import { useQuery, useMutation } from "@connectrpc/connect-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import type { FamilySetting } from "@/gen/family/v1/family_pb";
import { useAppForm } from "@/hooks/form";

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

    return (
        <div className="min-h-screen bg-gray-50 py-8">
            <div className="max-w-4xl mx-auto px-4">
                <div className="bg-white rounded-lg shadow-sm border">
                    <div className="px-6 py-4 border-b">
                        <h1 className="text-2xl font-semibold text-gray-900">Settings</h1>
                        <p className="text-sm text-gray-600 mt-1">
                            Manage your family settings and preferences
                        </p>
                    </div>
                    <div className="p-6">
                        <SettingsContent settings={data.familySettings} />
                    </div>
                </div>
            </div>
        </div>
    );
}

function SettingsContent({ settings }: { settings: FamilySetting[] }) {
    const [isCreating, setIsCreating] = useState(false);
    const [editingId, setEditingId] = useState<bigint | null>(null);

    const createMutation = useMutation(createFamilySetting);
    const updateMutation = useMutation(updateFamilySetting);
    const deleteMutation = useMutation(deleteFamilySetting);

    const form = useAppForm({
        defaultValues: {
            settingKey: "",
            settingValue: "",
            dataType: "string",
        },
        onSubmit: async ({ value }) => {
            try {
                if (editingId) {
                    await updateMutation.mutateAsync({
                        id: editingId,
                        settingValue: value.settingValue,
                        dataType: value.dataType,
                    });
                    setEditingId(null);
                } else {
                    await createMutation.mutateAsync({
                        settingKey: value.settingKey,
                        settingValue: value.settingValue,
                        dataType: value.dataType,
                    });
                    setIsCreating(false);
                }
                form.reset();
                window.location.reload();
            } catch (error) {
                console.error("Error saving setting:", error);
            }
        },
    });

    const handleEdit = (setting: FamilySetting) => {
        setEditingId(setting.id);
        form.setFieldValue("settingKey", setting.settingKey);
        form.setFieldValue("settingValue", setting.settingValue || "");
        form.setFieldValue("dataType", setting.dataType);
        setIsCreating(true);
    };

    const handleDelete = async (id: bigint) => {
        if (confirm("Are you sure you want to delete this setting?")) {
            try {
                await deleteMutation.mutateAsync({ id });
                window.location.reload();
            } catch (error) {
                console.error("Error deleting setting:", error);
            }
        }
    };

    const handleCancel = () => {
        setIsCreating(false);
        setEditingId(null);
        form.reset();
    };

    const dataTypeOptions = [
        { label: "Text", value: "string" },
        { label: "Number", value: "number" },
        { label: "Boolean", value: "boolean" },
    ];

    const simplefinSetting = settings.find(s => s.settingKey === "simplefin_token");

    return (
        <div className="space-y-6">
            <SimpleFin setting={simplefinSetting} onEdit={handleEdit} />
            
            <div>
                <div className="flex justify-between items-center mb-4">
                    <h2 className="text-lg font-medium text-gray-900">All Settings</h2>
                    <button
                        onClick={() => setIsCreating(true)}
                        disabled={isCreating}
                        className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white px-4 py-2 rounded-md text-sm font-medium"
                    >
                        Add Setting
                    </button>
                </div>

                {settings.length === 0 ? (
                    <div className="text-center py-8 text-gray-500">
                        No settings configured yet.
                    </div>
                ) : (
                    <div className="bg-gray-50 rounded-lg">
                        {settings.map((setting) => (
                            <div key={setting.id.toString()} className="flex items-center justify-between p-4 border-b last:border-b-0">
                                <div>
                                    <div className="font-medium text-gray-900">{setting.settingKey}</div>
                                    <div className="text-sm text-gray-600">
                                        {setting.settingValue || <em>No value</em>} 
                                        <span className="ml-2 text-xs text-gray-400">({setting.dataType})</span>
                                    </div>
                                </div>
                                <div className="flex gap-2">
                                    <button
                                        onClick={() => handleEdit(setting)}
                                        className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                                    >
                                        Edit
                                    </button>
                                    <button
                                        onClick={() => handleDelete(setting.id)}
                                        className="text-red-600 hover:text-red-800 text-sm font-medium"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}

                {isCreating && (
                    <div className="mt-6 bg-gray-50 p-6 rounded-lg border">
                        <h3 className="text-lg font-medium text-gray-900 mb-4">
                            {editingId ? "Edit Setting" : "Create New Setting"}
                        </h3>

                        <form
                            onSubmit={(e) => {
                                e.preventDefault();
                                e.stopPropagation();
                                form.handleSubmit();
                            }}
                            className="space-y-4"
                        >
                            {!editingId && (
                                <form.AppField name="settingKey">
                                    {(field) => (
                                        <field.TextField 
                                            label="Setting Key" 
                                            placeholder="e.g., api_token, max_users"
                                        />
                                    )}
                                </form.AppField>
                            )}
                            {editingId && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">
                                        Setting Key
                                    </label>
                                    <div className="w-full px-3 py-2 bg-gray-100 border border-gray-300 rounded-md text-gray-600">
                                        {form.getFieldValue("settingKey")}
                                    </div>
                                </div>
                            )}

                            <form.AppField name="settingValue">
                                {(field) => (
                                    <field.TextField 
                                        label="Setting Value" 
                                        placeholder="Enter the value"
                                    />
                                )}
                            </form.AppField>

                            <form.AppField name="dataType">
                                {(field) => (
                                    <field.Select 
                                        label="Data Type" 
                                        values={dataTypeOptions}
                                    />
                                )}
                            </form.AppField>

                            <div className="flex gap-3 pt-4">
                                <button
                                    type="submit"
                                    disabled={createMutation.isPending || updateMutation.isPending}
                                    className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white px-4 py-2 rounded-md text-sm font-medium"
                                >
                                    {createMutation.isPending || updateMutation.isPending 
                                        ? "Saving..." 
                                        : editingId ? "Update Setting" : "Create Setting"
                                    }
                                </button>
                                <button
                                    type="button"
                                    onClick={handleCancel}
                                    className="bg-gray-200 hover:bg-gray-300 text-gray-800 px-4 py-2 rounded-md text-sm font-medium"
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    </div>
                )}
            </div>
        </div>
    );
}

function SimpleFin({ setting, onEdit }: { 
    setting?: FamilySetting; 
    onEdit: (setting: FamilySetting) => void;
}) {
    const createMutation = useMutation(createFamilySetting);

    const handleAddToken = async () => {
        try {
            const response = await createMutation.mutateAsync({
                settingKey: "simplefin_token",
                settingValue: "",
                dataType: "string",
            });
            if (response.familySetting) {
                onEdit(response.familySetting);
            }
        } catch (error) {
            console.error("Error creating SimpleFin token:", error);
        }
    };
    return (
        <div className="border rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
                <div>
                    <h2 className="text-lg font-medium text-gray-900">SimpleFin Integration</h2>
                    <p className="text-sm text-gray-600">
                        Connect your bank accounts via SimpleFin for automatic transaction import
                    </p>
                </div>
                <div className="flex items-center">
                    {setting ? (
                        <div className="flex items-center gap-2">
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                Connected
                            </span>
                            <button
                                onClick={() => onEdit(setting)}
                                className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                            >
                                Edit Token
                            </button>
                        </div>
                    ) : (
                        <div className="flex items-center gap-2">
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                                Not Connected
                            </span>
                            <button
                                onClick={handleAddToken}
                                disabled={createMutation.isPending}
                                className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white px-3 py-1 rounded text-sm font-medium"
                            >
                                {createMutation.isPending ? "Adding..." : "Add Token"}
                            </button>
                        </div>
                    )}
                </div>
            </div>
            
            {setting ? (
                <div className="bg-gray-50 rounded p-3">
                    <div className="text-sm text-gray-600">
                        <strong>Token:</strong> {setting.settingValue ? `${setting.settingValue.substring(0, 8)}...` : "No value"}
                    </div>
                </div>
            ) : (
                <div className="bg-yellow-50 border border-yellow-200 rounded p-3">
                    <p className="text-sm text-yellow-800">
                        You haven't configured your SimpleFin token yet. This is required to import bank transactions automatically.
                    </p>
                </div>
            )}
        </div>
    );
}
