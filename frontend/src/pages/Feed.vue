<template>
  <main class="feed-page">
    <div class="feed-layout">
      <div class="feed-main">
        <header class="feed-heading">
          <h1>Your Feed</h1>
        </header>

        <!-- CREATE POST -->
        <section class="post-trigger-card">
          <UserAvatar
            :avatar-path="auth.user.avatar_path"
            :name="`${auth.user.first_name} ${auth.user.last_name}`"
            class="user-avatar"
          />

          <button
            type="button"
            class="post-trigger-prompt"
            @click="openPostModal"
          >
            Start a post
          </button>
        </section>

        <!-- POSTS -->

        <section class="posts-feed">
          <div class="section-heading-row">
            <h2>Posts</h2>

            <div class="feed-sort-control">
              <select
                id="feed-sort"
                v-model="feedSort"
                class="feed-sort-select"
                :disabled="loading || loadingMorePosts"
              >
                <option value="newest">Newest</option>

                <option value="oldest">Oldest</option>
              </select>
            </div>
          </div>

          <!-- LOADING -->

          <div v-if="loading" class="feed-state">
            <span class="loading-spinner"></span>

            Loading posts...
          </div>

          <div v-else-if="loadError" class="feed-state feed-state-error">
            {{ loadError }}
          </div>

          <div v-else-if="posts.length === 0" class="empty-state">
            <div class="empty-state-icon">
              <img src="/social_network_logo.png" alt="" />
            </div>

            <h3>No posts yet</h3>
          </div>

          <!-- POST CARDS -->

          <article v-for="post in posts" :key="post.id" class="post-card">
            <header class="post-card-header">
              <router-link
                :to="`/profiles/${post.user_id}`"
                class="user-avatar-link"
              >
                <UserAvatar
                  :avatar-path="post.author_avatar_path"
                  :name="post.author_name || post.author"
                  class="user-avatar"
                />
              </router-link>
              <div class="post-author">
                <div class="post-author-row">
                  <router-link
                    :to="`/profiles/${post.user_id}`"
                    class="post-author-name"
                  >
                    {{ post.author_name || post.author }}
                  </router-link>

                  <span
                    class="privacy-badge"
                    :class="`privacy-${post.privacy}`"
                  >
                    {{ privacyLabel(post.privacy) }}
                  </span>
                </div>

                <div class="post-author-meta">
                  <span v-if="post.author_nickname">
                    {{ post.author_nickname }}
                  </span>

                  <span v-if="post.author_nickname" class="meta-separator">
                    -
                  </span>

                  <time>
                    {{ formatDateTime(post.created_at) }}
                  </time>
                </div>
              </div>
            </header>

            <div class="post-card-body">
              <p class="post-content">
                {{ post.content }}
              </p>

              <img
                v-if="post.image_path"
                :src="post.image_path"
                alt="Post image"
                class="post-image"
              />
            </div>

            <div class="post-actions">
              <button
                type="button"
                class="comments-toggle"
                :aria-expanded="Boolean(openComments[post.id])"
                @click="toggleComments(post.id)"
              >
                <span> Comments </span>

                <span class="comments-toggle-arrow" aria-hidden="true">
                  <i
                    :class="
                      openComments[post.id]
                        ? 'fa-solid fa-chevron-up'
                        : 'fa-solid fa-chevron-down'
                    "
                  ></i>
                </span>
              </button>
            </div>

            <!-- COMMENTS -->

            <section v-if="openComments[post.id]" class="comments-section">
              <div v-if="loadingComments[post.id]" class="comments-loading">
                Loading comments...
              </div>

              <p v-if="commentErrors[post.id]" class="form-error">
                {{ commentErrors[post.id] }}
              </p>

              <div
                v-if="
                  !loadingComments[post.id] &&
                  (commentsByPost[post.id] || []).length === 0
                "
                class="no-comments"
              >
                No comments yet
              </div>

              <button
                v-if="!loadingComments[post.id] && commentsHasMore[post.id]"
                type="button"
                class="comments-earlier-button"
                :disabled="loadingEarlierComments[post.id]"
                @click="loadEarlierComments(post.id)"
              >
                {{
                  loadingEarlierComments[post.id]
                    ? "Loading..."
                    : "View earlier comments"
                }}
              </button>

              <div
                v-for="comment in commentsByPost[post.id] || []"
                :key="comment.id"
                class="comment-item"
              >
                <router-link
                  :to="`/profiles/${comment.user_id}`"
                  class="mini-avatar-link"
                >
                  <UserAvatar
                    :avatar-path="comment.author_avatar_path"
                    :name="comment.author_name"
                    class="mini-avatar"
                  />
                </router-link>

                <div class="comment-content">
                  <div class="comment-header">
                    <span
                      v-if="comment.author_nickname"
                      class="comment-nickname"
                    >
                      {{ comment.author_nickname }}
                    </span>

                    <router-link
                      :to="`/profiles/${comment.user_id}`"
                      class="comment-author"
                    >
                      {{ comment.author_name }}
                    </router-link>

                    <time>
                      {{ formatDateTime(comment.created_at) }}
                    </time>
                  </div>

                  <p>
                    {{ comment.content }}
                  </p>

                  <img
                    v-if="comment.image_path"
                    :src="comment.image_path"
                    alt="Comment image"
                    class="comment-image"
                  />
                </div>
              </div>

              <form
                class="comment-form"
                @submit.prevent="createComment(post.id)"
              >
                <UserAvatar
                  :avatar-path="auth.user.avatar_path"
                  :name="`${auth.user.first_name} ${auth.user.last_name}`"
                  class="mini-avatar"
                />

                <div class="comment-input-wrap">
                  <input
                    v-model="newComments[post.id]"
                    type="text"
                    placeholder="Write a comment..."
                    required
                  />

                  <div class="comment-tools">
                    <label
                      :for="`comment-image-${post.id}`"
                      class="comment-file-button"
                    >
                      <i class="fa-solid fa-image" aria-hidden="true"></i>

                      Image
                    </label>

                    <input
                      :id="`comment-image-${post.id}`"
                      :ref="
                        (el) => {
                          if (el) {
                            commentImageInputs[post.id] = el;
                          }
                        }
                      "
                      class="visually-hidden"
                      type="file"
                      accept="image/png, image/jpeg, image/gif"
                      @change="handleCommentImageChange(post.id, $event)"
                    />

                    <button type="submit" class="button-primary button-small">
                      Comment
                    </button>
                  </div>
                </div>
              </form>

              <div
                v-if="newCommentImages[post.id]"
                class="comment-selected-file"
              >
                Attached:
                {{ newCommentImages[post.id].name }}
              </div>
            </section>
          </article>

          <div v-if="posts.length > 0" class="feed-pagination">
            <!-- LOADING MORE -->

            <div v-if="loadingMorePosts" class="feed-load-more">
              <span class="loading-spinner"></span>

              Loading more posts...
            </div>

            <div v-else-if="loadMoreError" class="feed-load-more-error">
              <span>
                {{ loadMoreError }}
              </span>

              <button
                type="button"
                class="button button-ghost button-small"
                @click="loadPosts()"
              >
                Try again
              </button>
            </div>

            <!-- Load more trigger -->

            <div
              v-else-if="hasMorePosts"
              ref="loadMoreTrigger"
              class="feed-load-trigger"
              aria-hidden="true"
            ></div>

            <p v-else class="feed-end-message">You're all caught up</p>
          </div>
        </section>
      </div>
    </div>
  </main>
  <Teleport to="body">
    <div
      v-if="postModalOpen"
      class="post-modal-overlay"
      @click.self="closePostModal"
    >
      <section
        class="post-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-post-title"
      >
        <header class="post-modal-header">
          <div>
            <h2 id="create-post-title">Create post</h2>
          </div>

          <button
            type="button"
            class="post-modal-close"
            aria-label="Close create post"
            :disabled="posting"
            @click="closePostModal"
          >
            <i class="fa-solid fa-xmark" aria-hidden="true"></i>
          </button>
        </header>

        <!-- POST FORM -->

        <form class="post-form post-modal-form" @submit.prevent="createPost">
          <div class="post-modal-body">
            <div class="post-modal-author">
              <UserAvatar
                :avatar-path="auth.user.avatar_path"
                :name="`${auth.user.first_name} ${auth.user.last_name}`"
                class="user-avatar"
              />

              <div>
                <strong>
                  {{ auth.user.first_name }}
                  {{ auth.user.last_name }}
                </strong>

                <small v-if="auth.user.nickname">
                  {{ auth.user.nickname }}
                </small>
              </div>
            </div>

            <!-- CONTENT -->
            <textarea
              v-model="newPostContent"
              class="post-textarea"
              placeholder="What's on your mind?"
              required
            ></textarea>

            <!-- PRIVATE AUDIENCE -->
            <div v-if="newPostPrivacy === 'private'" class="audience-picker">
              <div class="audience-picker-header">
                <strong> Choose your audience </strong>
              </div>

              <p v-if="followersError" class="form-error">
                {{ followersError }}
              </p>

              <div v-if="myFollowers.length === 0" class="empty-inline">
                You don't have any followers to select yet. Only you will be
                able to see this post.
              </div>

              <div v-else class="audience-list">
                <label
                  v-for="follower in myFollowers"
                  :key="follower.id"
                  class="audience-user"
                >
                  <input
                    v-model="selectedAllowedUserIDs"
                    type="checkbox"
                    :value="follower.id"
                  />

                  <UserAvatar
                    :avatar-path="follower.avatar_path"
                    :name="`${follower.first_name} ${follower.last_name}`"
                    class="mini-avatar"
                  />

                  <span class="audience-user-info">
                    <strong>
                      {{ follower.first_name }}
                      {{ follower.last_name }}
                    </strong>

                    <small v-if="follower.nickname">
                      {{ follower.nickname }}
                    </small>
                  </span>
                </label>
              </div>
            </div>

            <div v-if="newPostImage" class="selected-file">
              <span> Attached: </span>

              <strong>
                {{ newPostImage.name }}
              </strong>
            </div>

            <p v-if="postError" class="form-error">
              {{ postError }}
            </p>
          </div>

          <footer class="post-modal-footer">
            <div class="post-tools">
              <label for="post-image" class="toolbar-button">
                <span class="toolbar-icon">
                  <i class="fa-solid fa-image" aria-hidden="true"></i>
                </span>

                Image or GIF
              </label>

              <input
                id="post-image"
                ref="postImageInput"
                class="visually-hidden"
                type="file"
                accept="image/png,image/jpeg,image/gif"
                @change="handlePostImageChange"
              />

              <!-- PRIVACY -->
              <div class="privacy-control">
                <label for="post-privacy" class="visually-hidden">
                  Post visibility
                </label>

                <select
                  id="post-privacy"
                  v-model="newPostPrivacy"
                  class="privacy-select"
                >
                  <option value="public">Public</option>

                  <option value="followers">Followers</option>

                  <option value="private">Selected followers</option>
                </select>
              </div>
            </div>

            <div class="post-modal-actions">
              <button
                type="button"
                class="button button-ghost"
                :disabled="posting"
                @click="closePostModal"
              >
                Cancel
              </button>

              <button type="submit" class="button-primary" :disabled="posting">
                {{ posting ? "Posting..." : "Create post" }}
              </button>
            </div>
          </footer>
        </form>
      </section>
    </div>
  </Teleport>
</template>

<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import { useAuthStore } from "../stores/auth";
import { apiRequest } from "../services/api";
import { formatDateTime } from "../utils/date";
import UserAvatar from "../components/UserAvatar.vue";
const auth = useAuthStore();
// Posts
const posts = ref([]);
const loading = ref(false);
const loadError = ref("");
const postError = ref("");
const feedSort = ref("newest");
const POSTS_PAGE_SIZE = 10;
const postOffset = ref(0);
const hasMorePosts = ref(true);
const loadingMorePosts = ref(false);
const loadMoreError = ref("");
const loadMoreTrigger = ref(null);
let postObserver = null;
const newPostContent = ref("");
const newPostPrivacy = ref("public");
const postModalOpen = ref(false);
const posting = ref(false);
// Comments
const commentsByPost = ref({});
const newComments = ref({});
const commentErrors = ref({});
const loadingComments = ref({});
const openComments = ref({});
const commentsHasMore = ref({});
const commentsBeforeID = ref({});
const loadingEarlierComments = ref({});
const commentImageInputs = ref({});
const COMMENTS_PAGE_SIZE = 5;
// Images
const newPostImage = ref(null);
const postImageInput = ref(null);
const newCommentImages = ref({});
// private posts
const myFollowers = ref([]);
const selectedAllowedUserIDs = ref([]);
const followersError = ref("");

async function loadMyFollowers() {
  if (!auth.user) return;
  try {
    followersError.value = "";
    myFollowers.value = await apiRequest(`/users/${auth.user.id}/followers`);
  } catch (err) {
    followersError.value = err.message;
    myFollowers.value = [];
  }
}

async function loadPosts(reset = false) {
  if (!auth.user) {
    return;
  }
  if (
    !reset &&
    (loading.value || loadingMorePosts.value || !hasMorePosts.value)
  ) {
    return;
  }
  if (reset) {
    posts.value = [];
    postOffset.value = 0;
    hasMorePosts.value = true;
    loadMoreError.value = "";
  }
  const initialLoad = posts.value.length === 0;
  try {
    if (initialLoad) {
      loading.value = true;
    } else {
      loadingMorePosts.value = true;
    }
    if (initialLoad) {
      loadError.value = "";
    } else {
      loadMoreError.value = "";
    }
    const result = await apiRequest(
      `/posts` +
        `?limit=${POSTS_PAGE_SIZE}` +
        `&offset=${postOffset.value}` +
        `&sort=${feedSort.value}`,
    );
    const incomingPosts = result.posts || [];
    if (reset) {
      posts.value = incomingPosts;
    } else {
      const existingIDs = new Set(posts.value.map((post) => post.id));
      posts.value.push(
        ...incomingPosts.filter((post) => !existingIDs.has(post.id)),
      );
    }
    postOffset.value =
      result.next_offset ?? postOffset.value + incomingPosts.length;
    hasMorePosts.value = Boolean(result.has_more);
  } catch (err) {
    if (initialLoad) {
      loadError.value = err.message;
    } else {
      loadMoreError.value = err.message;
    }
  } finally {
    loading.value = false;
    loadingMorePosts.value = false;
  }
}

function observeLoadMoreTrigger(element) {
  if (postObserver) {
    postObserver.disconnect();
    postObserver = null;
  }
  if (!element) {
    return;
  }
  postObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (
        entry.isIntersecting &&
        hasMorePosts.value &&
        !loading.value &&
        !loadingMorePosts.value
      ) {
        loadPosts();
      }
    },
    {
      root: null,

      rootMargin: "300px 0px",

      threshold: 0,
    },
  );
  postObserver.observe(element);
}

function openPostModal() {
  postError.value = "";
  postModalOpen.value = true;
}

function resetPostForm() {
  newPostContent.value = "";
  newPostPrivacy.value = "public";
  newPostImage.value = null;
  selectedAllowedUserIDs.value = [];
  if (postImageInput.value) {
    postImageInput.value.value = "";
  }
}

function closePostModal() {
  if (posting.value) {
    return;
  }
  postModalOpen.value = false;
  postError.value = "";
  resetPostForm();
}

async function createPost() {
  try {
    posting.value = true;
    postError.value = "";
    const formData = new FormData();
    formData.append("content", newPostContent.value);
    formData.append("privacy", newPostPrivacy.value);
    if (newPostPrivacy.value === "private") {
      formData.append(
        "allowed_user_ids",
        JSON.stringify(selectedAllowedUserIDs.value),
      );
    }
    if (newPostImage.value) {
      formData.append("image", newPostImage.value);
    }
    await apiRequest("/posts", {
      method: "POST",
      body: formData,
    });
    resetPostForm();
    postModalOpen.value = false;
    await loadPosts(true);
  } catch (err) {
    postError.value = err.message;
  } finally {
    posting.value = false;
  }
}

async function loadComments(postId, loadEarlier = false) {
  try {
    if (loadEarlier) {
      loadingEarlierComments.value[postId] = true;
    } else {
      loadingComments.value[postId] = true;
    }
    commentErrors.value[postId] = "";
    let path = `/posts/${postId}/comments` + `?limit=${COMMENTS_PAGE_SIZE}`;
    if (loadEarlier && commentsBeforeID.value[postId]) {
      path += `&before_id=` + commentsBeforeID.value[postId];
    }
    const result = await apiRequest(path);
    const incoming = result.comments || [];
    if (loadEarlier) {
      commentsByPost.value[postId] = [
        ...incoming,
        ...(commentsByPost.value[postId] || []),
      ];
    } else {
      commentsByPost.value[postId] = incoming;
    }
    commentsHasMore.value[postId] = Boolean(result.has_more);
    commentsBeforeID.value[postId] = result.next_before_id || 0;
  } catch (err) {
    commentErrors.value[postId] = err.message;
  } finally {
    loadingComments.value[postId] = false;
    loadingEarlierComments.value[postId] = false;
  }
}

async function toggleComments(postId) {
  if (openComments.value[postId]) {
    openComments.value[postId] = false;
    return;
  }
  openComments.value[postId] = true;
  const alreadyLoaded = Object.prototype.hasOwnProperty.call(
    commentsByPost.value,
    postId,
  );
  if (!alreadyLoaded) {
    await loadComments(postId);
  }
}

async function loadEarlierComments(postId) {
  if (loadingEarlierComments.value[postId] || !commentsHasMore.value[postId]) {
    return;
  }
  await loadComments(postId, true);
}

async function createComment(postId) {
  try {
    commentErrors.value[postId] = "";
    const formData = new FormData();
    formData.append("content", newComments.value[postId] || "");
    if (newCommentImages.value[postId]) {
      formData.append("image", newCommentImages.value[postId]);
    }
    await apiRequest(`/posts/${postId}/comments`, {
      method: "POST",
      body: formData,
    });
    newComments.value[postId] = "";
    newCommentImages.value[postId] = null;
    if (commentImageInputs.value[postId]) {
      commentImageInputs.value[postId].value = "";
    }
    await loadComments(postId);
  } catch (err) {
    commentErrors.value[postId] = err.message;
  }
}

function privacyLabel(privacy) {
  switch (privacy) {
    case "public":
      return "Public";
    case "followers":
      return "Followers";
    case "private":
      return "Selected";
    default:
      return privacy;
  }
}

function handlePostImageChange(event) {
  newPostImage.value = event.target.files[0] || null;
}

function handleCommentImageChange(postId, event) {
  newCommentImages.value[postId] = event.target.files[0] || null;
}

function clearFeed() {
  posts.value = [];
  commentsByPost.value = {};
  newComments.value = {};
  commentErrors.value = {};
  loadingComments.value = {};
  loadError.value = "";
  postError.value = "";
  newCommentImages.value = {};
  myFollowers.value = [];
  selectedAllowedUserIDs.value = [];
  followersError.value = "";
  openComments.value = {};
  commentsHasMore.value = {};
  commentsBeforeID.value = {};
  loadingEarlierComments.value = {};
  commentImageInputs.value = {};
  //rest pagination
  feedSort.value = "newest";
  postOffset.value = 0;
  hasMorePosts.value = true;
  loadingMorePosts.value = false;
  loadMoreError.value = "";
}

watch(
  () => auth.user,
  async (user) => {
    if (user) {
      await loadPosts(true);
      await loadMyFollowers();
    } else {
      clearFeed();
    }
  },
  { immediate: true },
);
watch(feedSort, async () => {
  if (!auth.user) {
    return;
  }
  await loadPosts(true);
});
watch(loadMoreTrigger, (element) => {
  observeLoadMoreTrigger(element);
});
watch(newPostPrivacy, (privacy) => {
  if (privacy !== "private") {
    selectedAllowedUserIDs.value = [];
  }
});
onBeforeUnmount(() => {
  if (postObserver) {
    postObserver.disconnect();
    postObserver = null;
  }
});
</script>
