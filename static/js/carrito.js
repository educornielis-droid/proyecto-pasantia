/* ============================================================
   SYPAGO STORE - LOGICA DEL CARRITO DE COMPRAS
   Archivo independiente: solo carrito (no toca Swiper).
   Persistencia: sessionStorage
   ============================================================ */

console.log("[Carrito] carrito.js se cargó correctamente.");

// Clave usada para guardar el carrito en sessionStorage
const CLAVE_STORAGE_CARRITO = "carritoSypago";

// Endpoint de nuestro propio backend que arma la transacción
// (el navegador NUNCA llama directo a la API de Sypago)
const ENDPOINT_INICIAR_CHECKOUT = "/api/checkout/iniciar";

/* ------------------------------------------------------------
   1. REFERENCIAS AL DOM
------------------------------------------------------------- */
let cuerpoTablaCarrito;
let mensajeCarritoVacio;
let contenedorAccionesCarrito;
let elementoTotalMonto;
let botonVaciarCarrito;
let botonPagarCarrito;

/* ------------------------------------------------------------
   2. INICIALIZACION
------------------------------------------------------------- */
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", inicializarCarrito);
} else {
    inicializarCarrito();
}

function inicializarCarrito() {
    try {
        cuerpoTablaCarrito = document.querySelector("#lista-carrito tbody");
        mensajeCarritoVacio = document.getElementById("carrito-vacio-msg");
        contenedorAccionesCarrito = document.getElementById("carrito-acciones");
        elementoTotalMonto = document.getElementById("total-monto");
        botonVaciarCarrito = document.getElementById("vaciar-carrito");
        botonPagarCarrito = document.getElementById("pagar-carrito");

        if (!cuerpoTablaCarrito || !mensajeCarritoVacio || !contenedorAccionesCarrito || !elementoTotalMonto) {
            console.error("[Carrito] Faltan elementos obligatorios en el DOM.");
            return;
        }

        renderizarCarrito();

        document.addEventListener("click", function (evento) {
            const botonAgregar = evento.target.closest(".agregar-carrito");
            if (botonAgregar) {
                evento.preventDefault();
                manejarClickAgregarProducto(botonAgregar);
            }
        });

        cuerpoTablaCarrito.addEventListener("click", function (evento) {
            const botonEliminar = evento.target.closest(".borrar-producto");
            if (botonEliminar) {
                evento.preventDefault();
                const filaProducto = botonEliminar.closest("tr");
                eliminarProductoDelCarrito(filaProducto.dataset.nombre);
            }
        });

        cuerpoTablaCarrito.addEventListener("change", function (evento) {
            if (evento.target.classList.contains("input-cantidad")) {
                const filaProducto = evento.target.closest("tr");
                actualizarCantidadProducto(filaProducto.dataset.nombre, evento.target.value);
            }
        });

        if (botonVaciarCarrito) {
            botonVaciarCarrito.addEventListener("click", function (evento) {
                evento.preventDefault();
                vaciarCarrito();
            });
        }

        // El botón "Pagar" ahora inicia la transacción en el backend
        // y redirige a checkout.html
        if (botonPagarCarrito) {
            botonPagarCarrito.addEventListener("click", function (evento) {
                evento.preventDefault();
                manejarClickPagar();
            });
        }

        console.log("[Carrito] Listeners registrados correctamente.");
    } catch (error) {
        console.error("[Carrito] Error crítico durante la inicialización:", error);
    }
}

/* ------------------------------------------------------------
   3. LECTURA / ESCRITURA EN sessionStorage
------------------------------------------------------------- */
function obtenerCarritoDesdeStorage() {
    try {
        const carritoGuardado = sessionStorage.getItem(CLAVE_STORAGE_CARRITO);
        return carritoGuardado ? JSON.parse(carritoGuardado) : [];
    } catch (error) {
        console.error("[Carrito] Error al leer sessionStorage.", error);
        return [];
    }
}

function guardarCarritoEnStorage(carrito) {
    try {
        sessionStorage.setItem(CLAVE_STORAGE_CARRITO, JSON.stringify(carrito));
    } catch (error) {
        console.error("[Carrito] Error al guardar en sessionStorage.", error);
    }
}

/* ------------------------------------------------------------
   4. OPERACIONES SOBRE EL CARRITO
------------------------------------------------------------- */
function manejarClickAgregarProducto(botonAgregar) {
    const nombreProducto = botonAgregar.dataset.nombre;
    const precioProducto = parseFloat(botonAgregar.dataset.precio);
    const imagenProducto = botonAgregar.dataset.imagen || "";

    if (!nombreProducto || isNaN(precioProducto)) {
        console.warn("[Carrito] Producto sin data-nombre o data-precio válidos.", botonAgregar);
        return;
    }

    agregarProductoAlCarrito(nombreProducto, precioProducto, 1, imagenProducto);
}

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

function eliminarProductoDelCarrito(nombreProducto) {
    const carritoActual = obtenerCarritoDesdeStorage();
    const carritoFiltrado = carritoActual.filter(function (producto) {
        return producto.nombre !== nombreProducto;
    });

    guardarCarritoEnStorage(carritoFiltrado);
    renderizarCarrito();
}

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

function vaciarCarrito() {
    const carritoActual = obtenerCarritoDesdeStorage();

    if (carritoActual.length === 0) {
        return;
    }

    if (!confirm("¿Seguro que deseas vaciar el carrito?")) {
        return;
    }

    guardarCarritoEnStorage([]);
    renderizarCarrito();
}

function calcularTotalCarrito(carrito) {
    return carrito.reduce(function (totalAcumulado, producto) {
        return totalAcumulado + (producto.precio * producto.cantidad);
    }, 0);
}

/* ------------------------------------------------------------
   5. RENDERIZADO DE LA TABLA DEL CARRITO
------------------------------------------------------------- */
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

function crearFilaProducto(producto) {
    const filaProducto = document.createElement("tr");
    filaProducto.dataset.nombre = producto.nombre;

    const celdaImagen = document.createElement("td");
    const imagenProducto = document.createElement("img");
    imagenProducto.src = producto.imagen || "";
    imagenProducto.alt = producto.nombre;
    celdaImagen.appendChild(imagenProducto);

    const celdaNombre = document.createElement("td");
    celdaNombre.textContent = producto.nombre;

    const celdaPrecio = document.createElement("td");
    celdaPrecio.textContent = "$" + producto.precio.toFixed(2);

    const celdaCantidad = document.createElement("td");
    const inputCantidad = document.createElement("input");
    inputCantidad.type = "number";
    inputCantidad.min = "1";
    inputCantidad.value = producto.cantidad;
    inputCantidad.classList.add("input-cantidad");
    celdaCantidad.appendChild(inputCantidad);

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

/* ------------------------------------------------------------
   6. INICIAR CHECKOUT EN EL BACKEND Y REDIRIGIR
------------------------------------------------------------- */
async function manejarClickPagar() {
    const carritoActual = obtenerCarritoDesdeStorage();

    if (carritoActual.length === 0) {
        alert("Tu carrito está vacío. Agrega productos antes de pagar.");
        return;
    }

    if (botonPagarCarrito) {
        botonPagarCarrito.style.pointerEvents = "none";
        botonPagarCarrito.style.opacity = "0.6";
        botonPagarCarrito.textContent = "Procesando...";
    }

    // OJO: solo mandamos nombre y cantidad. El precio, la imagen y el
    // total los recalcula el backend consultando la base de datos.
    const productosParaEnviar = carritoActual.map(function (producto) {
        return {
            nombre: producto.nombre,
            cantidad: producto.cantidad
        };
    });

    try {
        const respuesta = await fetch(ENDPOINT_INICIAR_CHECKOUT, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ productos: productosParaEnviar })
        });

        if (!respuesta.ok) {
            const detalleError = await respuesta.json().catch(function () { return {}; });
            throw new Error(detalleError.error || "No se pudo iniciar el checkout.");
        }

        const datosRespuesta = await respuesta.json();

        if (!datosRespuesta.redirect_url) {
            throw new Error("La respuesta del servidor no incluyó la URL de checkout.");
        }

        console.log("[Carrito] Redirigiendo a checkout:", datosRespuesta.redirect_url);
        window.location.href = datosRespuesta.redirect_url;
    } catch (error) {
        console.error("[Carrito] Error al iniciar el checkout:", error);
        alert("Hubo un problema al procesar tu pago: " + error.message);

        if (botonPagarCarrito) {
            botonPagarCarrito.style.pointerEvents = "auto";
            botonPagarCarrito.style.opacity = "1";
            botonPagarCarrito.textContent = "Pagar";
        }
    }
}