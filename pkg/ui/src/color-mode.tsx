import { useColorMode } from "@kobalte/core";
import Moon from "lucide-solid/icons/moon";
import Sun from "lucide-solid/icons/sun";
import { Show } from "solid-js";
import { Button } from "./components/ui/button";

export {
  ColorModeProvider,
  ColorModeScript,
  useColorMode,
  cookieStorageManagerSSR,
} from "@kobalte/core";

/** Sun/moon button that flips Kobalte's color mode (writes `[data-kb-theme]`). */
export function ThemeToggle() {
  const { colorMode, toggleColorMode } = useColorMode();
  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleColorMode}
      aria-label="Toggle color mode"
    >
      <Show when={colorMode() === "dark"} fallback={<Sun class="size-4" />}>
        <Moon class="size-4" />
      </Show>
    </Button>
  );
}
