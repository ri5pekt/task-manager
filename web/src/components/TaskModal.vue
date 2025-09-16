<template>
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40" @click="$emit('close')"></div>

        <div class="relative w-full max-w-2xl rounded-2xl bg-white p-5 shadow-xl">
            <div class="mb-4 flex items-center justify-between">
                <h3 class="text-lg font-semibold">
                    {{ mode === "create" ? "Create task" : "Task details" }}
                </h3>
                <button class="rounded px-2 py-1 text-sm hover:bg-gray-100" @click="$emit('close')">✕</button>
            </div>

            <!-- CREATE -->
            <div v-if="mode === 'create'" class="space-y-3">
                <label class="block text-sm font-medium">Title</label>
                <input
                    v-model.trim="formTitle"
                    class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-emerald-400"
                    placeholder="Short summary"
                />

                <label class="block text-sm font-medium">Description</label>
                <RichEditor v-model="formDesc" />

                <div class="mt-4 flex justify-end gap-2">
                    <button class="rounded px-3 py-1.5 hover:bg-gray-100" @click="$emit('close')">Cancel</button>
                    <button
                        :disabled="saving || !formTitle"
                        class="rounded bg-gray-900 px-3 py-1.5 text-white hover:bg-gray-800 disabled:opacity-60"
                        @click="createTask"
                    >
                        <span v-if="saving">Saving…</span><span v-else>Save</span>
                    </button>
                </div>
            </div>

            <!-- VIEW -->
            <div v-else class="space-y-4">
                <div>
                    <div class="text-xl font-semibold">{{ task?.title }}</div>
                    <div class="mt-2 prose max-w-none">
                        <SafeHtml :html="task?.description || ''" />
                    </div>
                </div>

                <div class="border-t pt-3">
                    <div class="mb-2 text-sm font-medium text-gray-700">Comments</div>

                    <div class="mb-3 flex gap-2">
                        <input
                            v-model="commentBody"
                            placeholder="Add a comment…"
                            class="flex-1 rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-emerald-400"
                        />
                        <button
                            :disabled="savingComment || !commentBody.trim()"
                            class="rounded bg-gray-900 px-3 py-1.5 text-white hover:bg-gray-800 disabled:opacity-60"
                            @click="addComment"
                        >
                            {{ savingComment ? "Posting…" : "Post" }}
                        </button>
                    </div>

                    <div v-if="loadingComments" class="text-sm text-gray-500">Loading…</div>
                    <ul v-else class="space-y-2">
                        <li v-for="c in comments" :key="c.id" class="rounded border bg-white p-2 text-sm">
                            <div class="text-gray-800">{{ c.body }}</div>
                            <div class="mt-1 text-xs text-gray-500">{{ new Date(c.created_at).toLocaleString() }}</div>
                        </li>
                        <li v-if="!comments.length" class="text-sm text-gray-500">No comments yet.</li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted } from "vue";
import { api } from "../lib/api";
import RichEditor from "./RichEditor.vue";
import SafeHtml from "./SafeHtml.vue";

const props = defineProps({
    open: { type: Boolean, default: false },
    mode: { type: String, default: "create" }, // 'create' | 'view'
    listId: { type: String, default: "" }, // used in create mode
    task: { type: Object, default: null }, // used in view mode
});

const emit = defineEmits(["close", "created", "commented"]);

// --- create mode state
const formTitle = ref("");
const formDesc = ref("");
const saving = ref(false);

// --- view mode state
const comments = ref([]);
const loadingComments = ref(false);
const commentBody = ref("");
const savingComment = ref(false);

// load comments when opening in view mode
watch(
    () => [props.open, props.mode, props.task?.id],
    async ([open, mode]) => {
        if (!open || mode !== "view" || !props.task?.id) return;
        loadingComments.value = true;
        try {
            console.log("[TaskModal] open view for task:", props.task?.id, {
                title: props.task?.title,
                hasDescription: !!props.task?.description,
                descriptionPreview: (props.task?.description || "").slice(0, 80),
            });
            comments.value = await api.get(`/api/comments?task_id=${encodeURIComponent(props.task.id)}`);
        } catch (e) {
            console.error(e);
        } finally {
            loadingComments.value = false;
        }
    }
);

// create task
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
        console.log("[TaskModal] createTask response:", res);
        console.log("[TaskModal] createTask payload:", payload);
        emit("created", { ...res, description: formDesc.value || "" });
        emit("close");
    } catch (e) {
        alert(e?.message || "Create failed");
    } finally {
        saving.value = false;
    }
}

// add comment
async function addComment() {
    const body = commentBody.value.trim();
    if (!body || !props.task?.id) return;
    savingComment.value = true;
    try {
        const c = await api.post("/api/comments", { task_id: props.task.id, body });
        comments.value.unshift(c);
        commentBody.value = "";
        emit("commented", c); // let Board bump comment_count if it wants
    } catch (e) {
        alert(e?.message || "Failed to comment");
    } finally {
        savingComment.value = false;
    }
}
</script>
