<template>
  <main class="error-page">
    <section class="error-card">
      <div class="error-code">Error</div>

      <div class="error-icon">
        <i class="fa-solid fa-circle-exclamation" aria-hidden="true"></i>
      </div>

      <h1>Something went wrong</h1>

      <p>
        The server encountered an unexpected problem while processing your
        request.
      </p>

      <div class="error-actions">
        <button type="button" class="button button-ghost" @click="goBack">
          Go back
        </button>

        <button type="button" class="button-primary" @click="retry">
          Try again
        </button>
      </div>
    </section>
  </main>
</template>

<script setup>
import { useRoute, useRouter } from "vue-router";
const route = useRoute();
const router = useRouter();

function retry() {
  window.location.reload();
}

async function goBack() {
  const from = typeof route.query.from === "string" ? route.query.from : "/";
  await router.replace(from);
}
</script>

<style scoped>
.error-page {
  min-height: calc(100vh - 58px);

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 48px 16px 80px;
}

.error-card {
  width: min(480px, 100%);

  padding: 42px 34px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-md);

  text-align: center;
}

.error-code {
  margin-bottom: 17px;

  color: var(--text-muted);

  font-size: 11px;
  font-weight: 750;

  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.error-icon {
  width: 58px;
  height: 58px;

  display: grid;
  place-items: center;

  margin: 0 auto 18px;

  border: 1px solid var(--primary-border);

  border-radius: 50%;

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 22px;
  font-weight: 800;
}

.error-card h1 {
  margin-bottom: 10px;
}

.error-card p {
  max-width: 370px;

  margin: 0 auto 24px;

  color: var(--text-secondary);

  line-height: 1.65;
}

.error-actions {
  display: flex;
  justify-content: center;

  gap: 8px;
}

@media (max-width: 560px) {
  .error-card {
    padding: 34px 21px;
  }

  .error-actions {
    flex-direction: column;
  }

  .error-actions button {
    width: 100%;
  }
}
</style>
