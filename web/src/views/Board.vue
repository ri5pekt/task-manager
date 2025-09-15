<template>
    <div class="mx-auto max-w-7xl p-6">
        <div class="mb-4 flex items-center gap-3">
            <h1 class="text-2xl font-semibold">Board</h1>
            <button
                @click="loadBoard"
                :disabled="loading"
                class="rounded-md bg-gray-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-60"
            >
                {{ loading ? "Loading…" : "Refresh" }}
            </button>
            <div v-if="error" class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                <span v-if="error.toLowerCase().includes('unauthorized')">
                    You need to log in to view boards.
                    <RouterLink to="/login" class="underline">Go to Login</RouterLink>
                </span>
                <span v-else>{{ error }}</span>
            </div>
        </div>

        <div v-if="board" class="space-y-2">
            <h2 class="text-lg font-medium text-gray-700">{{ board.name }}</h2>

            <!-- Lists grid -->
            <div class="flex gap-4 overflow-x-auto pb-2">
                <div
                    v-for="l in board.lists"
                    :key="l.id"
                    class="w-72 shrink-0 rounded-lg border border-gray-200 bg-white"
                >
                    <div class="border-b border-gray-200 px-4 py-2">
                        <div class="flex items-center justify-between">
                            <span class="font-medium">{{ l.name }}</span>
                            <span class="text-xs text-gray-500">pos {{ l.position }}</span>
                        </div>
                    </div>

                    <ul class="space-y-2 p-3">
                        <li
                            v-for="t in l.tasks"
                            :key="t.id"
                            class="rounded-md border border-emerald-200 bg-emerald-50 p-3"
                        >
                            <div class="flex items-center justify-between">
                                <span class="font-medium">{{ t.title }}</span>
                                <span class="text-xs uppercase tracking-wide text-gray-600">{{ t.status }}</span>
                            </div>
                            <div class="mt-2 flex items-center gap-3 text-sm text-gray-600">
                                <span>🗨️ {{ t.comment_count }}</span>
                                <span>👤 x{{ t.assignees?.length || 0 }}</span>
                                <span class="ml-auto text-xs text-gray-400">pos {{ t.position }}</span>
                            </div>
                        </li>
                        <li v-if="!l.tasks?.length" class="px-3 py-2 text-sm text-gray-500">No tasks</li>
                    </ul>
                </div>
            </div>
        </div>

        <p v-else class="text-gray-600">No board loaded yet.</p>
    </div>
</template>

<script setup>
import { ref, watchEffect } from "vue";
import { useRoute } from "vue-router";
import { api } from "../lib/api";

const route = useRoute();
const loading = ref(false);
const error = ref("");
const board = ref(null);

// Loads first board by default, or ?id=<uuid> if provided
async function loadBoard() {
    loading.value = true;
    error.value = "";
    try {
        const id = route.query.id;
        const url = id ? `/api/boards?id=${encodeURIComponent(id)}` : "/api/boards";
        board.value = await api.get(url);
    } catch (e) {
        error.value = e?.message || "Failed to load board";
    } finally {
        loading.value = false;
    }
}

// auto-load when the view mounts and when ?id changes
watchEffect(() => {
    route.query.id; // track dependency
    loadBoard();
});
</script>

<style scoped></style>
