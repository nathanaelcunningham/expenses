import { useStore } from "@tanstack/react-form";

import { useFieldContext } from "../hooks/form-context";

function ErrorMessages({
    errors,
}: {
    errors: Array<string | { message: string }>;
}) {
    return (
        <>
            {errors.map((error) => (
                <div
                    key={typeof error === "string" ? error : error.message}
                    className="text-sm text-red-600 mt-1"
                >
                    {typeof error === "string" ? error : error.message}
                </div>
            ))}
        </>
    );
}

export function TextField({
    label,
    placeholder,
}: {
    label: string;
    placeholder?: string;
}) {
    const field = useFieldContext<string>();
    const errors = useStore(field.store, (state) => state.meta.errors);

    return (
        <div>
            <label htmlFor={label} className="block text-sm font-medium text-gray-700 mb-2">
                {label}
            </label>
            <input
                value={field.state.value}
                placeholder={placeholder}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            {field.state.meta.isTouched && <ErrorMessages errors={errors} />}
        </div>
    );
}

export function TextArea({
    label,
    rows = 3,
}: {
    label: string;
    rows?: number;
}) {
    const field = useFieldContext<string>();
    const errors = useStore(field.store, (state) => state.meta.errors);

    return (
        <div>
            <label htmlFor={label} className="block text-sm font-medium text-gray-700 mb-2">
                {label}
            </label>
            <textarea
                value={field.state.value}
                onBlur={field.handleBlur}
                rows={rows}
                onChange={(e) => field.handleChange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            {field.state.meta.isTouched && <ErrorMessages errors={errors} />}
        </div>
    );
}

export function Select({
    label,
    values,
}: {
    label: string;
    values: Array<{ label: string; value: string }>;
    placeholder?: string;
}) {
    const field = useFieldContext<string>();
    const errors = useStore(field.store, (state) => state.meta.errors);

    return (
        <div>
            <label htmlFor={label} className="block text-sm font-medium text-gray-700 mb-2">
                {label}
            </label>
            <select
                name={field.name}
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
                {values.map((value) => (
                    <option key={value.value} value={value.value}>
                        {value.label}
                    </option>
                ))}
            </select>
            {field.state.meta.isTouched && <ErrorMessages errors={errors} />}
        </div>
    );
}

export function NumberField({
    label,
    placeholder,
    step,
    min,
    max,
}: {
    label: string;
    placeholder?: string;
    step?: string;
    min?: string;
    max?: string;
}) {
    const field = useFieldContext<number>();
    const errors = useStore(field.store, (state) => state.meta.errors);

    return (
        <div>
            <label htmlFor={label} className="block text-sm font-medium text-gray-700 mb-2">
                {label}
            </label>
            <input
                type="number"
                step={step}
                min={min}
                max={max}
                value={field.state.value}
                placeholder={placeholder}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(Number.parseFloat(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            {field.state.meta.isTouched && <ErrorMessages errors={errors} />}
        </div>
    );
}

export function CheckboxField({
    label,
}: {
    label: string;
}) {
    const field = useFieldContext<boolean>();

    return (
        <div className="flex items-center">
            <input
                type="checkbox"
                checked={field.state.value}
                onChange={(e) => field.handleChange(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <label className="ml-2 block text-sm text-gray-900">
                {label}
            </label>
        </div>
    );
}
