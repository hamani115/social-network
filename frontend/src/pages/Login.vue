<template>
  <main class="auth-page">
    <div class="auth-layout">
      <!-- INTRO -->
      <section class="auth-intro">
        <h1>Welcome back!</h1>
      </section>

      <!-- LOGIN CARD -->
      <section class="auth-card">
        <header class="auth-card-header">
          <h2>Log in</h2>
        </header>

        <!-- REGISTER SUCCESS -->
        <div v-if="registeredSuccessfully" class="auth-success">
          Account created successfully. You can now log in.
        </div>

        <form class="auth-form" @submit.prevent="submitLogin">
          <!-- EMAIL -->
          <div class="auth-field">
            <label for="login-email"> Email </label>

            <input
              id="login-email"
              v-model.trim="email"
              type="email"
              placeholder="example@gmail.com"
              autocomplete="email"
              required
              :disabled="submitting"
            />
          </div>

          <!-- PASSWORD -->
          <div class="auth-field">
            <label for="login-password"> Password </label>

            <div class="password-field">
              <input
                id="login-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Enter your password"
                autocomplete="current-password"
                required
                :disabled="submitting"
              />

              <button
                type="button"
                class="password-toggle"
                :aria-label="showPassword ? 'Hide password' : 'Show password'"
                :disabled="submitting"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? "Hide" : "Show" }}
              </button>
            </div>
          </div>

          <!-- ERROR -->

          <p v-if="error" class="auth-error" role="alert">
            {{ error }}
          </p>

          <!-- SUBMIT -->

          <button
            type="submit"
            class="button-primary auth-submit"
            :disabled="submitting || !email.trim() || !password"
          >
            <span v-if="submitting" class="loading-spinner"></span>

            {{ submitting ? "Logging in..." : "Log in" }}
          </button>
        </form>

        <footer class="auth-card-footer">
          <span> Don't have an account? </span>

          <router-link to="/register"> Create one </router-link>
        </footer>
      </section>
    </div>
  </main>
</template>

<script setup>
import { computed, ref } from "vue";

import { useRoute, useRouter } from "vue-router";

import { useAuthStore } from "../stores/auth";

const route = useRoute();

const router = useRouter();

const auth = useAuthStore();

const email = ref("");

const password = ref("");

const error = ref("");

const submitting = ref(false);

const showPassword = ref(false);

const registeredSuccessfully = computed(() => route.query.registered === "1");


// LOGIN


async function submitLogin() {
  if (submitting.value) {
    return;
  }

  try {
    submitting.value = true;

    error.value = "";

    await auth.login(email.value.trim(), password.value);

    await router.push("/");
  } catch (err) {
    error.value = err.message || "Could not log in";
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
/* PAGE */

.auth-page {
  width: min(940px, calc(100% - 32px));

  min-height: calc(100vh - 58px);

  display: flex;
  align-items: center;

  padding: 48px 0 72px;
}

.auth-layout {
  width: 100%;

  display: grid;

  grid-template-columns: minmax(0, 1fr) minmax(340px, 420px);

  align-items: center;

  gap: 70px;
}

/* INTRO */

.auth-intro {
  max-width: 470px;
}

.auth-intro h1 {
  margin-bottom: 15px;

  font-size: clamp(2.1rem, 5vw, 3.4rem);

  line-height: 1.04;
}


/* CARD */

.auth-card {
  padding: 26px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-md);
}

.auth-card-header {
  margin-bottom: 22px;
}

.auth-card-header h2 {
  margin-bottom: 5px;

  font-size: 1.45rem;
}

/* FORM */

.auth-form {
  display: grid;

  gap: 17px;
}

.auth-field {
  display: grid;

  gap: 7px;
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

.auth-submit {
  width: 100%;

  margin-top: 3px;
}

/* FEEDBACK */

.auth-error,
.auth-success {
  padding: 10px 12px;

  border-radius: var(--radius-md);

  font-size: 12px;
}

.auth-error {
  margin: 0;

  border: 1px solid rgba(255, 95, 109, 0.22);

  background: var(--danger-soft);

  color: var(--danger);
}

.auth-success {
  margin-bottom: 18px;

  border: 1px solid rgba(54, 201, 143, 0.25);

  background: var(--success-soft);

  color: var(--success);
}

/* FOOTER */

.auth-card-footer {
  display: flex;
  justify-content: center;

  gap: 5px;

  margin-top: 20px;
  padding-top: 18px;

  border-top: 1px solid var(--border-soft);

  color: var(--text-muted);

  font-size: 12px;
}

/* MOBILE */

@media (max-width: 760px) {
  .auth-page {
    min-height: auto;

    padding-top: 32px;
  }

  .auth-layout {
    grid-template-columns: 1fr;

    gap: 28px;
  }

  .auth-intro {
    max-width: 560px;
  }

  .auth-intro h1 {
    font-size: 2.2rem;
  }

  .auth-card {
    padding: 20px;
  }
}
</style>
