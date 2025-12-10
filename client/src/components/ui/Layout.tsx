import { Outlet, Link } from "react-router-dom";


export default function Layout() {
return (
<div className="flex h-screen bg-gray-100">
{/* Sidebar */}
<aside className="w-64 bg-white p-4 shadow-xl">
<h2 className="text-2xl font-bold mb-6">EduTrack</h2>
<nav className="flex flex-col gap-2">
<Link to="/" className="hover:text-blue-500">Dashboard</Link>
<Link to="/alumnos" className="hover:text-blue-500">Alumnos</Link>
<Link to="/grupos" className="hover:text-blue-500">Grupos</Link>
<Link to="/asistencias" className="hover:text-blue-500">Asistencias</Link>
<Link to="/evaluaciones" className="hover:text-blue-500">Evaluaciones</Link>
<Link to="/calificaciones" className="hover:text-blue-500">Calificaciones</Link>
<Link to="/reportes" className="hover:text-blue-500">Reportes</Link>
<Link to="/perfil" className="hover:text-blue-500">Perfil</Link>
</nav>
</aside>


{/* Contenido */}
<main className="flex-1 p-6 overflow-auto">
<Outlet />
</main>
</div>
);
}