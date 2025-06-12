import { createFormHook } from "@tanstack/react-form";

import { Select, TextArea, TextField } from "../components/FormComponents";
import { fieldContext, formContext } from "./form-context";

export const { useAppForm } = createFormHook({
    fieldComponents: {
        TextField,
        Select,
        TextArea,
    },
    formComponents: {},
    fieldContext,
    formContext,
});
