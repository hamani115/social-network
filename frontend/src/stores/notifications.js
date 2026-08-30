import { defineStore } from "pinia";
import { apiRequest } from "../services/api";

export const useNotificationsStore = defineStore("notifications", {
  state: () => ({
    notifications: [],

    loading: false,
    loadingMore: false,

    error: "",
    loadMoreError: "",

    filter: "all",

    limit: 20,
    offset: 0,

    hasMore: true,

    unreadCount: 0,

    initialized: false,

    requestVersion: 0,
  }),

  actions: {
    async fetchNotifications(reset = true) {
      if (!reset && (this.loading || this.loadingMore || !this.hasMore)) {
        return;
      }

      if (reset) {
        this.requestVersion += 1;

        this.notifications = [];

        this.offset = 0;

        this.hasMore = true;

        this.error = "";

        this.loadMoreError = "";
      }

      const requestVersion = this.requestVersion;

      const initialLoad = this.offset === 0;

      try {
        if (initialLoad) {
          this.loading = true;

          this.error = "";
        } else {
          this.loadingMore = true;

          this.loadMoreError = "";
        }

        const params = new URLSearchParams();

        params.set("limit", String(this.limit));

        params.set("offset", String(this.offset));

        params.set("filter", this.filter);

        const result = await apiRequest(`/notifications?${params.toString()}`);

        if (requestVersion !== this.requestVersion) {
          return;
        }

        const incoming = result.notifications || [];

        if (reset) {
          this.notifications = incoming;
        } else {
          const existingIDs = new Set(
            this.notifications.map((notification) => notification.id),
          );

          this.notifications.push(
            ...incoming.filter(
              (notification) => !existingIDs.has(notification.id),
            ),
          );
        }

        this.offset = result.next_offset ?? this.offset + incoming.length;

        this.hasMore = Boolean(result.has_more);

        this.unreadCount = Number(result.unread_count || 0);

        this.initialized = true;
      } catch (err) {
        if (requestVersion !== this.requestVersion) {
          return;
        }

        if (initialLoad) {
          this.error = err.message;
        } else {
          this.loadMoreError = err.message;
        }
      } finally {
        if (requestVersion === this.requestVersion) {
          this.loading = false;

          this.loadingMore = false;
        }
      }
    },

    async setFilter(filter) {
      if (filter !== "all" && filter !== "unread") {
        return;
      }

      if (this.filter === filter && this.initialized) {
        return;
      }

      this.filter = filter;

      await this.fetchNotifications(true);
    },

    async markAsRead(notificationId) {
      const notification = this.notifications.find(
        (item) => item.id === notificationId,
      );

      if (!notification || notification.is_read) {
        return;
      }

      try {
        this.error = "";

        await apiRequest(`/notifications/${notificationId}/read`, {
          method: "POST",
        });

        notification.is_read = true;

        this.unreadCount = Math.max(0, this.unreadCount - 1);

        if (this.filter === "unread") {
          this.notifications = this.notifications.filter(
            (item) => item.id !== notificationId,
          );

          this.offset = Math.max(0, this.offset - 1);
        }
      } catch (err) {
        this.error = err.message;

        throw err;
      }
    },

    async markAllAsRead() {
      if (this.unreadCount === 0) {
        return;
      }

      try {
        this.error = "";

        await apiRequest("/notifications/read-all", {
          method: "POST",
        });

        this.unreadCount = 0;

        if (this.filter === "unread") {
          this.notifications = [];

          this.offset = 0;

          this.hasMore = false;

          return;
        }

        this.notifications = this.notifications.map((notification) => ({
          ...notification,

          is_read: true,
        }));
      } catch (err) {
        this.error = err.message;

        throw err;
      }
    },
    addNotification(notification) {
      if (!notification) {
        return;
      }

      const alreadyExists = this.notifications.some(
        (existing) => existing.id === notification.id,
      );

      if (alreadyExists) {
        return;
      }

      if (!notification.is_read) {
        this.unreadCount += 1;
      }

      if (this.filter === "all" || !notification.is_read) {
        this.notifications.unshift(notification);

        if (this.initialized) {
          this.offset += 1;
        }
      }
    },

    // LOGOUT
    clear() {
      this.requestVersion += 1;

      this.notifications = [];

      this.loading = false;
      this.loadingMore = false;

      this.error = "";
      this.loadMoreError = "";

      this.filter = "all";

      this.offset = 0;

      this.hasMore = true;

      this.unreadCount = 0;

      this.initialized = false;
    },
  },
});
