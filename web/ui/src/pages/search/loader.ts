import * as api from "@/lib/api/api";

// searchLoader loads search data from the api
const searchLoader = async ({ request }: { request: Request; }) => {
  const url = new URL(request.url);
  const searchQuery = url.searchParams.get("query");

  if (!searchQuery) {
    return { error: "No search query provided" };
  }

  return await api.post('search', { query: decodeURIComponent(searchQuery) });
};

export { searchLoader };