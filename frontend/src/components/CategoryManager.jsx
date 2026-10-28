import { useState } from "react";

export default function CategoryManager({
  categories,
  selected,
  onSelect,
  onCreate,
  onRename,
  onDelete,
}) {
  const [newName, setNewName] = useState("");
  const [editingId, setEditingId] = useState(null);
  const [editName, setEditName] = useState("");

  async function handleCreate(e) {
    e.preventDefault();

    const name = newName.trim();
    if (!name) return;

    if (await onCreate(name)) setNewName("");
  }

  function startRename(category) {
    setEditingId(category.id);
    setEditName(category.name);
  }

  async function handleRename(e) {
    e.preventDefault();

    const name = editName.trim();
    if (!name) return;

    if (await onRename(editingId, name)) setEditingId(null);
  }

  function handleDelete(category) {
    const message = `Delete category "${category.name}"? Its todos will be kept and become uncategorized.`;

    if (window.confirm(message)) onDelete(category.id);
  }

  return (
    <nav aria-label="Categories">
      <h2>Categories</h2>

      <ul className="cat-list">
        <li className="cat-row">
          <button
            type="button"
            className={`cat-select${selected === "" ? " active" : ""}`}
            onClick={() => onSelect("")}
          >
            All todos
          </button>
        </li>

        <li className="cat-row">
          <button
            type="button"
            className={`cat-select${selected === "none" ? " active" : ""}`}
            onClick={() => onSelect("none")}
          >
            Uncategorized
          </button>
        </li>

        {categories.map((c) =>
          editingId === c.id ? (
            <li key={c.id} className="cat-row">
              <form className="cat-edit" onSubmit={handleRename}>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  maxLength={50}
                  autoFocus
                  aria-label="Category name"
                  onKeyDown={(e) => {
                    if (e.key === "Escape") setEditingId(null);
                  }}
                />
                <button
                  type="submit"
                  className="icon-btn"
                  disabled={!editName.trim()}
                >
                  Save
                </button>
                <button
                  type="button"
                  className="icon-btn"
                  onClick={() => setEditingId(null)}
                >
                  Cancel
                </button>
              </form>
            </li>
          ) : (
            <li key={c.id} className="cat-row">
              <button
                type="button"
                className={`cat-select${
                  selected === String(c.id) ? " active" : ""
                }`}
                onClick={() => onSelect(String(c.id))}
              >
                {c.name}
              </button>
              <button
                type="button"
                className="icon-btn"
                onClick={() => startRename(c)}
                aria-label={`Rename category ${c.name}`}
              >
                Rename
              </button>
              <button
                type="button"
                className="icon-btn"
                onClick={() => handleDelete(c)}
                aria-label={`Delete category ${c.name}`}
              >
                Delete
              </button>
            </li>
          )
        )}
      </ul>

      <form className="cat-add" onSubmit={handleCreate}>
        <input
          type="text"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder="New category"
          maxLength={50}
          aria-label="New category name"
        />
        <button type="submit" className="btn" disabled={!newName.trim()}>
          Add
        </button>
      </form>
    </nav>
  );
}