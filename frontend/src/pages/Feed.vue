<template>
  <main>
    <h1>Feed</h1>

    <section v-if="auth.user">
      <h2>Create Post</h2>

      <form @submit.prevent="createPost">
        <div>
          <label>Post Content</label>
          <textarea v-model="newPostContent"></textarea>
        </div>

        <div>
          <label>Visibility</label>
          <select v-model="newPostPrivacy">
            <option value="public">Public</option>
            <option value="followers">Followers Only</option>
            <option value="private">Private</option>
          </select>
        </div>

        <button type="submit">Post</button>
      </form>

      <p v-if="postError">{{ postError }}</p>
    </section>

    <section v-else>
      <p>You need to login before creating posts.</p>
    </section>

    <hr />

    <section>
      <h2>Posts</h2>

      <p v-if="loading">Loading posts...</p>
      <p v-if="loadError">{{ loadError }}</p>

      <article v-for="post in posts" :key="post.id">
        <h3>{{ post.author_name || post.author }}</h3>

        <p v-if="post.author_nickname">
          Nickname: {{ post.author_nickname }}
        </p>

        <p>{{ post.content }}</p>

        <small>
          Privacy: {{ post.privacy }} |
          Created at: {{ post.created_at }}
        </small>

        <section>
          <h4>Comments</h4>

          <p v-if="loadingComments[post.id]">Loading comments...</p>
          <p v-if="commentErrors[post.id]">{{ commentErrors[post.id] }}</p>

          <div
            v-for="comment in commentsByPost[post.id] || []"
            :key="comment.id"
          >
            <strong>{{ comment.author_name }}</strong>
            <p>{{ comment.content }}</p>
            <small>{{ comment.created_at }}</small>
            <hr />
          </div>

          <p v-if="(commentsByPost[post.id] || []).length === 0">
            No comments yet.
          </p>

          <form v-if="auth.user" @submit.prevent="createComment(post.id)">
            <input
              v-model="newComments[post.id]"
              type="text"
              placeholder="Write a comment..."
            />

            <button type="submit">Comment</button>
          </form>
        </section>

        <hr />
      </article>
    </section>

  </main>
</template>

<script setup>
import { ref, watch } from "vue";
import { useAuthStore } from "../stores/auth";
import { apiRequest } from "../services/api";

const auth = useAuthStore();
// Posts
const posts = ref([]);
const loading = ref(false);
const loadError = ref("");
const postError = ref("");
// Create Post
const newPostContent = ref("");
const newPostPrivacy = ref("public");
// Create Comments
const commentsByPost = ref({});
const newComments = ref({});
const commentErrors = ref({});
const loadingComments = ref({});

async function loadPosts() {
  try {
    loading.value = true;
    loadError.value = "";

    posts.value = await apiRequest("/posts");

    for (const post of posts.value) {
      await loadComments(post.id);
    }
  } catch (err) {
    loadError.value = err.message;
  } finally {
    loading.value = false;
  }
}

async function createPost() {
  try {
    postError.value = "";

    await apiRequest("/posts", {
      method: "POST",
      body: JSON.stringify({
        content: newPostContent.value,
        privacy: newPostPrivacy.value,
      }),
    });

    newPostContent.value = "";
    newPostPrivacy.value = "public";

    await loadPosts();
  } catch (err) {
    postError.value = err.message;
  }
}

async function loadComments(postId) {
  try {
    loadingComments.value[postId] = true;
    commentErrors.value[postId] = "";

    commentsByPost.value[postId] = await apiRequest(
      `/posts/${postId}/comments`
    );
  } catch (err) {
    commentErrors.value[postId] = err.message;
  } finally {
    loadingComments.value[postId] = false;
  }
}

async function createComment(postId) {
  try {
    commentErrors.value[postId] = "";

    const content = newComments.value[postId] || "";

    await apiRequest(`/posts/${postId}/comments`, {
      method: "POST",
      body: JSON.stringify({
        content,
      }),
    });

    newComments.value[postId] = "";

    await loadComments(postId);
  } catch (err) {
    commentErrors.value[postId] = err.message;
  }
}

function clearFeed() {
  posts.value = [];
  commentsByPost.value = {};
  newComments.value = {};
  commentErrors.value = {};
  loadingComments.value = {};
  loadError.value = "";
  postError.value = "";
}

watch(
  () => auth.user,
  async (user) => {
    if (user) {
      await loadPosts();
    } else {
      clearFeed();
    }
  },
  { immediate: true }
);

</script>
