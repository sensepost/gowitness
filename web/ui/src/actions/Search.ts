import * as api from "@/lib/api/api";
import { redirect } from "react-router-dom";

// searchAction grabs the form to search, encodes the data and
// redirects to the search URI that will trigger the loader
const searchAction = async ({ request }: { request: Request; }) => {
  const formData = await request.formData();
  const searchQuery = formData.get("query"); // Extract the search term

  if (!searchQuery) {
    return { error: "Search query is missing" };
  }

  return redirect(`/search?query=${encodeURIComponent(searchQuery as string)}`);
};

// searchLoader loads search data from the api
const searchLoader = async ({ request }: { request: Request; }) => {
  const url = new URL(request.url);
  const searchQuery = url.searchParams.get("query");

  if (!searchQuery) {
    return { error: "No search query provided" };
  }

  return await api.post('search', { query: decodeURIComponent(searchQuery) });
};

export { searchAction, searchLoader };