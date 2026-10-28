import { useEffect, useState } from "react";
import TodoFields from "./TodoFields";
import { EMPTY_VALUES, valuesToPayload } from "../utils";

export default function TodoForm({ categories, defaultCategory, onAdd }) {
  const preset = /^\d+$/.test(defaultCategory) ? defaultCategory : "";
  const [values, setValues] = useState({ ...EMPTY_VALUES, categoryId: preset });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setValues((v) => ({ ...v, categoryId: preset }));
  }, [preset]);

  function change(field, value) {
    setValues((v) => ({ ...v, [field]: value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();

    if (!values.title.trim()) return;

    setSaving(true);
    const ok = await onAdd(valuesToPayload(values));
    setSaving(false);

    if (ok) setValues({ ...EMPTY_VALUES, categoryId: preset });
  }

  return (
    <form className="card add-form" onSubmit={handleSubmit}>
      <div className="add-row">
        <input
          type="text"
          value={values.title}
          onChange={(e) => change("title", e.target.value)}
          placeholder="What needs to be done?"
          maxLength={200}
          aria-label="New todo title"
        />
        <button
          type="submit"
          className="btn primary"
          disabled={saving || !values.title.trim()}
        >
          Add
        </button>
      </div>

      <details className="more">
        <summary>Details</summary>
        <TodoFields values={values} onChange={change} categories={categories} />
      </details>
    </form>
  );
}