import { PRIORITIES, capitalize } from "../utils";

export default function TodoFields({ values, onChange, categories }) {
  return (
    <div className="fields">
      <label className="field">
        <span>Priority</span>
        <select
          value={values.priority}
          onChange={(e) => onChange("priority", e.target.value)}
        >
          {PRIORITIES.map((p) => (
            <option key={p} value={p}>
              {capitalize(p)}
            </option>
          ))}
        </select>
      </label>

      <label className="field">
        <span>Due date</span>
        <input
          type="date"
          value={values.dueDate}
          onChange={(e) => onChange("dueDate", e.target.value)}
        />
      </label>

      <label className="field">
        <span>Category</span>
        <select
          value={values.categoryId}
          onChange={(e) => onChange("categoryId", e.target.value)}
        >
          <option value="">No category</option>
          {categories.map((c) => (
            <option key={c.id} value={String(c.id)}>
              {c.name}
            </option>
          ))}
        </select>
      </label>

      <label className="field wide">
        <span>Notes</span>
        <textarea
          rows={3}
          maxLength={2000}
          value={values.notes}
          onChange={(e) => onChange("notes", e.target.value)}
        />
      </label>
    </div>
  );
}