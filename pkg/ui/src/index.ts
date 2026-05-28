// Utilities
export { cn } from "./lib/utils";

// Vendored solid-ui components (Kobalte-based, Tailwind v4)
export * from "./components/ui/button";
export * from "./components/ui/card";
export * from "./components/ui/text-field";
export * from "./components/ui/label";
export * from "./components/ui/badge";
export * from "./components/ui/skeleton";
export * from "./components/ui/switch";
export * from "./components/ui/checkbox";
export * from "./components/ui/select";
export * from "./components/ui/dialog";
export * from "./components/ui/dropdown-menu";
export * from "./components/ui/popover";
export * from "./components/ui/tabs";
export * from "./components/ui/separator";
export * from "./components/ui/table";
export * from "./components/ui/sonner";
export * from "./components/data-table/DataTable";

// Color mode (Kobalte) + theme toggle
export * from "./color-mode";

// Form kit (modular-forms + valibot, bound to solid-ui controls)
export * from "./form/TextFormField";

// Convenience re-exports so apps import forms from one place
export {
  createForm,
  createFormStore,
  Form,
  Field,
  FieldArray,
  getValue,
  getValues,
  setValue,
  setValues,
  reset,
  setError,
  valiForm,
  valiField,
} from "@modular-forms/solid";
export { toast } from "solid-sonner";
