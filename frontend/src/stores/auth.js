import { defineStore } from "pinia";
import { apiRequest } from "../services/api";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null,
  }),

  actions: {
    async register(form, avatar) {
      const formData = new FormData();
      formData.append("email", form.email);
      formData.append("password", form.password);
      formData.append("first_name", form.first_name);
      formData.append("last_name", form.last_name);
      formData.append("date_of_birth", form.date_of_birth);
      formData.append("nickname", form.nickname);
      formData.append("about_me", form.about_me);

      if (avatar) {
        formData.append("avatar", avatar);
      }

      return await apiRequest("/register", {
        method: "POST",
        body: formData,
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
