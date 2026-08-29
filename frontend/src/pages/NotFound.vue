<template>
  <main class="not-found-page">
    <section class="not-found-card">
      <div class="not-found-code">404</div>

      <h1>Page not found</h1>

      <p>The page you're looking for doesn't exist or may have been moved</p>

      <router-link
        class="button button-primary"
        :to="auth.user ? '/' : '/login'"
      >
        {{ auth.user ? "Back to Feed" : "Go to Login" }}
      </router-link>
    </section>
  </main>
</template>

<script setup>
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
</script>

<style scoped>
.not-found-page {
  min-height: calc(100vh - 58px);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 48px 16px 80px;
}

.not-found-card {
  width: min(480px, 100%);

  padding: 42px 34px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-md);

  text-align: center;
}

.not-found-code {
  margin-bottom: 18px;

  color: var(--primary);

  font-size: clamp(4.5rem, 10vw, 6.5rem);
  font-weight: 850;

  line-height: 1;

  letter-spacing: -0.04em;
}

.not-found-card h1 {
  margin-bottom: 10px;
}

.not-found-card p {
  max-width: 350px;

  margin: 0 auto 24px;

  color: var(--text-secondary);

  line-height: 1.65;
}

@media (max-width: 560px) {
  .not-found-card {
    padding: 34px 21px;
  }
}
</style>
