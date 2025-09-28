import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// https://vitejs.dev/config/
export default defineConfig(({ command, mode, ssrBuild }) => {
  const apiUrl = "http://localhost:3000";
  const ret = {
    plugins: [vue()],
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },
    define: {
      __API_URL__: JSON.stringify(apiUrl),
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: "false",
    },
  };

  return ret;
});
