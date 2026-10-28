async function request(path, options = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  if (res.status === 204) return null;

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    throw new Error(data?.error || `Request failed (${res.status})`);
  }

  return data;
}

function buildQuery(filters) {
  const params = new URLSearchParams();

  if (filters.status && filters.status !== "all") params.set("status", filters.status);
  if (filters.priority) params.set("priority", filters.priority);
  if (filters.q) params.set("q", filters.q);
  if (filters.category) params.set("category", filters.category);
  if (filters.sort && filters.sort !== "created") params.set("sort", filters.sort);
  if (filters.order && filters.order !== "asc") params.set("order", filters.order);

  const query = params.toString();

  return query ? `?${query}` : "";
}

export const listTodos = (filters = {}, signal) =>
  request(`/api/todos${buildQuery(filters)}`, { signal });

export const createTodo = (payload) =>
  request("/api/todos", {
    method: "POST",
    body: JSON.stringify(payload),
  });

export const updateTodo = (id, payload) =>
  request(`/api/todos/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });

export const deleteTodo = (id) =>
  request(`/api/todos/${id}`, { method: "DELETE" });

export const listCategories = () => request("/api/categories");

export const createCategory = (name) =>
  request("/api/categories", {
    method: "POST",
    body: JSON.stringify({ name }),
  });

export const updateCategory = (id, name) =>
  request(`/api/categories/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name }),
  });

export const deleteCategory = (id) =>
  request(`/api/categories/${id}`, { method: "DELETE" });