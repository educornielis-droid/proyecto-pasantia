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
    tipo_documento VARCHAR(2) NOT NULL DEFAULT 'V',       --PYEDE SER V, J, E, G
    numero_documento VARCHAR(20) NOT NULL,
    tipo_cuenta VARCHAR(4) NOT NULL DEFAULT 'CELE',      -- 'CELE' (Pago Móvil) o 'CNTA' (Cuenta 20 digitos)
    cuenta_o_telefono VARCHAR(20) NOT NULL,     
    banco_origen VARCHAR(4) NOT NULL,                    --CODIGO DEL IBP

    monto_final_usd NUMERIC(6, 2) NOT NULL CHECK (monto_final_usd > 0),
    monto_final_ves NUMERIC(10, 2) NOT NULL CHECK (monto_final_ves > 0),
    tasa_cambio NUMERIC(8, 2) NOT NULL CHECK (tasa_cambio > 0),

    estado_transaccion VARCHAR(20) DEFAULT 'PENDIENTE',
    referencia_sypago VARCHAR(100),
    fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE detalles (
    detalle_id SERIAL PRIMARY KEY,
    transaccion_id INT NOT NULL REFERENCES transacciones(transaccion_id) ON DELETE CASCADE,
    producto_id INT NOT NULL REFERENCES productos(producto_id) ON DELETE RESTRICT,
    cantidad_producto INT NOT NULL CHECK (cantidad_producto > 0)
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



INSERT INTO usuarios (nombre, apellido, contraseña)
VALUES ('admin', 'admin', 'admin123');

INSERT INTO productos (categoria_id, nombre, descripcion, precio, stock) VALUES 
(1, 'Camisa Oficial SyPago', 'Camisa 100% de algodon con logo estampado', 15.00, 20),
(1, 'Taza Termica SyPago', 'Taza de acero inoxidable para cafe', 8.50, 15);


-- Vista para observar mejor en qué categoría está cada producto
CREATE VIEW v_productos 
AS SELECT 
    p.producto_id,
    c.nombre AS nombre_categoria,
    p.nombre,
    p.descripcion,
    p.precio,
    p.stock,
    p.imagen_url
FROM productos p
JOIN categorias c ON p.categoria_id = c.categoria_id;

INSERT INTO categorias (categoria_id, nombre) VALUES
(1, 'Ropa y Accesorios'),
(2, 'Tecnología'),
(3, 'Periféricos'),
(4, 'Audio'),
(5, 'Móviles'),
(6, 'Accesorios'),
(7, 'Gaming'),
(8, 'Almacenamiento'),
(9, 'Smart Home'),
(10, 'Wearables'),
(11, 'Redes'),
(12, 'Oficina')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO productos (categoria_id, nombre, descripcion, precio, stock) VALUES
(2, 'Laptop Pro 15"', 'Computadora portátil con procesador de alto rendimiento y 16GB RAM.', 850.00, 10),
(2, 'Monitor Gamer 27"', 'Monitor 144Hz IPS 1ms ideal para trabajo y videojuegos.', 220.00, 15),
(3, 'Teclado Mecánico RGB', 'Teclado mecánico con switches azules y retroiluminación RGB.', 45.00, 30),
(3, 'Mouse Inalámbrico Ergonómico', 'Mouse óptico con sensor de alta precisión y batería recargable.', 25.00, 40),
(4, 'Auriculares Bluetooth Noise Cancelling', 'Auriculares Over-Ear con cancelación activa de ruido y 30h de batería.', 110.00, 20),
(4, 'Corneta Portátil Waterproof', 'Altavoz Bluetooth resistente al agua IPX7 con graves reforzados.', 35.00, 25),
(5, 'Smartphone X200 128GB', 'Teléfono inteligente con pantalla AMOLED y triple cámara de 64MP.', 320.00, 12),
(6, 'Cargador Rápido 65W GaN', 'Cargador compacto USB-C dual apto para laptops y teléfonos.', 28.00, 50),
(6, 'Hub USB-C 7 en 1', 'Adaptador multiporta con HDMI 4K, USB 3.0 y lector SD.', 32.00, 35),
(7, 'Silla Gamer Ergonómica', 'Silla con soporte lumbar reclinable y reposabrazos 3D.', 180.00, 8),
(7, 'Control Inalámbrico PC/Console', 'Mando ergonómico compatible con PC, Android y consolas.', 40.00, 22),
(4, 'Micrófono Condensador USB', 'Micrófono cardioide ideal para podcasts, streaming y llamadas.', 55.00, 18),
(2, 'Tablet 10" 64GB', 'Tablet pantalla HD ideal para lectura, consumo multimedia y clases.', 140.00, 14),
(6, 'Mochila Antirrobo para Laptop', 'Mochila impermeable con puerto de carga USB y compartimentos.', 30.00, 45),
(8, 'Disco Duro Externo 1TB', 'Almacenamiento portátil USB 3.0 resistente a impactos.', 50.00, 28),
(8, 'SSD M.2 NVMe 500GB', 'Unidad de estado sólido de alta velocidad de lectura/escritura.', 45.00, 30),
(9, 'Cámara de Seguridad Wi-Fi 1080p', 'Cámara IP de visión nocturna y detección de movimiento.', 38.00, 25),
(9, 'Lámpara LED de Escritorio Inteligente', 'Lámpara con ajuste de brillo, temperatura de color y temporizador.', 22.00, 30),
(10, 'Smartwatch Deportivo', 'Reloj inteligente con monitor de ritmo cardíaco, spO2 y GPS.', 65.00, 20),
(11, 'Router Wi-Fi 6 Doble Banda', 'Router Gigabit de alta velocidad para transmisión 4K y gaming.', 75.00, 15),
(12, 'Soporte de Aluminio para Laptop', 'Soporte plegable y regulable en altura para mejorar la postura.', 20.00, 50),
(3, 'Pad Mouse XL Gaming', 'Alfombrilla antideslizante bordada de 90x40 cm.', 15.00, 60),
(4, 'Barra de Sonido para TV/PC', 'Barra de sonido compacta con conexión Bluetooth y Auxiliar.', 60.00, 12),
(7, 'WebCam Full HD 1080p', 'Cámara web con micrófono integrado y tapa de privacidad.', 35.00, 25),
(6, 'Cable HDMI 2.1 8K', 'Cable de alta velocidad braided de 2 metros.', 12.00, 80);

INSERT INTO productos (categoria_id, nombre, descripcion, precio, stock) VALUES
(2, 'Mini PC Ryzen 7', 'Computadora ultracompacta con 32GB RAM y 1TB SSD ideal para trabajo pesado.', 520.00, 8),
(2, 'Monitor Portátil 15.6"', 'Pantalla secundaria IPS Full HD tipo C para laptop o consola.', 160.00, 12),
(3, 'Teclado Mecánico 60% Inalámbrico', 'Teclado compacto con conectividad Bluetooth/2.4G y switches rojos silenciosos.', 65.00, 20),
(3, 'Mouse Vertical Ergonómico', 'Mouse diseñado para prevenir fatiga en la muñeca con ajuste DPI.', 35.00, 25),
(4, 'Micrófono Lavarier Inalámbrico', 'Par de micrófonos de solapa tipo C/Lighting para grabaciones y entrevistas.', 42.00, 18),
(4, 'Auriculares In-Ear TWS Gamer', 'Auriculares intrauditivos de baja latencia con estuche de carga RGB.', 30.00, 35),
(5, 'Smartphone Pro Max 256GB', 'Teléfono gama alta con cámara de 108MP y carga ultra rápida de 120W.', 680.00, 6),
(5, 'Feature Phone Básico 4G', 'Teléfono celular clásico resistente con linterna y batería de larga duración.', 25.00, 40),
(6, 'Power Bank 20000mAh 22.5W', 'Batería externa de alta capacidad con pantalla LED e indicador de carga.', 38.00, 30),
(6, 'Organizador de Cables Rígido', 'Estuche de viaje impermeable para accesorios, cables y cargadores.', 18.00, 50),
(7, 'Volante y Pedales de Carreras', 'Set de simulación con respuesta de fuerza y rotación de 900 grados.', 240.00, 5),
(7, 'Escritorio Gamer con Luces LED', 'Mesa de juego ergonómica con soporte para audífonos y portavasos.', 150.00, 7),
(8, 'Memoria USB 3.2 128GB', 'Pendrive metálico de alta velocidad de transferencia.', 16.00, 60),
(8, 'Tarjeta MicroSD XC 256GB', 'Memoria de alta velocidad Clase 10 A2 ideal para videos en 4K.', 28.00, 45),
(9, 'Enchufe Inteligente Wi-Fi', 'Tomacorriente programable compatible con asistentes de voz.', 14.00, 50),
(9, 'Tira LED RGBIC 5 Metros', 'Cinta de luces inteligetes personalizable por segmentos desde la app.', 25.00, 30),
(10, 'Pulsera de Actividad Smart Band', 'Monitoreo de pasos, sueño, frecuencia cardíaca y 12 modos deportivos.', 35.00, 40),
(10, 'Reloj Inteligente Militar', 'Smartwatch ultrarresistente al agua, golpes y temperaturas extremas.', 85.00, 15),
(11, 'Repetidor Wi-Fi Dual Band', 'Extensor de cobertura inalámbrica con puertos Ethernet para mayor estabilidad.', 32.00, 22),
(11, 'Switch Gigabit 8 Puertos', 'Conmutador de red metálico Plug and Play para conectar múltiples equipos.', 28.00, 20),
(12, 'Lámpara Monitor Bar', 'Barra de luz LED para monitor con control táctil y sin reflejos en pantalla.', 40.00, 25),
(12, 'Soporte Ajustable para Smartphone/Tablet', 'Base metálica de escritorio multiposición plegable.', 15.00, 55),
(3, 'Webcam 2K con Aro de Luz', 'Cámara HD con iluminación regulable integrada para videollamadas.', 48.00, 18),
(4, 'Amplificador de Auriculares DAC USB', 'Convertidor de audio portátil para alta fidelidad de sonido.', 50.00, 10),
(6, 'Funda Protectora para Laptop 15.6"', 'Estuche acolchado e impermeable con bolsillo frontal para accesorios.', 18.00, 40);



UPDATE productos 
SET imagen_url = '/static/img/Gemini_Generated_Image_gl3lkjgl3lkjgl3l.png'
WHERE producto_id = 1;

UPDATE productos 
SET imagen_url = '/static/img/Gemini_Generated_Image_jq6re6jq6re6jq6r.png'
WHERE producto_id = 2;

-- no hay id=3

UPDATE productos 
SET imagen_url = '/static/img/Laptop_Pro_15.png'
WHERE producto_id = 4;

UPDATE productos 
SET imagen_url = '/static/img/Monitor_Gamer_27.png'
WHERE producto_id = 5;

UPDATE productos 
SET imagen_url = '/static/img/Teclado_Mecánico_RGB.png'
WHERE producto_id = 6;

UPDATE productos 
SET imagen_url = '/static/img/Mouse_Inalámbrico_Ergonómico.png'
WHERE producto_id = 7;

UPDATE productos 
SET imagen_url = '/static/img/Auriculares_Bluetooth_Noise_Cancelling.png'
WHERE producto_id = 8;

UPDATE productos 
SET imagen_url = '/static/img/Corneta_Portátil_Waterproof.png'
WHERE producto_id = 9;

UPDATE productos 
SET imagen_url = '/static/img/Smartphone_X200_128GB.png'
WHERE producto_id = 10;

UPDATE productos 
SET imagen_url = '/static/img/Cargador_Rápido_65W_GaN.png'
WHERE producto_id = 11;

UPDATE productos 
SET imagen_url = '/static/img/Laptop_Pro_15.png'
WHERE producto_id = 4;

UPDATE productos 
SET imagen_url = '/static/img/Monitor_Gamer_27.png'
WHERE producto_id = 5;

UPDATE productos 
SET imagen_url = '/static/img/Teclado_Mecánico_RGB.png'
WHERE producto_id = 6;

UPDATE productos 
SET imagen_url = '/static/img/Mouse_Inalámbrico_Ergonómico.png'
WHERE producto_id = 7;

UPDATE productos 
SET imagen_url = '/static/img/Auriculares_Bluetooth_Noise_Cancelling.png'
WHERE producto_id = 8;

UPDATE productos 
SET imagen_url = '/static/img/Corneta_Portátil_Waterproof.png'
WHERE producto_id = 9;

UPDATE productos 
SET imagen_url = '/static/img/Smartphone_X200_128GB.png'
WHERE producto_id = 10;

UPDATE productos 
SET imagen_url = '/static/img/Cargador_Rápido_65W_GaN.png'
WHERE producto_id = 11;

UPDATE productos 
SET imagen_url = '/static/img/Hub_USB-C_7_en_1.png'
WHERE producto_id = 12;

UPDATE productos 
SET imagen_url = '/static/img/Silla_Gamer_Ergonómica.png'
WHERE producto_id = 13;

UPDATE productos 
SET imagen_url = '/static/img/Control_Inalámbrico_PC_Console.png'
WHERE producto_id = 14;

UPDATE productos 
SET imagen_url = '/static/img/Micrófono_Condensador_USB.png'
WHERE producto_id = 15;

UPDATE productos 
SET imagen_url = '/static/img/Tablet_10_64GB.png'
WHERE producto_id = 16;

UPDATE productos 
SET imagen_url = '/static/img/Mochila_Antirrobo_para_Laptop.png'
WHERE producto_id = 17;

UPDATE productos 
SET imagen_url = '/static/img/Disco_Duro_Externo_1TB.png'
WHERE producto_id = 18;

UPDATE productos 
SET imagen_url = '/static/img/SSD_M.2_NVMe_500GB.png'
WHERE producto_id = 19;

UPDATE productos 
SET imagen_url = '/static/img/Cámara_Seguridad_Wi-Fi_1080p.png'
WHERE producto_id = 20;


UPDATE productos 
SET imagen_url = '/static/img/Lámpara_LED_Escritorio_Inteligente.png'
WHERE producto_id = 21;

UPDATE productos 
SET imagen_url = '/static/img/Smartwatch_Deportivo.png'
WHERE producto_id = 22;

UPDATE productos 
SET imagen_url = '/static/img/Router_Wi-Fi_6_Doble_Banda.png'
WHERE producto_id = 23;

UPDATE productos 
SET imagen_url = '/static/img/Soporte_Aluminio_para_Laptop.png'
WHERE producto_id = 24;

UPDATE productos 
SET imagen_url = '/static/img/Pad_Mouse_XL_Gaming.png'
WHERE producto_id = 25;

UPDATE productos 
SET imagen_url = '/static/img/Barra_Sonido_para_TVPC.png'
WHERE producto_id = 26;

UPDATE productos 
SET imagen_url = '/static/img/WebCam_Full_HD_1080p.png'
WHERE producto_id = 27;

UPDATE productos 
SET imagen_url = '/static/img/Cable_HDMI_2.1_8K.png'
WHERE producto_id = 28;

-- UPDATE productos 
-- SET imagen_url = '/static/img/Cámara_Seguridad_Wi-Fi_1080p.png'
-- WHERE producto_id = 29;

-- UPDATE productos 
-- SET imagen_url = '/static/img/Cámara_Seguridad_Wi-Fi_1080p.png'
-- WHERE producto_id = 30;

UPDATE productos 
SET imagen_url = '/static/img/Gorra_Sypago.png'
WHERE producto_id = 31;

UPDATE productos 
SET imagen_url = '/static/img/Mini_PC_Ryzen_7.png'
WHERE producto_id = 32;

UPDATE productos 
SET imagen_url = '/static/img/Monitor_Portátil_15.6.png'
WHERE producto_id = 33;

UPDATE productos 
SET imagen_url = '/static/img/Teclado_Mecánico_60%_Inalámbrico.png'
WHERE producto_id = 34;



-- Reemplazo de las tablas "detalles" y "transacciones"

DROP TABLE IF EXISTS detalles CASCADE;
DROP TABLE IF EXISTS transacciones CASCADE;

-- Cabecera de cada compra (una fila por transacción de pago)
CREATE TABLE transacciones (
    transaccion_id       TEXT PRIMARY KEY,              -- el idTransaccion que genera mi comercio (12 hex)
    tipo_documento        TEXT NOT NULL,                 -- V, E, J, G
    numero_documento      TEXT NOT NULL,
    tipo_cuenta           TEXT NOT NULL,                 -- CELE o CNTA
    cuenta_o_telefono     TEXT NOT NULL,
    banco_origen          TEXT NOT NULL,                 -- del banco pagador
    monto_final_usd       NUMERIC(12,2) NOT NULL,
    monto_final_ves       NUMERIC(14,2) NOT NULL,
    tasa_cambio           NUMERIC(10,4) NOT NULL,        
    estado_transaccion    TEXT NOT NULL DEFAULT 'PEND',  -- PEND, PROC, ACCP, RJCT, CANC
    referencia_sypago     TEXT,                          -- transaction_id que devuelve Sypago
    fecha_creacion        TIMESTAMP NOT NULL DEFAULT now(),
    fecha_actualizacion   TIMESTAMP NOT NULL DEFAULT now(), -- para saber cuándo cambió de estado por última vez (útil para el polling)
    codigo_rechazo        TEXT                              -- (ej. AC04, AM04) cuando estado_transaccion = 'RJCT', para poder mostrarle al usuario o revisarlo después
);


-- Detalle: qué productos y cantidades tenía cada transacción
CREATE TABLE detalles (
    detalle_id         SERIAL PRIMARY KEY,
    transaccion_id      TEXT NOT NULL REFERENCES transacciones(transaccion_id) ON DELETE CASCADE,
    producto_id          INTEGER NOT NULL REFERENCES productos(producto_id),
    cantidad_producto    INTEGER NOT NULL
);

-- Índice para consultar rápido "todas las transacciones de un producto" o "el detalle de una transacción" sin escanear toda la tabla.
CREATE INDEX idx_detalles_transaccion ON detalles(transaccion_id);
CREATE INDEX idx_detalles_producto ON detalles(producto_id);