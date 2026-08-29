<template>
  <section v-show="active" class="group-tab-panel group-chat-panel">
    <!-- Header -->

    <div class="group-panel-heading">
      <h2>Group chat</h2>
    </div>

    <!-- Error -->

    <p v-if="groupMessagesError" class="group-page-error">
      {{ groupMessagesError }}
    </p>

    <!-- Chat window -->

    <div class="group-chat-shell">
      <div
        ref="groupMessagesContainer"
        class="group-chat-messages"
        @scroll.passive="handleGroupChatScroll"
      >
        <!-- Initial loading -->

        <div
          v-if="loadingGroupMessages && groupMessages.length === 0"
          class="group-chat-loading"
        >
          <span class="loading-spinner"></span>

          Loading messages...
        </div>

        <!-- Loading older -->

        <div v-if="loadingOlderGroupMessages" class="group-chat-loading-older">
          <span class="loading-spinner"></span>

          Loading older messages...
        </div>

        <!-- Start of chat -->

        <p
          v-else-if="!hasMoreGroupMessages && groupMessages.length > 0"
          class="group-chat-start"
        >
          Beginning of conversation
        </p>

        <!-- Empty -->

        <div
          v-if="!loadingGroupMessages && groupMessages.length === 0"
          class="group-chat-empty"
        >
          <div class="group-chat-empty-icon">
            <i class="fa-solid fa-comments" aria-hidden="true"></i>
          </div>

          <h3>No messages yet</h3>
        </div>

        <!-- Messages -->

        <div
          v-for="message in groupMessages"
          :key="message.id"
          class="group-chat-message-row"
          :class="{
            mine: message.sender_id === auth.user?.id,
            theirs: message.sender_id !== auth.user?.id,
          }"
        >
          <UserAvatar
            v-if="message.sender_id !== auth.user?.id"
            :avatar-path="message.sender_avatar_path"
            :name="message.sender_name"
            class="group-chat-message-avatar"
          />

          <div class="group-chat-message">
            <strong>
              {{
                message.sender_id === auth.user?.id ? "Me" : message.sender_name
              }}
            </strong>

            <p>{{ message.content }}</p>

            <small>
              {{ formatDateTime(message.created_at) }}
            </small>
          </div>
        </div>
      </div>

      <!-- New message indicator -->

      <button
        v-if="newGroupMessageCount > 0"
        type="button"
        class="group-chat-new-messages"
        @click="jumpToNewestGroupMessage"
      >
        {{
          newGroupMessageCount === 1
            ? "1 new message"
            : `${newGroupMessageCount} new messages`
        }}

        <i class="fa-solid fa-arrow-down" aria-hidden="true"></i>
      </button>
    </div>

    <!-- Message form -->

    <form class="group-chat-form" @submit.prevent="sendGroupMessage">
      <input
        v-model="groupMessageInput"
        type="text"
        placeholder="Type a group message..."
        autocomplete="off"
      />

      <button
        type="submit"
        class="button-primary"
        :disabled="!websocket.connected || !groupMessageInput.trim()"
      >
        Send
      </button>
    </form>
  </section>
</template>

<script setup>
import { nextTick, onUnmounted, ref, watch } from "vue";

import { apiRequest } from "../../services/api";
import { useAuthStore } from "../../stores/auth";
import { useWebSocketStore } from "../../stores/websocket";
import { formatDateTime } from "../../utils/date";

import UserAvatar from "../UserAvatar.vue";

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

const auth = useAuthStore();
const websocket = useWebSocketStore();

// Messages

const groupMessages = ref([]);
const groupMessageInput = ref("");

const loadingGroupMessages = ref(false);
const loadingOlderGroupMessages = ref(false);

const groupMessagesError = ref("");

const GROUP_MESSAGES_PAGE_SIZE = 30;

const hasMoreGroupMessages = ref(true);
const groupMessagesBeforeID = ref(0);

let groupMessagesRequestVersion = 0;

const messagesLoaded = ref(false);

// Scroll

const groupMessagesContainer = ref(null);

const GROUP_CHAT_TOP_THRESHOLD = 80;
const GROUP_CHAT_BOTTOM_THRESHOLD = 100;

const groupChatNearBottom = ref(true);

const newGroupMessageCount = ref(0);

// Scroll helpers

function isGroupChatNearBottom() {
  const container = groupMessagesContainer.value;

  if (!container) {
    return true;
  }

  const distanceFromBottom =
    container.scrollHeight - container.scrollTop - container.clientHeight;

  return distanceFromBottom <= GROUP_CHAT_BOTTOM_THRESHOLD;
}

function handleGroupChatScroll() {
  const container = groupMessagesContainer.value;

  if (!container) {
    return;
  }

  // Bottom

  groupChatNearBottom.value = isGroupChatNearBottom();

  if (groupChatNearBottom.value) {
    newGroupMessageCount.value = 0;
  }

  // Top

  if (
    container.scrollTop <= GROUP_CHAT_TOP_THRESHOLD &&
    hasMoreGroupMessages.value &&
    !loadingGroupMessages.value &&
    !loadingOlderGroupMessages.value
  ) {
    loadGroupMessages(false);
  }
}

async function scrollGroupChatToBottom() {
  await nextTick();

  const container = groupMessagesContainer.value;

  if (!container) {
    return;
  }

  container.scrollTop = container.scrollHeight;

  groupChatNearBottom.value = true;

  newGroupMessageCount.value = 0;
}

async function jumpToNewestGroupMessage() {
  await scrollGroupChatToBottom();
}

// Load messages

async function loadGroupMessages(reset = false) {
  if (
    !reset &&
    (loadingGroupMessages.value ||
      loadingOlderGroupMessages.value ||
      !hasMoreGroupMessages.value)
  ) {
    return;
  }

  if (reset) {
    groupMessagesRequestVersion += 1;

    groupMessages.value = [];

    groupMessagesBeforeID.value = 0;

    hasMoreGroupMessages.value = true;

    newGroupMessageCount.value = 0;

    groupMessagesError.value = "";
  }

  const requestVersion = groupMessagesRequestVersion;

  const loadingOlder = !reset && groupMessages.value.length > 0;

  const container = groupMessagesContainer.value;

  const previousScrollHeight = container ? container.scrollHeight : 0;

  const previousScrollTop = container ? container.scrollTop : 0;

  try {
    if (loadingOlder) {
      loadingOlderGroupMessages.value = true;
    } else {
      loadingGroupMessages.value = true;
    }

    groupMessagesError.value = "";

    const params = new URLSearchParams();

    params.set("limit", String(GROUP_MESSAGES_PAGE_SIZE));

    if (loadingOlder && groupMessagesBeforeID.value) {
      params.set("before_id", String(groupMessagesBeforeID.value));
    }

    const result = await apiRequest(
      `/groups/${props.groupId}/messages` + `?${params.toString()}`,
    );

    if (requestVersion !== groupMessagesRequestVersion) {
      return;
    }

    const incomingMessages = result.messages || [];

    if (reset) {
      groupMessages.value = incomingMessages;
    } else {
      const existingIDs = new Set(
        groupMessages.value.map((message) => message.id),
      );

      const olderMessages = incomingMessages.filter(
        (message) => !existingIDs.has(message.id),
      );

      groupMessages.value = [...olderMessages, ...groupMessages.value];
    }

    hasMoreGroupMessages.value = Boolean(result.has_more);

    groupMessagesBeforeID.value = result.next_before_id || 0;

    await nextTick();

    if (reset) {
      await scrollGroupChatToBottom();
    } else {
      const updatedContainer = groupMessagesContainer.value;

      if (updatedContainer) {
        const newScrollHeight = updatedContainer.scrollHeight;

        updatedContainer.scrollTop =
          previousScrollTop + (newScrollHeight - previousScrollHeight);
      }
    }
  } catch (err) {
    if (requestVersion === groupMessagesRequestVersion) {
      groupMessagesError.value = err.message;
    }
  } finally {
    if (requestVersion === groupMessagesRequestVersion) {
      loadingGroupMessages.value = false;

      loadingOlderGroupMessages.value = false;
    }
  }
}

// Incoming WebSocket message

function handleIncomingGroupMessage(message) {
  if (!message) {
    return;
  }

  if (message.group_id !== Number(props.groupId)) {
    return;
  }

  const alreadyExists = groupMessages.value.some(
    (existingMessage) => existingMessage.id === message.id,
  );

  if (alreadyExists) {
    return;
  }

  groupMessages.value.push(message);

  if (props.active) {
    scrollGroupChatToBottom();
  }
}

// Send

function sendGroupMessage() {
  groupMessagesError.value = "";

  const content = groupMessageInput.value.trim();

  if (!content) {
    return;
  }

  const sent = websocket.send({
    type: "group_message",
    group_id: Number(props.groupId),
    content,
  });

  if (!sent) {
    groupMessagesError.value = websocket.error;

    return;
  }

  groupMessageInput.value = "";
}

// Reset

function resetGroupChat() {
  groupMessagesRequestVersion += 1;

  messagesLoaded.value = false;

  groupMessages.value = [];
  groupMessageInput.value = "";

  loadingGroupMessages.value = false;
  loadingOlderGroupMessages.value = false;

  groupMessagesError.value = "";

  hasMoreGroupMessages.value = true;
  groupMessagesBeforeID.value = 0;

  newGroupMessageCount.value = 0;
  groupChatNearBottom.value = true;
}

// Active tab

watch(
  () => props.active,
  async (active) => {
    if (!active) {
      websocket.setActiveGroup(null);

      return;
    }

    websocket.setActiveGroup(Number(props.groupId));

    newGroupMessageCount.value = 0;

    if (!messagesLoaded.value) {
      await loadGroupMessages(true);

      messagesLoaded.value = true;
    } else {
      await scrollGroupChatToBottom();
    }
  },
  {
    immediate: true,
  },
);

// Different group

watch(
  () => props.groupId,
  async () => {
    resetGroupChat();

    if (!props.active) {
      websocket.setActiveGroup(null);

      return;
    }

    websocket.setActiveGroup(Number(props.groupId));

    await loadGroupMessages(true);

    messagesLoaded.value = true;
  },
);

// WebSocket events

watch(
  () => websocket.eventVersion,
  () => {
    const event = websocket.lastEvent;

    if (!event) {
      return;
    }

    if (event.type === "group_message") {
      handleIncomingGroupMessage(event.data);
    }
  },
);

onUnmounted(() => {
  groupMessagesRequestVersion += 1;

  websocket.setActiveGroup(null);
});
</script>

<style scoped>
/* Panel */

.group-tab-panel {
  padding: 22px;

  border: 1px solid var(--border-soft);
  border-radius: var(--radius-lg);

  background: var(--surface);

  box-shadow: var(--shadow-sm);
}

.group-panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 18px;

  margin-bottom: 20px;
}

.group-panel-heading h2 {
  margin-bottom: 4px;
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

/* Chat window */

.group-chat-shell {
  position: relative;
}

.group-chat-messages {
  height: 500px;

  overflow-y: auto;

  padding: 16px;

  border: 1px solid var(--border-soft);
  border-radius: var(--radius-lg);

  background: var(--bg-secondary);

  scroll-behavior: auto;
}

/* History */

.group-chat-loading,
.group-chat-loading-older {
  display: flex;
  align-items: center;
  justify-content: center;

  gap: 7px;

  padding: 12px;

  color: var(--text-muted);

  font-size: 11px;
}

.group-chat-start {
  margin: 2px 0 16px;

  color: var(--text-muted);

  font-size: 10px;

  text-align: center;
}

/* Empty */

.group-chat-empty {
  min-height: 380px;

  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;

  color: var(--text-muted);

  text-align: center;
}

.group-chat-empty-icon {
  width: 50px;
  height: 50px;

  display: grid;
  place-items: center;

  margin-bottom: 10px;

  border-radius: 50%;

  background: var(--primary-soft);

  font-size: 20px;
}

.group-chat-empty h3 {
  margin-bottom: 4px;

  color: var(--text);
}

/* Message rows */

.group-chat-message-row {
  display: flex;
  align-items: flex-end;

  gap: 8px;

  margin-bottom: 9px;
}

.group-chat-message-row.mine {
  justify-content: flex-end;
}

.group-chat-message-row.theirs {
  justify-content: flex-start;
}

.group-chat-message-avatar {
  width: 30px;
  height: 30px;

  flex: 0 0 30px;

  display: grid;
  place-items: center;

  overflow: hidden;

  border: 1px solid var(--primary-border);
  border-radius: 50%;

  background: var(--primary-soft);
  color: var(--primary);

  font-size: 9px;
  font-weight: 800;
}

/* Message */

.group-chat-message {
  max-width: 72%;

  padding: 9px 12px;

  border: 1px solid var(--border);
  border-radius: 14px;

  background: var(--surface-2);
}

.group-chat-message-row.mine .group-chat-message {
  border-color: var(--primary-border);

  background: var(--primary-soft);
}

.group-chat-message strong {
  display: block;

  margin-bottom: 3px;

  font-size: 10px;
}

.group-chat-message p {
  margin: 0 0 4px;

  white-space: pre-wrap;

  overflow-wrap: anywhere;

  line-height: 1.45;
}

.group-chat-message small {
  color: var(--text-muted);

  font-size: 9px;
}

/* New messages */

.group-chat-new-messages {
  position: absolute;

  left: 50%;
  bottom: 14px;

  z-index: 5;

  min-height: 32px;

  padding: 5px 12px;

  border: 1px solid var(--primary-border);

  border-radius: var(--radius-round);

  background: var(--primary);

  color: white;

  font-size: 11px;
  font-weight: 700;

  box-shadow: var(--shadow-md);

  transform: translateX(-50%);
}

/* Input */

.group-chat-form {
  display: flex;

  gap: 9px;

  margin-top: 8px;
}

.group-chat-form input {
  flex: 1;
}

/* Mobile */

@media (max-width: 700px) {
  .group-tab-panel {
    padding: 16px;
  }

  .group-panel-heading {
    flex-direction: column;
  }

  .group-chat-messages {
    height: min(55vh, 480px);
  }

  .group-chat-message {
    max-width: 88%;
  }

  .group-chat-form {
    align-items: stretch;

    flex-direction: column;
  }

  .group-chat-form button {
    width: 100%;
  }
}
</style>
