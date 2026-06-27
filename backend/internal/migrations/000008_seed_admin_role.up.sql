INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES ('b7e2a1f3-9c4d-4e8b-a056-3f9d72c1e845', 'admin', 'Rol de administrador del sistema', NOW(), NOW())
ON CONFLICT (id) DO UPDATE
  SET name        = EXCLUDED.name,
      description = EXCLUDED.description,
      updated_at  = NOW();
