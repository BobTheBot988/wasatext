<script>
import ErrorMsg from './ErrorMsg.vue'
import { errorState, errorManager } from '@/services/axios'

export default {
    name: 'ErrorDisplay',
    components: {
        ErrorMsg
    },
    setup() {
        return {
            errorState,
            errorManager
        }
    },
    methods: {
        dismissError(errorId) {
            this.errorManager.removeError(errorId)
        },
        dismissAllErrors() {
            this.errorManager.clearErrors()
        }
    }
}
</script>

<template>
  <div v-if="errorState.errors.length > 0" class="error-container mb-3">
    <div v-for="error in errorState.errors" :key="error.id" class="error-item mb-2">
      <div class="d-flex align-items-start">
        <div class="flex-grow-1">
          <ErrorMsg :msg="error.message" />
        </div>
        <button
          type="button" class="btn-close ms-2 mt-1" aria-label="Close"
          @click="dismissError(error.id)"
        />
      </div>
    </div>

    <!-- Dismiss all button if multiple errors -->
    <div v-if="errorState.errors.length > 1" class="text-end">
      <button type="button" class="btn btn-outline-danger btn-sm" @click="dismissAllErrors">
        Dismiss All
      </button>
    </div>
  </div>
</template>

<style scoped>
.error-container {
    position: relative;
}

.error-item {
    position: relative;
}

.btn-close {
    flex-shrink: 0;
}
</style>