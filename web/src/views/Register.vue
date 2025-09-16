<template>
    <div class="mx-auto max-w-sm py-10">
        <h1 class="mb-6 text-2xl font-semibold">Register</h1>

        <form @submit.prevent="onSubmit" class="space-y-4">
            <div>
                <label class="mb-1 block text-sm font-medium">Name</label>
                <input
                    v-model.trim="name"
                    required
                    class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-emerald-400"
                />
            </div>

            <div>
                <label class="mb-1 block text-sm font-medium">Email</label>
                <input
                    v-model.trim="email"
                    type="email"
                    required
                    class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-emerald-400"
                />
            </div>

            <div>
                <label class="mb-1 block text-sm font-medium">Password</label>
                <input
                    v-model="password"
                    type="password"
                    required
                    class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-emerald-400"
                />
            </div>

            <button
                type="submit"
                :disabled="loading"
                class="w-full rounded-md bg-gray-900 px-4 py-2 font-medium text-white hover:bg-gray-800 disabled:opacity-60"
            >
                <span v-if="loading">Creating account…</span>
                <span v-else>Register</span>
            </button>

            <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        </form>

        <p class="mt-6 text-sm text-gray-600">
            Already have an account?
            <RouterLink to="/login" class="underline">Login</RouterLink>
        </p>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "../lib/api";

const name = ref("");
const email = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");
const router = useRouter();

async function onSubmit() {
    loading.value = true;
    error.value = "";
    try {
        await api.post("/api/register", { name: name.value, email: email.value, password: password.value });
        await api.post("/api/login", { email: email.value, password: password.value });
        router.push("/board");
    } catch (e) {
        error.value = e?.message || "Registration failed";
    } finally {
        loading.value = false;
    }
}
</script>
