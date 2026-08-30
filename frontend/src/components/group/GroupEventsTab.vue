<template>
  <section v-show="active" class="group-tab-panel group-events-panel">
    <div class="group-panel-heading">
      <div>
        <h2>Events</h2>
      </div>

      <button type="button" class="button-primary" @click="openEventModal">
        <i class="fa-solid fa-plus" aria-hidden="true"></i>
        Create event
      </button>
    </div>

    <!-- Upcoming + Past Events-->
    <div class="event-scope-tabs">
      <button
        type="button"
        class="event-scope-button"
        :class="{ active: eventScope === 'upcoming' }"
        @click="changeEventScope('upcoming')"
      >
        Upcoming
      </button>

      <button
        type="button"
        class="event-scope-button"
        :class="{ active: eventScope === 'past' }"
        @click="changeEventScope('past')"
      >
        Past
      </button>
    </div>

    <div v-if="loadingGroupEvents" class="group-section-state">
      <span class="loading-spinner"></span>
      Loading events...
    </div>

    <p v-else-if="groupEventsError" class="group-page-error">
      {{ groupEventsError }}
    </p>

    <div v-else-if="groupEvents.length === 0" class="group-events-empty">
      <div class="group-events-empty-icon">
        <i class="fa-solid fa-calendar-days" aria-hidden="true"></i>
      </div>

      <h3>
        {{
          eventScope === "upcoming" ? "No upcoming events" : "No past events"
        }}
      </h3>
    </div>

    <!-- Events -->
    <div v-else class="group-events-list">
      <article
        v-for="event in groupEvents"
        :key="event.id"
        class="group-event-card"
      >
        <div class="group-event-main">
          <div class="group-event-date-block">
            <span class="group-event-date-icon">
              <i class="fa-solid fa-calendar-days" aria-hidden="true"></i>
            </span>
          </div>

          <div class="group-event-content">
            <div class="group-event-title-row">
              <div>
                <h3>{{ event.title }}</h3>

                <div class="group-event-created-by">
                  <UserAvatar
                    :avatar-path="event.creator_avatar_path"
                    :name="event.creator_name"
                    class="group-event-creator-avatar"
                  />

                  <span>
                    Created by
                    <strong>{{ event.creator_name }}</strong>
                  </span>
                </div>
              </div>

              <span
                class="group-event-scope-badge"
                :class="{ past: eventScope === 'past' }"
              >
                {{ eventScope === "upcoming" ? "Upcoming" : "Past" }}
              </span>
            </div>

            <p v-if="event.description" class="group-event-description">
              {{ event.description }}
            </p>

            <div class="group-event-time">
              <strong>{{ formatEventDate(event.event_time) }}</strong>
              <span>{{ formatEventClock(event.event_time) }}</span>
            </div>

            <div class="group-event-response-summary">
              <span>
                <strong>{{ event.going_count }}</strong>
                going
              </span>

              <span>
                <strong>{{ event.not_going_count }}</strong>
                not going
              </span>
            </div>

            <div class="group-event-response-area">
              <span class="group-event-response-label">
                Your response:
                <strong>
                  {{
                    event.my_response === "going"
                      ? "Going"
                      : event.my_response === "not_going"
                        ? "Not going"
                        : "Not answered"
                  }}
                </strong>
              </span>

              <div class="group-event-actions">
                <button
                  type="button"
                  class="button event-going-button"
                  :class="{ selected: event.my_response === 'going' }"
                  :disabled="
                    changingEventResponseId === event.id ||
                    event.my_response === 'going'
                  "
                  @click="respondToEvent(event.id, 'going')"
                >
                  <i class="fa-solid fa-check" aria-hidden="true"></i>
                  Going
                </button>

                <button
                  type="button"
                  class="button event-not-going-button"
                  :class="{ selected: event.my_response === 'not_going' }"
                  :disabled="
                    changingEventResponseId === event.id ||
                    event.my_response === 'not_going'
                  "
                  @click="respondToEvent(event.id, 'not_going')"
                >
                  Not going
                </button>
              </div>
            </div>
          </div>
        </div>
      </article>

      <!-- Pagination -->
      <div class="group-events-pagination">
        <div v-if="loadingMoreGroupEvents" class="group-events-loading-more">
          <span class="loading-spinner"></span>
          Loading more events...
        </div>

        <div
          v-else-if="groupEventsLoadMoreError"
          class="group-events-load-error"
        >
          <span>{{ groupEventsLoadMoreError }}</span>

          <button
            type="button"
            class="button button-ghost button-small"
            @click="loadGroupEvents()"
          >
            Try again
          </button>
        </div>

        <div
          v-else-if="hasMoreGroupEvents"
          ref="groupEventsLoadTrigger"
          class="group-events-load-trigger"
          aria-hidden="true"
        ></div>

        <p v-else class="group-events-end">
          {{
            eventScope === "upcoming"
              ? "No more upcoming events"
              : "No more past events"
          }}
        </p>
      </div>
    </div>
  </section>

  <Teleport to="body">
    <div
      v-if="eventModalOpen"
      class="event-modal-overlay"
      @click.self="closeEventModal"
    >
      <section
        class="event-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-event-title"
      >
        <header class="event-modal-header">
          <div>
            <h2 id="create-event-title">Create event</h2>
          </div>

          <button
            type="button"
            class="event-modal-close"
            aria-label="Close create event"
            :disabled="creatingGroupEvent"
            @click="closeEventModal"
          >
            <i class="fa-solid fa-xmark" aria-hidden="true"></i>
          </button>
        </header>

        <form @submit.prevent="createGroupEvent">
          <div class="event-modal-body">
            <div>
              <label for="modal-event-title">Event title</label>

              <input
                id="modal-event-title"
                v-model="newGroupEvent.title"
                type="text"
                placeholder="Event title"
                required
              />
            </div>

            <div>
              <label for="modal-event-description">Description</label>

              <textarea
                id="modal-event-description"
                v-model="newGroupEvent.description"
                rows="5"
                placeholder="Tell members about the event..."
              ></textarea>
            </div>

            <div>
              <label for="modal-event-time">Date and time</label>

              <input
                id="modal-event-time"
                v-model="newGroupEvent.event_time"
                type="datetime-local"
                required
              />
            </div>

            <p v-if="groupEventsError" class="group-page-error">
              {{ groupEventsError }}
            </p>
          </div>

          <footer class="event-modal-footer">
            <button
              type="button"
              class="button button-ghost"
              :disabled="creatingGroupEvent"
              @click="closeEventModal"
            >
              Cancel
            </button>

            <button
              type="submit"
              class="button-primary"
              :disabled="creatingGroupEvent"
            >
              {{ creatingGroupEvent ? "Creating..." : "Create event" }}
            </button>
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
import {
  currentLocalDateTimeForBackend,
  formatEventClock,
  formatEventDate,
  formatEventTimeForBackend,
} from "../../utils/date";
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
const groupEvents = ref([]);
const loadingGroupEvents = ref(false);
const groupEventsError = ref("");
const GROUP_EVENTS_PAGE_SIZE = 10;
const eventScope = ref("upcoming");
const eventOffset = ref(0);
const hasMoreGroupEvents = ref(true);
const loadingMoreGroupEvents = ref(false);
const groupEventsLoadMoreError = ref("");
const groupEventsLoadTrigger = ref(null);
let groupEventsObserver = null;
let groupEventsRequestVersion = 0;
const eventsLoaded = ref(false);
const newGroupEvent = ref({
  title: "",
  description: "",
  event_time: "",
});
const eventModalOpen = ref(false);
const creatingGroupEvent = ref(false);
const changingEventResponseId = ref(null);

async function loadGroupEvents(reset = false) {
  if (
    !reset &&
    (loadingGroupEvents.value ||
      loadingMoreGroupEvents.value ||
      !hasMoreGroupEvents.value)
  ) {
    return;
  }
  if (reset) {
    groupEventsRequestVersion += 1;
    groupEvents.value = [];
    eventOffset.value = 0;
    hasMoreGroupEvents.value = true;
    groupEventsLoadMoreError.value = "";
  }
  const requestVersion = groupEventsRequestVersion;
  const initialLoad = eventOffset.value === 0;
  try {
    if (initialLoad) {
      loadingGroupEvents.value = true;
      groupEventsError.value = "";
    } else {
      loadingMoreGroupEvents.value = true;
      groupEventsLoadMoreError.value = "";
    }
    const params = new URLSearchParams();
    params.set("scope", eventScope.value);
    params.set("limit", String(GROUP_EVENTS_PAGE_SIZE));
    params.set("offset", String(eventOffset.value));
    params.set("now", currentLocalDateTimeForBackend());
    const result = await apiRequest(
      `/groups/${props.groupId}/events?${params.toString()}`,
    );
    if (requestVersion !== groupEventsRequestVersion) {
      return;
    }
    const incomingEvents = result.events || [];
    if (reset) {
      groupEvents.value = incomingEvents;
    } else {
      const existingIDs = new Set(groupEvents.value.map((event) => event.id));
      groupEvents.value.push(
        ...incomingEvents.filter((event) => !existingIDs.has(event.id)),
      );
    }
    eventOffset.value =
      result.next_offset ?? eventOffset.value + incomingEvents.length;
    hasMoreGroupEvents.value = Boolean(result.has_more);
  } catch (err) {
    if (requestVersion !== groupEventsRequestVersion) {
      return;
    }
    if (initialLoad) {
      groupEventsError.value = err.message;
    } else {
      groupEventsLoadMoreError.value = err.message;
    }
  } finally {
    if (requestVersion === groupEventsRequestVersion) {
      loadingGroupEvents.value = false;
      loadingMoreGroupEvents.value = false;
    }
  }
}

async function changeEventScope(scope) {
  if (scope !== "upcoming" && scope !== "past") {
    return;
  }
  if (eventScope.value === scope) {
    return;
  }
  eventScope.value = scope;
  await loadGroupEvents(true);
}

function observeGroupEventsTrigger(element) {
  if (groupEventsObserver) {
    groupEventsObserver.disconnect();
    groupEventsObserver = null;
  }
  if (!element) {
    return;
  }
  groupEventsObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (
        entry.isIntersecting &&
        props.active &&
        hasMoreGroupEvents.value &&
        !loadingGroupEvents.value &&
        !loadingMoreGroupEvents.value
      ) {
        loadGroupEvents();
      }
    },
    {
      root: null,
      rootMargin: "300px 0px",
      threshold: 0,
    },
  );
  groupEventsObserver.observe(element);
}

// Create event

function openEventModal() {
  groupEventsError.value = "";
  eventModalOpen.value = true;
}

function resetEventForm() {
  newGroupEvent.value = {
    title: "",
    description: "",
    event_time: "",
  };
}

function closeEventModal() {
  if (creatingGroupEvent.value) {
    return;
  }
  eventModalOpen.value = false;
  groupEventsError.value = "";
  resetEventForm();
}

async function createGroupEvent() {
  try {
    creatingGroupEvent.value = true;
    groupEventsError.value = "";
    const eventTime = newGroupEvent.value.event_time;
    const eventDate = new Date(eventTime);
    const createdScope = eventDate < new Date() ? "past" : "upcoming";
    await apiRequest(`/groups/${props.groupId}/events`, {
      method: "POST",
      body: JSON.stringify({
        title: newGroupEvent.value.title,
        description: newGroupEvent.value.description,
        event_time: formatEventTimeForBackend(eventTime),
      }),
    });
    resetEventForm();
    eventModalOpen.value = false;
    eventScope.value = createdScope;
    await loadGroupEvents(true);
  } catch (err) {
    groupEventsError.value = err.message;
  } finally {
    creatingGroupEvent.value = false;
  }
}

async function respondToEvent(eventId, response) {
  try {
    changingEventResponseId.value = eventId;
    groupEventsError.value = "";
    const event = groupEvents.value.find((item) => item.id === eventId);
    if (!event) {
      return;
    }
    const previousResponse = event.my_response;
    if (previousResponse === response) {
      return;
    }
    const action = response === "going" ? "going" : "not-going";
    await apiRequest(`/groups/${props.groupId}/events/${eventId}/${action}`, {
      method: "POST",
    });
    if (previousResponse === "going") {
      event.going_count = Math.max(0, event.going_count - 1);
    }
    if (previousResponse === "not_going") {
      event.not_going_count = Math.max(0, event.not_going_count - 1);
    }
    if (response === "going") {
      event.going_count += 1;
    } else {
      event.not_going_count += 1;
    }
    event.my_response = response;
  } catch (err) {
    groupEventsError.value = err.message;
  } finally {
    changingEventResponseId.value = null;
  }
}

function resetEvents() {
  groupEventsRequestVersion += 1;
  eventsLoaded.value = false;
  groupEvents.value = [];
  loadingGroupEvents.value = false;
  groupEventsError.value = "";
  eventScope.value = "upcoming";
  eventOffset.value = 0;
  hasMoreGroupEvents.value = true;
  loadingMoreGroupEvents.value = false;
  groupEventsLoadMoreError.value = "";
  eventModalOpen.value = false;
  creatingGroupEvent.value = false;
  changingEventResponseId.value = null;
  resetEventForm();
}

watch(
  () => props.active,
  async (active) => {
    if (active && !eventsLoaded.value) {
      await loadGroupEvents(true);
      eventsLoaded.value = true;
    }
  },
  { immediate: true },
);
watch(
  () => props.groupId,
  async () => {
    resetEvents();
    if (props.active) {
      await loadGroupEvents(true);
      eventsLoaded.value = true;
    }
  },
);
watch(groupEventsLoadTrigger, (element) => {
  observeGroupEventsTrigger(element);
});
onUnmounted(() => {
  if (groupEventsObserver) {
    groupEventsObserver.disconnect();
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

.event-scope-tabs {
  width: fit-content;

  display: flex;

  gap: 4px;

  margin-bottom: 18px;
  padding: 4px;

  border: 1px solid var(--border-soft);
  border-radius: var(--radius-round);

  background: var(--bg-secondary);
}

.event-scope-button {
  min-height: 34px;

  padding: 5px 13px;

  border: 0;
  border-radius: var(--radius-round);

  background: transparent;

  color: var(--text-muted);

  font-size: 12px;
  font-weight: 650;
}

.event-scope-button:hover {
  background: var(--surface-2);

  color: var(--text);
}

.event-scope-button.active {
  background: var(--primary-soft);

  color: var(--primary);
}

.group-events-list {
  display: grid;

  gap: 12px;
}

.group-event-card {
  margin: 0;

  padding: 18px;

  background: var(--bg-secondary);
}

.group-event-main {
  display: flex;

  gap: 15px;
}

.group-event-date-block {
  width: 48px;
  height: 48px;

  flex: 0 0 48px;

  display: grid;
  place-items: center;

  border: 1px solid var(--primary-border);
  border-radius: 13px;

  background: var(--primary-soft);
}

.group-event-date-icon {
  font-size: 19px;
}

.group-event-content {
  min-width: 0;

  flex: 1;
}

.group-event-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 14px;

  margin-bottom: 10px;
}

.group-event-title-row h3 {
  margin-bottom: 3px;
}

.group-event-created-by {
  display: flex;
  align-items: center;
  gap: 7px;

  margin: 0;

  color: var(--text-muted);

  font-size: 11px;
}

.group-event-creator-avatar {
  width: 24px;
  height: 24px;

  flex: 0 0 24px;

  display: grid;
  place-items: center;

  overflow: hidden;

  border: 1px solid var(--primary-border);
  border-radius: 50%;

  background: var(--primary-soft);
  color: var(--primary);

  font-size: 8px;
  font-weight: 800;
}

.group-event-scope-badge {
  flex-shrink: 0;

  padding: 3px 8px;

  border: 1px solid rgba(54, 201, 143, 0.25);
  border-radius: var(--radius-round);

  background: rgba(54, 201, 143, 0.08);

  color: var(--success);

  font-size: 10px;
  font-weight: 700;
}

.group-event-scope-badge.past {
  border-color: var(--border);

  background: var(--surface-2);

  color: var(--text-muted);
}

.group-event-description {
  color: var(--text-secondary);

  line-height: 1.55;
}

.group-event-time {
  display: flex;
  align-items: baseline;

  flex-wrap: wrap;

  gap: 8px;

  margin: 12px 0;

  color: var(--text);
}

.group-event-time span {
  color: var(--text-muted);

  font-size: 12px;
}

.group-event-response-summary {
  display: flex;

  flex-wrap: wrap;

  gap: 14px;

  padding: 10px 0;

  border-top: 1px solid var(--border-soft);

  color: var(--text-muted);

  font-size: 12px;
}

.group-event-response-summary strong {
  color: var(--text-secondary);
}

.group-event-response-area {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  padding-top: 11px;

  border-top: 1px solid var(--border-soft);
}

.group-event-response-label {
  color: var(--text-muted);

  font-size: 12px;
}

.group-event-response-label strong {
  color: var(--text-secondary);
}

.group-event-actions {
  display: flex;

  gap: 7px;
}

.event-going-button.selected {
  border-color: rgba(54, 201, 143, 0.3);

  background: rgba(54, 201, 143, 0.1);

  color: var(--success);
}

.event-not-going-button.selected {
  border-color: rgba(255, 95, 109, 0.28);

  background: var(--danger-soft);

  color: var(--danger);
}

.group-events-empty {
  padding: 42px 20px;

  text-align: center;
}

.group-events-empty-icon {
  width: 50px;
  height: 50px;

  display: grid;
  place-items: center;

  margin: 0 auto 12px;

  border-radius: 50%;

  background: var(--primary-soft);

  font-size: 20px;
}

.group-events-empty h3 {
  margin-bottom: 5px;
}

.group-events-pagination {
  padding: 16px 0 2px;

  text-align: center;
}

.group-events-loading-more,
.group-events-load-error {
  min-height: 42px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 8px;

  color: var(--text-muted);

  font-size: 12px;
}

.group-events-load-error {
  color: var(--danger);
}

.group-events-load-trigger {
  height: 2px;
}

.group-events-end {
  color: var(--text-muted);

  font-size: 12px;
}

.event-modal-overlay {
  position: fixed;
  inset: 0;

  z-index: 2200;

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;

  background: rgba(3, 7, 18, 0.76);

  backdrop-filter: blur(5px);
}

.event-modal {
  width: min(600px, 100%);
  max-height: calc(100vh - 48px);

  overflow-y: auto;

  border: 1px solid var(--border);
  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.55);
}

.event-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 18px;

  padding: 20px 22px 10px;

  border-bottom: 1px solid var(--border-soft);
}

.event-modal-header h2 {
  margin: 0;
}

.event-modal-close {
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

.event-modal-close:hover {
  background: var(--surface-2);

  color: var(--text);
}

.event-modal-body {
  display: grid;

  gap: 17px;

  padding: 22px;
}

.event-modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;

  gap: 8px;

  padding: 16px 22px;

  border-top: 1px solid var(--border-soft);
}

@media (max-width: 700px) {
  .group-tab-panel {
    padding: 16px;
  }

  .group-panel-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .group-event-main {
    flex-direction: column;
  }

  .group-event-title-row {
    flex-direction: column;
  }

  .group-event-response-area {
    align-items: stretch;

    flex-direction: column;
  }

  .group-event-actions {
    width: 100%;
  }

  .group-event-actions button {
    flex: 1;
  }

  .event-modal-overlay {
    padding: 12px;
  }

  .event-modal {
    max-height: calc(100vh - 24px);
  }

  .event-modal-footer button {
    flex: 1;
  }
}
</style>
