/* ============================================================
   PRODUCTOS.HTML - Showcase de un producto + filtro de categorías
   No depende de ninguna librería externa (Swiper, etc.) - es un
   slider propio, liviano, hecho a la medida de este layout.
   No toca carrito.js: el botón "Agregar al Carrito" que se genera
   aquí usa la misma clase .agregar-carrito que carrito.js ya
   escucha por delegación de eventos en document, así que funciona
   automáticamente sin cambiar nada de ese archivo.
   ============================================================ */

console.log("[Productos] productos.js se cargó correctamente.");

let listaProductos = [];
let categoriaActual = "todas";
let indiceActual = 0;

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", inicializarProductos);
} else {
    inicializarProductos();
}

function inicializarProductos() {
    listaProductos = leerProductosDesdeElDOM();

    if (listaProductos.length === 0) {
        mostrarSinProductos();
        return;
    }

    construirDropdownCategorias();
    renderizarProductoActual();

    const botonMenuProductos = document.getElementById("btn-menu-productos");
    const dropdownCategorias = document.getElementById("categorias-dropdown");

    if (botonMenuProductos && dropdownCategorias) {
        botonMenuProductos.addEventListener("click", function (evento) {
            evento.preventDefault();
            evento.stopPropagation();
            dropdownCategorias.classList.toggle("abierto");
        });

        // Cerrar el dropdown si el usuario hace click afuera
        document.addEventListener("click", function (evento) {
            if (!dropdownCategorias.contains(evento.target) && evento.target !== botonMenuProductos) {
                dropdownCategorias.classList.remove("abierto");
            }
        });
    }

    const flechaAnterior = document.getElementById("flecha-anterior");
    const flechaSiguiente = document.getElementById("flecha-siguiente");

    if (flechaAnterior) {
        flechaAnterior.addEventListener("click", function () {
            moverShowcase(-1);
        });
    }

    if (flechaSiguiente) {
        flechaSiguiente.addEventListener("click", function () {
            moverShowcase(1);
        });
    }
}

/* ------------------------------------------------------------
   LEER LOS PRODUCTOS QUE EL BACKEND YA RENDERIZÓ EN EL DOM
------------------------------------------------------------- */
function leerProductosDesdeElDOM() {
    const nodosProducto = document.querySelectorAll("#datos-productos .producto-dato");

    return Array.from(nodosProducto).map(function (nodo) {
        return {
            nombre: nodo.dataset.nombre,
            descripcion: nodo.dataset.descripcion,
            precio: parseFloat(nodo.dataset.precio),
            imagen: nodo.dataset.imagen || "",
            categoria: nodo.dataset.categoria || "Sin categoría",
            stock: parseInt(nodo.dataset.stock, 10) || 0
        };
    });
}

/* ------------------------------------------------------------
   DROPDOWN DE CATEGORÍAS (se arma solo, a partir de los productos)
------------------------------------------------------------- */
function construirDropdownCategorias() {
    const dropdownCategorias = document.getElementById("categorias-dropdown");
    if (!dropdownCategorias) return;

    const categoriasUnicas = Array.from(new Set(listaProductos.map(function (producto) {
        return producto.categoria;
    })));

    dropdownCategorias.innerHTML = "";

    const botonTodas = crearBotonCategoria("todas", "Todas las categorías");
    dropdownCategorias.appendChild(botonTodas);

    categoriasUnicas.forEach(function (categoria) {
        dropdownCategorias.appendChild(crearBotonCategoria(categoria, categoria));
    });

    actualizarBotonCategoriaActiva();
}

function crearBotonCategoria(valorCategoria, textoVisible) {
    const boton = document.createElement("button");
    boton.type = "button";
    boton.textContent = textoVisible;
    boton.dataset.categoria = valorCategoria;

    boton.addEventListener("click", function () {
        categoriaActual = valorCategoria;
        indiceActual = 0;
        actualizarBotonCategoriaActiva();

        const tarjeta = document.getElementById("showcase-contenido");
        tarjeta.classList.add("showcase-anim-salida-izq");
        setTimeout(function () {
            renderizarProductoActual();
            tarjeta.classList.remove("showcase-anim-salida-izq");
            tarjeta.classList.add("showcase-anim-entrada-der");
            void tarjeta.offsetWidth;
            tarjeta.classList.remove("showcase-anim-entrada-der");
        }, DURACION_ANIMACION_MS);

        document.getElementById("categorias-dropdown").classList.remove("abierto");
    });

    return boton;
}

function actualizarBotonCategoriaActiva() {
    document.querySelectorAll("#categorias-dropdown button").forEach(function (boton) {
        boton.classList.toggle("categoria-activa", boton.dataset.categoria === categoriaActual);
    });
}

/* ------------------------------------------------------------
   FILTRADO POR CATEGORÍA
------------------------------------------------------------- */
function productosFiltrados() {
    if (categoriaActual === "todas") {
        return listaProductos;
    }
    return listaProductos.filter(function (producto) {
        return producto.categoria === categoriaActual;
    });
}

/* ------------------------------------------------------------
   RENDERIZAR EL PRODUCTO ACTUAL EN EL SHOWCASE
------------------------------------------------------------- */
function renderizarProductoActual() {
    const showcaseTarjeta = document.getElementById("showcase-contenido");
    const showcaseIndicador = document.getElementById("showcase-indicador");
    const flechaAnterior = document.getElementById("flecha-anterior");
    const flechaSiguiente = document.getElementById("flecha-siguiente");
    const tituloCategoria = document.getElementById("titulo-categoria-actual");

    const productos = productosFiltrados();

    tituloCategoria.textContent = categoriaActual === "todas" ? "Todos los productos" : categoriaActual;

    if (productos.length === 0) {
        showcaseTarjeta.innerHTML = '<p class="showcase-sin-productos">No hay productos en esta categoría todavía.</p>';
        showcaseIndicador.textContent = "";
        flechaAnterior.disabled = true;
        flechaSiguiente.disabled = true;
        return;
    }

    // Nos aseguramos de que el índice esté dentro de rango (por si venías
    // de una categoría con más productos que la nueva).
    if (indiceActual >= productos.length) {
        indiceActual = 0;
    }

    const producto = productos[indiceActual];

    showcaseTarjeta.innerHTML = `
        <div class="showcase-izquierda">
            <h3 class="showcase-nombre">${escaparTexto(producto.nombre)}</h3>
            <p class="showcase-descripcion">${escaparTexto(producto.descripcion)}</p>
            <p class="showcase-stock">Unidades disponibles: ${producto.stock}</p>
            <a href="#" class="agregar-carrito btn-3 showcase-boton"
               data-nombre="${escaparAtributo(producto.nombre)}"
               data-precio="${producto.precio}"
               data-imagen="${escaparAtributo(producto.imagen)}">
               Agregar al Carrito
            </a>
        </div>
        <div class="showcase-centro">
            <img class="showcase-imagen" src="${escaparAtributo(producto.imagen)}" alt="${escaparAtributo(producto.nombre)}">
            <div class="showcase-sombra"></div>
        </div>
        <div class="showcase-derecha">
            <span class="showcase-precio-label">Precio</span>
            <span class="showcase-precio">$${producto.precio.toFixed(2)}</span>
        </div>
    `;

    showcaseIndicador.textContent = (indiceActual + 1) + " de " + productos.length;

    // Con 1 solo producto no tiene sentido mostrar flechas activas
    const soloUnProducto = productos.length <= 1;
    flechaAnterior.disabled = soloUnProducto;
    flechaSiguiente.disabled = soloUnProducto;
}

const DURACION_ANIMACION_MS = 220;
let animandoShowcase = false;

function moverShowcase(direccion) {
    const productos = productosFiltrados();
    if (productos.length <= 1 || animandoShowcase) return;

    const tarjeta = document.getElementById("showcase-contenido");
    animandoShowcase = true;

    // 1. El producto actual "sale" hacia el lado contrario a donde apunta la flecha
    tarjeta.classList.add(direccion > 0 ? "showcase-anim-salida-izq" : "showcase-anim-salida-der");

    setTimeout(function () {
        indiceActual = (indiceActual + direccion + productos.length) % productos.length;
        renderizarProductoActual();

        // 2. Colocamos el nuevo contenido ya desplazado, del lado por donde "entra"
        tarjeta.classList.remove("showcase-anim-salida-izq", "showcase-anim-salida-der");
        tarjeta.classList.add(direccion > 0 ? "showcase-anim-entrada-der" : "showcase-anim-entrada-izq");

        // 3. Forzamos que el navegador registre esa posición inicial antes de animar
        void tarjeta.offsetWidth;

        // 4. Quitamos la clase de "entrada": como .showcase-contenido ya tiene
        //    transition definida, esto anima suavemente hasta la posición normal
        tarjeta.classList.remove("showcase-anim-entrada-der", "showcase-anim-entrada-izq");

        setTimeout(function () {
            animandoShowcase = false;
        }, DURACION_ANIMACION_MS);
    }, DURACION_ANIMACION_MS);
}

function mostrarSinProductos() {
    const showcaseTarjeta = document.getElementById("showcase-contenido");
    if (showcaseTarjeta) {
        showcaseTarjeta.innerHTML = '<p class="showcase-sin-productos">No hay productos disponibles por el momento.</p>';
    }
    const flechaAnterior = document.getElementById("flecha-anterior");
    const flechaSiguiente = document.getElementById("flecha-siguiente");
    if (flechaAnterior) flechaAnterior.disabled = true;
    if (flechaSiguiente) flechaSiguiente.disabled = true;
}

/* ------------------------------------------------------------
   PEQUEÑAS UTILIDADES DE ESCAPE
   (los datos ya vienen de tu propia BD, pero nunca está de más
   evitar que un nombre/descripción con comillas rompa el HTML)
------------------------------------------------------------- */
function escaparTexto(texto) {
    const contenedorTemporal = document.createElement("div");
    contenedorTemporal.textContent = texto || "";
    return contenedorTemporal.innerHTML;
}

function escaparAtributo(texto) {
    return (texto || "").replace(/"/g, "&quot;");
}