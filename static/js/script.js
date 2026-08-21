/* ============================================================
   INDEX.HTML - Carrusel del hero (sin Swiper, propio y liviano)
   No depende de ninguna librería externa. El botón "Agregar al
   Carrito" de cada slide usa la clase .agregar-carrito, que
   carrito.js ya escucha por delegación de eventos en document -
   no hace falta ningún cableado extra aquí para que funcione.
   ============================================================ */

console.log("[Inicio] script.js se cargó correctamente.");

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", inicializarHeroCarrusel);
} else {
    inicializarHeroCarrusel();
}

function inicializarHeroCarrusel() {
    const carrusel = document.getElementById("hero-carrusel");
    const track = document.getElementById("hero-track");
    const contenedorPuntos = document.getElementById("hero-puntos");
    const botonAnterior = document.getElementById("hero-anterior");
    const botonSiguiente = document.getElementById("hero-siguiente");

    if (!carrusel || !track) {
        return; // esta página no tiene el hero carrusel, no hay nada que hacer
    }

    const diapositivas = track.querySelectorAll(".hero-slide");
    if (diapositivas.length === 0) {
        return;
    }

    const DURACION_AUTOPLAY_MS = 5000;
    let indiceActual = 0;
    let temporizadorAutoplay = null;

    // Construye los puntos de navegación, uno por diapositiva
    diapositivas.forEach(function (_, indice) {
        const punto = document.createElement("button");
        punto.type = "button";
        punto.className = "hero-punto";
        punto.setAttribute("aria-label", "Ir al producto " + (indice + 1));
        punto.addEventListener("click", function () {
            irADiapositiva(indice);
            reiniciarAutoplay();
        });
        contenedorPuntos.appendChild(punto);
    });

    const puntos = contenedorPuntos.querySelectorAll(".hero-punto");

    function actualizarPuntos() {
        puntos.forEach(function (punto, indice) {
            punto.classList.toggle("hero-punto-activo", indice === indiceActual);
        });
    }

    function irADiapositiva(indice) {
        indiceActual = (indice + diapositivas.length) % diapositivas.length;
        track.style.transform = "translateX(-" + (indiceActual * 100) + "%)";
        actualizarPuntos();
    }

    function irASiguiente() {
        irADiapositiva(indiceActual + 1);
    }

    function irAAnterior() {
        irADiapositiva(indiceActual - 1);
    }

    function iniciarAutoplay() {
        temporizadorAutoplay = setInterval(irASiguiente, DURACION_AUTOPLAY_MS);
    }

    function detenerAutoplay() {
        if (temporizadorAutoplay) {
            clearInterval(temporizadorAutoplay);
            temporizadorAutoplay = null;
        }
    }

    function reiniciarAutoplay() {
        detenerAutoplay();
        iniciarAutoplay();
    }

    if (botonSiguiente) {
        botonSiguiente.addEventListener("click", function () {
            irASiguiente();
            reiniciarAutoplay();
        });
    }

    if (botonAnterior) {
        botonAnterior.addEventListener("click", function () {
            irAAnterior();
            reiniciarAutoplay();
        });
    }

    // El usuario puede mover el carrusel, y también se detiene solo
    // mientras el mouse está encima (para que no se mueva justo cuando
    // alguien está por hacer clic en "Agregar al Carrito").
    carrusel.addEventListener("mouseenter", detenerAutoplay);
    carrusel.addEventListener("mouseleave", iniciarAutoplay);

    actualizarPuntos();
    iniciarAutoplay();
}