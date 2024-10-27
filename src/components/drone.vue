<template>
  <div>
    <div @click="toggleExpand" :class="{ expanded: isExpanded }" class="image-container">
      <h2>Infrared Images</h2>
      <input type="date" @change="fetchImages" />
      <div v-if="selectedImages.length">
        <img
          v-for="image in selectedImages"
          :key="image.id"
          :src="image.url"
          :alt="image.description"
          @click="openImage(image)"
        />
      </div>
    </div>
    <div v-if="isExpanded" class="full-screen-view">
      <button @click="toggleExpand">Close</button>
      <img :src="currentImage.url" alt="Zoomed Image" class="zoomable" />
      <button @click="zoomIn">Zoom In</button>
      <button @click="zoomOut">Zoom Out</button>
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
    };
  },
  methods: {
    fetchImages(event) {
      const selectedDate = event.target.value;
      // Fetch images based on the selected date
      // Set selectedImages based on the fetched data
    },
    toggleExpand() {
      this.isExpanded = !this.isExpanded;
    },
    openImage(image) {
      this.currentImage = image;
      this.zoomLevel = 1; // Reset zoom level when opening an image
    },
    zoomIn() {
      this.zoomLevel += 0.1; // Increase zoom level
    },
    zoomOut() {
      if (this.zoomLevel > 0.1) {
        this.zoomLevel -= 0.1; // Decrease zoom level
      }
    },
  },
};
</script>

<style scoped>
.image-container {
  cursor: pointer;
  transition: all 0.3s ease;
}

.image-container.expanded {
  width: 100%; /* Make it full width when expanded */
  height: auto; /* Adjust height as needed */
}

.full-screen-view {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.zoomable {
  transform: scale(var(--zoom-level));
  transition: transform 0.3s ease;
}
</style>
