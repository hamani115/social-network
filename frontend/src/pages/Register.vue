<template>
  <main class="register-page">
    <div class="register-layout">
      <!-- INTRO -->

      <aside class="register-intro">
        <h1>Create your account!</h1>
      </aside>

      <!-- CARD -->

      <section class="register-card">
        <header class="register-card-header">
          <h2>Sign up</h2>
        </header>

        <form class="register-form" @submit.prevent="submitRegister">
          <!-- NAME -->

          <div class="register-name-grid">
            <div class="register-field">
              <label for="register-first-name"> First name * </label>

              <input
                id="register-first-name"
                v-model.trim="form.first_name"
                type="text"
                placeholder="First name"
                autocomplete="given-name"
                required
                :disabled="submitting"
              />
            </div>

            <div class="register-field">
              <label for="register-last-name"> Last name * </label>

              <input
                id="register-last-name"
                v-model.trim="form.last_name"
                type="text"
                placeholder="Last name"
                autocomplete="family-name"
                required
                :disabled="submitting"
              />
            </div>
          </div>

          <!-- EMAIL -->

          <div class="register-field">
            <label for="register-email"> Email * </label>

            <input
              id="register-email"
              v-model.trim="form.email"
              type="email"
              placeholder="example@gmail.com"
              autocomplete="email"
              required
              :disabled="submitting"
            />
          </div>

          <!-- PASSWORD -->

          <div class="register-field">
            <label for="register-password"> Password * </label>

            <div class="password-field">
              <input
                id="register-password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Create a password"
                autocomplete="new-password"
                required
                :disabled="submitting"
              />

              <button
                type="button"
                class="password-toggle"
                :disabled="submitting"
                :aria-label="showPassword ? 'Hide password' : 'Show password'"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? "Hide" : "Show" }}
              </button>
            </div>
          </div>

          <!-- DOB -->

          <div class="register-field">
            <label for="register-date-of-birth"> Date of birth * </label>

            <input
              id="register-date-of-birth"
              v-model="form.date_of_birth"
              type="date"
              :max="today"
              autocomplete="bday"
              required
              :disabled="submitting"
            />
          </div>

          <!-- NICKNAME -->

          <div class="register-field">
            <label for="register-nickname">
              Nickname
              <span> Optional </span>
            </label>

            <input
              id="register-nickname"
              v-model.trim="form.nickname"
              type="text"
              placeholder="What would you like people to call you?"
              :disabled="submitting"
            />
          </div>

          <!-- ABOUT -->

          <div class="register-field">
            <label for="register-about">
              About Me
              <span> Optional </span>
            </label>

            <textarea
              id="register-about"
              v-model.trim="form.about_me"
              rows="4"
              placeholder="Tell people about yourself..."
              :disabled="submitting"
            ></textarea>
          </div>

          <!-- AVATAR -->

          <div class="register-field">
            <label for="register-avatar">
              Avatar
              <span> Optional </span>
            </label>

            <div class="avatar-picker">
              <UserAvatar
                :avatar-path="avatarPreview"
                :name="`${form.first_name} ${form.last_name}`"
                class="avatar-preview"
              />

              <div class="avatar-picker-content">
                <label
                  for="register-avatar"
                  class="button button-ghost button-small avatar-button"
                >
                  Choose image
                </label>

                <input
                  id="register-avatar"
                  class="visually-hidden"
                  type="file"
                  accept="
                    image/jpeg,
                    image/png,
                    image/gif
                  "
                  :disabled="submitting"
                  @change="handleAvatarChange"
                />

                <span class="avatar-file-name">
                  {{ avatar ? avatar.name : "JPG, PNG or GIF" }}
                </span>
              </div>
            </div>

            <button
              v-if="avatar"
              type="button"
              class="remove-avatar-button"
              :disabled="submitting"
              @click="removeAvatar"
            >
              Remove selected image
            </button>
          </div>

          

          <p v-if="error" class="register-error" role="alert">
            {{ error }}
          </p>

          <!-- SUBMIT -->

          <button
            type="submit"
            class="button-primary register-submit"
            :disabled="submitting"
          >
            <span v-if="submitting" class="loading-spinner"></span>

            {{ submitting ? "Creating account..." : "Create account" }}
          </button>
        </form>

        <footer class="register-footer">
          <span> Already have an account? </span>

          <router-link to="/login"> Log in </router-link>
        </footer>
      </section>
    </div>
  </main>
</template>

<script setup>
import UserAvatar from "../components/UserAvatar.vue";

import { onUnmounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { todayDateInput } from "../utils/date";

const router = useRouter();
const auth = useAuthStore();
const error = ref("");
const submitting = ref(false);
const showPassword = ref(false);

const form = reactive({
  email: "",
  password: "",
  first_name: "",
  last_name: "",
  date_of_birth: "",
  nickname: "",
  about_me: "",
});

const avatar = ref(null);
const avatarPreview = ref("");

const today = todayDateInput();

function clearAvatarPreview() {
  if (avatarPreview.value) {
    URL.revokeObjectURL(avatarPreview.value);

    avatarPreview.value = "";
  }
}

function handleAvatarChange(event) {
  const file = event.target.files?.[0];

  error.value = "";

  clearAvatarPreview();

  if (!file) {
    avatar.value = null;

    return;
  }

  const allowedTypes = ["image/jpeg", "image/png", "image/gif"];

  if (!allowedTypes.includes(file.type)) {
    avatar.value = null;

    event.target.value = "";

    error.value = "Avatar must be a JPG, PNG or GIF image";

    return;
  }

  avatar.value = file;

  avatarPreview.value = URL.createObjectURL(file);
}

function removeAvatar() {
  avatar.value = null;

  clearAvatarPreview();

  const input = document.getElementById("register-avatar");

  if (input) {
    input.value = "";
  }
}

async function submitRegister() {
  if (submitting.value) {
    return;
  }

  try {
    submitting.value = true;

    error.value = "";

    await auth.register(form, avatar.value);

    await router.push({
      path: "/login",

      query: {
        registered: "1",
      },
    });
  } catch (err) {
    error.value = err.message || "Could not create account";
  } finally {
    submitting.value = false;
  }
}

onUnmounted(() => {
  clearAvatarPreview();
});
</script>

<style scoped>
.register-page {
  width: min(1040px, calc(100% - 32px));

  padding: 42px 0 72px;
}

.register-layout {
  display: grid;

  grid-template-columns: minmax(250px, 0.7fr) minmax(480px, 1.3fr);

  align-items: start;

  gap: 55px;
}

.register-intro {
  position: sticky;

  top: 100px;

  padding-top: 24px;
}

.register-intro h1 {
  margin-bottom: 14px;

  font-size: clamp(2rem, 4vw, 3rem);

  line-height: 1.05;
}

.register-card {
  padding: 27px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-md);
}

.register-card-header {
  margin-bottom: 23px;
}

.register-card-header h2 {
  margin-bottom: 5px;

  font-size: 1.45rem;
}

.register-form {
  display: grid;

  gap: 17px;
}

.register-name-grid {
  display: grid;

  grid-template-columns: repeat(2, minmax(0, 1fr));

  gap: 13px;
}

.register-field {
  display: grid;

  gap: 7px;
}

.register-field label span {
  margin-left: 4px;

  color: var(--text-muted);

  font-size: 10px;
  font-weight: 500;
}

.password-field {
  position: relative;
}

.password-field input {
  padding-right: 67px;
}

.password-toggle {
  position: absolute;

  right: 5px;
  top: 50%;

  min-height: 32px;

  padding: 4px 9px;

  border: 0;

  background: transparent;

  color: var(--primary);

  font-size: 11px;

  transform: translateY(-50%);
}

.password-toggle:hover {
  background: var(--primary-soft);
}

.avatar-picker {
  display: flex;
  align-items: center;

  gap: 13px;

  padding: 12px;

  border: 1px solid var(--border);

  border-radius: var(--radius-lg);

  background: var(--bg-secondary);
}

.avatar-preview {
  width: 58px;
  height: 58px;

  flex: 0 0 58px;

  display: grid;
  place-items: center;

  overflow: hidden;

  border: 1px solid var(--primary-border);

  border-radius: 50%;

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 16px;
  font-weight: 800;
}

.avatar-picker-content {
  min-width: 0;

  display: flex;
  align-items: center;
  flex-wrap: wrap;

  gap: 8px;
}

.avatar-button {
  cursor: pointer;
}

.avatar-file-name {
  max-width: 240px;

  overflow: hidden;

  color: var(--text-muted);

  font-size: 10px;

  text-overflow: ellipsis;
  white-space: nowrap;
}

.remove-avatar-button {
  width: fit-content;

  min-height: auto;

  padding: 2px 0;

  border: 0;

  background: transparent;

  color: var(--danger);

  font-size: 10px;
}

.remove-avatar-button:hover {
  background: transparent;

  color: var(--danger-hover);
}

.register-error {
  margin: 0;

  padding: 10px 12px;

  border: 1px solid rgba(255, 95, 109, 0.22);

  border-radius: var(--radius-md);

  background: var(--danger-soft);

  color: var(--danger);

  font-size: 12px;
}

.register-submit {
  width: 100%;

  margin-top: 3px;
}

.register-footer {
  display: flex;
  justify-content: center;

  gap: 5px;

  margin-top: 21px;
  padding-top: 18px;

  border-top: 1px solid var(--border-soft);

  color: var(--text-muted);

  font-size: 12px;
}

@media (max-width: 800px) {
  .register-layout {
    grid-template-columns: 1fr;

    gap: 26px;
  }

  .register-intro {
    position: static;

    padding-top: 0;
  }
}

@media (max-width: 560px) {
  .register-page {
    padding-top: 28px;
  }

  .register-card {
    padding: 19px;
  }

  .register-name-grid {
    grid-template-columns: 1fr;
  }

  .avatar-picker {
    align-items: flex-start;
  }

  .avatar-picker-content {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
