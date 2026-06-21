import { defineStore } from "pinia";
import { apiRequest } from "../services/api";

export const useNotificationsStore = defineStore("notifications", {
  state: () => ({
    notifications: [],
    loading: false,
    error: "",
  }),

  getters: {
    unreadCount: (state) => {
      return state.notifications.filter((notification) => !notification.is_read)
        .length;
    },
  },

  actions: {
    async fetchNotifications() {
      try {
        this.loading = true;
        this.error = "";

        this.notifications = await apiRequest("/notifications");
      } catch (err) {
        this.error = err.message;
      } finally {
        this.loading = false;
      }
    },

    async markAsRead(notificationId) {
      await apiRequest(`/notifications/${notificationId}/read`, {
        method: "POST",
      });

      await this.fetchNotifications();
    },

    async markAllAsRead() {
      await apiRequest("/notifications/read-all", {
        method: "POST",
      });

      await this.fetchNotifications();
    },

    clear() {
      this.notifications = [];
      this.loading = false;
      this.error = "";
    },
  },
});