import { useEffect, useMemo, useState } from "react";
import {
  createCategory,
  createTodo,
  deleteCategory,
  deleteTodo,
  listCategories,
  listTodos,
  updateCategory,
  updateTodo,
} from "./api";
import CategoryManager from "./components/CategoryManager";
import Filters from "./components/Filters";
import TodoForm from "./components/TodoForm";
import TodoItem from "./components/TodoItem";
import { toPayload } from "./utils";

const DEFAULT_FILTERS = {
  status: "all",
  priority: "",
  q: "",
  sort: "created",
  order: "asc",
  category: "",
};

export default function App() {
  const [todos, setTodos] = useState(null);
  const [categories, setCategories] = useState([]);
  const [filters, setFilters] = useState(DEFAULT_FILTERS);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [reloadKey, setReloadKey] = useState(0);

  const categoryNames = useMemo(
    () => new Map(categories.map((c) => [c.id, c.name])),
    [categories]
  );

  const hasActiveFilters =
    filters.status !== "all" || filters.priority !== "" || filters.q !== "";

  useEffect(() => {
    const timer = setTimeout(() => {
      const q = searchText.trim();
      setFilters((prev) => (prev.q === q ? prev : { ...prev, q }));
    }, 300);

    return () => clearTimeout(timer);
  }, [searchText]);

  useEffect(() => {
    const controller = new AbortController();

    setLoading(true);

    listTodos(filters, controller.signal)
      .then((items) => {
        setTodos(items);
        setError("");
      })
      .catch((err) => {
        if (err.name !== "AbortError") setError(err.message);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });

    return () => controller.abort();
  }, [filters, reloadKey]);

  useEffect(() => {
    loadCategories().catch((err) => setError(err.message));
  }, []);

  async function loadCategories() {
    setCategories(await listCategories());
  }

  function reload() {
    setReloadKey((k) => k + 1);
  }

  async function run(action) {
    setError("");

    try {
      await action();
      return true;
    } catch (err) {
      setError(err.message);
      return false;
    }
  }

  function changeFilters(partial) {
    setFilters((prev) => ({ ...prev, ...partial }));
  }

  function clearFilters() {
    setSearchText("");
    setFilters((prev) => ({ ...prev, status: "all", priority: "", q: "" }));
  }

  function handleAdd(payload) {
    return run(async () => {
      await createTodo(payload);
      reload();
    });
  }

  function handleSave(todo, payload) {
    return run(async () => {
      await updateTodo(todo.id, payload);
      reload();
    });
  }

  function handleToggle(todo) {
    return run(async () => {
      await updateTodo(todo.id, toPayload(todo, { completed: !todo.completed }));
      reload();
    });
  }

  function handleDelete(id) {
    return run(async () => {
      await deleteTodo(id);
      reload();
    });
  }

  function handleCreateCategory(name) {
    return run(async () => {
      await createCategory(name);
      await loadCategories();
    });
  }

  function handleRenameCategory(id, name) {
    return run(async () => {
      await updateCategory(id, name);
      await loadCategories();
    });
  }

  function handleDeleteCategory(id) {
    return run(async () => {
      await deleteCategory(id);
      await loadCategories();

      if (filters.category === String(id)) {
        changeFilters({ category: "" });
      } else {
        reload();
      }
    });
  }

  function renderList() {
    if (todos === null) {
      return loading ? <p className="status">Loading…</p> : null;
    }

    if (todos.length === 0) {
      return (
        <p className="empty">
          {hasActiveFilters ? "No todos match your filters." : "No todos here yet."}
        </p>
      );
    }

    return (
      <>
        <p className="count">
          {todos.length} {todos.length === 1 ? "todo" : "todos"}
        </p>

        <ul className="todo-list" aria-busy={loading}>
          {todos.map((todo) => (
            <TodoItem
              key={todo.id}
              todo={todo}
              categories={categories}
              categoryName={categoryNames.get(todo.category_id)}
              onToggle={handleToggle}
              onSave={handleSave}
              onDelete={handleDelete}
            />
          ))}
        </ul>
      </>
    );
  }

  return (
    <div className="layout">
      <aside className="sidebar">
        <CategoryManager
          categories={categories}
          selected={filters.category}
          onSelect={(value) => changeFilters({ category: value })}
          onCreate={handleCreateCategory}
          onRename={handleRenameCategory}
          onDelete={handleDeleteCategory}
        />
      </aside>

      <main className="main">
        <h1>Todo List</h1>

        <TodoForm
          categories={categories}
          defaultCategory={filters.category}
          onAdd={handleAdd}
        />

        {error && (
          <div className="error" role="alert">
            <span>{error}</span>
            <button type="button" onClick={() => setError("")} aria-label="Dismiss error">
              ×
            </button>
          </div>
        )}

        <Filters
          filters={filters}
          searchText={searchText}
          onSearchChange={setSearchText}
          onChange={changeFilters}
          onClear={clearFilters}
          hasActive={hasActiveFilters}
        />

        {renderList()}
      </main>
    </div>
  );
}