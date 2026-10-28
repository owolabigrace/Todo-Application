import { useState } from "react";
import TodoFields from "./TodoFields";
import { todoToValues, valuesToPayload } from "../utils";

export default function TodoEditor({ todo, categories, onSave, onCancel }) {
  const [values, setValues] = useState(() => todoToValues(todo));
  const [saving, setSaving] = useState(false);

  function change(field, value) {
    setValues((v) => ({ ...v, [field]: value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();

    if (!values.title.trim()) return;

    setSaving(true);
    await onSave(valuesToPayload(values, todo.completed));
    setSaving(false);
  }

  return (
    <form
      className="edit-form"
      onSubmit={handleSubmit}
      onKeyDown={(e) => {
        if (e.key === "Escape") onCancel();
      }}
    >
      <input
        type="text"
        className="edit-title"
        value={values.title}
        onChange={(e) => change("title", e.target.value)}
        maxLength={200}
        autoFocus
        aria-label="Todo title"
      />

      <TodoFields values={values} onChange={change} categories={categories} />

      <div className="form-actions">
        <button
          type="submit"
          className="btn primary"
          disabled={saving || !values.title.trim()}
        >
          Save
        </button>
        <button type="button" className="btn" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </form>
  );
}