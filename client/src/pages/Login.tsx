export default function Login() {
return (
<div className="flex justify-center items-center h-screen bg-gray-200">
<div className="bg-white p-10 rounded-lg shadow-xl w-96">
<h1 className="text-2xl font-bold mb-6">Iniciar Sesión</h1>
<input className="w-full p-2 border mb-4" placeholder="Email" />
<input className="w-full p-2 border mb-4" placeholder="Contraseña" type="password" />
<button className="w-full bg-blue-600 text-white py-2 rounded">Entrar</button>
</div>
</div>
);
}