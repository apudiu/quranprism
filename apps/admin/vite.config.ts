import { defineConfig } from "vite";
import { nitroV2Plugin as nitro } from "@solidjs/vite-plugin-nitro-2";
import { solidStart } from "@solidjs/start/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  server: {
    port: 3001,
    strictPort: true,
  },
  preview: {
    port: 3001,
    strictPort: true,
  },
  // Workspace packages ship Solid .tsx source (no build step). Force vite to
  // bundle + run the Solid transform over them instead of externalizing.
  ssr: {
    noExternal: ["@qp/ui", "@qp/http"],
  },
  plugins: [
    solidStart(),
    tailwindcss(),
    nitro()
  ]
});
