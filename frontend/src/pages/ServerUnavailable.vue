<template>
  <main class="error-page">
    <section class="error-card">
      <div class="error-code">Offline</div>

      <div class="error-icon">
        <i class="fa-solid fa-triangle-exclamation" aria-hidden="true"></i>
      </div>

      <h1>Server unavailable</h1>

      <p>Could not connect to the server</p>

      <button type="button" class="button-primary" @click="retry">
        Try again
      </button>
    </section>
  </main>
</template>

<script setup>
function retry() {
  window.location.replace("/");
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

  border: 1px solid rgba(255, 95, 109, 0.25);

  border-radius: 50%;

  background: var(--danger-soft);

  color: var(--danger);

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

@media (max-width: 560px) {
  .error-card {
    padding: 34px 21px;
  }
}
</style>
