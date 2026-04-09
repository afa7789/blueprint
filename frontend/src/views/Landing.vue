<script setup lang="ts">
import WaitlistForm from '../components/landing/WaitlistForm.vue'
</script>

<template>
  <div class="landing">
    <section class="hero">
      <h1>Build Something Amazing</h1>
      <p class="hero-description">
        A modern platform to create, manage, and scale your digital presence.
        Join the waitlist to get early access.
      </p>
      <WaitlistForm />
    </section>

    <section class="features">
      <div class="feature">
        <h2>Easy Setup</h2>
        <p>Get started in minutes with our intuitive onboarding process.</p>
      </div>
      <div class="feature">
        <h2>Powerful Features</h2>
        <p>Everything you need to build and manage your digital presence.</p>
      </div>
      <div class="feature">
        <h2>Secure by Default</h2>
        <p>Enterprise-grade security to keep your data safe.</p>
      </div>
    </section>

    <section class="cta">
      <h2>Ready to get started?</h2>
      <p>Join the waitlist and be the first to know when we launch.</p>
    </section>
  </div>
</template>

<style scoped>
.landing {
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

.hero {
  padding: 80px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.hero h1 {
  font-size: 56px;
  font-weight: 600;
  letter-spacing: -1.5px;
  margin: 0;
}

.hero-description {
  font-size: 20px;
  color: var(--text);
  max-width: 600px;
  line-height: 1.6;
}

.features {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 32px;
  padding: 60px 20px;
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
}

.feature {
  text-align: left;
}

.feature h2 {
  font-size: 24px;
  margin-bottom: 12px;
}

.feature p {
  font-size: 16px;
  color: var(--text);
  line-height: 1.5;
}

.cta {
  padding: 80px 20px;
  text-align: center;
}

.cta h2 {
  font-size: 32px;
  margin-bottom: 16px;
}

.cta p {
  font-size: 18px;
  color: var(--text);
}

@media (max-width: 768px) {
  .hero h1 {
    font-size: 36px;
  }

  .features {
    grid-template-columns: 1fr;
    gap: 24px;
  }
}
</style>