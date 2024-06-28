import axios from "axios";

const API_URL = "http://localhost:8000";

export default {
  async getMovies() {
    const response = await axios.get(`${API_URL}/movies`);
    return response.data;
  },
  async getMovie(id) {
    const response = await axios.get(`${API_URL}/movies/${id}`);
    return response.data;
  },
  async createMovie(movie) {
    const response = await axios.post(`${API_URL}/movies`, movie);
    return response.data;
  },
  async updateMovie(id, movie) {
    const response = await axios.put(`${API_URL}/movies/${id}`, movie);
    return response.data;
  },
  async deleteMovie(id) {
    await axios.delete(`${API_URL}/movies/${id}`);
  },
  async searchMovies(query) {
    const response = await axios.get(`${API_URL}/movies/search?q=${query}`);
    return response.data;
  },
};
