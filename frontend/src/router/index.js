import { createRouter, createWebHistory } from "vue-router";

import Login from "../pages/Login.vue";
import Register from "../pages/Register.vue";
import Feed from "../pages/Feed.vue";
import Users from "../pages/Users.vue";
import Profile from "../pages/Profile.vue";
import Notifications from "../pages/Notifications.vue";
import Groups from "../pages/Groups.vue";
import GroupDetail from "../pages/GroupDetail.vue";

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
  {
    path: "/users",
    component: Users,
  },
  {
    path: "/profile/me",
    component: Profile,
  },
  {
    path: "/profiles/:id",
    component: Profile,
  },
  {
    path: "/notifications",
    component: Notifications,
  },
  {
    path: "/groups",
    component: Groups,
  },
  {
    path: "/groups/:id",
    component: GroupDetail,
  },
];


export default createRouter({
  history: createWebHistory(),
  routes,
});