<template>
    <div class="space-y-4">
        <h1 class="text-2xl font-semibold">Welcome</h1>

        <div v-if="auth.isAuthed" class="flex items-center gap-3">
            <span class="text-gray-700"
                >Signed in as <strong>{{ auth.user?.name }}</strong></span
            >
            <button @click="onLogout" class="rounded-md bg-gray-900 px-3 py-1.5 text-white hover:bg-gray-800">
                Logout
            </button>
        </div>

        <div v-else>
            <RouterLink to="/login" class="underline">Login</RouterLink>
            <span class="mx-1">·</span>
            <RouterLink to="/register" class="underline">Register</RouterLink>
        </div>
    </div>
</template>

<script setup>
import { useRouter } from "vue-router";
import { useAuth } from "../stores/auth";
import { api } from "../lib/api";

const router = useRouter();
const auth = useAuth();

async function onLogout() {
    try {
        await api.post("/api/logout", {}); // clears cookies server-side
    } catch (_) {}
    // reset client state
    auth.$reset?.(); // if defined; otherwise:
    auth.user = null;
    auth.isAuthed = false;
    router.push("/login");
}
</script>
