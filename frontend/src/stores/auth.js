import { defineStore } from "pinia";
import { apiRequest } from "../services/api";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null,
  }),

  actions: {
    async register(form) {
      return apiRequest("/register", {
        method: "POST",
        body: JSON.stringify(form),
      });
    },

    async login(email, password) {
      await apiRequest("/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });

      await this.fetchMe();
    },

    // !TEST THOROUGLY
    async fetchMe() {
      this.user = await apiRequest("/me");
    },

    async logout() {
      await apiRequest("/logout", {
        method: "POST",
      });

      this.user = null;
    },
  },
});