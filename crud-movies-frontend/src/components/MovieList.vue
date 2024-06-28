<template>
  <div class="movie-list-container">
    <h1>Movies</h1>
    <div class="search-container">
      <input
        v-model="searchQuery"
        @input="debouncedSearch"
        placeholder="Search movies..."
        class="search-input"
      />
    </div>
    <div class="movie-grid">
      <div v-for="movie in sortedMovies" :key="movie.id" class="movie-card">
        <h3>{{ movie.title }}</h3>
        <p>{{ movie.director.firstname }} {{ movie.director.lastname }}</p>
        <p>MID: {{ movie.mid }}</p>
        <div class="button-group">
          <button @click="editMovie(movie.id)" class="edit-btn">Edit</button>
          <button @click="deleteMovie(movie.id)" class="delete-btn">
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import { debounce } from "lodash";

export default {
  data() {
    return {
      movies: [],
      searchQuery: "",
    };
  },
  computed: {
    sortedMovies() {
      return [...this.movies].sort((a, b) => a.title.localeCompare(b.title));
    },
  },
  created() {
    this.fetchMovies();
    // create a debounced version of searchMovies
    this.debouncedSearch = debounce(this.searchMovies, 300);
  },
  methods: {
    async fetchMovies() {
      try {
        const response = await axios.get("http://localhost:8000/movies");
        this.movies = response.data;
      } catch (error) {
        console.error("Error fetching movies:", error);
      }
    },
    editMovie(id) {
      this.$router.push(`/edit/${id}`);
    },
    async deleteMovie(id) {
      try {
        await axios.delete(`http://localhost:8000/movies/${id}`);
        this.fetchMovies();
      } catch (error) {
        console.error("Error deleting movie:", error);
      }
    },
    async searchMovies() {
      if (this.searchQuery.trim() === "") {
        await this.fetchMovies();
        return;
      }
      try {
        console.log("Searching for:", this.searchQuery);
        const response = await axios.get(
          `http://localhost:8000/movies/search?q=${this.searchQuery}`
        );
        console.log("Search response:", response.data);
        this.movies = response.data;
      } catch (error) {
        console.error(
          "Error searching movies:",
          error.response ? error.response.data : error.message
        );
      }
    },
  },
};
</script>

<style scoped>
.movie-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  text-align: center;
  color: #333;
}

.search-container {
  margin-bottom: 4rem;
  text-align: center;
}

.search-input {
  width: 100%;
  max-width: 400px;
  padding: 10px;
  font-size: 16px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.movie-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
}

.movie-card {
  background-color: #f5f5f5;
  border-radius: 8px;
  padding: 15px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.movie-card h3 {
  margin-top: 0;
  color: #333;
}

.movie-card p {
  color: #666;
}

.button-group {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
}

.edit-btn,
.delete-btn {
  padding: 5px 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.edit-btn {
  background-color: #4caf50;
  color: white;
}

.edit-btn:hover {
  background-color: #45a049;
}

.delete-btn {
  background-color: #f44336;
  color: white;
}

.delete-btn:hover {
  background-color: #d32f2f;
}
</style>
