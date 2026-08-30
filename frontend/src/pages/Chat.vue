<template>
  <main class="chat-page">
    <header class="chat-page-header">
      <h1>Private Chat</h1>
    </header>

    <p v-if="websocket.error" class="chat-error">
      {{ websocket.error }}
    </p>


    <div class="chat-layout">
      <!-- USERS -->
      <aside class="chat-users-panel">
        <header class="chat-users-header">
          <div>
            <h2>Chats</h2>

            <span>
              {{ chatUsers.length }}
              chats
            </span>
          </div>
        </header>

        <!-- SEARCH -->
        <div class="chat-user-search">
          <span aria-hidden="true">
            <i class="fa-solid fa-magnifying-glass"></i>
          </span>

          <input
            v-model="chatSearchQuery"
            type="search"
            placeholder="Search chats..."
            autocomplete="off"
          />
        </div>

        <!-- Loading -->

        <div v-if="loadingUsers" class="chat-users-state">
          <span class="loading-spinner"></span>

          Loading chats...
        </div>

        <p v-else-if="usersError" class="chat-error">
          {{ usersError }}
        </p>

        <!-- NO USERS -->
        <div v-else-if="chatUsers.length === 0" class="chat-users-empty">
          <strong> No chats available </strong>

          <p>
            Follow someone, or have someone follow you, to start a private chat
          </p>
        </div>

        <!-- NO RESULTS -->
        <div
          v-else-if="filteredChatUsers.length === 0"
          class="chat-users-empty"
        >
          <strong> No matches </strong>

          <p>No chat users match your search</p>
        </div>

        <!-- USER LIST -->

        <div v-else class="chat-user-list">
          <button
            v-for="user in filteredChatUsers"
            :key="user.id"
            type="button"
            class="chat-user-button"
            :class="{
              active: selectedUser?.id === user.id,
            }"
            @click="selectUser(user)"
          >
            <div class="chat-user-avatar-wrap">
              <UserAvatar
                :avatar-path="user.avatar_path"
                :name="`${user.first_name} ${user.last_name}`"
                class="chat-user-avatar"
              />

              <span
                class="user-presence-dot"
                :class="{ online: websocket.isUserOnline(user.id) }"
              ></span>
            </div>

            <div class="chat-user-info">
              <strong>
                {{ user.first_name }}
                {{ user.last_name }}
              </strong>

              <span v-if="user.nickname">
                {{ user.nickname }}
              </span>
            </div>

            <span
              v-if="websocket.privateUnreadForUser(user.id) > 0"
              class="chat-unread-badge"
            >
              {{ websocket.privateUnreadForUser(user.id) }}
            </span>
          </button>
        </div>
      </aside>

      <!-- CONVERSATION -->
      <section class="conversation-panel">

        <div v-if="!selectedUser" class="conversation-placeholder">
          <div class="conversation-placeholder-icon">
            <i class="fa-solid fa-comments" aria-hidden="true"></i>
          </div>

          <h2>Select a conversation</h2>
        </div>

        <template v-else>
          <header class="conversation-header">
            <router-link
              :to="`/profiles/${selectedUser.id}`"
              class="conversation-user"
            >
              <div class="conversation-avatar-wrap">
                <UserAvatar
                  :avatar-path="selectedUser.avatar_path"
                  :name="`${selectedUser.first_name} ${selectedUser.last_name}`"
                  class="conversation-avatar"
                />

                <span
                  class="user-presence-dot"
                  :class="{
                    online: websocket.isUserOnline(selectedUser.id),
                  }"
                ></span>
              </div>

              <div class="conversation-user-info">
                <h2>
                  {{ selectedUser.first_name }}
                  {{ selectedUser.last_name }}
                </h2>

                <span v-if="selectedUser.nickname">
                  {{ selectedUser.nickname }}
                </span>

                <span v-if="selectedUserTyping" class="conversation-typing">
                  typing...
                </span>

                <span
                  v-else
                  class="conversation-presence"
                  :class="{
                    online: websocket.isUserOnline(selectedUser.id),
                  }"
                >
                  {{
                    websocket.isUserOnline(selectedUser.id)
                      ? "Online"
                      : "Offline"
                  }}
                </span>
              </div>
            </router-link>
          </header>

          <p v-if="messagesError" class="chat-error">
            {{ messagesError }}
          </p>

          <!-- MESSAGE WINDOW -->

          <div class="messages-shell">
            <div
              ref="messagesContainer"
              class="messages"
              @scroll.passive="handleMessagesScroll"
            >
              <!-- LOADING -->

              <div
                v-if="loadingMessages && messages.length === 0"
                class="messages-loading"
              >
                <span class="loading-spinner"></span>

                Loading messages...
              </div>

              <!-- LOADING ORDER -->

              <div v-if="loadingOlderMessages" class="messages-loading-older">
                <span class="loading-spinner"></span>

                Loading older messages...
              </div>

              <p
                v-else-if="!hasMoreMessages && messages.length > 0"
                class="conversation-beginning"
              >
                Beginning of conversation
              </p>

              <div
                v-if="!loadingMessages && messages.length === 0"
                class="conversation-empty"
              >
                <div>
                  <i class="fa-solid fa-comments" aria-hidden="true"></i>
                </div>

                <h3>No messages yet</h3>
              </div>

              <!-- MESSAGES -->

              <div
                v-for="message in messages"
                :key="message.id"
                class="message-row"
                :class="{
                  mine: message.sender_id === auth.user?.id,

                  theirs: message.sender_id !== auth.user?.id,
                }"
              >
                <div class="message-bubble">
                  <strong>
                    {{
                      message.sender_id === auth.user?.id
                        ? "Me"
                        : message.sender_name
                    }}
                  </strong>

                  <p>
                    {{ message.content }}
                  </p>

                  <small>
                    {{ formatDateTime(message.created_at) }}
                  </small>
                </div>
              </div>
            </div>

            <button
              v-if="newPrivateMessageCount > 0"
              type="button"
              class="private-new-message-button"
              @click="jumpToNewestMessage"
            >
              {{
                newPrivateMessageCount === 1
                  ? "1 new message"
                  : `${newPrivateMessageCount} new messages`
              }}
              <i class="fa-solid fa-arrow-down" aria-hidden="true"></i>
            </button>
          </div>

          <!-- SEND -->

          <form class="message-form" @submit.prevent="sendMessage">
            <input
              v-model="messageInput"
              type="text"
              placeholder="Type a message..."
              autocomplete="off"
              @input="handleTypingInput"
              @blur="stopTyping"
            />

            <button
              type="submit"
              class="button-primary"
              :disabled="!websocket.connected || !messageInput.trim()"
            >
              Send
            </button>
          </form>

          <p v-if="sendError" class="chat-error">
            {{ sendError }}
          </p>
        </template>
      </section>
    </div>
  </main>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";

import { apiRequest } from "../services/api";
import { useAuthStore } from "../stores/auth";
import { useWebSocketStore } from "../stores/websocket";
import { formatDateTime } from "../utils/date";

import UserAvatar from "../components/UserAvatar.vue";

const auth = useAuthStore();
const websocket = useWebSocketStore();

const chatUsers = ref([]);
const selectedUser = ref(null);
const messages = ref([]);

const loadingUsers = ref(false);
const loadingMessages = ref(false);

const usersError = ref("");
const messagesError = ref("");
const sendError = ref("");

const messageInput = ref("");

const selectedUserTyping = ref(false);
let typingStopTimer = null;
let receivedTypingTimer = null;
let typingSent = false;
let lastTypingSentAt = 0;

const TYPING_STOP_DELAY = 1500;
const TYPING_REFRESH_INTERVAL = 1000;
const RECEIVED_TYPING_TIMEOUT = 2500;

const messagesContainer = ref(null);
const PRIVATE_MESSAGES_PAGE_SIZE = 30;
const hasMoreMessages = ref(true);
const messagesBeforeID = ref(0);
const loadingOlderMessages = ref(false);
const newPrivateMessageCount = ref(0);
const privateChatNearBottom = ref(true);
let messagesRequestVersion = 0;
const CHAT_TOP_THRESHOLD = 80;
const CHAT_BOTTOM_THRESHOLD = 100;

const chatSearchQuery = ref("");
const filteredChatUsers = computed(() => {
  const query = chatSearchQuery.value.trim().toLowerCase();

  if (!query) {
    return chatUsers.value;
  }

  return chatUsers.value.filter((user) => {
    const searchableText = [user.first_name, user.last_name, user.nickname]
      .filter(Boolean)
      .join(" ")
      .toLowerCase();

    return searchableText.includes(query);
  });
});

function displayUserName(user) {
  if (!user) {
    return "";
  }

  const fullName = `${user.first_name} ${user.last_name}`;

  if (user.nickname) {
    return `${fullName} (${user.nickname})`;
  }

  return fullName;
}

async function loadChatUsers() {
  try {
    loadingUsers.value = true;
    usersError.value = "";

    chatUsers.value = await apiRequest("/chat/users");
  } catch (err) {
    usersError.value = err.message;
  } finally {
    loadingUsers.value = false;
  }
}

async function selectUser(user) {
  if (!user) {
    return;
  }

  const sameUser = selectedUser.value?.id === user.id;

  stopTyping();
  clearReceivedTyping();

  selectedUser.value = user;

  websocket.setActivePrivateChat(user.id);

  newPrivateMessageCount.value = 0;
  sendError.value = "";

  if (sameUser && messages.value.length > 0) {
    await scrollToBottom();
    return;
  }

  messageInput.value = "";

  await loadMessages(user.id, true);
}

async function loadMessages(userId, reset = false) {
  if (!userId) {
    return;
  }

  if (
    !reset &&
    (loadingMessages.value ||
      loadingOlderMessages.value ||
      !hasMoreMessages.value)
  ) {
    return;
  }

  if (reset) {
    messagesRequestVersion += 1;

    messages.value = [];

    messagesBeforeID.value = 0;

    hasMoreMessages.value = true;

    newPrivateMessageCount.value = 0;

    messagesError.value = "";
  }

  const requestVersion = messagesRequestVersion;

  const loadingOlder = !reset && messages.value.length > 0;

  const container = messagesContainer.value;

  const previousScrollHeight = container ? container.scrollHeight : 0;

  const previousScrollTop = container ? container.scrollTop : 0;

  try {
    if (loadingOlder) {
      loadingOlderMessages.value = true;
    } else {
      loadingMessages.value = true;
    }

    messagesError.value = "";

    const params = new URLSearchParams();

    params.set("limit", String(PRIVATE_MESSAGES_PAGE_SIZE));

    if (loadingOlder && messagesBeforeID.value) {
      params.set("before_id", String(messagesBeforeID.value));
    }

    const result = await apiRequest(
      `/chat/${userId}/messages` + `?${params.toString()}`,
    );

    if (
      requestVersion !== messagesRequestVersion ||
      selectedUser.value?.id !== userId
    ) {
      return;
    }

    const incomingMessages = result.messages || [];

    // MERGE
    if (reset) {
      messages.value = incomingMessages;
    } else {
      const existingIDs = new Set(messages.value.map((message) => message.id));

      const olderMessages = incomingMessages.filter(
        (message) => !existingIDs.has(message.id),
      );

      messages.value = [...olderMessages, ...messages.value];
    }

    hasMoreMessages.value = Boolean(result.has_more);

    messagesBeforeID.value = result.next_before_id || 0;

    // SCROLL
    await nextTick();

    if (reset) {
      await scrollToBottom();
    } else {
      const updatedContainer = messagesContainer.value;

      if (updatedContainer) {
        const newScrollHeight = updatedContainer.scrollHeight;

        updatedContainer.scrollTop =
          previousScrollTop + (newScrollHeight - previousScrollHeight);
      }
    }
  } catch (err) {
    if (requestVersion === messagesRequestVersion) {
      messagesError.value = err.message;
    }
  } finally {
    if (requestVersion === messagesRequestVersion) {
      loadingMessages.value = false;

      loadingOlderMessages.value = false;
    }
  }
}

async function scrollToBottom() {
  await nextTick();

  const container = messagesContainer.value;

  if (!container) {
    return;
  }

  container.scrollTop = container.scrollHeight;

  privateChatNearBottom.value = true;

  newPrivateMessageCount.value = 0;
}

function isPrivateChatNearBottom() {
  const container = messagesContainer.value;

  if (!container) {
    return true;
  }

  const distanceFromBottom =
    container.scrollHeight - container.scrollTop - container.clientHeight;

  return distanceFromBottom <= CHAT_BOTTOM_THRESHOLD;
}

function handleMessagesScroll() {
  const container = messagesContainer.value;

  if (!container) {
    return;
  }

  privateChatNearBottom.value = isPrivateChatNearBottom();

  if (privateChatNearBottom.value) {
    newPrivateMessageCount.value = 0;
  }

  if (
    container.scrollTop <= CHAT_TOP_THRESHOLD &&
    selectedUser.value &&
    hasMoreMessages.value &&
    !loadingMessages.value &&
    !loadingOlderMessages.value
  ) {
    loadMessages(selectedUser.value.id, false);
  }
}

function clearReceivedTyping() {
  if (receivedTypingTimer) {
    clearTimeout(receivedTypingTimer);
    receivedTypingTimer = null;
  }

  selectedUserTyping.value = false;
}

function handleIncomingPrivateTyping(data) {
  if (!data) {
    return;
  }

  const senderID = Number(data.sender_id);

  const selectedUserID = Number(selectedUser.value?.id);

  if (!senderID || !selectedUserID || senderID !== selectedUserID) {
    return;
  }

  if (!data.typing) {
    clearReceivedTyping();
    return;
  }

  selectedUserTyping.value = true;

  if (receivedTypingTimer) {
    clearTimeout(receivedTypingTimer);
  }

  receivedTypingTimer = setTimeout(() => {
    selectedUserTyping.value = false;
    receivedTypingTimer = null;
  }, RECEIVED_TYPING_TIMEOUT);
}

function handleIncomingPrivateMessage(message) {
  if (!message) {
    return;
  }

  const currentUserID = auth.user?.id;

  const selectedUserID = selectedUser.value?.id;

  if (!currentUserID || !selectedUserID) {
    return;
  }

  const belongsToOpenConversation =
    (message.sender_id === currentUserID &&
      message.receiver_id === selectedUserID) ||
    (message.sender_id === selectedUserID &&
      message.receiver_id === currentUserID);

  if (!belongsToOpenConversation) {
    return;
  }

  const alreadyExists = messages.value.some(
    (existingMessage) => existingMessage.id === message.id,
  );

  if (alreadyExists) {
    return;
  }

  const wasNearBottom = isPrivateChatNearBottom();

  const isMyMessage = message.sender_id === currentUserID;

  if (!isMyMessage) {
    clearReceivedTyping();
  }

  messages.value.push(message);

  if (wasNearBottom || isMyMessage) {
    newPrivateMessageCount.value = 0;

    scrollToBottom();

    return;
  }

  newPrivateMessageCount.value += 1;
}

async function jumpToNewestMessage() {
  await scrollToBottom();
}

function sendTypingStatus(typing) {
  if (!selectedUser.value || !websocket.connected) {
    return;
  }

  const sent = websocket.send({
    type: "private_typing",
    receiver_id: selectedUser.value.id,
    typing,
  });

  if (!sent) {
    return;
  }

  typingSent = typing;

  if (typing) {
    lastTypingSentAt = Date.now();
  } else {
    lastTypingSentAt = 0;
  }
}

function stopTyping() {
  if (typingStopTimer) {
    clearTimeout(typingStopTimer);
    typingStopTimer = null;
  }

  if (typingSent) {
    sendTypingStatus(false);
  }

  typingSent = false;
  lastTypingSentAt = 0;
}

function handleTypingInput() {
  if (!selectedUser.value || !websocket.connected) {
    return;
  }

  const hasText = Boolean(messageInput.value.trim());

  if (!hasText) {
    stopTyping();
    return;
  }

  const now = Date.now();

  if (!typingSent || now - lastTypingSentAt >= TYPING_REFRESH_INTERVAL) {
    sendTypingStatus(true);
  }

  if (typingStopTimer) {
    clearTimeout(typingStopTimer);
  }

  typingStopTimer = setTimeout(() => {
    sendTypingStatus(false);
    typingStopTimer = null;
  }, TYPING_STOP_DELAY);
}

function sendMessage() {
  sendError.value = "";

  if (!selectedUser.value) {
    sendError.value = "select a user first";

    return;
  }

  const content = messageInput.value.trim();

  if (!content) {
    return;
  }

  const sent = websocket.send({
    type: "private_message",
    receiver_id: selectedUser.value.id,
    content,
  });

  if (!sent) {
    sendError.value = websocket.error;
    return;
  }

  stopTyping();

  messageInput.value = "";
}

watch(
  () => websocket.eventVersion,
  () => {
    const event = websocket.lastEvent;

    if (!event) {
      return;
    }

    if (event.type === "private_message") {
      handleIncomingPrivateMessage(event.data);

      return;
    }

    if (event.type === "private_typing") {
      handleIncomingPrivateTyping(event.data);
    }
  },
);
onMounted(async () => {
  await loadChatUsers();
});

onUnmounted(() => {
  messagesRequestVersion += 1;

  stopTyping();
  clearReceivedTyping();

  websocket.setActivePrivateChat(null);
});
</script>

<style scoped>
.chat-page {
  width: min(1120px, 100%);
}

.chat-page-header {
  margin-bottom: 20px;
}

.chat-page-header h1 {
  margin-bottom: 5px;
}

.chat-layout {
  min-height: 650px;

  display: grid;

  grid-template-columns: 285px minmax(0, 1fr);

  overflow: hidden;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-sm);
}

.chat-users-panel {
  min-width: 0;

  padding: 17px;

  border-right: 1px solid var(--border-soft);

  background: var(--bg-secondary);
}

.chat-users-header {
  margin-bottom: 14px;
}

.chat-users-header > div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;

  gap: 10px;
}

.chat-users-header h2 {
  margin: 0;
}

.chat-users-header span {
  color: var(--text-muted);

  font-size: 10px;
}

.chat-user-search {
  position: relative;

  margin-bottom: 12px;
}

.chat-user-search > span {
  position: absolute;

  left: 12px;
  top: 50%;

  color: var(--text-muted);

  transform: translateY(-50%);

  pointer-events: none;

  font-size: 14px;
}

.chat-user-search input {
  width: 100%;

  padding-left: 37px;

  background: var(--surface);
}

.chat-user-avatar-wrap,
.conversation-avatar-wrap {
  position: relative;

  width: 40px;
  height: 40px;

  flex: 0 0 40px;
}

.user-presence-dot {
  position: absolute;

  right: -1px;
  bottom: -1px;

  width: 11px;
  height: 11px;

  border: 2px solid var(--surface);
  border-radius: 50%;

  background: var(--text-muted);
}

.user-presence-dot.online {
  background: var(--success);
}

.chat-user-list {
  max-height: 560px;

  overflow-y: auto;
}

.chat-user-button {
  width: 100%;
  min-height: 62px;

  display: flex;
  align-items: center;

  gap: 10px;

  margin: 0;
  padding: 9px;

  border: 0;
  border-radius: var(--radius-md);

  background: transparent;

  color: var(--text);

  text-align: left;
}

.chat-user-button:hover {
  background: var(--surface-2);
}

.chat-user-button.active {
  background: var(--primary-soft);
}

.chat-user-avatar,
.conversation-avatar {
  width: 40px;
  height: 40px;

  flex: 0 0 40px;

  display: grid;
  place-items: center;

  overflow: hidden;

  border: 1px solid var(--primary-border);

  border-radius: 50%;

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 11px;
  font-weight: 800;
}

.chat-user-info {
  min-width: 0;
  flex: 1;

  display: flex;
  flex-direction: column;
}

.chat-user-info strong {
  overflow: hidden;

  font-size: 12px;

  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-user-info span {
  overflow: hidden;

  color: var(--text-muted);

  font-size: 10px;

  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-unread-badge {
  min-width: 20px;
  height: 20px;

  display: inline-flex;
  align-items: center;
  justify-content: center;

  padding: 0 5px;

  border-radius: var(--radius-round);

  background: var(--primary);

  color: white;

  font-size: 9px;
  font-weight: 750;
}

.chat-users-state,
.chat-users-empty {
  padding: 22px 8px;

  color: var(--text-muted);

  font-size: 11px;

  text-align: center;
}

.chat-users-state {
  display: flex;
  align-items: center;
  justify-content: center;

  gap: 7px;
}

.chat-users-empty strong {
  display: block;

  margin-bottom: 5px;
}

.chat-users-empty p {
  margin: 0;
}

.conversation-panel {
  min-width: 0;

  display: flex;
  flex-direction: column;

  padding: 0;
}

.conversation-user-info {
  min-width: 0;

  display: flex;
  flex-direction: column;
}

.conversation-presence {
  color: var(--text-muted);
}

.conversation-presence.online {
  color: var(--success);
}

.conversation-placeholder {
  min-height: 650px;

  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;

  padding: 30px;

  color: var(--text-muted);

  text-align: center;
}

.conversation-placeholder-icon {
  width: 58px;
  height: 58px;

  display: grid;
  place-items: center;

  margin-bottom: 12px;

  border-radius: 50%;

  background: var(--primary-soft);

  font-size: 23px;
}

.conversation-placeholder h2 {
  margin-bottom: 5px;

  color: var(--text);
}

.conversation-user .conversation-typing {
  color: var(--primary);

  font-weight: 650;
}

.conversation-header {
  min-height: 72px;

  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 14px;

  padding: 13px 18px;

  border-bottom: 1px solid var(--border-soft);
}

.conversation-user {
  min-width: 0;

  display: flex;
  align-items: center;

  gap: 10px;

  color: inherit;
}

.conversation-user h2 {
  margin: 0;

  font-size: 15px;
}

.conversation-user span {
  color: var(--text-muted);

  font-size: 10px;
}

.messages-shell {
  position: relative;

  flex: 1;

  min-height: 0;
}

.messages {
  height: 510px;

  overflow-y: auto;

  padding: 18px;

  background: var(--bg-secondary);

  scroll-behavior: auto;
}

.messages-loading,
.messages-loading-older {
  display: flex;
  align-items: center;
  justify-content: center;

  gap: 7px;

  padding: 12px;

  color: var(--text-muted);

  font-size: 10px;
}

.conversation-beginning {
  margin: 0 0 16px;

  color: var(--text-muted);

  font-size: 9px;

  text-align: center;
}

.conversation-empty {
  min-height: 430px;

  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;

  color: var(--text-muted);

  text-align: center;
}

.conversation-empty > div {
  width: 48px;
  height: 48px;

  display: grid;
  place-items: center;

  margin-bottom: 9px;

  border-radius: 50%;

  background: var(--primary-soft);

  font-size: 19px;
}

.conversation-empty h3 {
  margin-bottom: 4px;

  color: var(--text);
}

.message-row {
  display: flex;

  margin-bottom: 9px;
}

.message-row.mine {
  justify-content: flex-end;
}

.message-row.theirs {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 72%;

  padding: 9px 12px;

  border: 1px solid var(--border);

  border-radius: 14px;

  background: var(--surface-2);
}

.message-row.mine .message-bubble {
  border-color: var(--primary-border);

  background: var(--primary-soft);
}

.message-bubble strong {
  display: block;

  margin-bottom: 3px;

  font-size: 10px;
}

.message-bubble p {
  margin: 0 0 4px;

  white-space: pre-wrap;

  overflow-wrap: anywhere;

  line-height: 1.45;
}

.message-bubble small {
  color: var(--text-muted);

  font-size: 9px;
}

.private-new-message-button {
  position: absolute;

  left: 50%;
  bottom: 14px;

  z-index: 5;

  min-height: 31px;

  padding: 5px 11px;

  border: 1px solid var(--primary-border);

  border-radius: var(--radius-round);

  background: var(--primary);

  color: white;

  font-size: 10px;
  font-weight: 700;

  transform: translateX(-50%);

  box-shadow: var(--shadow-md);
}

.message-form {
  display: flex;

  gap: 9px;

  padding: 14px;

  border-top: 1px solid var(--border-soft);
}

.message-form input {
  min-width: 0;
  flex: 1;
}

.chat-error {
  margin: 10px 0;

  padding: 9px 11px;

  border: 1px solid rgba(255, 95, 109, 0.22);

  border-radius: var(--radius-md);

  background: var(--danger-soft);

  color: var(--danger);

  font-size: 11px;
}

@media (max-width: 760px) {
  .chat-page-header {
    align-items: flex-start;

    flex-direction: column;
  }

  .chat-layout {
    display: block;
  }

  .chat-users-panel {
    border-right: 0;

    border-bottom: 1px solid var(--border-soft);
  }

  .chat-user-list {
    max-height: 230px;
  }

  .conversation-placeholder {
    min-height: 360px;
  }

  .messages {
    height: min(58vh, 510px);
  }

  .message-bubble {
    max-width: 88%;
  }

  .message-form {
    align-items: stretch;

    flex-direction: column;
  }

  .message-form button {
    width: 100%;
  }
}
</style>
