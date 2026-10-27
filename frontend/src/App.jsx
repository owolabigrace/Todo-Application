import { useEffect, useState } from "react";
import { listTodos, createTodo, updateTodo, deleteTodo } from "./api";

export default function App() {
  const [todos, setTodos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [newTitle, setNewTitle] = useState("");
  const [editingId, setEditingId] = useState(null);
  const [editTitle, setEditTitle] = useState("");

  useEffect(() => {
    let cancelled = false;

    listTodos()
      .then((items) => {
        if (!cancelled) setTodos(items);
      })
      .catch((err) => {
        if (!cancelled) setError(err.message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  async function run(action) {
    setError("");
    try {
      await action();
    } catch (err) {
      setError(err.message);
    }
  }

  function replaceTodo(updated) {
    setTodos((prev) => prev.map((t) => (t.id === updated.id ? updated : t)));
  }

  function handleAdd(e) {
    e.preventDefault();
    const title = newTitle.trim();
    if (!title) return;

    run(async () => {
      const created = await createTodo(title);
      setTodos((prev) => [...prev, created]);
      setNewTitle("");
    });
  }

  function handleToggle(todo) {
    run(async () => {
      const updated = await updateTodo(todo.id, todo.title, !todo.completed);
      replaceTodo(updated);
    });
  }

  function startEdit(todo) {
    setEditingId(todo.id);
    setEditTitle(todo.title);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditTitle("");
  }

  function handleSave(todo) {
    const title = editTitle.trim();
    if (!title) {
      setError("Title is required");
      return;
    }

    run(async () => {
      const updated = await updateTodo(todo.id, title, todo.completed);
      replaceTodo(updated);
      cancelEdit();
    });
  }

  function handleDelete(id) {
    run(async () => {
      await deleteTodo(id);
      setTodos((prev) => prev.filter((t) => t.id !== id));
    });
  }

  return (
    <main className="app">
      <h1>Todo List</h1>

      <form className="add-form" onSubmit={handleAdd}>
        <input
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          placeholder="What needs to be done?"
          maxLength={200}
          aria-label="New todo title"
        />
        <button type="submit">Add</button>
      </form>

      {error && <p className="error">{error}</p>}

      {loading ? (
        <p>Loading…</p>
      ) : todos.length === 0 ? (
        <p className="empty">No todos yet.</p>
      ) : (
        <ul className="todo-list">
          {todos.map((todo) => (
            <li key={todo.id} className={todo.completed ? "done" : ""}>
              <input
                type="checkbox"
                checked={todo.completed}
                onChange={() => handleToggle(todo)}
                aria-label={`Mark "${todo.title}" as complete`}
              />

              {editingId === todo.id ? (
                <>
                  <input
                    className="edit-input"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    maxLength={200}
                    autoFocus
                    onKeyDown={(e) => {
                      if (e.key === "Enter") handleSave(todo);
                      if (e.key === "Escape") cancelEdit();
                    }}
                  />
                  <button onClick={() => handleSave(todo)}>Save</button>
                  <button onClick={cancelEdit}>Cancel</button>
                </>
              ) : (
                <>
                  <span className="title">{todo.title}</span>
                  <button onClick={() => startEdit(todo)}>Edit</button>
                  <button onClick={() => handleDelete(todo.id)}>Delete</button>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}