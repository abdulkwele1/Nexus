<template>
  <div class="admin-container">
    <!-- Navbar -->
    <nav class="navbar">
      <div class="max-w-screen-xl flex flex-wrap items-center justify-center mx-auto py-2 px-4">
        <div class="flex items-center justify-center w-full">
          <a @click="goHome" class="flex items-center space-x-3 rtl:space-x-reverse cursor-pointer">
            <img alt="Vue logo" class="logo" src="@/assets/nexus.png" width="80" height="32" />
          </a>
          <div class="ml-8">
            <ul class="font-medium flex flex-row space-x-8">
              <li>
                <a @click="goHome" class="text-white hover:text-[#90EE90] cursor-pointer">Home</a>
              </li>
              <li>
                <a @click="goToSettings" class="text-white hover:text-[#90EE90] cursor-pointer">Settings</a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </nav>

    <!-- Admin Content -->
    <div class="admin-content">
      <div class="admin-header">
        <h1>Admin Panel</h1>
        <p>Manage user accounts and permissions</p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading">
        <p>Loading users...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-message">
        <p>{{ error }}</p>
        <button @click="loadUsers" class="retry-button">Retry</button>
      </div>

      <!-- Users List -->
      <div v-else class="users-section">
        <div class="section-header">
          <h2>All Users</h2>
          <div class="header-actions">
            <button @click="showCreateForm = !showCreateForm" class="create-button">Create User</button>
            <button @click="loadUsers" class="refresh-button">Refresh</button>
          </div>
        </div>

        <!-- Create User Form -->
        <div v-if="showCreateForm" class="create-user-form">
          <h3>Create New User</h3>
          <form @submit.prevent="handleCreateUser">
            <div class="form-group">
              <label>Username:</label>
              <input 
                v-model="newUser.username" 
                @input="validateUsername"
                type="text" 
                required 
                placeholder="Enter username (no spaces)"
                :class="['form-input', usernameValidationClass]"
              />
              <div v-if="usernameValidationMessage" :class="['validation-message', usernameValidationClass]">
                {{ usernameValidationMessage }}
              </div>
            </div>
            <div class="form-group">
              <label>Password:</label>
              <input 
                v-model="newUser.password" 
                type="password" 
                required 
                placeholder="Enter password"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Role:</label>
              <select v-model="newUser.role" class="form-select">
                <option value="user">User</option>
                <option value="admin">Admin</option>
                <option v-if="currentUserRole === 'root_admin'" value="root_admin">Root Admin</option>
              </select>
            </div>
            <div class="form-actions">
              <button 
                type="submit" 
                class="submit-button" 
                :disabled="creatingUser || !isUsernameValid || checkingUsername"
              >
                {{ creatingUser ? 'Creating...' : 'Create User' }}
              </button>
              <button type="button" @click="cancelCreateUser" class="cancel-button">Cancel</button>
            </div>
            <div v-if="createError" class="error-text">{{ createError }}</div>
          </form>
        </div>

        <div v-if="users.length === 0" class="no-users">
          <p>No users found.</p>
        </div>

        <div v-else class="users-list">
          <div v-for="user in users" :key="user.username" class="user-card">
            <div class="user-info">
              <h3>{{ user.username }}</h3>
              <span class="user-role" :class="getRoleClass(user.role)">{{ user.role }}</span>
            </div>
            
            <div class="user-actions">
              <!-- Role Management -->
              <div v-if="user.role !== 'root_admin' || currentUserRole === 'root_admin'" class="role-management">
                <label>Role:</label>
                <select 
                  :value="user.role" 
                  @change="updateUserRole(user.username, $event.target.value)"
                  :disabled="user.role === 'root_admin' && currentUserRole !== 'root_admin'"
                  class="role-select"
                >
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                  <option v-if="currentUserRole === 'root_admin'" value="root_admin">Root Admin</option>
                </select>
              </div>

              <!-- Remove Admin Button (only for root_admin) -->
              <button 
                v-if="currentUserRole === 'root_admin' && user.role === 'admin' && user.username !== currentUsername"
                @click="removeAdminPermissions(user.username)"
                class="remove-admin-button"
              >
                Remove Admin
              </button>

              <!-- Delete User Button -->
              <button 
                v-if="user.username !== currentUsername && (user.role !== 'root_admin' || currentUserRole === 'root_admin')"
                @click="deleteUser(user.username)"
                class="delete-button"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useNexusStore } from '@/stores/nexus';

const store = useNexusStore();
const router = useRouter();

// Reactive state
const users = ref([]);
const loading = ref(false);
const error = ref('');
const currentUsername = ref('');
const currentUserRole = ref('');
const showCreateForm = ref(false);
const creatingUser = ref(false);
const createError = ref('');
const checkingUsername = ref(false);
const usernameValidationMessage = ref('');
const usernameValidationClass = ref('');
const isUsernameValid = ref(false);
let usernameCheckTimeout = null;

const newUser = ref({
  username: '',
  password: '',
  role: 'user'
});

// Load users on component mount
onMounted(async () => {
  await loadUsers();
  await loadCurrentUserInfo();
});

// Load current user information
const loadCurrentUserInfo = async () => {
  try {
    const settings = await store.user.getUserSettings();
    currentUsername.value = settings.username;
    currentUserRole.value = settings.role;
  } catch (error) {
    console.error('Error loading current user info:', error);
  }
};

// Load all users
const loadUsers = async () => {
  loading.value = true;
  error.value = '';
  
  try {
    const response = await store.user.getAllUsers();
    users.value = response.users || [];
  } catch (err) {
    console.error('Error loading users:', err);
    error.value = 'Failed to load users. You may not have admin permissions.';
  } finally {
    loading.value = false;
  }
};

// Update user role
const updateUserRole = async (username, newRole) => {
  try {
    await store.user.updateUserRole(username, newRole);
    // Update local state
    const user = users.value.find(u => u.username === username);
    if (user) {
      user.role = newRole;
    }
  } catch (err) {
    console.error('Error updating user role:', err);
    alert('Failed to update user role');
  }
};

// Remove admin permissions
const removeAdminPermissions = async (username) => {
  if (confirm(`Are you sure you want to remove admin permissions from ${username}?`)) {
    try {
      await store.user.removeAdminPermissions(username);
      // Update local state
      const user = users.value.find(u => u.username === username);
      if (user) {
        user.role = 'user';
      }
    } catch (err) {
      console.error('Error removing admin permissions:', err);
      alert('Failed to remove admin permissions');
    }
  }
};

// Validate username in real-time
const validateUsername = async () => {
  const username = newUser.value.username.trim();
  
  // Clear previous timeout
  if (usernameCheckTimeout) {
    clearTimeout(usernameCheckTimeout);
  }
  
  // Reset validation state
  usernameValidationMessage.value = '';
  usernameValidationClass.value = '';
  isUsernameValid.value = false;
  
  // Check for spaces
  if (username.includes(' ')) {
    usernameValidationMessage.value = 'Username cannot contain spaces';
    usernameValidationClass.value = 'error';
    isUsernameValid.value = false;
    return;
  }
  
  // If username is empty, don't check
  if (!username) {
    return;
  }
  
  // Debounce the API call
  usernameCheckTimeout = setTimeout(async () => {
    checkingUsername.value = true;
    try {
      const result = await store.user.checkUsernameAvailable(username);
      if (result.available) {
        usernameValidationMessage.value = 'Username is available';
        usernameValidationClass.value = 'success';
        isUsernameValid.value = true;
      } else {
        usernameValidationMessage.value = result.error || 'Username is already taken';
        usernameValidationClass.value = 'error';
        isUsernameValid.value = false;
      }
    } catch (err) {
      console.error('Error checking username:', err);
      // Don't block user if check fails
      usernameValidationMessage.value = '';
      usernameValidationClass.value = '';
      isUsernameValid.value = true;
    } finally {
      checkingUsername.value = false;
    }
  }, 500); // 500ms debounce
};

// Create new user
const handleCreateUser = async () => {
  // Final validation before submitting
  if (!isUsernameValid.value || newUser.value.username.includes(' ')) {
    createError.value = 'Please fix the username errors before creating the user.';
    return;
  }
  
  creatingUser.value = true;
  createError.value = '';
  
  try {
    await store.user.createUser(newUser.value.username.trim(), newUser.value.password, newUser.value.role);
    // Reset form
    newUser.value = { username: '', password: '', role: 'user' };
    usernameValidationMessage.value = '';
    usernameValidationClass.value = '';
    isUsernameValid.value = false;
    showCreateForm.value = false;
    // Reload users list
    await loadUsers();
  } catch (err) {
    console.error('Error creating user:', err);
    const errorMessage = err.message || err.error || 'Failed to create user. User may already exist.';
    createError.value = errorMessage;
    // Update validation if username is taken
    if (errorMessage.includes('already exists') || errorMessage.includes('taken')) {
      usernameValidationMessage.value = 'Username is already taken';
      usernameValidationClass.value = 'error';
      isUsernameValid.value = false;
    }
  } finally {
    creatingUser.value = false;
  }
};

// Cancel create user form
const cancelCreateUser = () => {
  showCreateForm.value = false;
  newUser.value = { username: '', password: '', role: 'user' };
  createError.value = '';
  usernameValidationMessage.value = '';
  usernameValidationClass.value = '';
  isUsernameValid.value = false;
  if (usernameCheckTimeout) {
    clearTimeout(usernameCheckTimeout);
    usernameCheckTimeout = null;
  }
};

// Delete user
const deleteUser = async (username) => {
  if (confirm(`Are you sure you want to delete user "${username}"? This action cannot be undone.`)) {
    try {
      await store.user.deleteUser(username);
      // Remove from local state
      users.value = users.value.filter(u => u.username !== username);
    } catch (err) {
      console.error('Error deleting user:', err);
      alert('Failed to delete user');
    }
  }
};

// Get CSS class for role styling
const getRoleClass = (role) => {
  switch (role) {
    case 'root_admin':
      return 'role-root-admin';
    case 'admin':
      return 'role-admin';
    default:
      return 'role-user';
  }
};

// Navigation methods
const goHome = () => {
  router.push('/home');
};

const goToSettings = () => {
  router.push('/settings');
};
</script>

<style scoped>
/* Navbar styles */
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1000;
  margin-top: -1px;
  height: 60px;
  background-color: #000000;
  border-bottom: 2px solid #90EE90;
  box-shadow: 0 0 10px rgba(144, 238, 144, 0.3);
}

.max-w-screen-xl {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
}

.logo {
  width: 80px;
  height: 32px;
  object-fit: contain;
}

/* Admin container */
.admin-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  padding: 80px 20px 20px;
  background-color: #121212;
}

.admin-content {
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.admin-header {
  text-align: center;
  margin-bottom: 40px;
}

.admin-header h1 {
  color: white;
  font-size: 2.5rem;
  margin-bottom: 10px;
}

.admin-header p {
  color: #e0e0e0;
  font-size: 1.1rem;
}

/* Loading and error states */
.loading, .error-message {
  text-align: center;
  padding: 40px;
  color: white;
}

.error-message {
  color: #ff6b6b;
}

.retry-button {
  background-color: #90EE90;
  color: black;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 10px;
}

.retry-button:hover {
  background-color: #7CCD7C;
}

/* Users section */
.users-section {
  background: #1a1a1a;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #333;
}

.section-header h2 {
  color: white;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.refresh-button {
  background-color: #90EE90;
  color: black;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.refresh-button:hover {
  background-color: #7CCD7C;
}

.create-button {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.create-button:hover {
  background-color: #45a049;
}

/* Users list */
.no-users {
  text-align: center;
  color: #e0e0e0;
  padding: 40px;
}

.users-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.user-card {
  background: #2a2a2a;
  border-radius: 6px;
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid #333;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.user-info h3 {
  color: white;
  margin: 0;
  font-size: 1.2rem;
}

.user-role {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: bold;
}

.role-user {
  background-color: #4a4a4a;
  color: #e0e0e0;
}

.role-admin {
  background-color: #ffa500;
  color: black;
}

.role-root-admin {
  background-color: #ff4444;
  color: white;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 15px;
}

.role-management {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-management label {
  color: #e0e0e0;
  font-size: 0.9rem;
}

.role-select {
  background: #3a3a3a;
  color: white;
  border: 1px solid #555;
  border-radius: 4px;
  padding: 6px 10px;
  font-size: 0.9rem;
}

.role-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.remove-admin-button {
  background-color: #ff6b6b;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.remove-admin-button:hover {
  background-color: #ff5252;
}

.delete-button {
  background-color: #dc3545;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.delete-button:hover {
  background-color: #c82333;
}

/* Create User Form */
.create-user-form {
  background: #2a2a2a;
  border-radius: 6px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #333;
}

.create-user-form h3 {
  color: white;
  margin-top: 0;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  color: #e0e0e0;
  margin-bottom: 5px;
  font-size: 0.9rem;
}

.form-input,
.form-select {
  width: 100%;
  padding: 8px 12px;
  background: #3a3a3a;
  color: white;
  border: 1px solid #555;
  border-radius: 4px;
  font-size: 0.9rem;
  box-sizing: border-box;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #90EE90;
}

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.submit-button {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.submit-button:hover:not(:disabled) {
  background-color: #45a049;
}

.submit-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cancel-button {
  background-color: #6c757d;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.cancel-button:hover {
  background-color: #5a6268;
}

.error-text {
  color: #ff6b6b;
  margin-top: 10px;
  font-size: 0.9rem;
}

.validation-message {
  margin-top: 5px;
  font-size: 0.85rem;
  padding: 4px 0;
}

.validation-message.success {
  color: #4CAF50;
}

.validation-message.error {
  color: #ff6b6b;
}

.form-input.error {
  border-color: #ff6b6b;
}

.form-input.success {
  border-color: #4CAF50;
}

/* Navigation text color */
.navbar a {
  color: white;
}

.navbar a:hover {
  color: #90EE90;
}
</style>
