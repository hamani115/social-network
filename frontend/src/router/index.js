import { createRouter, createWebHistory } from "vue-router";

import Login from "../pages/Login.vue";
import Register from "../pages/Register.vue";
import Feed from "../pages/Feed.vue";

const routes = [
  {
    path: "/",
    component: Feed,
  },
  {
    path: "/login",
    component: Login,
  },
  {
    path: "/register",
    component: Register,
  },
];


export default createRouter({
  history: createWebHistory(),
  routes,
});