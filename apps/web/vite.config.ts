import { defineConfig } from "vite";
import { nitroV2Plugin as nitro } from "@solidjs/vite-plugin-nitro-2";
import { solidStart } from "@solidjs/start/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  server: {
    port: 3002,
    strictPort: true,
  },
  preview: {
    port: 3002,
    strictPort: true,
  },
  plugins: [
    solidStart(),
    tailwindcss(),
    nitro()
  ]
});
