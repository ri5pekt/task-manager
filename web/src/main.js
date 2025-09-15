import { createApp } from "vue";
import { createPinia } from "pinia";
import router from "./router";
import "./style.css";
import App from "./App.vue";
import { api } from "./lib/api";

if (import.meta.env.DEV) {
    // dev convenience: use in browser console:  api.get('/api/boards')
    // eslint-disable-next-line no-undef
    window.api = api;
}

const app = createApp(App);
app.use(createPinia());
app.use(router);
app.mount("#app");
