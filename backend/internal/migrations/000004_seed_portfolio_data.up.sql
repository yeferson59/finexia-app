INSERT INTO risks (id, name, description, created_at, updated_at)
VALUES
  ('25549e04-2eb7-4a05-9f07-4698324588ce', 'Bajo Riesgo', 'Inversiones conservadoras', NOW(), NOW()),
  ('62f30795-c2d5-4cdd-8c67-aef7d588aefa', 'Riesgo Moderado', 'Balance entre riesgo y retorno', NOW(), NOW()),
  ('5bb911c5-a14f-4a50-aa8d-3f032baf1cf5', 'Alto Riesgo', 'Busca máximo crecimiento', NOW(), NOW())
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      description = EXCLUDED.description,
      updated_at = NOW();

INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES
  ('25549e04-2eb7-4a05-9f07-4698324588ce', 'customer', 'Rol por defecto para usuarios autenticados', NOW(), NOW())
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name,
      description = EXCLUDED.description,
      updated_at = NOW();
