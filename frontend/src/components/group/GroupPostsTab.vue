<template>
  <section v-show="active" class="group-tab-panel group-posts-panel">
    

    <div class="group-panel-heading">
      <div>
        <h2>Posts</h2>
      </div>

      <div class="group-posts-toolbar">
        <select
          id="group-post-sort"
          v-model="groupPostSort"
          class="group-post-sort"
          :disabled="loadingGroupPosts || loadingMoreGroupPosts"
          @change="changeGroupPostSort"
        >
          <option value="newest">Newest</option>

          <option value="oldest">Oldest</option>
        </select>

        <button
          type="button"
          class="button-primary"
          @click="openGroupPostModal"
        >
          <i class="fa-solid fa-plus" aria-hidden="true"></i>

          Create post
        </button>
      </div>
    </div>

    <!-- Loading -->

    <div v-if="loadingGroupPosts" class="group-section-state">
      <span class="loading-spinner"></span>

      Loading group posts...
    </div>

    

    <p v-else-if="groupPostsError" class="group-page-error">
      {{ groupPostsError }}
    </p>

    

    <div v-else-if="groupPosts.length === 0" class="group-empty-posts">
      <h3>No posts yet</h3>
    </div>

    <!-- Posts -->

    <div v-else class="group-post-list">
      <article
        v-for="post in groupPosts"
        :key="post.id"
        class="group-post-card"
      >
        <header class="group-post-author">
          <UserAvatar
            :avatar-path="post.author_avatar_path"
            :name="post.author_name"
            class="group-post-avatar"
          />

          <div class="group-post-author-info">
            <strong>
              {{ post.author_name }}
            </strong>

            <span v-if="post.author_nickname">
              {{ post.author_nickname }}
            </span>

            <small>
              {{ formatDateTime(post.created_at) }}
            </small>
          </div>
        </header>

        <!-- Content -->

        <p class="group-post-content">
          {{ post.content }}
        </p>

        <div v-if="post.image_path" class="group-post-image-wrapper">
          <img
            :src="post.image_path"
            alt="Group post image"
            class="group-post-image"
          />
        </div>

        <!-- Comments -->

        <button
          type="button"
          class="group-comments-toggle"
          :aria-expanded="Boolean(openGroupComments[post.id])"
          @click="toggleGroupComments(post.id)"
        >
          <span>Comments</span>

          <span aria-hidden="true">
            <i
              :class="
                openGroupComments[post.id]
                  ? 'fa-solid fa-chevron-up'
                  : 'fa-solid fa-chevron-down'
              "
            ></i>
          </span>
        </button>

        <section v-if="openGroupComments[post.id]" class="group-post-comments">
          <div v-if="loadingGroupComments[post.id]" class="group-comment-state">
            <span class="loading-spinner"></span>

            Loading comments...
          </div>

          <p v-if="groupCommentErrors[post.id]" class="group-page-error">
            {{ groupCommentErrors[post.id] }}
          </p>

          <button
            v-if="
              !loadingGroupComments[post.id] && groupCommentsHasMore[post.id]
            "
            type="button"
            class="group-earlier-comments"
            :disabled="loadingEarlierGroupComments[post.id]"
            @click="loadEarlierGroupComments(post.id)"
          >
            {{
              loadingEarlierGroupComments[post.id]
                ? "Loading..."
                : "View earlier comments"
            }}
          </button>

          <p
            v-if="
              !loadingGroupComments[post.id] &&
              (groupCommentsByPost[post.id] || []).length === 0
            "
            class="group-comment-empty"
          >
            No comments yet
          </p>

          <div class="group-comment-list">
            <article
              v-for="comment in groupCommentsByPost[post.id] || []"
              :key="comment.id"
              class="group-comment-card"
            >
              <div class="group-comment-layout">
                <UserAvatar
                  :avatar-path="comment.author_avatar_path"
                  :name="comment.author_name"
                  class="group-comment-avatar"
                />

                <div class="group-comment-body">
                  <div class="group-comment-header">
                    <strong>
                      {{ comment.author_name }}
                    </strong>

                    <span v-if="comment.author_nickname">
                      {{ comment.author_nickname }}
                    </span>

                    <small>
                      {{ formatDateTime(comment.created_at) }}
                    </small>
                  </div>

                  <p>
                    {{ comment.content }}
                  </p>

                  <img
                    v-if="comment.image_path"
                    :src="imageUrl(comment.image_path)"
                    alt="Group comment image"
                  />
                </div>
              </div>
            </article>
          </div>


          <form
            class="group-comment-form"
            @submit.prevent="createGroupComment(post.id)"
          >
            <input
              v-model="newGroupComments[post.id]"
              type="text"
              placeholder="Write a comment..."
              required
            />

            <label
              :for="`group-comment-image-${post.id}`"
              class="button button-ghost group-comment-image-button"
            >
              Image
            </label>

            <input
              :id="`group-comment-image-${post.id}`"
              :ref="
                (element) => {
                  if (element) {
                    groupCommentImageInputs[post.id] = element;
                  }
                }
              "
              class="visually-hidden"
              type="file"
              accept="
                image/png,
                image/jpeg,
                image/gif
              "
              @change="handleGroupCommentImageChange(post.id, $event)"
            />

            <button type="submit" class="button-primary">Comment</button>
          </form>

          <div
            v-if="newGroupCommentImages[post.id]"
            class="group-selected-file"
          >
            Attached:
            {{ newGroupCommentImages[post.id].name }}
          </div>
        </section>
      </article>

      <!-- Pagination -->

      <div class="group-post-pagination">
        <div v-if="loadingMoreGroupPosts" class="group-post-loading-more">
          <span class="loading-spinner"></span>

          Loading more posts...
        </div>

        <div v-else-if="groupPostsLoadMoreError" class="group-post-load-error">
          <span>
            {{ groupPostsLoadMoreError }}
          </span>

          <button
            type="button"
            class="button button-ghost button-small"
            @click="loadGroupPosts()"
          >
            Try again
          </button>
        </div>

        <div
          v-else-if="hasMoreGroupPosts"
          ref="groupPostsLoadTrigger"
          class="group-post-load-trigger"
          aria-hidden="true"
        ></div>

        <p v-else class="group-post-end">You're all caught up</p>
      </div>
    </div>
  </section>

  <Teleport to="body">
    <div
      v-if="groupPostModalOpen"
      class="group-post-modal-overlay"
      @click.self="closeGroupPostModal"
    >
      <section class="group-post-modal">
        <header class="group-post-modal-header">
          <h2>Create group post</h2>

          <button
            type="button"
            class="group-post-modal-close"
            :disabled="creatingGroupPost"
            @click="closeGroupPostModal"
          >
            <i class="fa-solid fa-xmark" aria-hidden="true"></i>
          </button>
        </header>

        <form @submit.prevent="createGroupPost">
          <div class="group-post-modal-body">
            <textarea
              v-model="newGroupPostContent"
              rows="6"
              placeholder="
                Share something with the group...
              "
              required
            ></textarea>

            <div v-if="newGroupPostImage" class="group-selected-file">
              Attached:
              {{ newGroupPostImage.name }}
            </div>

            <p v-if="groupPostsError" class="group-page-error">
              {{ groupPostsError }}
            </p>
          </div>

          <footer class="group-post-modal-footer">
            <div>
              <label
                for="
                  group-post-modal-image
                "
                class="button button-ghost"
              >
                Image or GIF
              </label>

              <input
                id="
                  group-post-modal-image
                "
                ref="
                  groupPostImageInput
                "
                class="visually-hidden"
                type="file"
                accept="
                  image/png,
                  image/jpeg,
                  image/gif
                "
                @change="handleGroupPostImageChange"
              />
            </div>

            <div class="group-post-modal-actions">
              <button
                type="button"
                class="button button-ghost"
                :disabled="creatingGroupPost"
                @click="closeGroupPostModal"
              >
                Cancel
              </button>

              <button
                type="submit"
                class="button-primary"
                :disabled="creatingGroupPost"
              >
                {{ creatingGroupPost ? "Posting..." : "Create post" }}
              </button>
            </div>
          </footer>
        </form>
      </section>
    </div>
  </Teleport>
</template>

<script setup>
import UserAvatar from "../UserAvatar.vue";

import { onUnmounted, ref, watch } from "vue";

import { apiRequest } from "../../services/api";
import { formatDateTime } from "../../utils/date";

const props = defineProps({
  groupId: {
    type: [String, Number],
    required: true,
  },

  active: {
    type: Boolean,
    default: false,
  },
});

const groupPosts = ref([]);

const loadingGroupPosts = ref(false);
const groupPostsError = ref("");

const GROUP_POSTS_PAGE_SIZE = 10;

const groupPostSort = ref("newest");
const groupPostOffset = ref(0);

const hasMoreGroupPosts = ref(true);
const loadingMoreGroupPosts = ref(false);

const groupPostsLoadMoreError = ref("");

const groupPostsLoadTrigger = ref(null);

let groupPostsObserver = null;

const postsLoaded = ref(false);

const newGroupPostContent = ref("");
const newGroupPostImage = ref(null);
const groupPostImageInput = ref(null);

const groupPostModalOpen = ref(false);
const creatingGroupPost = ref(false);

// Comments
const GROUP_COMMENTS_PAGE_SIZE = 5;

const groupCommentsByPost = ref({});

const newGroupComments = ref({});
const newGroupCommentImages = ref({});
const groupCommentImageInputs = ref({});

const groupCommentErrors = ref({});

const openGroupComments = ref({});

const loadingGroupComments = ref({});
const loadingEarlierGroupComments = ref({});

const groupCommentsHasMore = ref({});
const groupCommentsBeforeID = ref({});

// Posts

async function loadGroupPosts(reset = false) {
  if (
    !reset &&
    (loadingGroupPosts.value ||
      loadingMoreGroupPosts.value ||
      !hasMoreGroupPosts.value)
  ) {
    return;
  }

  if (reset) {
    groupPosts.value = [];

    groupPostOffset.value = 0;

    hasMoreGroupPosts.value = true;

    groupPostsLoadMoreError.value = "";

    openGroupComments.value = {};

    groupCommentsByPost.value = {};

    groupCommentsHasMore.value = {};

    groupCommentsBeforeID.value = {};

    groupCommentErrors.value = {};
  }

  const initialLoad = groupPostOffset.value === 0;

  try {
    if (initialLoad) {
      loadingGroupPosts.value = true;
      groupPostsError.value = "";
    } else {
      loadingMoreGroupPosts.value = true;

      groupPostsLoadMoreError.value = "";
    }

    const result = await apiRequest(
      `/groups/${props.groupId}/posts` +
        `?limit=${GROUP_POSTS_PAGE_SIZE}` +
        `&offset=${groupPostOffset.value}` +
        `&sort=${groupPostSort.value}`,
    );

    const incomingPosts = result.posts || [];

    if (reset) {
      groupPosts.value = incomingPosts;
    } else {
      const existingIDs = new Set(groupPosts.value.map((post) => post.id));

      groupPosts.value.push(
        ...incomingPosts.filter((post) => !existingIDs.has(post.id)),
      );
    }

    groupPostOffset.value =
      result.next_offset ?? groupPostOffset.value + incomingPosts.length;

    hasMoreGroupPosts.value = Boolean(result.has_more);
  } catch (err) {
    if (initialLoad) {
      groupPostsError.value = err.message;
    } else {
      groupPostsLoadMoreError.value = err.message;
    }
  } finally {
    loadingGroupPosts.value = false;

    loadingMoreGroupPosts.value = false;
  }
}

async function changeGroupPostSort() {
  await loadGroupPosts(true);
}

// Comments

async function loadGroupComments(postId, loadEarlier = false) {
  try {
    if (loadEarlier) {
      loadingEarlierGroupComments.value[postId] = true;
    } else {
      loadingGroupComments.value[postId] = true;
    }

    groupCommentErrors.value[postId] = "";

    let path =
      `/groups/${props.groupId}` +
      `/posts/${postId}/comments` +
      `?limit=${GROUP_COMMENTS_PAGE_SIZE}`;

    if (loadEarlier && groupCommentsBeforeID.value[postId]) {
      path += `&before_id=` + groupCommentsBeforeID.value[postId];
    }

    const result = await apiRequest(path);

    const incoming = result.comments || [];

    if (loadEarlier) {
      groupCommentsByPost.value[postId] = [
        ...incoming,
        ...(groupCommentsByPost.value[postId] || []),
      ];
    } else {
      groupCommentsByPost.value[postId] = incoming;
    }

    groupCommentsHasMore.value[postId] = Boolean(result.has_more);

    groupCommentsBeforeID.value[postId] = result.next_before_id || 0;
  } catch (err) {
    groupCommentErrors.value[postId] = err.message;
  } finally {
    loadingGroupComments.value[postId] = false;

    loadingEarlierGroupComments.value[postId] = false;
  }
}

async function toggleGroupComments(postId) {
  if (openGroupComments.value[postId]) {
    openGroupComments.value[postId] = false;

    return;
  }

  openGroupComments.value[postId] = true;

  const alreadyLoaded = Object.prototype.hasOwnProperty.call(
    groupCommentsByPost.value,
    postId,
  );

  if (!alreadyLoaded) {
    await loadGroupComments(postId);
  }
}

async function loadEarlierGroupComments(postId) {
  if (
    loadingEarlierGroupComments.value[postId] ||
    !groupCommentsHasMore.value[postId]
  ) {
    return;
  }

  await loadGroupComments(postId, true);
}

async function createGroupComment(postId) {
  try {
    groupCommentErrors.value[postId] = "";

    const content = newGroupComments.value[postId] || "";

    const formData = new FormData();

    formData.append("content", content);

    if (newGroupCommentImages.value[postId]) {
      formData.append("image", newGroupCommentImages.value[postId]);
    }

    await apiRequest(`/groups/${props.groupId}` + `/posts/${postId}/comments`, {
      method: "POST",
      body: formData,
    });

    newGroupComments.value[postId] = "";

    newGroupCommentImages.value[postId] = null;

    if (groupCommentImageInputs.value[postId]) {
      groupCommentImageInputs.value[postId].value = "";
    }

    await loadGroupComments(postId);
  } catch (err) {
    groupCommentErrors.value[postId] = err.message;
  }
}


function handleGroupPostImageChange(event) {
  const file = event.target.files[0];

  newGroupPostImage.value = file || null;
}

function handleGroupCommentImageChange(postId, event) {
  const file = event.target.files[0];

  newGroupCommentImages.value[postId] = file || null;
}

function openGroupPostModal() {
  groupPostsError.value = "";

  groupPostModalOpen.value = true;
}

function resetGroupPostForm() {
  newGroupPostContent.value = "";
  newGroupPostImage.value = null;

  if (groupPostImageInput.value) {
    groupPostImageInput.value.value = "";
  }
}

function closeGroupPostModal() {
  if (creatingGroupPost.value) {
    return;
  }

  groupPostModalOpen.value = false;

  resetGroupPostForm();
}

async function createGroupPost() {
  try {
    creatingGroupPost.value = true;

    groupPostsError.value = "";

    const formData = new FormData();

    formData.append("content", newGroupPostContent.value);

    if (newGroupPostImage.value) {
      formData.append("image", newGroupPostImage.value);
    }

    await apiRequest(`/groups/${props.groupId}/posts`, {
      method: "POST",
      body: formData,
    });

    resetGroupPostForm();

    groupPostModalOpen.value = false;

    groupPostSort.value = "newest";

    await loadGroupPosts(true);
  } catch (err) {
    groupPostsError.value = err.message;
  } finally {
    creatingGroupPost.value = false;
  }
}

// Infinite scrolling

function observeGroupPostsTrigger(element) {
  if (groupPostsObserver) {
    groupPostsObserver.disconnect();
    groupPostsObserver = null;
  }

  if (!element) {
    return;
  }

  groupPostsObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];

      if (
        entry.isIntersecting &&
        props.active &&
        hasMoreGroupPosts.value &&
        !loadingGroupPosts.value &&
        !loadingMoreGroupPosts.value
      ) {
        loadGroupPosts();
      }
    },
    {
      root: null,
      rootMargin: "300px 0px",
      threshold: 0,
    },
  );

  groupPostsObserver.observe(element);
}

function resetPosts() {
  postsLoaded.value = false;

  groupPosts.value = [];

  groupPostSort.value = "newest";

  groupPostOffset.value = 0;

  hasMoreGroupPosts.value = true;

  loadingMoreGroupPosts.value = false;

  groupPostsLoadMoreError.value = "";

  groupPostsError.value = "";

  openGroupComments.value = {};

  groupCommentsByPost.value = {};

  groupCommentsHasMore.value = {};

  groupCommentsBeforeID.value = {};

  groupCommentErrors.value = {};

  newGroupComments.value = {};

  newGroupCommentImages.value = {};

  closeGroupPostModal();
}

watch(
  () => props.active,
  async (active) => {
    if (active && !postsLoaded.value) {
      await loadGroupPosts(true);

      postsLoaded.value = true;
    }
  },
  {
    immediate: true,
  },
);

watch(
  () => props.groupId,
  async () => {
    resetPosts();

    if (props.active) {
      await loadGroupPosts(true);

      postsLoaded.value = true;
    }
  },
);

watch(groupPostsLoadTrigger, (element) => {
  observeGroupPostsTrigger(element);
});

onUnmounted(() => {
  if (groupPostsObserver) {
    groupPostsObserver.disconnect();
  }
});
</script>

<style scoped>
.group-tab-panel {
  padding: 22px;

  border: 1px solid var(--border-soft);
  border-radius: var(--radius-lg);

  background: var(--surface);

  box-shadow: var(--shadow-sm);
}

.group-panel-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 18px;

  margin-bottom: 20px;
}

.group-panel-heading h2 {
  margin-bottom: 4px;
}

.group-section-state {
  color: var(--text-muted);

  font-size: 13px;

  text-align: center;
}

.group-page-error {
  margin: 0;

  padding: 11px 13px;

  border: 1px solid rgba(255, 95, 109, 0.2);
  border-radius: var(--radius-md);

  background: var(--danger-soft);

  color: var(--danger);

  font-size: 13px;
}

.group-posts-toolbar {
  display: flex;
  align-items: center;

  gap: 8px;
}

.group-posts-toolbar label {
  color: var(--text-muted);

  font-size: 12px;
}

.group-post-sort {
  width: auto;
  min-width: 105px;
}

.group-post-list {
  display: grid;

  gap: 14px;
}

.group-post-card {
  margin: 0;
  padding: 0;

  overflow: hidden;

  background: var(--bg-secondary);
}

.group-post-author {
  display: flex;
  align-items: center;

  gap: 11px;

  padding: 16px 17px 8px;
}

.group-post-avatar {
  width: 40px;
  height: 40px;

  flex: 0 0 40px;

  display: grid;
  place-items: center;
  overflow: hidden;

  border-radius: 50%;

  background: var(--primary-soft);

  color: var(--primary);

  font-weight: 800;
}

.group-comment-layout {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.group-comment-avatar {
  width: 32px;
  height: 32px;

  flex: 0 0 32px;

  display: grid;
  place-items: center;

  overflow: hidden;

  border: 1px solid var(--primary-border);
  border-radius: 50%;

  background: var(--primary-soft);
  color: var(--primary);

  font-size: 10px;
  font-weight: 800;
}

.group-comment-body {
  min-width: 0;
  flex: 1;
}

.group-post-author-info {
  display: flex;
  flex-direction: column;

  gap: 1px;
}

.group-post-author-info span,
.group-post-author-info small {
  color: var(--text-muted);

  font-size: 11px;
}

.group-post-content {
  padding: 4px 17px 12px;

  white-space: pre-wrap;
}

.group-post-image-wrapper {
  padding: 0 17px 15px;
}

.group-post-image {
  width: 100%;
  max-height: 600px;

  margin: 0;

  object-fit: contain;
}

.group-comments-toggle {
  width: 100%;
  min-height: 42px;

  justify-content: space-between;

  padding: 8px 17px;

  border: 0;
  border-top: 1px solid var(--border-soft);
  border-radius: 0;

  background: transparent;

  color: var(--text-secondary);
}

.group-comments-toggle:hover {
  background: var(--primary-soft);

  color: var(--primary);
}

.group-post-comments {
  margin: 16px 0 0;

  padding: 14px 17px 17px;

  border-top: 1px solid var(--border-soft);
}

.group-comment-state {
  display: flex;
  align-items: center;
  justify-content: center;

  gap: 7px;

  padding: 12px;

  color: var(--text-muted);

  font-size: 12px;
}

.group-earlier-comments {
  width: 100%;

  margin-bottom: 10px;

  border: 0;

  background: transparent;

  color: var(--primary);
}

.group-comment-card {
  margin-bottom: 8px;
  padding: 12px;

  background: var(--surface);
}

.group-comment-header {
  display: flex;
  align-items: baseline;

  flex-wrap: wrap;

  gap: 6px;

  margin-bottom: 5px;
}

.group-comment-header span,
.group-comment-header small {
  color: var(--text-muted);

  font-size: 11px;
}

.group-comment-card p {
  margin: 0;
}

.group-comment-card img {
  max-width: 220px;
}

.group-comment-form {
  display: flex;
  align-items: center;

  gap: 8px;

  margin-top: 12px;
}

.group-comment-form input {
  flex: 1;
}

.group-selected-file {
  margin-top: 7px;

  color: var(--text-muted);

  font-size: 11px;
}

.group-post-pagination {
  padding: 15px 0 2px;

  text-align: center;
}

.group-post-loading-more,
.group-post-load-error {
  min-height: 42px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 8px;

  color: var(--text-muted);

  font-size: 12px;
}

.group-post-load-trigger {
  height: 2px;
}

.group-post-end {
  color: var(--text-muted);

  font-size: 12px;
}

.group-empty-posts {
  padding: 40px 20px;

  text-align: center;
}

.group-post-modal-overlay {
  position: fixed;
  inset: 0;

  z-index: 2100;

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;

  background: rgba(3, 7, 18, 0.76);

  backdrop-filter: blur(5px);
}

.group-post-modal {
  width: min(650px, 100%);
  max-height: calc(100vh - 48px);

  overflow-y: auto;

  border: 1px solid var(--border);
  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.55);
}

.group-post-modal-header {
  display: flex;
  justify-content: space-between;

  gap: 16px;

  padding: 20px 22px 10px;

  border-bottom: 1px solid var(--border-soft);
}

.group-post-modal-header h2 {
  margin: 0;
}

.group-post-modal-close {
  width: 36px;
  height: 36px;
  min-height: 0;

  padding: 0;

  border: 0;
  border-radius: 50%;

  background: transparent;

  color: var(--text-secondary);

  font-size: 15px;
}

.group-post-modal-body {
  padding: 22px;
}

.group-post-modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 12px;

  padding: 16px 22px;

  border-top: 1px solid var(--border-soft);
}

.group-post-modal-actions {
  display: flex;

  gap: 8px;
}

@media (max-width: 700px) {
  .group-tab-panel {
    padding: 16px;
  }

  .group-panel-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .group-posts-toolbar {
    width: 100%;
  }

  .group-posts-toolbar .button-primary {
    margin-left: auto;
  }

  .group-posts-toolbar label {
    display: none;
  }

  .group-comment-form {
    align-items: stretch;

    flex-direction: column;
  }

  .group-comment-form > * {
    width: 100%;
  }

  .group-post-modal-overlay {
    padding: 12px;
  }

  .group-post-modal {
    max-height: calc(100vh - 24px);
  }

  .group-post-modal-footer {
    align-items: stretch;

    flex-direction: column;
  }

  .group-post-modal-actions {
    width: 100%;
  }

  .group-post-modal-actions button {
    flex: 1;
  }
}
</style>
