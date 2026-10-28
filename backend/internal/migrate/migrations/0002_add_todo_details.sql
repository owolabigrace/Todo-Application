ALTER TABLE todos
    ADD COLUMN notes    TEXT NOT NULL DEFAULT '' CHECK (char_length(notes) <= 2000),
    ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    ADD COLUMN due_date DATE;
