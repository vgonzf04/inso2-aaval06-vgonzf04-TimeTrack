CREATE TABLE IF NOT EXISTS empleados (
    id SERIAL PRIMARY KEY,
    nombre TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    cargo TEXT NOT NULL,
    fecha_contratacion DATE NOT NULL,
    supervisor_id INTEGER
);

CREATE TABLE IF NOT EXISTS fichajes (
    id SERIAL PRIMARY KEY,
    empleado_id INTEGER NOT NULL REFERENCES empleados(id),
    entrada TIMESTAMP NOT NULL,
    salida TIMESTAMP,
    ubicacion TEXT
);

CREATE TABLE IF NOT EXISTS vacaciones (
    id SERIAL PRIMARY KEY,
    empleado_id INTEGER NOT NULL REFERENCES empleados(id),
    inicio DATE NOT NULL,
    fin DATE NOT NULL,
    estado TEXT NOT NULL CHECK (estado IN ('pendiente', 'aprobada', 'rechazada'))
);
