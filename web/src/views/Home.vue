<template>
    <div class="p-6">
        <h1 class="mb-2 text-xl font-semibold">Home</h1>
        <p v-if="auth.isAuthed">
            Logged in as <strong>{{ auth.user.email }}</strong>
        </p>
        <p v-else>Not logged in.</p>
    </div>
</template>

<script setup>
import { useAuth } from "../stores/auth";
const auth = useAuth();
</script>

<style scoped></style>
