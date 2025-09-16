<template>
    <div v-html="clean" />
</template>

<script setup>
import { computed } from "vue";
import DOMPurify from "dompurify";

const props = defineProps({
    html: { type: String, default: "" },
});

// Make it reactive + allow basic tags/attrs (incl. <img src>)
const options = {
    ALLOWED_TAGS: [
        "p",
        "br",
        "strong",
        "b",
        "em",
        "i",
        "u",
        "ul",
        "ol",
        "li",
        "a",
        "img",
        "code",
        "pre",
        "span",
        "div",
        "h1",
        "h2",
        "h3",
        "blockquote",
    ],
    ALLOWED_ATTR: ["href", "target", "rel", "src", "alt", "title", "class"],
    ALLOW_DATA_ATTR: true, // allow data-* if present
};

const clean = computed(() => DOMPurify.sanitize(props.html || "", options));
</script>
