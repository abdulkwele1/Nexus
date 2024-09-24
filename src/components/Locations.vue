<template>
    <div class="container">
        <button class="home-button" @click="goHome">Home</button>
        </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
const router = useRouter(); // Create router instance
const goHome = () => {
    router.push('/home'); // Adjust the path as necessary
    };
    
</script>

<style scoped>
 /* Style for the home button */
    .home-button {
    position: absolute; /* Keep button in top left */
    top: 20px;
    left: 20px;
    padding: 10px 20px;
    font-size: 16px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s;
    }

    .home-button:hover {
    background-color: #0056b3;
    }   
</style>
