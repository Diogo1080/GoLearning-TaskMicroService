DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'tasks' AND column_name = 'duedate'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'tasks' AND column_name = 'due_date'
    ) THEN
        ALTER TABLE tasks RENAME COLUMN duedate TO due_date;
    END IF;
END $$;

ALTER TABLE tasks
    ALTER COLUMN priority SET DEFAULT 1,
    ALTER COLUMN completed SET DEFAULT false;

ALTER TABLE tasks
    ADD CONSTRAINT tasks_priority_check CHECK (priority BETWEEN 1 AND 3);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks (user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_due_date ON tasks (user_id, due_date);
