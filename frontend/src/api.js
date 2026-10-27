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

export const listTodos = () => request("/api/todos");

export const createTodo = (title) =>
  request("/api/todos", {
    method: "POST",
    body: JSON.stringify({ title }),
  });

export const updateTodo = (id, title, completed) =>
  request(`/api/todos/${id}`, {
    method: "PUT",
    body: JSON.stringify({ title, completed }),
  });

export const deleteTodo = (id) =>
  request(`/api/todos/${id}`, { method: "DELETE" });