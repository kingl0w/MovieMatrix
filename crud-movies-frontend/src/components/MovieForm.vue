<template>
  <div class="form-container">
    <h1>{{ isEditing ? "Edit Movie" : "Add Movie" }}</h1>
    <form @submit.prevent="saveMovie" class="movie-form">
      <div class="form-group">
        <label for="title">Title:</label>
        <input id="title" v-model="movie.title" required />
      </div>
      <div class="form-group">
        <label for="director-firstname">Director First Name:</label>
        <input
          id="director-firstname"
          v-model="movie.director.firstname"
          required
        />
      </div>
      <div class="form-group">
        <label for="director-lastname">Director Last Name:</label>
        <input
          id="director-lastname"
          v-model="movie.director.lastname"
          required
        />
      </div>
      <div class="form-group">
        <label for="cover">Cover URL:</label>
        <input id="cover" v-model="movie.cover" />
      </div>
      <div class="form-group">
        <label for="categories">Categories:</label>
        <input
          id="categories"
          v-model="categoriesInput"
          placeholder="Comma-separated categories"
        />
      </div>
      <button type="submit" class="submit-btn">Save</button>
    </form>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      movie: {
        mid: "", // Changed from isbn to mid
        title: "",
        director: { firstname: "", lastname: "" },
        cover: "",
        categories: [],
      },
      categoriesInput: "",
      isEditing: false,
    };
  },
  created() {
    const id = this.$route.params.id;
    if (id) {
      this.isEditing = true;
      this.fetchMovie(id);
    }
  },
  methods: {
    async fetchMovie(id) {
      try {
        const response = await axios.get(`http://localhost:8000/movies/${id}`);
        this.movie = response.data;
        this.categoriesInput = this.movie.categories
          .map((c) => c.name)
          .join(", ");
      } catch (error) {
        console.error("Error fetching movie:", error);
      }
    },
    async saveMovie() {
      try {
        console.log("Starting saveMovie method");
        console.log("Current movie data:", this.movie);
        console.log("Categories input:", this.categoriesInput);

        this.movie.categories = this.categoriesInput
          .split(",")
          .map((c) => ({ name: c.trim() }))
          .filter((c) => c.name !== "");

        console.log("Processed categories:", this.movie.categories);

        if (this.isEditing) {
          console.log("Updating existing movie with ID:", this.movie.id);
          const response = await axios.put(
            `http://localhost:8000/movies/${this.movie.id}`,
            this.movie
          );
          console.log("Update response:", response.data);
        } else {
          console.log("Creating new movie");
          const response = await axios.post(
            "http://localhost:8000/movies",
            this.movie
          );
          console.log("Create response:", response.data);
        }

        console.log("Movie saved successfully");
        this.$router.push("/");
      } catch (error) {
        console.error("Error in saveMovie method:", error);
        if (error.response) {
          console.error("Response data:", error.response.data);
          console.error("Response status:", error.response.status);
          console.error("Response headers:", error.response.headers);
        } else if (error.request) {
          console.error("No response received:", error.request);
        } else {
          console.error("Error setting up request:", error.message);
        }
      }
    },
  },
};
</script>

<style scoped>
.form-container {
  max-width: 500px;
  margin: 0 auto;
  padding: 20px;
  background-color: #f5f5f5;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h1 {
  text-align: center;
  color: #333;
}

.movie-form {
  display: flex;
  flex-direction: column;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
  color: #555;
}

input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.submit-btn {
  background-color: #4caf50;
  color: white;
  padding: 10px 15px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.3s;
}

.submit-btn:hover {
  background-color: #45a049;
}
</style>
