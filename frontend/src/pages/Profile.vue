<template>
  <main class="profile-page">
    <p v-if="loading" class="status-message">Loading profile...</p>

    <p v-if="error" class="error-message">
      {{ error }}
    </p>

    <section v-if="profile" class="profile-container">
      <div v-if="updateMessage" class="profile-success-message" role="status">
        <span>
          {{ updateMessage }}
        </span>

        <button
          type="button"
          class="profile-success-dismiss"
          aria-label="Dismiss profile update message"
          @click="updateMessage = ''"
        >
          <i class="fa-solid fa-xmark" aria-hidden="true"></i>
        </button>
      </div>

      <ProfileHeader
        :profile="profile"
        @follow="followUser"
        @unfollow="unfollowUser"
        @edit="openEditModal"
      />

      <ProfileConnections v-if="profile.can_view_profile" :profile="profile" />

      <ProfileActivity v-if="profile.can_view_profile" :profile="profile" />
    </section>
  </main>
  <EditProfileModal
    v-if="editModalOpen"
    :profile="profile"
    @close="editModalOpen = false"
    @updated="handleProfileUpdated"
  />
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiRequest } from "../services/api";
import { useAuthStore } from "../stores/auth";

import ProfileHeader from "../components/profile/ProfileHeader.vue";
import ProfileConnections from "../components/profile/ProfileConnections.vue";
import ProfileActivity from "../components/profile/ProfileActivity.vue";
import EditProfileModal from "../components/profile/EditProfileModal.vue";

const route = useRoute();
const auth = useAuthStore();

const profile = ref(null);
const loading = ref(false);
const error = ref("");

const updateMessage = ref("");
const editModalOpen = ref(false);

function openEditModal() {
  updateMessage.value = "";
  editModalOpen.value = true;
}

function isMyProfileRoute() {
  return route.path === "/profile/me";
}

async function handleProfileUpdated() {
  editModalOpen.value = false;

  await loadProfile();

  if (isMyProfileRoute() && auth.user && profile.value) {
    Object.assign(auth.user, {
      email: profile.value.email,
      first_name: profile.value.first_name,
      last_name: profile.value.last_name,
      nickname: profile.value.nickname,
      avatar_path: profile.value.avatar_path,
    });
  }

  updateMessage.value = "Profile updated successfully";
}

function profileApiPath() {
  if (isMyProfileRoute()) {
    return "/profile/me";
  }

  return `/profiles/${route.params.id}`;
}

async function loadProfile() {
  try {
    loading.value = true;
    error.value = "";
    updateMessage.value = "";

    profile.value = await apiRequest(profileApiPath());
  } catch (err) {
    error.value = err.message;
  } finally {
    loading.value = false;
  }
}

async function followUser() {
  if (!profile.value) return;

  try {
    error.value = "";

    await apiRequest(`/users/${profile.value.id}/follow`, {
      method: "POST",
    });

    await loadProfile();
  } catch (err) {
    error.value = err.message;
  }
}

async function unfollowUser() {
  if (!profile.value) return;

  try {
    error.value = "";

    await apiRequest(`/users/${profile.value.id}/unfollow`, {
      method: "POST",
    });

    await loadProfile();
  } catch (err) {
    error.value = err.message;
  }
}

onMounted(() => {
  loadProfile();
});

watch(
  () => route.fullPath,
  () => {
    loadProfile();
  },
);
</script>

<style scoped>
.profile-page {
  width: 100%;
  padding: 28px 20px 60px;
}

.profile-container {
  width: min(1100px, 100%);
  margin: 0 auto;
}

/* GENERAL MESSAGES */

.status-message,
.error-message {
  width: min(1100px, 100%);

  margin: 20px auto;
}

.error-message {
  color: #f87171;
}

/* PROFILE UPDATE MESSAGE */

.profile-success-message {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  margin: 0 0 16px;

  padding: 10px 12px 10px 16px;

  border: 1px solid rgba(74, 222, 128, 0.2);

  border-radius: 10px;

  color: #86efac;

  background: rgba(74, 222, 128, 0.07);

  font-size: 14px;
}

.profile-success-dismiss {
  flex-shrink: 0;

  width: 30px;
  height: 30px;
  min-height: 0;

  padding: 0;

  border: none;
  border-radius: 50%;

  color: #86efac;
  background: transparent;

  font-size: 21px;
  line-height: 1;

  cursor: pointer;
}

.profile-success-dismiss:hover {
  color: #dcfce7;

  background: rgba(74, 222, 128, 0.1);
}

@media (max-width: 700px) {
  .profile-page {
    padding: 16px 12px 40px;
  }

}
</style>
