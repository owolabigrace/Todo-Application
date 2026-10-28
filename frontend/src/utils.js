export const PRIORITIES = ["low", "medium", "high"];

export const EMPTY_VALUES = {
  title: "",
  notes: "",
  priority: "medium",
  dueDate: "",
  categoryId: "",
};

export function capitalize(text) {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

export function todayISO() {
  const d = new Date();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");

  return `${d.getFullYear()}-${month}-${day}`;
}

export function formatDate(iso) {
  const [year, month, day] = iso.split("-").map(Number);

  return new Date(year, month - 1, day).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function dueStatus(todo, today = todayISO()) {
  if (!todo.due_date || todo.completed) return "";
  if (todo.due_date < today) return "overdue";
  if (todo.due_date === today) return "today";

  return "";
}

export function dueLabel(todo, today = todayISO()) {
  if (!todo.due_date) return "";

  const status = dueStatus(todo, today);

  if (status === "today") return "Due today";
  if (status === "overdue") return `Overdue · ${formatDate(todo.due_date)}`;

  return `Due ${formatDate(todo.due_date)}`;
}

export function todoToValues(todo) {
  return {
    title: todo.title,
    notes: todo.notes || "",
    priority: todo.priority,
    dueDate: todo.due_date || "",
    categoryId: todo.category_id ? String(todo.category_id) : "",
  };
}

export function valuesToPayload(values, completed = false) {
  return {
    title: values.title.trim(),
    notes: values.notes,
    completed,
    priority: values.priority,
    due_date: values.dueDate || null,
    category_id: values.categoryId ? Number(values.categoryId) : null,
  };
}

export function toPayload(todo, overrides = {}) {
  return {
    title: todo.title,
    notes: todo.notes,
    completed: todo.completed,
    priority: todo.priority,
    due_date: todo.due_date,
    category_id: todo.category_id,
    ...overrides,
  };
}