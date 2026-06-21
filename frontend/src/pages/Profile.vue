<template>
  <main>
    <p v-if="loading">Loading profile...</p>
    <p v-if="error">{{ error }}</p>

    <section v-if="profile">
      <h1>
        {{ profile.first_name }} {{ profile.last_name }}
      </h1>

      <p v-if="profile.nickname">
        Nickname: {{ profile.nickname }}
      </p>

      <p>
        Profile:
        <strong>{{ profile.is_public ? "Public" : "Private" }}</strong>
      </p>

      <p>
        Followers: <strong>{{ profile.followers_count }}</strong>
        |
        Following: <strong>{{ profile.following_count }}</strong>
      </p>

      <section v-if="profile.can_view_profile">
        <h2>Followers</h2>

        <p v-if="loadingFollowLists">Loading followers...</p>
        <p v-if="followListsError">{{ followListsError }}</p>

        <p v-if="followers.length === 0">
          No followers yet.
        </p>

        <ul>
          <li v-for="user in followers" :key="user.id">
            <router-link :to="`/profiles/${user.id}`">
              {{ user.first_name }} {{ user.last_name }}
            </router-link>
            <span v-if="user.nickname">
              - {{ user.nickname }}
            </span>
          </li>
        </ul>

        <h2>Following</h2>

        <p v-if="following.length === 0">
          Not following anyone yet.
        </p>

        <ul>
          <li v-for="user in following" :key="user.id">
            <router-link :to="`/profiles/${user.id}`">
              {{ user.first_name }} {{ user.last_name }}
            </router-link>
            <span v-if="user.nickname">
              - {{ user.nickname }}
            </span>
          </li>
        </ul>
      </section>

      <p v-if="profile.can_view_profile && profile.about_me">
        About me: {{ profile.about_me }}
      </p>

      <p v-if="!profile.can_view_profile">
        This profile is private. Follow this user to see more information.
      </p>

      <section v-if="!profile.is_owner">
        <p>
          Follow status:
          <strong>{{ profile.follow_status }}</strong>
        </p>

        <button v-if="profile.follow_status === 'none'" @click="followUser">
          Follow
        </button>

        <button v-else-if="profile.follow_status === 'following'" @click="unfollowUser">
          Unfollow
        </button>

        <button v-else-if="profile.follow_status === 'pending'" @click="unfollowUser">
          Cancel Request
        </button>
      </section>

      <section v-if="profile.is_owner">
        <h2>Edit Profile</h2>

        <form @submit.prevent="updateProfile">
          <div>
            <label for="nickname">Nickname</label>
            <input id="nickname" v-model="editForm.nickname" type="text" />
          </div>

          <div>
            <label for="about_me">About Me</label>
            <textarea id="about_me" v-model="editForm.about_me"></textarea>
          </div>

          <div>
            <label>
              <input v-model="editForm.is_public" type="checkbox" />
              Public profile
            </label>
          </div>

          <button type="submit">Save Profile</button>
        </form>

        <p v-if="updateMessage">{{ updateMessage }}</p>
      </section>

      <hr />

      <section>
        <h2>Posts</h2>

        <p v-if="loadingPosts">Loading posts...</p>
        <p v-if="postsError">{{ postsError }}</p>

        <p v-if="posts.length === 0">
          No posts to show.
        </p>

        <article v-for="post in posts" :key="post.id">
          <h3>{{ post.author_name }}</h3>

          <p>
            Privacy:
            <strong>{{ post.privacy }}</strong>
          </p>

          <p>{{ post.content }}</p>

          <img v-if="post.image_path" :src="imageUrl(post.image_path)" alt="Post image" style="max-width: 300px" />

          <p>{{ post.created_at }}</p>

          <hr />
        </article>
      </section>
    </section>
  </main>
</template>

<script setup>
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiRequest } from "../services/api";

const route = useRoute();

const profile = ref(null);
const posts = ref([]);
const followers = ref([]);
const following = ref([]);

const loading = ref(false);
const loadingPosts = ref(false);
const loadingFollowLists = ref(false);

const error = ref("");
const postsError = ref("");
const followListsError = ref("");

const updateMessage = ref("");

const editForm = ref({
  nickname: "",
  about_me: "",
  is_public: true,
});

function isMyProfileRoute() {
  return route.path === "/profile/me";
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

    editForm.value.nickname = profile.value.nickname || "";
    editForm.value.about_me = profile.value.about_me || "";
    editForm.value.is_public = profile.value.is_public;

    await loadProfilePosts();
    await loadFollowLists();
  } catch (err) {
    error.value = err.message;
  } finally {
    loading.value = false;
  }
}

async function loadProfilePosts() {
  if (!profile.value) return;

  try {
    loadingPosts.value = true;
    postsError.value = "";

    posts.value = await apiRequest(`/profiles/${profile.value.id}/posts`);
  } catch (err) {
    postsError.value = err.message;
  } finally {
    loadingPosts.value = false;
  }
}

async function loadFollowLists() {
  if (!profile.value) return;

  try {
    loadingFollowLists.value = true;
    followListsError.value = "";

    followers.value = await apiRequest(`/users/${profile.value.id}/followers`);
    following.value = await apiRequest(`/users/${profile.value.id}/following`);
  } catch (err) {
    followers.value = [];
    following.value = [];
    followListsError.value = err.message;
  } finally {
    loadingFollowLists.value = false;
  }
}

async function updateProfile() {
  try {
    error.value = "";
    updateMessage.value = "";

    await apiRequest("/profile/me", {
      method: "PUT",
      body: JSON.stringify({
        nickname: editForm.value.nickname,
        about_me: editForm.value.about_me,
        is_public: editForm.value.is_public,
      }),
    });

    updateMessage.value = "Profile updated successfully.";

    await loadProfile();
  } catch (err) {
    error.value = err.message;
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

function imageUrl(path) {
  return `http://localhost:8080${path}`;
}

onMounted(() => {
  loadProfile();
});

watch(
  () => route.fullPath,
  () => {
    loadProfile();
  }
);
</script>