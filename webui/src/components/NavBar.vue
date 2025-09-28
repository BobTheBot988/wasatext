<script>
import { RouterLink } from "vue-router";
import Search from "../components/Search.vue";
import BaseModal from "../components/BaseModal.vue";
import CreateConversationModal from "./CreateConversationModal.vue";
import CreateGroupModal from "./CreateGroupModal.vue";

export default {
  components: {
    Search,
    BaseModal,
    CreateConversationModal,
    CreateGroupModal,
  },
  props: ["isAuthed", "user"],
  emits: ["turn-main", "pause", "unpause"],
  data() {
    return {
      search_selected: false,
      users: null,
      k: 0,
    };
  },
  mounted() {
    if (this.localAuth) {
      this.getUser();
    }
    this.$forceUpdate();
  },
  methods: {
    getUser() {
      this.$emit("getUser");
    },

    handleLogOut() {
      if (!this.isAuthed) {
        alert("Not logged in");
        return;
      }
      this.$emit("logOut");
      this.$forceUpdate();
    },
    HandleReload() {
      this.k += 1;
    },
  },
};
</script>
<template>
  <nav id="sidebarMenu">
    <div>
      <ul>
        <template v-if="isAuthed">
          <li title="User" class="username">
            <RouterLink to="/user">
              <svg class="feather">
                <use href="/feather-sprite-v4.29.0.svg#user" />
              </svg>

              <span>{{ user.userName }}</span>
            </RouterLink>
          </li>
          <li title="Chats">
            <RouterLink to="/conversations/">
              <svg class="feather">
                <use href="/feather-sprite-v4.29.0.svg#layout" />
              </svg>
              <span>Chats</span>
            </RouterLink>
          </li>
          <li title="Create New Chat">
            <CreateConversationModal
              @opened="$emit('pause')"
              @close="$emit('unpause')"
            />
          </li>
          <li title="Create New Group">
            <svg class="feather">
              <use href="/feather-sprite-v4.29.0.svg#plus" />
            </svg>
            <CreateGroupModal
              :key="k"
              :user-id="user.userId"
              @opened="$emit('pause')"
              @close="
                HandleReload();
                $emit('unpause');
              "
            />
          </li>
          <h6>
            <span>Secondary menu</span>
          </h6>
          <li title="Log Out">
            <RouterLink :to="'/'" @click="handleLogOut">
              <svg class="feather">
                <use href="/feather-sprite-v4.29.0.svg#log-out" />
              </svg>
              <span>Logout</span>
            </RouterLink>
          </li>
        </template>
        <template v-else>
          <li title="Login">
            <RouterLink to="/session">
              <svg class="feather">
                <use href="/feather-sprite-v4.29.0.svg#log-in" />
              </svg>
              Login
            </RouterLink>
          </li>
        </template>
      </ul>
    </div>
    <Teleport to="#modal-container" :disabled="!search_selected">
      <Transition name="modal">
        <BaseModal
          v-if="search_selected"
          @exit-search="exitHandler"
          @disable_main="$emit('disable_main')"
        >
          <Search :users="users" />
        </BaseModal>
      </Transition>
    </Teleport>
  </nav>
</template>
