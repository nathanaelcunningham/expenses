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
            
            <MonthlyIncome settings={settings} />
            
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
                        {settings.map((setting) => {
                            const renderSettingValue = () => {
                                if (!setting.settingValue) {
                                    return <em>No value</em>;
                                }
                                
                                // Special handling for different setting types
                                if (setting.settingKey === "simplefin_token") {
                                    return `${setting.settingValue.substring(0, 12)}...`;
                                }
                                
                                if (setting.settingKey === "monthly_income") {
                                    try {
                                        const income = JSON.parse(setting.settingValue);
                                        return `Total: $${income.total_amount?.toLocaleString() || 0} (${income.sources?.length || 0} sources)`;
                                    } catch {
                                        return "Invalid JSON";
                                    }
                                }
                                
                                // For other settings, truncate if too long
                                if (setting.settingValue.length > 50) {
                                    return `${setting.settingValue.substring(0, 50)}...`;
                                }
                                
                                return setting.settingValue;
                            };
                            
                            return (
                                <div key={setting.id.toString()} className="flex items-center justify-between p-4 border-b last:border-b-0">
                                    <div className="flex-1 min-w-0 mr-4">
                                        <div className="font-medium text-gray-900">{setting.settingKey}</div>
                                        <div className="text-sm text-gray-600 break-words">
                                            {renderSettingValue()}
                                            <span className="ml-2 text-xs text-gray-400">({setting.dataType})</span>
                                        </div>
                                    </div>
                                    <div className="flex gap-2 flex-shrink-0">
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
                            );
                        })}
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

function MonthlyIncome({ settings }: { settings: FamilySetting[] }) {
    const [isEditing, setIsEditing] = useState(false);
    const [incomeSources, setIncomeSources] = useState<Array<{ name: string; amount: number; description: string; is_active: boolean }>>([]);
    
    const createMutation = useMutation(createFamilySetting);
    const updateMutation = useMutation(updateFamilySetting);
    
    const incomeSetting = settings.find(s => s.settingKey === "monthly_income");
    
    // Parse existing income data
    const existingIncome = incomeSetting?.settingValue 
        ? JSON.parse(incomeSetting.settingValue)
        : { total_amount: 0, sources: [], updated_at: new Date().toISOString() };
    
    // Initialize income sources when editing starts
    const startEditing = () => {
        setIncomeSources(existingIncome.sources || []);
        setIsEditing(true);
    };
    
    const cancelEditing = () => {
        setIncomeSources([]);
        setIsEditing(false);
    };
    
    const saveIncome = async () => {
        console.log("Submitting income data:", incomeSources);
        try {
            const incomeData = {
                total_amount: incomeSources.reduce((sum: number, source: any) => sum + (source.is_active ? source.amount : 0), 0),
                sources: incomeSources,
                updated_at: new Date().toISOString(),
            };
            
            console.log("Prepared income data:", incomeData);
            const incomeValue = JSON.stringify(incomeData);
            
            if (incomeSetting) {
                console.log("Updating existing setting:", incomeSetting.id);
                await updateMutation.mutateAsync({
                    id: incomeSetting.id,
                    settingValue: incomeValue,
                    dataType: "json",
                });
            } else {
                console.log("Creating new setting");
                await createMutation.mutateAsync({
                    settingKey: "monthly_income",
                    settingValue: incomeValue,
                    dataType: "json",
                });
            }
            
            setIsEditing(false);
            window.location.reload();
        } catch (error) {
            console.error("Error saving income:", error);
            alert("Error saving income: " + (error as Error).message);
        }
    };
    
    const addIncomeSource = () => {
        const newSources = [...incomeSources, { name: "", amount: 0, description: "", is_active: true }];
        console.log("Adding income source, new sources:", newSources);
        setIncomeSources(newSources);
    };
    
    const removeIncomeSource = (index: number) => {
        setIncomeSources(incomeSources.filter((_: any, i: number) => i !== index));
    };
    
    const updateIncomeSource = (index: number, field: string, value: any) => {
        const newSources = [...incomeSources];
        newSources[index] = { ...newSources[index], [field]: value };
        setIncomeSources(newSources);
    };
    
    const totalIncome = existingIncome.sources.reduce((sum: number, source: any) => sum + (source.is_active ? source.amount : 0), 0);
    
    return (
        <div className="border rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
                <div>
                    <h2 className="text-lg font-medium text-gray-900">Monthly Income</h2>
                    <p className="text-sm text-gray-600">
                        Track your family's monthly income sources
                    </p>
                </div>
                <div className="flex items-center gap-4">
                    <div className="text-right">
                        <div className="text-2xl font-bold text-green-600">
                            ${totalIncome.toLocaleString()}
                        </div>
                        <div className="text-sm text-gray-500">Total Monthly</div>
                    </div>
                    <button
                        onClick={isEditing ? cancelEditing : startEditing}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                    >
                        {isEditing ? "Cancel" : "Edit Income"}
                    </button>
                </div>
            </div>
            
            {!isEditing ? (
                <div className="space-y-3">
                    {existingIncome.sources.length === 0 ? (
                        <div className="bg-yellow-50 border border-yellow-200 rounded p-3">
                            <p className="text-sm text-yellow-800">
                                No income sources configured yet. Click "Edit Income" to add your income sources.
                            </p>
                        </div>
                    ) : (
                        <div className="space-y-2">
                            {existingIncome.sources.map((source: any, index: number) => (
                                <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                                    <div className="flex-1">
                                        <div className="font-medium text-gray-900">{source.name}</div>
                                        {source.description && (
                                            <div className="text-sm text-gray-600">{source.description}</div>
                                        )}
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                            source.is_active 
                                                ? 'bg-green-100 text-green-800' 
                                                : 'bg-gray-100 text-gray-800'
                                        }`}>
                                            {source.is_active ? 'Active' : 'Inactive'}
                                        </span>
                                        <div className="text-lg font-semibold text-gray-900">
                                            ${source.amount.toLocaleString()}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            ) : (
                <div className="space-y-4">
                    <div>
                        <div className="space-y-4">
                            <div className="flex items-center justify-between">
                                <h3 className="text-lg font-medium text-gray-900">Income Sources</h3>
                                <button
                                    type="button"
                                    onClick={addIncomeSource}
                                    className="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm font-medium"
                                >
                                    Add Source
                                </button>
                            </div>
                            
                            {incomeSources.map((source: any, index: number) => (
                                <div key={index} className="border rounded-lg p-4 bg-gray-50">
                                    <div className="flex items-center justify-between mb-3">
                                        <h4 className="font-medium text-gray-900">Income Source {index + 1}</h4>
                                        <button
                                            type="button"
                                            onClick={() => removeIncomeSource(index)}
                                            className="text-red-600 hover:text-red-800 text-sm font-medium"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                    
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                                Source Name
                                            </label>
                                            <input
                                                type="text"
                                                value={source.name}
                                                onChange={(e) => updateIncomeSource(index, "name", e.target.value)}
                                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                                                placeholder="e.g., Salary, Freelance"
                                            />
                                        </div>
                                        
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                                Monthly Amount
                                            </label>
                                            <input
                                                type="number"
                                                value={source.amount}
                                                onChange={(e) => updateIncomeSource(index, "amount", parseFloat(e.target.value) || 0)}
                                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                                                placeholder="0.00"
                                                min="0"
                                                step="0.01"
                                            />
                                        </div>
                                        
                                        <div className="md:col-span-2">
                                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                                Description (optional)
                                            </label>
                                            <input
                                                type="text"
                                                value={source.description}
                                                onChange={(e) => updateIncomeSource(index, "description", e.target.value)}
                                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                                                placeholder="Additional details about this income source"
                                            />
                                        </div>
                                        
                                        <div className="md:col-span-2">
                                            <label className="flex items-center">
                                                <input
                                                    type="checkbox"
                                                    checked={source.is_active}
                                                    onChange={(e) => updateIncomeSource(index, "is_active", e.target.checked)}
                                                    className="mr-2"
                                                />
                                                <span className="text-sm text-gray-700">Active income source</span>
                                            </label>
                                        </div>
                                    </div>
                                </div>
                            ))}
                            
                            <div className="flex gap-3 pt-4">
                                <button
                                    type="button"
                                    onClick={saveIncome}
                                    disabled={createMutation.isPending || updateMutation.isPending}
                                    className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white px-4 py-2 rounded-md text-sm font-medium"
                                >
                                    {createMutation.isPending || updateMutation.isPending ? "Saving..." : "Save Income"}
                                </button>
                                <button
                                    type="button"
                                    onClick={cancelEditing}
                                    className="bg-gray-200 hover:bg-gray-300 text-gray-800 px-4 py-2 rounded-md text-sm font-medium"
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
