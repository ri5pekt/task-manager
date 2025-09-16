<template>
    <div class="min-h-screen bg-gray-50 text-gray-900">
        <header class="bg-gray-900 text-white">
            <nav class="mx-auto max-w-5xl px-4 py-3 flex items-center justify-between">
                <RouterLink to="/" class="font-semibold tracking-wide">Task Manager</RouterLink>
                <div class="flex gap-4 items-center">
                    <RouterLink to="/" class="hover:text-emerald-300">Home</RouterLink>
                    <RouterLink to="/board" class="hover:text-emerald-300">Board</RouterLink>
                    <span v-if="auth.isAuthed" class="font-medium text-emerald-300">
                        {{ auth.user?.name }}
                    </span>
                    <RouterLink v-else to="/login" class="hover:text-emerald-300">Login</RouterLink>
                </div>
            </nav>
        </header>
        <main class="mx-auto max-w-5xl px-4 py-6">
            <RouterView />
        </main>
    </div>
</template>

<script setup>
import { useAuth } from "./stores/auth";
const auth = useAuth();
</script>
