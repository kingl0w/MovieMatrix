<template>
  <div class="category-list-container">
    <h1>Categories</h1>
    <ul class="category-grid" v-if="categories.length > 0">
      <li
        v-for="category in categories"
        :key="category.id"
        class="category-item"
      >
        <router-link
          :to="{ name: 'CategoryMovies', params: { id: category.id } }"
        >
          {{ category.name }}
        </router-link>
      </li>
    </ul>
    <p v-else>No categories with movies available.</p>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "CategoryList",
  data() {
    return {
      categories: [],
    };
  },
  async created() {
    try {
      const response = await axios.get("http://localhost:8000/categories");
      this.categories = response.data;
    } catch (error) {
      console.error("Error fetching categories:", error);
    }
  },
};
</script>

<style scoped>
.category-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  text-align: center;
  color: #333;
  margin-bottom: 4rem;
}

.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
  list-style-type: none;
  padding: 0;
  justify-content: center;
  max-width: 1000px;
  margin: 0 auto;
}

.category-item {
  background-color: #f5f5f5;
  border-radius: 8px;
  padding: 15px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.category-item a {
  color: #333;
  text-decoration: none;
  font-weight: bold;
}

.category-item a:hover {
  color: #4caf50;
}
</style>
