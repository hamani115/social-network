import { createRouter, createWebHistory } from "vue-router";
import Login from "../pages/Login.vue";
import Register from "../pages/Register.vue";
import Feed from "../pages/Feed.vue";
import Users from "../pages/Users.vue";
import Profile from "../pages/Profile.vue";
import Notifications from "../pages/Notifications.vue";
import Groups from "../pages/Groups.vue";
import GroupDetail from "../pages/GroupDetail.vue";
import Chat from "../pages/Chat.vue";
import NotFound from "../pages/NotFound.vue";
import ServerUnavailable from "../pages/ServerUnavailable.vue";
import GenericError from "../pages/GenericError.vue";
import { isServerUnavailableError } from "../services/api";
import { useAuthStore } from "../stores/auth";
const routes = [
  {
    path: "/",
    component: Feed,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/login",
    component: Login,
    meta: {
      guestOnly: true,
    },
  },
  {
    path: "/register",
    component: Register,
    meta: {
      guestOnly: true,
    },
  },
  {
    path: "/users",
    component: Users,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/profile/me",
    component: Profile,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/profiles/:id",
    component: Profile,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/notifications",
    component: Notifications,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/groups",
    component: Groups,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/groups/:id",
    component: GroupDetail,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/chat",
    component: Chat,
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: "/server-unavailable",
    component: ServerUnavailable,
  },
  {
    path: "/error",
    component: GenericError,
  },
  {
    path: "/:pathMatch(.*)*",
    component: NotFound,
  },
];
const router = createRouter({
  history: createWebHistory(),

  routes,
});
let sessionChecked = false;
router.beforeEach(async (to) => {
  const auth = useAuthStore();
  const needsSession = Boolean(to.meta.requiresAuth || to.meta.guestOnly);
  if (needsSession && !sessionChecked) {
    try {
      await auth.fetchMe();
      sessionChecked = true;
    } catch (err) {
      if (isServerUnavailableError(err)) {
        return {
          path: "/server-unavailable",

          query: {
            from: to.fullPath,
          },
        };
      }
      if (err?.status >= 500) {
        return {
          path: "/error",

          query: {
            from: to.fullPath,
          },
        };
      }
      auth.user = null;
      sessionChecked = true;
    }
  }
  if (to.meta.requiresAuth && !auth.user) {
    return {
      path: "/login",

      query: {
        redirect: to.fullPath,
      },
    };
  }
  if (to.meta.guestOnly && auth.user) {
    return "/";
  }
  return true;
});
export default router;
