<template>
    <div class="mx-auto max-w-sm py-10">
        <h1 class="mb-6 text-2xl font-semibold">Login</h1>

        <form @submit.prevent="onSubmit" class="space-y-4">
            <div>
                <label class="mb-1 block text-sm font-medium">Email</label>
                <input
                    v-model.trim="email"
                    type="email"
                    required
                    class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-emerald-400"
                    placeholder="you@example.com"
                />
            </div>

            <div>
                <label class="mb-1 block text-sm font-medium">Password</label>
                <input
                    v-model="password"
                    type="password"
                    required
                    class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-emerald-400"
                    placeholder="••••••••"
                />
            </div>

            <button
                type="submit"
                :disabled="auth.loading"
                class="w-full rounded-md bg-gray-900 px-4 py-2 font-medium text-white hover:bg-gray-800 disabled:opacity-60"
            >
                <span v-if="auth.loading">Signing in…</span>
                <span v-else>Sign in</span>
            </button>

            <p v-if="auth.error" class="text-sm text-red-600">{{ auth.error }}</p>
        </form>

        <p class="mt-6 text-sm text-gray-600">
            Don’t have an account?
            <RouterLink to="/register" class="underline">Register</RouterLink>
        </p>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuth } from "../stores/auth";

const email = ref("");
const password = ref("");
const auth = useAuth();
const router = useRouter();

async function onSubmit() {
    try {
        await auth.login(email.value, password.value);
        router.push("/");
    } catch {
        /* error shown by store */
    }
}
</script>

<style scoped></style>
