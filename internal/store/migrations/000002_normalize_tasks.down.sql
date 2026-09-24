DROP INDEX IF EXISTS idx_tasks_user_due_date;
DROP INDEX IF EXISTS idx_tasks_user_id;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_priority_check,
    ALTER COLUMN priority DROP DEFAULT,
    ALTER COLUMN completed DROP DEFAULT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'tasks' AND column_name = 'due_date'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'tasks' AND column_name = 'duedate'
    ) THEN
        ALTER TABLE tasks RENAME COLUMN due_date TO duedate;
    END IF;
END $$;
