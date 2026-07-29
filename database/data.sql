DROP VIEW IF EXISTS v_resumen_transacciones;
DROP TABLE IF EXISTS detalles;
DROP TABLE IF EXISTS transacciones;
DROP TABLE IF EXISTS productos;
DROP TABLE IF EXISTS categorias;
DROP TABLE IF EXISTS usuarios;

CREATE TABLE categorias (
    categoria_id SERIAL PRIMARY KEY,
    nombre VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE productos (
    producto_id SERIAL PRIMARY KEY,
    categoria_id INT NOT NULL REFERENCES categorias(categoria_id) ON DELETE RESTRICT,
    nombre VARCHAR(100) NOT NULL,
    descripcion VARCHAR(500),
    precio NUMERIC(6, 2) NOT NULL CHECK (precio > 0),
    stock INT NOT NULL CHECK (stock >= 0),
    imagen_url VARCHAR(255)
);

CREATE TABLE transacciones (
    transaccion_id SERIAL PRIMARY KEY,
    --DATOS REQUERIDOS POR SYPAGO
    tipo_documento VARCHAR(2) NOT NULL DEFAULT 'V'       --PYEDE SER V, J, E, G
    numero_documento VARCHAR(20) NOT NULL,
    tipo_cuenta VARCHAR(4) NOT NULL DEFAULT 'CELE',      -- 'CELE' (Pago Móvil) o 'CNTA' (Cuenta 20 digitos)
    cuenta_o_telefono VARCHAR(20) NOT NULL,     
    banco_origen VARCHAR(4) NOT NULL,                    --CODIGO DEL IBP

    monto_final_usd NUMERIC(6, 2) NOT NULL CHECK (monto_final_usd > 0),
    monto_final_ves NUMERIC(10, 2) NOT NULL CHECK (monto_final_ves > 0),
    tasa_cambio NUMERIC(8, 2) NOT NULL CHECK (tasa_cambio > 0)

    estado_transaccion VARCHAR(20) DEFAULT 'PENDIENTE',
    referencia_sypago VARCHAR(100),
    fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE detalles (
    detalle_id SERIAL PRIMARY KEY,
    transaccion_id INT NOT NULL REFERENCES transacciones(transaccion_id) ON DELETE CASCADE,
    producto_id INT NOT NULL REFERENCES productos(producto_id) ON DELETE RESTRICT,
    cantidad_producto INT NOT NULL CHECK (cantidad_producto > 0),
);

CREATE TABLE usuarios (
    usuario_id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    apellido VARCHAR(100) NOT NULL,
    correo VARCHAR(150), 
    contraseña VARCHAR(255) NOT NULL,
    fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);




-- Vista para consultar transacciones con detalles consolidados para el Panel Administrativo
CREATE VIEW v_resumen_transacciones 
AS SELECT 
    t.transaccion_id,
    CONCAT(t.tipo_documento, '-', t.numero_documento) AS documento_completo,
    t.cuenta_o_telefono,
    t.banco_origen,
    t.monto_final_usd,
    t.monto_final_ves,
    t.tasa_cambio,
    t.estado_transaccion,
    t.referencia_sypago,
    t.fecha_creacion,
    COUNT(d.detalle_id) AS total_items_comprados
FROM transacciones t
LEFT JOIN detalles d ON t.transaccion_id = d.transaccion_id
GROUP BY t.transaccion_id;



INSERT INTO usuarios (nombre, contraseña)
VALUES ('admin', 'admin123')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO productos (nombre, descripcion, precio, stock)
VALUES ('Camisa Oficial SyPago', 'Camisa 100% de algodon con logo estampado', 15.00, 20),
('Taza Termica SyPago', 'Taza de acero inoxidable para cafe', 8.50, 15);


-- Vista para observar mejor en que categoria esta cada producto
CREATE VIEW v_productos 
AS SELECT 
    p.producto_id,
    c.nombre AS nombre_categoria
    p.nombre,
    p.descripcion,
    p.precio,
    p.stock,
    p.imagen_url
FROM productos p
JOIN categorias c ON p.categoria_id = c.categoria_id;
