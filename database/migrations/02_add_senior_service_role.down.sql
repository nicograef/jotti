-- Reassign any senior_service users to service before removing the enum value
UPDATE users SET role = 'service' WHERE role = 'senior_service';

-- PostgreSQL does not support removing a single enum value.
-- Recreate the enum type without 'senior_service'.
ALTER TYPE UserRole RENAME TO UserRole_old;
CREATE TYPE UserRole AS ENUM ('admin', 'service');
ALTER TABLE users ALTER COLUMN role TYPE UserRole USING role::text::UserRole;
DROP TYPE UserRole_old;
