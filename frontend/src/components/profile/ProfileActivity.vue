<template>
  <section class="info-card activity-card">
    <div class="card-heading activity-heading">
      <div>
        <h2>Activity</h2>

        <p class="card-subtitle">
          {{ posts.length }}
          {{ posts.length === 1 ? "post" : "posts" }}
        </p>
      </div>
    </div>

    <p v-if="loadingPosts" class="activity-status">Loading activity...</p>

    <p v-else-if="postsError" class="activity-error">
      {{ postsError }}
    </p>

    <p v-else-if="posts.length === 0" class="empty-text">
      No posts to show yet.
    </p>

    <div v-else class="activity-list">
      <article v-for="post in posts" :key="post.id" class="activity-post">
        <div class="post-author-row">
          <router-link
            :to="`/profiles/${profile.id}`"
            class="post-author-avatar-link"
          >
            <img
              v-if="profile.avatar_path"
              :src="profile.avatar_path"
              :alt="`${post.author_name}'s avatar`"
              class="post-author-avatar"
            />

            <div v-else class="post-author-avatar post-author-placeholder">
              {{ userInitials(profile) }}
            </div>
          </router-link>

          <div class="post-author-info">
            <router-link
              :to="`/profiles/${profile.id}`"
              class="post-author-name"
            >
              {{ post.author_name }}
            </router-link>

            <span v-if="post.author_nickname" class="post-author-nickname">
              {{ post.author_nickname }}
            </span>

            <div class="post-meta">
              <span>
                {{ formatDate(post.created_at) }}
              </span>

              <span class="post-meta-divider"> - </span>

              <span>
                {{ privacyLabel(post.privacy) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Post content -->
        <p class="post-content">
          {{ post.content }}
        </p>

        <div v-if="post.image_path" class="post-image-wrapper">
          <img :src="post.image_path" alt="Post image" class="post-image" />
        </div>
      </article>
    </div>
  </section>
</template>

<script setup>
import { ref, watch } from "vue";

import { apiRequest } from "../../services/api";
import { formatDate } from "../../utils/date";

const props = defineProps({
  profile: {
    type: Object,
    required: true,
  },
});

const posts = ref([]);
const loadingPosts = ref(false);
const postsError = ref("");

async function loadProfilePosts() {
  if (!props.profile?.id) {
    return;
  }

  try {
    loadingPosts.value = true;
    postsError.value = "";

    posts.value = await apiRequest(`/profiles/${props.profile.id}/posts`);
  } catch (err) {
    posts.value = [];
    postsError.value = err.message;
  } finally {
    loadingPosts.value = false;
  }
}

function userInitials(user) {
  const firstInitial = user.first_name?.charAt(0) || "";

  const lastInitial = user.last_name?.charAt(0) || "";

  return (firstInitial + lastInitial).toUpperCase();
}

function privacyLabel(privacy) {
  switch (privacy) {
    case "public":
      return "Public";

    case "followers":
      return "Followers";

    case "private":
      return "Selected followers";

    default:
      return privacy;
  }
}

watch(
  () => props.profile.id,
  () => {
    loadProfilePosts();
  },
  {
    immediate: true,
  },
);
</script>

<style scoped>
.info-card {
  margin-top: 20px;
  padding: 26px 30px;

  border: 1px solid rgba(255, 255, 255, 0.08);

  border-radius: 16px;

  background: rgba(20, 25, 35, 0.95);

  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
}

.card-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  margin-bottom: 20px;
}

.card-heading h2 {
  margin: 0;

  color: #f8fafc;

  font-size: 22px;
}

.card-subtitle {
  margin: 5px 0 0;

  color: #64748b;

  font-size: 13px;
}

.empty-text {
  margin: 0 0 24px;

  color: #64748b;

  font-style: italic;
}

.activity-card {
  padding-bottom: 8px;
}

.activity-heading {
  margin-bottom: 8px;
}

.activity-list {
  max-height: 850px;

  margin: 0 -30px;

  overflow-y: auto;

  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.3) transparent;
}

.activity-post {
  margin: 0;
  padding: 22px 30px 26px;

  border: 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);

  border-radius: 0;

  background: transparent;
}

.activity-post:first-child {
  border-top: none;
}

.post-author-row {
  display: flex;
  align-items: flex-start;

  gap: 12px;
}

.post-author-avatar-link {
  flex-shrink: 0;

  text-decoration: none;
}

.post-author-avatar {
  width: 48px;
  height: 48px;

  border-radius: 50%;

  object-fit: cover;

  background: #1f2937;
}

.post-author-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;

  color: #eaf3ff;

  background: linear-gradient(135deg, #4f9cff, #2867d6);

  font-size: 14px;
  font-weight: 700;
}

.post-author-info {
  min-width: 0;

  display: flex;
  flex-direction: column;
}

.post-author-name {
  width: fit-content;

  color: #f1f5f9;

  font-size: 15px;
  font-weight: 700;

  text-decoration: none;
}

.post-author-name:hover {
  color: #80b7ff;

  text-decoration: underline;
}

.post-author-nickname {
  margin-top: 2px;

  color: #94a3b8;

  font-size: 13px;
}

.post-meta {
  display: flex;
  align-items: center;

  flex-wrap: wrap;

  gap: 6px;

  margin-top: 4px;

  color: #64748b;

  font-size: 12px;
}

.post-meta-divider {
  font-size: 10px;
}

.post-content {
  margin: 18px 0 0;

  color: #dbe3ee;

  font-size: 15px;
  line-height: 1.65;

  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.post-image-wrapper {
  overflow: hidden;

  margin-top: 18px;

  border: 1px solid rgba(255, 255, 255, 0.06);

  border-radius: 12px;

  background: #0d1117;
}

.post-image {
  width: 100%;
  max-height: 600px;

  display: block;

  object-fit: contain;

  background: #0d1117;
}
.activity-status {
  padding-bottom: 18px;

  color: #94a3b8;
}

.activity-error {
  padding-bottom: 18px;

  color: #f87171;
}

.activity-list::-webkit-scrollbar {
  width: 8px;
}

.activity-list::-webkit-scrollbar-track {
  background: transparent;
}

.activity-list::-webkit-scrollbar-thumb {
  border-radius: 999px;

  background: rgba(148, 163, 184, 0.25);
}

.activity-list::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.4);
}

@media (max-width: 700px) {
  .activity-list {
    max-height: none;
    overflow-y: visible;
  }
}
</style>
