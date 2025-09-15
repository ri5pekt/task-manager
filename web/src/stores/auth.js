import { defineStore } from "pinia";
import { api } from "../lib/api";

export const useAuth = defineStore("auth", {
    state: () => ({ user: null, loading: false, error: "" }),
    getters: { isAuthed: (s) => !!s.user },
    actions: {
        async fetchMe() {
            this.loading = true;
            this.error = "";
            try {
                const me = await api.get("/api/me");
                this.user = me;
                return me;
            } catch (e) {
                if (e?.status === 401) this.user = null;
                else this.error = e?.message || "Failed to load user";
                throw e;
            } finally {
                this.loading = false;
            }
        },
        async login(email, password) {
            this.loading = true;
            this.error = "";
            try {
                await api.post("/api/login", { email, password });
                await this.fetchMe();
            } catch (e) {
                this.error = e?.message || "Login failed";
                throw e;
            } finally {
                this.loading = false;
            }
        },
        clientLogout() {
            this.user = null;
            this.error = "";
        },
    },
});
