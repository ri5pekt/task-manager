<!-- web/src/views/Board.vue -->
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

            <!-- Lists (draggable as a row) -->
            <Draggable
                v-model="board.lists"
                item-key="id"
                group="lists"
                :animation="180"
                ghost-class="drag-list-ghost"
                chosen-class="drag-list-chosen"
                drag-class="drag-list-dragging"
                tag="div"
                class="flex gap-4 overflow-x-auto pb-2 items-start"
                @change="onListChange"
            >
                <template #item="{ element: l, index }">
                    <div :key="l.id" class="w-72 shrink-0 rounded-lg border border-gray-200 bg-white">
                        <div class="border-b border-gray-200 px-4 py-2 cursor-grab">
                            <div class="flex items-center justify-between">
                                <span class="font-medium">{{ l.name }}</span>
                                <!-- show live index to reflect client-side reorder instantly -->
                                <span class="text-xs text-gray-500">pos {{ index }}</span>
                            </div>
                        </div>

                        <!-- Tasks in this list (draggable, cross-list moves enabled) -->
                        <draggable
                            v-model="l.tasks"
                            group="tasks"
                            item-key="id"
                            class="space-y-2 p-3 list-none"
                            :data-list-id="l.id"
                            @add="onTaskAdd"
                            @change="(e) => onTaskOrderChange(e, l)"
                            ghost-class="drag-task-ghost"
                            chosen-class="drag-task-chosen"
                            drag-class="drag-task-dragging"
                        >
                            <template #item="{ element: t }">
                                <li
                                    class="rounded-md border border-emerald-200 bg-emerald-50 p-3 cursor-pointer hover:bg-emerald-100"
                                    @click="openTask(t)"
                                >
                                    <div class="flex items-center justify-between">
                                        <span class="font-medium">{{ t.title }}</span>
                                        <span class="text-xs uppercase tracking-wide text-gray-600">{{
                                            t.status
                                        }}</span>
                                    </div>
                                    <div class="mt-2 flex items-center gap-3 text-sm text-gray-600">
                                        <span class="flex items-center gap-1">
                                            <ChatBubbleLeftIcon class="h-4 w-4 text-gray-500" /> {{ t.comment_count }}
                                        </span>
                                        <span class="flex items-center gap-1">
                                            <UserIcon class="h-4 w-4 text-gray-500" /> x{{ t.assignees?.length || 0 }}
                                        </span>
                                        <span class="ml-auto text-xs text-gray-400">pos {{ t.position }}</span>
                                    </div>
                                </li>
                            </template>
                            <template #footer>
                                <li v-if="!l.tasks?.length" class="px-3 py-2 text-sm text-gray-500">No tasks</li>
                            </template>
                        </draggable>

                        <button
                            class="m-3 mt-1 w-[calc(100%-1.5rem)] rounded-md border border-dashed border-gray-300 py-2 text-sm text-gray-600 hover:border-gray-400 hover:bg-white"
                            @click="openCreate(l.id)"
                        >
                            + Add task
                        </button>
                    </div>
                </template>
            </Draggable>
            <!-- /Lists -->
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
import Draggable from "vuedraggable"; // Lists
import draggable from "vuedraggable"; // Tasks (same component; using both tags is fine)

const route = useRoute();
const loading = ref(false);
const error = ref("");

// Keep board reactive; null is fine initially
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
    route.query.id; // track dep
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

function onUpdated(updated) {
    console.log("[Board] onUpdated:", updated);
    if (!updated?.id || !board.value?.lists?.length) return;
    for (const list of board.value.lists) {
        const idx = list.tasks?.findIndex((t) => t.id === updated.id) ?? -1;
        if (idx !== -1) {
            const prev = list.tasks[idx];
            list.tasks.splice(idx, 1, {
                ...prev,
                title: updated.title ?? prev.title,
                description: updated.description ?? prev.description,
                status: updated.status ?? prev.status,
                position: typeof updated.position === "number" ? updated.position : prev.position,
            });
            break;
        }
    }
}

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

// ===== DND: Tasks =====
const lastFromListId = ref(""); // set by @add, used by @change(added)

function onTaskAdd(evt) {
    // evt is a CustomEvent from Sortable with DOM refs
    lastFromListId.value = evt?.from?.dataset?.listId || "";
    console.log("[DND] add: fromListId =", lastFromListId.value);
}

async function onTaskOrderChange(evt, destList) {
    const { moved, added } = evt;

    // Same-list reorder
    if (moved) {
        const payload = {
            task_id: moved.element?.id,
            to_list_id: destList.id,
            to_index: moved.newIndex ?? 0,
        };
        try {
            await api.post("/api/tasks/reorder", payload);
            console.log("[DND] persisted same-list reorder", payload);
        } catch (e) {
            console.error("persist failed", e);
            // optional: await loadBoard();
        }
    }

    // Cross-list move (destination side)
    if (added) {
        const payload = {
            task_id: added.element?.id,
            to_list_id: destList.id,
            to_index: added.newIndex ?? 0,
        };
        try {
            await api.post("/api/tasks/reorder", payload);
            console.log("[DND] persisted cross-list move", payload);
        } catch (e) {
            console.error("persist failed", e);
            // optional: await loadBoard();
        }
    }
}

// ===== DND: Lists =====
async function onListChange() {
    if (!board.value?.lists?.length) return;
    const order = board.value.lists.map((l) => l.id);
    console.log("[DND][lists] new order:", order);
    try {
        await api.post("/api/lists/reorder", { board_id: board.value.id, list_ids: order });
        console.log("[DND][lists] persisted");
    } catch (e) {
        console.error("List reorder failed:", e?.message || e);
        // optional: await loadBoard();
    }
}
</script>

<style scoped>
/* Task drag visuals */
.drag-task-ghost {
    background-color: #e5e7eb; /* gray-200 */
    border: 2px dashed #9ca3af; /* gray-400 */
    min-height: 3.5rem;
    border-radius: 0.375rem;
    opacity: 0.6;
}
.drag-task-chosen {
    opacity: 0.5;
}
.drag-task-dragging {
    cursor: grabbing;
}

/* List drag visuals */
.drag-list-ghost {
    opacity: 0.5;
    transform: scale(0.98);
}
.drag-list-chosen {
    opacity: 0.8;
}
.drag-list-dragging {
    cursor: grabbing;
}
</style>
