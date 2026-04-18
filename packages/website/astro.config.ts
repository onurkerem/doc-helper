import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  site: "https://doc-helper.dev",
  vite: {
    plugins: [tailwindcss()],
  },
});
