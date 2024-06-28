import { createRouter, createWebHistory } from "vue-router";
import MovieList from "../components/MovieList.vue";
import MovieForm from "../components/MovieForm.vue";
import CategoryList from "../views/CategoryList.vue";
import CategoryMovies from "../views/CategoryMovies.vue";

const routes = [
  { path: "/", name: "Home", component: MovieList },
  { path: "/add", name: "AddMovie", component: MovieForm },
  { path: "/edit/:id", name: "EditMovie", component: MovieForm },
  { path: "/categories", name: "CategoryList", component: CategoryList },
  {
    path: "/categories/:id/movies",
    name: "CategoryMovies",
    component: CategoryMovies,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
