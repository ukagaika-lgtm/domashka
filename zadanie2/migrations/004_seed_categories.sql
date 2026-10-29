INSERT OR IGNORE INTO categories (name) VALUES
    ('Еда'),
    ('Транспорт'),
    ('Жильё'),
    ('Развлечения'),
    ('Здоровье'),
    ('Другое');

UPDATE expenses
SET category_id = (
    SELECT id FROM categories WHERE name = 'Другое'
)
WHERE category_id IS NULL;
