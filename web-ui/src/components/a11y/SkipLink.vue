<template>
  <a
    href="#main-content"
    class="skip-link"
    @click.prevent="skipToMain"
  >
    <slot>Skip to main content</slot>
  </a>
</template>

<script setup lang="ts">
function skipToMain() {
  const main = document.getElementById('main-content') || document.querySelector('main')
  if (main) {
    main.setAttribute('tabindex', '-1')
    main.focus()
    main.scrollIntoView()
  }
}
</script>

<style scoped>
.skip-link {
  position: absolute;
  top: -100%;
  left: 0;
  z-index: 9999;
  padding: 8px 16px;
  background: #000;
  color: #fff;
  font-size: 14px;
  text-decoration: none;
  border-radius: 0 0 4px 0;
}

.skip-link:focus {
  top: 0;
  outline: 2px solid #fff;
  outline-offset: 2px;
}
</style>
