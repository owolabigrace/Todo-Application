import { PRIORITIES, capitalize } from "../utils";

const STATUSES = [
  ["all", "All"],
  ["active", "Active"],
  ["completed", "Completed"],
];

const SORTS = [
  ["created", "Date created"],
  ["due", "Due date"],
  ["priority", "Priority"],
  ["title", "Title"],
];

export default function Filters({
  filters,
  searchText,
  onSearchChange,
  onChange,
  onClear,
  hasActive,
}) {
  return (
    <section className="filters" aria-label="Filter and sort todos">
      <input
        type="search"
        className="search"
        placeholder="Search title and notes"
        value={searchText}
        maxLength={200}
        onChange={(e) => onSearchChange(e.target.value)}
        aria-label="Search todos"
      />

      <div className="segmented" role="group" aria-label="Status">
        {STATUSES.map(([value, label]) => (
          <button
            type="button"
            key={value}
            className={filters.status === value ? "active" : ""}
            aria-pressed={filters.status === value}
            onClick={() => onChange({ status: value })}
          >
            {label}
          </button>
        ))}
      </div>

      <select
        value={filters.priority}
        onChange={(e) => onChange({ priority: e.target.value })}
        aria-label="Filter by priority"
      >
        <option value="">Any priority</option>
        {PRIORITIES.map((p) => (
          <option key={p} value={p}>
            {capitalize(p)}
          </option>
        ))}
      </select>

      <select
        value={filters.sort}
        onChange={(e) => onChange({ sort: e.target.value })}
        aria-label="Sort by"
      >
        {SORTS.map(([value, label]) => (
          <option key={value} value={value}>
            {label}
          </option>
        ))}
      </select>

      <button
        type="button"
        className="btn"
        onClick={() =>
          onChange({ order: filters.order === "asc" ? "desc" : "asc" })
        }
        aria-label={`Sort order: ${
          filters.order === "asc" ? "ascending" : "descending"
        }`}
      >
        {filters.order === "asc" ? "↑ Asc" : "↓ Desc"}
      </button>

      {hasActive && (
        <button type="button" className="link" onClick={onClear}>
          Clear filters
        </button>
      )}
    </section>
  );
}