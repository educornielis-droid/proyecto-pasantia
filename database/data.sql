CREATE TABLE productos {
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    descripcion VARCHAR(500),
    categoria VARCHAR(50) NOT NULL, --trabajarla así o crear una tabla directamente para eso?
    precio NUMERIC(6, 2) NOT NULL CHECK (precio > 0), --4 enteros y dos decimales
    stock INT NOT NULL CHECK (stock >= 0), --precios y stocks
    imagen_url VARCHAR(255)
};

CREATE TABLE transacciones {
    id_orden SERIAL PRIMARY KEY,
    referencia_orden VARCHAR(100)
    fecha_orden VARCHAR(500)
    precio_producto NUMERIC(100)
    tipo_producto VARCHAR(70)
    imagen_producto 
};

CREATE TABLE detalles {
    id SERIAL PRIMARY KEY,
    nombre_producto VARCHAR(100)
    descripcion_producto VARCHAR(500)
    precio_producto NUMERIC(100)
    tipo_producto VARCHAR(70)
    imagen_producto 
};

CREATE TABLE usuarios {
    id SERIAL PRIMARY KEY,
    nombre_producto VARCHAR(100)
    descripcion_producto VARCHAR(500)
    precio_producto NUMERIC(100)
    tipo_producto VARCHAR(70)
    imagen_producto 
};


CREATE TABLE logs {
    id
}