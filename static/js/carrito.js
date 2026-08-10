/* ============================================================
   SYPAGO STORE - LÓGICA DEL CARRITO DE COMPRAS
   Archivo independiente: solo carrito (no toca Swiper ni nada
   relacionado con el resto de script.js).
   Persistencia: sessionStorage
   ============================================================ */

console.log("[Carrito] carrito.js se cargó correctamente.");

// Clave usada para guardar el carrito en sessionStorage
const CLAVE_STORAGE_CARRITO = "carritoSypago";

/* ------------------------------------------------------------
   1. REFERENCIAS AL DOM
   (se buscan una sola vez al cargar el documento)
------------------------------------------------------------- */
let cuerpoTablaCarrito;        // <tbody> dentro de #lista-carrito
let mensajeCarritoVacio;       // #carrito-vacio-msg
let contenedorAccionesCarrito; // #carrito-acciones
let elementoTotalMonto;        // #total-monto
let botonVaciarCarrito;        // #vaciar-carrito
let botonPagarCarrito;         // #pagar-carrito

/* ------------------------------------------------------------
   2. INICIALIZACIÓN
   Se ejecuta inmediatamente si el DOM ya terminó de cargar
   (por ejemplo si el script se inyecta o carga tarde), o espera
   a "DOMContentLoaded" en caso contrario. Esto evita el caso en
   que el evento ya se disparó antes de registrar el listener.
------------------------------------------------------------- */
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", inicializarCarrito);
} else {
    inicializarCarrito();
}

function inicializarCarrito() {
    try {
        console.log("[Carrito] Iniciando búsqueda de elementos del DOM...");

        cuerpoTablaCarrito = document.querySelector("#lista-carrito tbody");
        mensajeCarritoVacio = document.getElementById("carrito-vacio-msg");
        contenedorAccionesCarrito = document.getElementById("carrito-acciones");
        elementoTotalMonto = document.getElementById("total-monto");
        botonVaciarCarrito = document.getElementById("vaciar-carrito");
        botonPagarCarrito = document.getElementById("pagar-carrito");

        console.log("[Carrito] Elementos encontrados:", {
            cuerpoTablaCarrito: Boolean(cuerpoTablaCarrito),
            mensajeCarritoVacio: Boolean(mensajeCarritoVacio),
            contenedorAccionesCarrito: Boolean(contenedorAccionesCarrito),
            elementoTotalMonto: Boolean(elementoTotalMonto),
            botonVaciarCarrito: Boolean(botonVaciarCarrito),
            botonPagarCarrito: Boolean(botonPagarCarrito),
            botonesAgregarEnPagina: document.querySelectorAll(".agregar-carrito").length
        });

        // Si alguno de estos elementos no existe, evitamos que el script rompa la página
        if (!cuerpoTablaCarrito || !mensajeCarritoVacio || !contenedorAccionesCarrito || !elementoTotalMonto) {
            console.error("[Carrito] Faltan elementos obligatorios en el DOM. Revisa el HTML.");
            return;
        }

        // Pintamos el carrito con lo que ya hubiera guardado en la sesión
        renderizarCarrito();

        // Delegación de eventos para los botones "Agregar al Carrito"
        // (delegamos en document porque los productos vienen de un {{ range }} de Go)
        document.addEventListener("click", function (evento) {
            const botonAgregar = evento.target.closest(".agregar-carrito");
            if (botonAgregar) {
                evento.preventDefault();
                console.log("[Carrito] Click detectado en 'Agregar al Carrito':", botonAgregar.dataset);
                manejarClickAgregarProducto(botonAgregar);
            }
        });

        // Delegación de eventos dentro de la tabla del carrito
        // (cantidad y botón de eliminar se crean dinámicamente)
        cuerpoTablaCarrito.addEventListener("click", function (evento) {
            const botonEliminar = evento.target.closest(".borrar-producto");
            if (botonEliminar) {
                evento.preventDefault();
                const filaProducto = botonEliminar.closest("tr");
                const nombreProducto = filaProducto.dataset.nombre;
                eliminarProductoDelCarrito(nombreProducto);
            }
        });

        cuerpoTablaCarrito.addEventListener("change", function (evento) {
            if (evento.target.classList.contains("input-cantidad")) {
                const filaProducto = evento.target.closest("tr");
                const nombreProducto = filaProducto.dataset.nombre;
                actualizarCantidadProducto(nombreProducto, evento.target.value);
            }
        });

        // Vaciar carrito
        if (botonVaciarCarrito) {
            botonVaciarCarrito.addEventListener("click", function (evento) {
                evento.preventDefault();
                vaciarCarrito();
            });
        }

        // Pagar carrito
        if (botonPagarCarrito) {
            botonPagarCarrito.addEventListener("click", function (evento) {
                evento.preventDefault();
                procesarPagoCarrito();
            });
        }

        console.log("[Carrito] Listeners registrados correctamente. Carrito listo para usarse.");
    } catch (error) {
        console.error("[Carrito] Error crítico durante la inicialización:", error);
    }
}

/* ------------------------------------------------------------
   3. LECTURA / ESCRITURA EN sessionStorage
------------------------------------------------------------- */

/**
 * Obtiene el carrito guardado en sessionStorage.
 * Devuelve siempre un arreglo (vacío si no hay nada o hay un error).
 */
function obtenerCarritoDesdeStorage() {
    try {
        const carritoGuardado = sessionStorage.getItem(CLAVE_STORAGE_CARRITO);
        return carritoGuardado ? JSON.parse(carritoGuardado) : [];
    } catch (error) {
        console.error("Carrito: error al leer sessionStorage.", error);
        return [];
    }
}

/**
 * Guarda el arreglo del carrito en sessionStorage.
 */
function guardarCarritoEnStorage(carrito) {
    try {
        sessionStorage.setItem(CLAVE_STORAGE_CARRITO, JSON.stringify(carrito));
    } catch (error) {
        console.error("Carrito: error al guardar en sessionStorage.", error);
    }
}

/* ------------------------------------------------------------
   4. OPERACIONES SOBRE EL CARRITO
------------------------------------------------------------- */

/**
 * Lee los data-attributes del botón clickeado y agrega el producto.
 */
function manejarClickAgregarProducto(botonAgregar) {
    const nombreProducto = botonAgregar.dataset.nombre;
    const precioProducto = parseFloat(botonAgregar.dataset.precio);
    const imagenProducto = botonAgregar.dataset.imagen || "";

    if (!nombreProducto || isNaN(precioProducto)) {
        console.warn("Carrito: el producto no tiene data-nombre o data-precio válidos.", botonAgregar);
        return;
    }

    agregarProductoAlCarrito(nombreProducto, precioProducto, 1, imagenProducto);
}

/**
 * Agrega un producto al carrito. Si ya existe, incrementa la cantidad.
 */
function agregarProductoAlCarrito(nombreProducto, precioProducto, cantidadAgregada = 1, imagenProducto = "") {
    const carritoActual = obtenerCarritoDesdeStorage();
    const productoExistente = carritoActual.find(function (producto) {
        return producto.nombre === nombreProducto;
    });

    if (productoExistente) {
        productoExistente.cantidad += cantidadAgregada;
    } else {
        carritoActual.push({
            nombre: nombreProducto,
            precio: precioProducto,
            cantidad: cantidadAgregada,
            imagen: imagenProducto
        });
    }

    guardarCarritoEnStorage(carritoActual);
    renderizarCarrito();
}

/**
 * Elimina por completo un producto del carrito, sin importar su cantidad.
 */
function eliminarProductoDelCarrito(nombreProducto) {
    const carritoActual = obtenerCarritoDesdeStorage();
    const carritoFiltrado = carritoActual.filter(function (producto) {
        return producto.nombre !== nombreProducto;
    });

    guardarCarritoEnStorage(carritoFiltrado);
    renderizarCarrito();
}

/**
 * Actualiza la cantidad de un producto específico.
 * Si el valor ingresado no es válido (vacío, negativo, cero, no numérico),
 * se corrige automáticamente a 1.
 */
function actualizarCantidadProducto(nombreProducto, nuevaCantidadTexto) {
    const carritoActual = obtenerCarritoDesdeStorage();
    const productoAActualizar = carritoActual.find(function (producto) {
        return producto.nombre === nombreProducto;
    });

    if (!productoAActualizar) {
        return;
    }

    let nuevaCantidadNumerica = parseInt(nuevaCantidadTexto, 10);

    if (isNaN(nuevaCantidadNumerica) || nuevaCantidadNumerica < 1) {
        nuevaCantidadNumerica = 1;
    }

    productoAActualizar.cantidad = nuevaCantidadNumerica;

    guardarCarritoEnStorage(carritoActual);
    renderizarCarrito();
}

/**
 * Vacía completamente el carrito.
 */
function vaciarCarrito() {
    const carritoActual = obtenerCarritoDesdeStorage();

    if (carritoActual.length === 0) {
        return;
    }

    const confirmoVaciado = confirm("¿Seguro que deseas vaciar el carrito?");
    if (!confirmoVaciado) {
        return;
    }

    guardarCarritoEnStorage([]);
    renderizarCarrito();
}

/**
 * Calcula el total a pagar sumando precio * cantidad de cada producto.
 */
function calcularTotalCarrito(carrito) {
    return carrito.reduce(function (totalAcumulado, producto) {
        return totalAcumulado + (producto.precio * producto.cantidad);
    }, 0);
}

/**
 * Simula el proceso de pago. Aquí se debería integrar la pasarela
 * de pago real (o una llamada a tu backend en Go) más adelante.
 */
function procesarPagoCarrito() {
    const carritoActual = obtenerCarritoDesdeStorage();

    if (carritoActual.length === 0) {
        alert("Tu carrito está vacío. Agrega productos antes de pagar.");
        return;
    }

    const totalAPagar = calcularTotalCarrito(carritoActual);
    const confirmoPago = confirm(
        "Total a pagar: $" + totalAPagar.toFixed(2) + "\n¿Deseas continuar con el pago?"
    );

    if (!confirmoPago) {
        return;
    }

    // TODO: aquí se debería redirigir a la pasarela de pago real
    // o hacer un fetch/POST al backend con el detalle del carrito.
    alert("¡Gracias por tu compra! Pago procesado correctamente.");

    guardarCarritoEnStorage([]);
    renderizarCarrito();
}

/* ------------------------------------------------------------
   5. RENDERIZADO
------------------------------------------------------------- */

/**
 * Vuelve a dibujar toda la tabla del carrito, el total
 * y muestra/oculta el mensaje de "carrito vacío".
 */
function renderizarCarrito() {
    const carritoActual = obtenerCarritoDesdeStorage();

    cuerpoTablaCarrito.innerHTML = "";

    if (carritoActual.length === 0) {
        mensajeCarritoVacio.style.display = "block";
        contenedorAccionesCarrito.style.display = "none";
        elementoTotalMonto.textContent = "$0.00";
        return;
    }

    mensajeCarritoVacio.style.display = "none";
    contenedorAccionesCarrito.style.display = "block";

    carritoActual.forEach(function (producto) {
        cuerpoTablaCarrito.appendChild(crearFilaProducto(producto));
    });

    const totalCarrito = calcularTotalCarrito(carritoActual);
    elementoTotalMonto.textContent = "$" + totalCarrito.toFixed(2);
}

/**
 * Crea el elemento <tr> para un producto del carrito,
 * respetando las columnas definidas en el <thead> del HTML:
 * Imagen | Nombre | Precio | Cantidad | (acciones)
 */
function crearFilaProducto(producto) {
    const filaProducto = document.createElement("tr");
    filaProducto.dataset.nombre = producto.nombre;

    // Columna Imagen
    const celdaImagen = document.createElement("td");
    const imagenProducto = document.createElement("img");
    imagenProducto.src = producto.imagen || "";
    imagenProducto.alt = producto.nombre;
    celdaImagen.appendChild(imagenProducto);

    // Columna Nombre
    const celdaNombre = document.createElement("td");
    celdaNombre.textContent = producto.nombre;

    // Columna Precio (precio unitario)
    const celdaPrecio = document.createElement("td");
    celdaPrecio.textContent = "$" + producto.precio.toFixed(2);

    // Columna Cantidad (input editable)
    const celdaCantidad = document.createElement("td");
    const inputCantidad = document.createElement("input");
    inputCantidad.type = "number";
    inputCantidad.min = "1";
    inputCantidad.value = producto.cantidad;
    inputCantidad.classList.add("input-cantidad");
    celdaCantidad.appendChild(inputCantidad);

    // Columna Acciones (eliminar producto individual)
    const celdaAcciones = document.createElement("td");
    const enlaceEliminar = document.createElement("a");
    enlaceEliminar.href = "#";
    enlaceEliminar.classList.add("borrar-producto");
    enlaceEliminar.title = "Eliminar producto";
    enlaceEliminar.innerHTML = '<i class="fa-solid fa-trash"></i>';
    celdaAcciones.appendChild(enlaceEliminar);

    filaProducto.appendChild(celdaImagen);
    filaProducto.appendChild(celdaNombre);
    filaProducto.appendChild(celdaPrecio);
    filaProducto.appendChild(celdaCantidad);
    filaProducto.appendChild(celdaAcciones);

    return filaProducto;
}