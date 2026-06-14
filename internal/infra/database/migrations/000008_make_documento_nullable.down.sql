-- Cannot restore NOT NULL if there are existing rows with NULL, but we can try
ALTER TABLE mandados ALTER COLUMN tipo_documento SET NOT NULL;
