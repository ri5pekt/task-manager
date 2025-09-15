import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
    plugins: [vue(), tailwindcss()],
    server: {
        host: true,
        port: 5173,
        strictPort: true,
        hmr: {
            host: "localhost",
            clientPort: 5173,
            protocol: "ws",
        },
        watch: {
            usePolling: true,
            interval: 300,
        },
        // ⬇️ put the proxy back so /api/* goes to the Go service
        proxy: {
            "/api": { target: "http://api:8080", changeOrigin: true },
        },
    },
});
