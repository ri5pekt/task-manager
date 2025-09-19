<!-- TaskModal.vue -->
<template>
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40" @click="$emit('close')"></div>

        <div class="relative w-full max-w-2xl rounded-2xl bg-white p-5 shadow-xl">
            <!-- Header -->
            <div class="mb-4 flex items-center justify-between">
                <h3 class="text-lg font-semibold">
                    {{ mode === "create" ? "Create task" : isEditing ? "Edit task" : "Task details" }}
                </h3>
                <div class="flex items-center gap-2">
                    <!-- Create mode actions -->
                    <template v-if="mode === 'create'">
                        <button class="rounded px-2 py-1 text-sm hover:bg-gray-100" @click="$emit('close')">
                            Cancel
                        </button>
                        <button
                            :disabled="saving || !formTitle"
                            class="rounded bg-gray-900 px-3 py-1.5 text-sm text-white hover:bg-gray-800 disabled:opacity-60"
                            @click="createTask"
                        >
                            <span v-if="saving">Saving…</span><span v-else>Save</span>
                        </button>
                    </template>

                    <!-- Editing actions -->
                    <template v-else-if="isEditing">
                        <button class="rounded px-2 py-1 text-sm hover:bg-gray-100" @click="isEditing = false">
                            Cancel
                        </button>
                        <button
                            :disabled="!editTitle"
                            class="rounded bg-emerald-600 px-3 py-1.5 text-sm text-white hover:bg-emerald-500 disabled:opacity-60"
                            @click="saveEdit"
                        >
                            Save
                        </button>
                    </template>

                    <!-- View mode actions -->
                    <template v-else>
                        <button class="rounded px-2 py-1 text-sm hover:bg-gray-100" @click="startEdit">Edit</button>
                        <button
                            class="rounded bg-red-600 px-3 py-1.5 text-sm text-white hover:bg-red-500"
                            @click="removeTask"
                        >
                            Delete
                        </button>
                    </template>

                    <!-- Close (always) -->
                    <button class="rounded px-2 py-1 text-sm hover:bg-gray-100" @click="$emit('close')">✕</button>
                </div>
            </div>

            <!-- Body -->

            <!-- CREATE -->
            <div v-if="mode === 'create'" class="space-y-3">
                <label class="block text-sm font-medium">Title</label>
                <input
                    v-model.trim="formTitle"
                    class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-emerald-400"
                    placeholder="Short summary"
                    autofocus
                />

                <label class="block text-sm font-medium">Description</label>
                <RichEditor :key="editorKey" v-model="formDesc" />
            </div>

            <!-- NOT CREATE: either EDIT or VIEW -->
            <div v-else>
                <!-- EDIT -->
                <div v-if="isEditing" class="space-y-3">
                    <label class="block text-sm font-medium">Title</label>
                    <input
                        v-model.trim="editTitle"
                        class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-emerald-400"
                        placeholder="Task title"
                        autofocus
                    />

                    <label class="block text-sm font-medium">Description</label>
                    <RichEditor v-model="editDesc" />
                </div>

                <!-- VIEW -->
                <div v-else class="space-y-5">
                    <div>
                        <div class="flex items-start justify-between">
                            <div>
                                <div class="text-xl font-semibold">{{ task?.title }}</div>
                                <div class="mt-2 prose max-w-none">
                                    <SafeHtml :html="task?.description || ''" />
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Comments -->
                    <div class="border-t pt-4">
                        <div class="mb-2 text-sm font-medium text-gray-700">Comments</div>

                        <!-- Collapsed composer (Trello-style) -->
                        <button
                            v-if="!composerOpen"
                            class="w-full rounded-xl border px-4 py-3 text-left text-gray-500 hover:bg-gray-50"
                            @click="openComposer"
                        >
                            Write a comment…
                        </button>

                        <!-- Expanded rich composer -->
                        <div v-else class="space-y-2">
                            <RichEditor v-model="commentHtml" />
                            <div class="flex items-center gap-2">
                                <button
                                    :disabled="savingComment || !hasCommentContent"
                                    class="rounded bg-gray-900 px-3 py-1.5 text-sm text-white hover:bg-gray-800 disabled:opacity-60"
                                    @click="saveComment"
                                >
                                    {{ savingComment ? "Saving…" : "Save" }}
                                </button>
                                <button class="rounded px-2 py-1 text-sm hover:bg-gray-100" @click="cancelComment">
                                    Cancel
                                </button>
                            </div>
                        </div>

                        <!-- Comments list -->
                        <div v-if="loadingComments" class="mt-3 text-sm text-gray-500">Loading…</div>
                        <div v-else class="mt-3 max-h-60 overflow-y-auto pr-1">
                            <ul class="space-y-2">
                                <li v-for="c in comments" :key="c.id" class="rounded border bg-white p-2 text-sm">
                                    <div class="text-gray-800">
                                        <SafeHtml :html="c.body || ''" />
                                    </div>
                                    <div class="mt-1 text-xs text-gray-500">
                                        {{ new Date(c.created_at).toLocaleString() }}
                                    </div>
                                </li>
                                <li v-if="!comments.length" class="text-sm text-gray-500">No comments yet.</li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
            <!-- /NOT CREATE -->
        </div>
    </div>
</template>

<script setup>
import { ref, watch, computed, nextTick } from "vue";
import { api } from "../lib/api";
import RichEditor from "./RichEditor.vue";
import SafeHtml from "./SafeHtml.vue";

const props = defineProps({
    open: { type: Boolean, default: false },
    mode: { type: String, default: "create" }, // 'create' | 'view'
    listId: { type: String, default: "" }, // used in create mode
    task: { type: Object, default: null }, // used in view mode
});

const emit = defineEmits(["close", "created", "commented", "updated", "deleted"]);

// --- create mode state
const formTitle = ref("");
const formDesc = ref("");
const saving = ref(false);
const editorKey = ref("editor-boot");

// --- view mode / comments state
const comments = ref([]);
const loadingComments = ref(false);

// New Trello-style composer state
const composerOpen = ref(false);
const commentHtml = ref("");
const savingComment = ref(false);

const hasCommentContent = computed(() => {
    const html = (commentHtml.value || "").replace(/<br\s*\/?>/gi, "\n");
    const text = html
        .replace(/<[^>]+>/g, "")
        .replace(/&nbsp;/g, " ")
        .trim();
    return text.length > 0;
});

// --- edit mode state
const isEditing = ref(false);
const editTitle = ref("");
const editDesc = ref("");

function initForm() {
    if (props.mode === "create") {
        formTitle.value = "";
        formDesc.value = "";
    } else if (props.task) {
        formTitle.value = props.task.title || "";
        formDesc.value = props.task.description || "";
    }
    // force remount RichEditor so it doesn't keep stale content
    editorKey.value = `${props.task?.id || "new"}-${Date.now()}`;
}

watch(
    () => props.open,
    (o) => {
        if (o) initForm();
    }
);
watch(
    () => props.mode,
    () => initForm(),
    { immediate: true }
);
watch(
    () => props.task?.id,
    () => initForm()
);

// load comments when opening in view mode
watch(
    () => [props.open, props.mode, props.task?.id],
    async ([open, mode]) => {
        if (!open || mode !== "view" || !props.task?.id) return;
        loadingComments.value = true;
        try {
            comments.value = await api.get(`/api/comments?task_id=${encodeURIComponent(props.task.id)}`);
        } catch (e) {
            console.error(e);
        } finally {
            loadingComments.value = false;
        }
    }
);

// also reset edit/composer state on open/close
watch(
    () => props.open,
    (o) => {
        if (o) {
            isEditing.value = false;
            composerOpen.value = false;
            commentHtml.value = "";
        } else {
            isEditing.value = false;
            composerOpen.value = false;
            commentHtml.value = "";
        }
    }
);

// ----- create task
async function createTask() {
    if (!formTitle.value.trim() || !props.listId) return;
    saving.value = true;
    try {
        const payload = {
            list_id: props.listId,
            title: formTitle.value.trim(),
            description: formDesc.value || "",
        };
        const res = await api.post("/api/tasks", payload);
        emit("created", { ...res, description: formDesc.value || "" });
        emit("close");
    } catch (e) {
        alert(e?.message || "Create failed");
    } finally {
        saving.value = false;
    }
}

// ----- comments (Trello-style)
function openComposer() {
    composerOpen.value = true;
    nextTick(() => {
        // editor will autofocus caret itself; no-op here
    });
}
function cancelComment() {
    composerOpen.value = false;
    commentHtml.value = "";
}
async function saveComment() {
    const body = commentHtml.value || "";
    if (!props.task?.id || !hasCommentContent.value) return;
    savingComment.value = true;
    try {
        const c = await api.post("/api/comments", { task_id: props.task.id, body });
        comments.value.unshift(c);
        commentHtml.value = "";
        composerOpen.value = false;
        emit("commented", c); // let Board bump comment_count if it wants
    } catch (e) {
        alert(e?.message || "Failed to comment");
    } finally {
        savingComment.value = false;
    }
}

// ----- edit task
function startEdit() {
    if (!props.task) return;
    isEditing.value = true;
    editTitle.value = props.task.title || "";
    editDesc.value = props.task.description || "";
}
async function saveEdit() {
    if (!props.task?.id) return;
    try {
        const payload = { title: editTitle.value, description: editDesc.value };
        const res = await api.patch(`/api/tasks?id=${encodeURIComponent(props.task.id)}`, payload);
        emit("updated", res);
        isEditing.value = false;
        emit("close");
    } catch (e) {
        alert(e?.message || "Failed to update task");
    }
}
async function removeTask() {
    if (!props.task?.id) return;
    if (!confirm("Are you sure you want to delete this task?")) return;
    try {
        await api.del(`/api/tasks?id=${encodeURIComponent(props.task.id)}`);
        emit("deleted", props.task.id);
        emit("close");
    } catch (e) {
        alert(e?.message || "Failed to delete task");
    }
}
</script>

<style scoped>
/* Small UX helpers */
</style>
