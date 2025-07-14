<template>
  <div class="drone-container">
    <div class="upload-section">
      <h2>Upload Drone Images</h2>
      <input 
        type="file" 
        @change="handleFileUpload" 
        accept="image/*" 
        multiple 
        class="file-input"
        ref="fileInput"
      />
      
      <!-- Preview Section -->
      <div v-if="selectedFiles.length" class="preview-section">
        <h3>Selected Images ({{ selectedFiles.length }})</h3>
        <div class="preview-grid">
          <div v-for="(file, index) in selectedFilePreviews" :key="index" class="preview-item">
            <img :src="file.preview" :alt="file.name" class="preview-image" />
            <button @click="removeFile(index)" class="remove-btn" title="Remove image">
              <span>×</span>
            </button>
            <span class="file-name">{{ file.name }}</span>
          </div>
        </div>
      </div>

      <button @click="uploadImages" class="upload-btn" :disabled="!selectedFiles.length">
        Upload Selected Images
      </button>
    </div>

    <div class="date-navigation">
      <div class="calendar-select">
        <input 
          type="date" 
          v-model="selectedDate"
          @change="jumpToDate"
          class="calendar-input"
        />
        <button @click="goToToday" class="today-btn">Today</button>
      </div>

      <div class="week-selector">
        <button @click="previousWeek" class="week-nav-btn">&lt; Previous Week</button>
        <span class="week-display">
          {{ formatDateRange(weekStart, weekEnd) }}
        </span>
        <button @click="nextWeek" class="week-nav-btn">Next Week &gt;</button>
      </div>
    </div>

    <div class="days-container">
      <div v-for="day in weekDays" :key="day.date" class="day-section">
        <div 
          class="day-header" 
          @click="toggleDayExpanded(day.date)"
          :class="{ 
            'has-images': day.images.length > 0,
            'is-selected': isSelectedDay(day.date)
          }"
        >
          <div class="day-info">
            <span class="day-name">{{ formatDayName(day.date) }}</span>
            <span class="day-date">{{ formatDate(day.date) }}</span>
          </div>
          <div class="day-summary">
            <span class="image-count" v-if="day.images.length">
              {{ day.images.length }} image{{ day.images.length !== 1 ? 's' : '' }}
            </span>
            <span class="expand-icon">{{ expandedDays.includes(day.date) ? '▼' : '▶' }}</span>
          </div>
        </div>

        <div v-if="expandedDays.includes(day.date)" class="day-content">
          <div v-if="day.images.length" class="image-grid">
            <div v-for="image in day.images" :key="image.id" class="image-item">
              <img
                :src="image.url"
                :alt="image.description"
                @click.stop="openImage(image)"
              />
              <p class="image-time">{{ formatTime(image.date) }}</p>
            </div>
          </div>
          <p v-else class="no-images">No images uploaded on this day</p>
        </div>
      </div>
    </div>

    <div v-if="isExpanded" class="full-screen-view">
      <button @click="toggleExpand" class="close-btn">Close</button>
      <img 
        :src="currentImage?.url" 
        alt="Zoomed Image" 
        class="zoomable" 
        :style="{ transform: `scale(${zoomLevel})` }"
      />
      <div class="zoom-controls">
        <button @click="zoomOut" class="zoom-btn">-</button>
        <span>{{ Math.round(zoomLevel * 100) }}%</span>
        <button @click="zoomIn" class="zoom-btn">+</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      isExpanded: false,
      selectedImages: [],
      currentImage: null,
      zoomLevel: 1,
      selectedFiles: [],
      isUploading: false,
      weekStart: new Date(),
      weekEnd: new Date(),
      expandedDays: [],
      weekDays: [],
      selectedDate: new Date().toISOString().split('T')[0] // Today's date in YYYY-MM-DD format
    };
  },
  created() {
    this.initializeWeek();
    this.fetchWeekImages();
  },
  methods: {
    initializeWeek(startDate = null) {
      if (startDate) {
        this.weekStart = new Date(startDate);
      } else {
        this.weekStart = new Date();
      }
      // Set to the start of the week (Sunday)
      this.weekStart.setDate(this.weekStart.getDate() - this.weekStart.getDay());
      this.weekEnd = new Date(this.weekStart);
      this.weekEnd.setDate(this.weekStart.getDate() + 6);
      this.generateWeekDays();
    },
    generateWeekDays() {
      this.weekDays = [];
      for (let i = 0; i < 7; i++) {
        const currentDate = new Date(this.weekStart);
        currentDate.setDate(this.weekStart.getDate() + i);
        this.weekDays.push({
          date: currentDate.toISOString().split('T')[0],
          images: []
        });
      }
    },
    jumpToDate() {
      this.initializeWeek(this.selectedDate);
      this.fetchWeekImages();
      // Expand the selected day automatically
      this.expandedDays = [this.selectedDate];
    },
    goToToday() {
      const today = new Date().toISOString().split('T')[0];
      this.selectedDate = today;
      this.jumpToDate();
    },
    isSelectedDay(date) {
      return date === this.selectedDate;
    },
    async fetchWeekImages() {
      try {
        // TODO: Replace with your actual API endpoint
        const response = await fetch(
          `/api/drone/images/week?start=${this.weekStart.toISOString()}&end=${this.weekEnd.toISOString()}`
        );
        if (!response.ok) throw new Error('Failed to fetch images');
        
        const data = await response.json();
        // Group images by day
        this.weekDays = this.weekDays.map(day => ({
          ...day,
          images: data.images.filter(img => 
            new Date(img.date).toISOString().split('T')[0] === day.date
          )
        }));
      } catch (error) {
        console.error('Error fetching week images:', error);
      }
    },
    previousWeek() {
      this.weekStart.setDate(this.weekStart.getDate() - 7);
      this.weekEnd.setDate(this.weekEnd.getDate() - 7);
      this.generateWeekDays();
      this.fetchWeekImages();
    },
    nextWeek() {
      this.weekStart.setDate(this.weekStart.getDate() + 7);
      this.weekEnd.setDate(this.weekEnd.getDate() + 7);
      this.generateWeekDays();
      this.fetchWeekImages();
    },
    toggleDayExpanded(date) {
      const index = this.expandedDays.indexOf(date);
      if (index === -1) {
        this.expandedDays.push(date);
      } else {
        this.expandedDays.splice(index, 1);
      }
    },
    formatDayName(dateStr) {
      return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'long' });
    },
    formatDate(dateStr) {
      return new Date(dateStr).toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric' 
      });
    },
    formatTime(dateStr) {
      return new Date(dateStr).toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    },
    formatDateRange(start, end) {
      const startStr = start.toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric' 
      });
      const endStr = end.toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric',
        year: 'numeric'
      });
      return `${startStr} - ${endStr}`;
    },
    handleFileUpload(event) {
      const files = Array.from(event.target.files);
      this.selectedFiles = files;
      
      // Generate previews for selected files
      this.selectedFilePreviews = files.map(file => ({
        name: file.name,
        preview: URL.createObjectURL(file),
        file: file
      }));
    },
    removeFile(index) {
      // Remove from both arrays
      URL.revokeObjectURL(this.selectedFilePreviews[index].preview);
      this.selectedFilePreviews.splice(index, 1);
      this.selectedFiles.splice(index, 1);
      
      // Reset file input if all files are removed
      if (this.selectedFiles.length === 0) {
        this.$refs.fileInput.value = '';
      }
    },
    async uploadImages() {
      if (!this.selectedFiles.length || this.isUploading) return;
      
      this.isUploading = true;
      try {
        const formData = new FormData();
        this.selectedFiles.forEach((file, index) => {
          formData.append(`image${index}`, file);
        });

        // TODO: Replace with your actual API endpoint
        const response = await fetch('/api/drone/upload', {
          method: 'POST',
          body: formData
        });

        if (!response.ok) throw new Error('Upload failed');

        const result = await response.json();
        // Refresh the current week's images after upload
        await this.fetchWeekImages();
        
        // Clean up previews and reset
        this.selectedFilePreviews.forEach(preview => {
          URL.revokeObjectURL(preview.preview);
        });
        this.selectedFilePreviews = [];
        this.selectedFiles = [];
        this.$refs.fileInput.value = '';
        
      } catch (error) {
        console.error('Error uploading images:', error);
      } finally {
        this.isUploading = false;
      }
    },
    toggleExpand() {
      this.isExpanded = !this.isExpanded;
    },
    openImage(image) {
      this.currentImage = image;
      this.zoomLevel = 1;
    },
    zoomIn() {
      if (this.zoomLevel < 3) {
        this.zoomLevel += 0.1;
      }
    },
    zoomOut() {
      if (this.zoomLevel > 0.3) {
        this.zoomLevel -= 0.1;
      }
    }
  }
};
</script>

<style scoped>
.drone-container {
  padding: 20px;
  max-width: 1200px;
  margin: 80px auto 0;
}

.upload-section {
  background: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.file-input {
  margin: 10px 0;
  padding: 10px;
  border: 2px dashed #ccc;
  border-radius: 4px;
  width: 100%;
}

.upload-btn {
  background: #4CAF50;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.upload-btn:disabled {
  background: #cccccc;
  cursor: not-allowed;
}

.date-navigation {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-bottom: 20px;
}

.calendar-select {
  display: flex;
  gap: 10px;
  align-items: center;
}

.calendar-input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.today-btn {
  padding: 8px 16px;
  background: #2196F3;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.today-btn:hover {
  background: #1976D2;
}

.week-selector {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20px;
}

.week-nav-btn {
  background: #f0f0f0;
  border: 1px solid #ddd;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.week-nav-btn:hover {
  background: #e0e0e0;
}

.week-display {
  font-size: 1.1em;
  font-weight: 500;
}

.days-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  overflow: hidden;
}

.day-section {
  border-bottom: 1px solid #eee;
}

.day-section:last-child {
  border-bottom: none;
}

.day-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.day-header:hover {
  background-color: #f5f5f5;
}

.day-header.has-images {
  background-color: #f0f7ff;
}

.day-header.is-selected {
  background-color: #e3f2fd;
  border-left: 4px solid #2196F3;
}

.day-info {
  display: flex;
  gap: 10px;
  align-items: center;
}

.day-name {
  font-weight: 500;
}

.day-date {
  color: #666;
}

.day-summary {
  display: flex;
  align-items: center;
  gap: 10px;
}

.image-count {
  color: #4CAF50;
  font-size: 0.9em;
}

.expand-icon {
  color: #666;
  font-size: 0.8em;
}

.day-content {
  padding: 20px;
  background: #fafafa;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
}

.image-item {
  position: relative;
  cursor: pointer;
}

.image-item img {
  width: 100%;
  height: 200px;
  object-fit: cover;
  border-radius: 4px;
  transition: transform 0.2s;
}

.image-item img:hover {
  transform: scale(1.05);
}

.image-time {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0,0,0,0.7);
  color: white;
  padding: 5px;
  margin: 0;
  font-size: 0.8em;
  text-align: center;
  border-bottom-left-radius: 4px;
  border-bottom-right-radius: 4px;
}

.full-screen-view {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.close-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  background: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
}

.zoomable {
  max-width: 90%;
  max-height: 80vh;
  object-fit: contain;
  transition: transform 0.3s ease;
}

.zoom-controls {
  position: absolute;
  bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255,255,255,0.9);
  padding: 10px;
  border-radius: 20px;
}

.zoom-btn {
  background: #333;
  color: white;
  border: none;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.no-images {
  text-align: center;
  color: #666;
  padding: 20px;
}

.preview-section {
  margin: 20px 0;
  background: white;
  padding: 15px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.preview-section h3 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 1.1em;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 15px;
  margin-bottom: 15px;
}

.preview-item {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f5f5;
  padding: 5px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.preview-image {
  width: 100%;
  height: 150px;
  object-fit: cover;
  border-radius: 4px;
  display: block;
}

.remove-btn {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #ff4444;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  transition: all 0.2s ease;
}

.remove-btn:hover {
  background: #ff4444;
  color: white;
  transform: scale(1.1);
}

.remove-btn span {
  line-height: 1;
  font-weight: bold;
}

.file-name {
  display: block;
  font-size: 0.8em;
  color: #666;
  margin-top: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0 5px;
}
</style>
