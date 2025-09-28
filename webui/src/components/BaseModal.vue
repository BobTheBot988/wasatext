<script>
export default {
  props: ["val", "title", "icon"],
  emits: ["close", "opened"],
  data() {
    return { open: false };
  },
  methods: {
    closeModal() {
      this.$emit("close");
      this.open = false;
    },
    handleOpen() {
      this.$emit("opened");
      this.open = true;
    },
  },
};
</script>
<template>
  <label class="my-label">
    <svg v-if="icon && icon !== ''" class="feather">
      <use :href="icon" />
    </svg>

    <span> {{ val }}</span>
    <input type="button" class="btn" :title="title" @click="handleOpen" />
  </label>

  <Teleport to="#modal-container">
    <transition name="modal">
      <div v-if="open" id="modal-container-inner" @keyup.esc="closeModal">
        <div id="modal">
          <div id="modal-title">
            <slot name="title" />

            <input
              id="exit-button"
              type="button"
              title="Close modal"
              value="X"
              @click="closeModal"
            />
          </div>
          <slot />
        </div>
      </div>
    </transition>
  </Teleport>
</template>
