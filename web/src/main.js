import { createApp } from "vue";
import { createPinia } from "pinia";
import router from "./router";
import "./style.css";
import App from "./App.vue";
import { api } from "./lib/api";
import { useAuth } from "./stores/auth";

if (import.meta.env.DEV) {
    // dev convenience
    // eslint-disable-next-line no-undef
    window.api = api;
}

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
app.use(router);

// 🔐 Try to restore user session on first load
useAuth(pinia)
    .fetchMe()
    .catch(() => {
        /* not logged in is fine */
    });

app.mount("#app");
