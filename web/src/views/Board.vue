<!-- Board.vue -->
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
                            class="rounded-md border border-emerald-200 bg-emerald-50 p-3 cursor-pointer hover:bg-emerald-100"
                            @click="openTask(t)"
                        >
                            <div class="flex items-center justify-between">
                                <span class="font-medium">{{ t.title }}</span>
                                <span class="text-xs uppercase tracking-wide text-gray-600">{{ t.status }}</span>
                            </div>
                            <div class="mt-2 flex items-center gap-3 text-sm text-gray-600">
                                +
                                <span class="flex items-center gap-1">
                                    <ChatBubbleLeftIcon class="h-4 w-4 text-gray-500" /> {{ t.comment_count }}
                                </span>
                                <span class="flex items-center gap-1">
                                    <UserIcon class="h-4 w-4 text-gray-500" /> x{{ t.assignees?.length || 0 }}
                                </span>
                                <span class="ml-auto text-xs text-gray-400">pos {{ t.position }}</span>
                            </div>
                        </li>
                        <li v-if="!l.tasks?.length" class="px-3 py-2 text-sm text-gray-500">No tasks</li>
                    </ul>
                    <button
                        class="m-3 mt-1 w-[calc(100%-1.5rem)] rounded-md border border-dashed border-gray-300 py-2 text-sm text-gray-600 hover:border-gray-400 hover:bg-white"
                        @click="openCreate(l.id)"
                    >
                        + Add task
                    </button>
                </div>
            </div>
        </div>

        <p v-else class="text-gray-600">No board loaded yet.</p>
    </div>

    <TaskModal
        :open="showModal"
        :mode="modalMode"
        :listId="targetListId"
        :task="activeTask"
        @close="showModal = false"
        @created="onCreated"
        @commented="onCommented"
        @updated="onUpdated"
        @deleted="onDeleted"
    />
</template>

<script setup>
import { ref, watchEffect } from "vue";
import { useRoute } from "vue-router";
import { api } from "../lib/api";
import { ChatBubbleLeftIcon, UserIcon } from "@heroicons/vue/24/outline";
import TaskModal from "../components/TaskModal.vue";

const route = useRoute();
const loading = ref(false);
const error = ref("");

// Keep board reactive; null is fine initially, we replace with server payload
const board = ref(null);

// --- load board (first by default, or ?id=...)
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

// auto-load on mount and when ?id changes
watchEffect(() => {
    // dependency track
    route.query.id;
    loadBoard();
});

// ----- Modal orchestration -----
const showModal = ref(false);
const modalMode = ref("create"); // "create" | "view"
const targetListId = ref(""); // used for create
const activeTask = ref(null); // keep a REFERENCE to the task in board.lists

function openCreate(listId) {
    modalMode.value = "create";
    targetListId.value = listId;
    activeTask.value = null;
    showModal.value = true;
}

function openTask(task) {
    // IMPORTANT: keep reference, do NOT clone
    console.log("[Board] openTask:", { id: task.id, title: task.title, hasDescription: !!task.description });
    modalMode.value = "view";
    activeTask.value = task;
    targetListId.value = "";
    showModal.value = true;
}

// ----- Events from TaskModal -----

// When TaskModal creates a task, append it to the right list (reactively)
function onCreated(res) {
    console.log("[Board] onCreated:", res);
    if (!board.value?.lists?.length) return;
    const list =
        board.value.lists.find((l) => l.id === res.list_id) ||
        board.value.lists.find((l) => l.id === targetListId.value);
    if (!list) return;
    if (!Array.isArray(list.tasks)) list.tasks = [];
    list.tasks.push({
        id: res.id,
        list_id: res.list_id,
        title: res.title,
        description: res.description || "",
        status: res.status,
        position: res.position,
        assignees: res.assignees ?? [],
        comment_count: 0,
    });
}

// When a comment is posted in TaskModal, bump the card’s comment_count
function onCommented(c) {
    const taskId = activeTask.value?.id ?? c.task_id;
    if (!taskId || !board.value?.lists?.length) return;
    for (const list of board.value.lists) {
        const card = list.tasks?.find((t) => t.id === taskId);
        if (card) {
            card.comment_count = (card.comment_count || 0) + 1;
            break;
        }
    }
}

// PATCH result → update the task in-place using splice to preserve reactivity
function onUpdated(updated) {
    console.log("[Board] onUpdated:", updated);
    if (!updated?.id || !board.value?.lists?.length) return;
    for (const list of board.value.lists) {
        const idx = list.tasks?.findIndex((t) => t.id === updated.id) ?? -1;
        if (idx !== -1) {
            const prev = list.tasks[idx];
            // replace via splice to trigger reactivity
            list.tasks.splice(idx, 1, {
                ...prev, // keep fields not returned by PATCH (assignees, comment_count, list_id)
                title: updated.title ?? prev.title,
                description: updated.description ?? prev.description,
                status: updated.status ?? prev.status,
                position: typeof updated.position === "number" ? updated.position : prev.position,
            });
            break;
        }
    }
}

// DELETE result → remove the task reactively
function onDeleted(taskId) {
    console.log("[Board] onDeleted:", taskId);
    if (!taskId || !board.value?.lists?.length) return;
    for (const list of board.value.lists) {
        const idx = list.tasks?.findIndex((t) => t.id === taskId) ?? -1;
        if (idx !== -1) {
            list.tasks.splice(idx, 1);
            break;
        }
    }
}
</script>

<style scoped></style>
