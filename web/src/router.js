// web/src/router.js
import { createRouter, createWebHistory } from "vue-router";
import Home from "./views/Home.vue";

export default createRouter({
    history: createWebHistory(),
    routes: [
        { path: "/", name: "home", component: Home },
        { path: "/login", name: "login", component: () => import("./views/Login.vue") },
        { path: "/board", name: "board", component: () => import("./views/Board.vue") },
        { path: "/register", name: "register", component: () => import("./views/Register.vue") },
    ],
});
