import { createFormHook } from "@tanstack/react-form";

import { CheckboxField, NumberField, Select, TextArea, TextField } from "../components/FormComponents";
import { fieldContext, formContext } from "./form-context";

export const { useAppForm } = createFormHook({
    fieldComponents: {
        TextField,
        NumberField,
        CheckboxField,
        Select,
        TextArea,
    },
    formComponents: {},
    fieldContext,
    formContext,
});
