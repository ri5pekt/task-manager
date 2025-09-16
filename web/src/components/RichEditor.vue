<template>
    <Editor v-model="inner" :init="init" @blur="emitUpdate" />
</template>

<script setup>
import { ref, watch } from "vue";
import Editor from "@hugerte/hugerte-vue"; // official Vue wrapper (default export) :contentReference[oaicite:1]{index=1}

/* ===== HugeRTE bundling (per docs) ===== */
// IMPORTANT: the default export must be imported before other HugeRTE imports
import hugerte from "hugerte"; // global must come first
import "hugerte/models/dom"; // DOM model
import "hugerte/icons/default"; // icons
import "hugerte/themes/silver"; // theme
// skins + content CSS (so the editor actually renders styled)
import "hugerte/skins/ui/oxide/skin.js";
import "hugerte/skins/ui/oxide/content.js";
import "hugerte/skins/content/default/content.js";
// plugins you use (paste is core—don't import it)
import "hugerte/plugins/lists";
import "hugerte/plugins/link";
import "hugerte/plugins/image";
import "hugerte/plugins/code";
/* ======================================= */
// Docs show this exact bundling approach + the need to set skin_url/content_css="default". :contentReference[oaicite:2]{index=2}

const props = defineProps({ modelValue: { type: String, default: "" }, height: { type: Number, default: 280 } });
const emit = defineEmits(["update:modelValue"]);
const inner = ref(props.modelValue);

watch(inner, () => emit("update:modelValue", inner.value || ""), { flush: "post" });

watch(
    () => props.modelValue,
    (v) => {
        if (v !== inner.value) inner.value = v;
    }
);
function emitUpdate() {
    emit("update:modelValue", inner.value || "");
}

// Upload handler to your /api/uploads endpoint (reuses your CSRF/cookies)
async function images_upload_handler(blobInfo) {
    const fd = new FormData();
    fd.append("file", blobInfo.blob(), blobInfo.filename());
    const m = document.cookie.match(/(?:^|;\s*)csrf=([^;]+)/);
    const csrf = m ? decodeURIComponent(m[1]) : "";
    const res = await fetch("/api/uploads", {
        method: "POST",
        credentials: "include",
        headers: csrf ? { "X-CSRF-Token": csrf } : {},
        body: fd,
    });
    if (!res.ok) throw new Error("upload failed");
    const json = await res.json();

    const absUrl = json.url.startsWith("http") ? json.url : new URL(json.url, window.location.origin).toString();
    return absUrl;
}

const init = {
    height: props.height,
    menubar: false,
    // DO NOT list "paste" (it’s core in Tiny6/HugeRTE; listing it makes the loader look for a plugin file)
    plugins: "lists link image code",
    toolbar: "undo redo | bold italic underline | bullist numlist | link image | code",
    paste_data_images: true,
    automatic_uploads: true,
    images_upload_handler,
    // tell HugeRTE to use the bundled skin/content we imported above
    skin_url: "default",
    content_css: "default",
    branding: false,
    // optional: keep URLs as-is for your /uploads paths
    convert_urls: false,
};

// needed to keep tree-shakers happy; reference prevents import elision
void hugerte;
</script>
