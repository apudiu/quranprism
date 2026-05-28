import {
  Field,
  type FieldElementProps,
  type FieldPath,
  type FieldStore,
  type FieldValues,
  type FormStore,
} from "@modular-forms/solid";
import { Show, type JSX } from "solid-js";
import {
  TextField,
  TextFieldErrorMessage,
  TextFieldInput,
  TextFieldLabel,
  TextFieldTextArea,
} from "../components/ui/text-field";

/**
 * `Field` re-typed for string-valued fields. modular-forms' `Field` uses a
 * conditional prop type that doesn't resolve through a generic wrapper, so we
 * cast once here while keeping `field`/`props` fully typed for callers.
 */
const StringField = Field as unknown as <
  TValues extends FieldValues,
  TName extends FieldPath<TValues>,
>(props: {
  of: FormStore<TValues, any>;
  name: TName;
  children: (
    field: FieldStore<TValues, TName>,
    props: FieldElementProps<TValues, TName>,
  ) => JSX.Element;
}) => JSX.Element;

export interface TextFormFieldProps<
  TValues extends FieldValues,
  TName extends FieldPath<TValues>,
> {
  /** The modular-forms store from `createForm`/`createFormStore`. */
  of: FormStore<TValues, any>;
  name: TName;
  label?: string;
  type?: "text" | "email" | "password" | "tel" | "url" | "number" | "search";
  placeholder?: string;
  multiline?: boolean;
  required?: boolean;
}

/**
 * solid-ui TextField bound to a modular-forms field. The shared kit both web
 * and admin import — pairs a styled control with valibot validation errors.
 */
export function TextFormField<
  TValues extends FieldValues,
  TName extends FieldPath<TValues>,
>(props: TextFormFieldProps<TValues, TName>) {
  return (
    <StringField of={props.of} name={props.name}>
      {(field, fieldProps) => (
        <TextField
          class="w-full"
          validationState={field.error ? "invalid" : "valid"}
          required={props.required}
        >
          <Show when={props.label}>
            <TextFieldLabel>{props.label}</TextFieldLabel>
          </Show>
          <Show
            when={props.multiline}
            fallback={
              <TextFieldInput
                {...fieldProps}
                type={props.type ?? "text"}
                placeholder={props.placeholder}
                value={(field.value as string | undefined) ?? ""}
              />
            }
          >
            <TextFieldTextArea
              {...fieldProps}
              placeholder={props.placeholder}
              value={(field.value as string | undefined) ?? ""}
            />
          </Show>
          <Show when={field.error}>
            <TextFieldErrorMessage>{field.error}</TextFieldErrorMessage>
          </Show>
        </TextField>
      )}
    </StringField>
  );
}
