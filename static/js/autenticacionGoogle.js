// Importación desde el CDN oficial de Firebase para navegadores
import { initializeApp } from "https://www.gstatic.com/firebasejs/10.12.0/firebase-app.js";
import { getAnalytics } from "https://www.gstatic.com/firebasejs/10.12.0/firebase-analytics.js";
import { getAuth, signInWithPopup, GoogleAuthProvider } from "https://www.gstatic.com/firebasejs/10.12.0/firebase-auth.js";

const provider = new GoogleAuthProvider();

const firebaseConfig = {
    apiKey: "AIzaSyA6aQqm3ecxAEas8MOyIqulDU6oEcGo_uY",    
    authDomain: "pasantiaproyectogolang.firebaseapp.com",
    projectId: "pasantiaproyectogolang",
    storageBucket: "pasantiaproyectogolang.firebasestorage.app",
    messagingSenderId: "1075012396074",
    appId: "1:1075012396074:web:d460b826adf0709a564f53",
    measurementId: "G-0TDL83MJ74"
};

const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);
const auth = getAuth();

function call_login_google(e) {
    if (e) e.preventDefault();
    
    signInWithPopup(auth, provider)    
    .then((result) => {
        const credential = GoogleAuthProvider.credentialFromResult(result);
        const token = credential.accessToken;
        const user = result.user;
        
        alert(`¡Bienvenido ${user.displayName}!`);
    }).catch((error) => {
        console.error("Error en autenticación:", error);
        alert(`Error: ${error.message}`);
    });
}

// Vinculación de eventos al cargar el DOM
document.addEventListener("DOMContentLoaded", () => {
    const btnSignUp = document.getElementById("google-signup");
    const btnLogin = document.getElementById("google-login");

    if (btnSignUp) btnSignUp.addEventListener("click", call_login_google);
    if (btnLogin) btnLogin.addEventListener("click", call_login_google);
});