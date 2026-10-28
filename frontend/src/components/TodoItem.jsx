import { useState } from "react";
import TodoEditor from "./TodoEditor";
import { capitalize, dueLabel, dueStatus } from "../utils";

export default function TodoItem({
  todo,
  categories,
  categoryName,
  onToggle,
  onSave,
  onDelete,
}) {
  const [editing, setEditing] = useState(false);

  if (editing) {
    return (
      <li className="todo editing">
        <TodoEditor
          todo={todo}
          categories={categories}
          onCancel={() => setEditing(false)}
          onSave={async (payload) => {
            const ok = await onSave(todo, payload);
            if (ok) setEditing(false);
            return ok;
          }}
        />
      </li>
    );
  }

  const status = dueStatus(todo);

  return (
    <li className={`todo${todo.completed ? " done" : ""}`}>
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={() => onToggle(todo)}
        aria-label={`Mark "${todo.title}" as ${
          todo.completed ? "not complete" : "complete"
        }`}
      />

      <div className="todo-body">
        <span className="todo-title">{todo.title}</span>

        <div className="meta">
          <span className={`badge priority-${todo.priority}`}>
            {capitalize(todo.priority)}
          </span>

          {todo.due_date && (
            <span className={`badge due ${status}`}>{dueLabel(todo)}</span>
          )}

          {categoryName && <span className="badge category">{categoryName}</span>}
        </div>

        {todo.notes && <p className="notes">{todo.notes}</p>}
      </div>

      <div className="actions">
        <button
          type="button"
          className="btn"
          onClick={() => setEditing(true)}
          aria-label={`Edit "${todo.title}"`}
        >
          Edit
        </button>
        <button
          type="button"
          className="btn danger"
          onClick={() => onDelete(todo.id)}
          aria-label={`Delete "${todo.title}"`}
        >
          Delete
        </button>
      </div>
    </li>
  );
}