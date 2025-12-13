import { Outlet, Link, useNavigate, useLocation } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import { logout } from "../../api/auth";

export default function Layout() {
    const navigate = useNavigate();
    const location = useLocation();
    const { user, role } = useAuthStore();

    function handleLogout() {
        logout();
        navigate("/login");
    }

    function isActive(path: string) {
        return location.pathname === path;
    }

    function getLinkClass(path: string) {
        const baseClass =
            "flex items-center gap-2 px-3 py-2 rounded-lg transition-colors";
        if (isActive(path)) {
            return `${baseClass} bg-blue-100 text-blue-700 font-medium`;
        }
        return `${baseClass} hover:bg-gray-100 text-gray-700`;
    }

    function getRoleBadge() {
        if (role === "secretary") {
            return (
                <span className="px-2 py-1 text-xs font-medium bg-purple-100 text-purple-700 rounded-full">
                    Secretario
                </span>
            );
        }
        return (
            <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">
                Docente
            </span>
        );
    }

    return (
        <div className="flex h-screen bg-gray-100">
            {/* Sidebar */}
            <aside className="w-64 bg-white shadow-xl flex flex-col">
                {/* Logo */}
                <div className="p-4 border-b">
                    <h2 className="text-2xl font-bold text-blue-600">
                        EduTrack
                    </h2>
                    <p className="text-xs text-gray-500 mt-1">
                        Sistema de Gestión Educativa
                    </p>
                </div>

                {/* Navigation */}
                <nav className="flex-1 p-4 flex flex-col gap-1 overflow-y-auto">
                    <Link to="/" className={getLinkClass("/")}>
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
                            />
                        </svg>
                        Dashboard
                    </Link>

                    <Link to="/alumnos" className={getLinkClass("/alumnos")}>
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                            />
                        </svg>
                        Alumnos
                    </Link>

                    <Link to="/grupos" className={getLinkClass("/grupos")}>
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                            />
                        </svg>
                        Grupos
                    </Link>

                    <Link
                        to="/asistencias"
                        className={getLinkClass("/asistencias")}
                    >
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
                            />
                        </svg>
                        Asistencias
                    </Link>

                    <Link
                        to="/evaluaciones"
                        className={getLinkClass("/evaluaciones")}
                    >
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                            />
                        </svg>
                        Evaluaciones
                    </Link>

                    <Link
                        to="/calificaciones"
                        className={getLinkClass("/calificaciones")}
                    >
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"
                            />
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z"
                            />
                        </svg>
                        Calificaciones
                    </Link>

                    <Link to="/reportes" className={getLinkClass("/reportes")}>
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                            />
                        </svg>
                        Reportes
                    </Link>

                    <div className="border-t my-2"></div>

                    <Link to="/perfil" className={getLinkClass("/perfil")}>
                        <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                            />
                        </svg>
                        Mi Perfil
                    </Link>
                </nav>

                {/* User info and logout */}
                <div className="p-4 border-t bg-gray-50">
                    <div className="flex items-center gap-3 mb-3">
                        <div className="w-10 h-10 bg-blue-500 rounded-full flex items-center justify-center text-white font-bold">
                            {user?.name?.charAt(0).toUpperCase() || "U"}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">
                                {user?.name || "Usuario"}
                            </p>
                            <p className="text-xs text-gray-500 truncate">
                                {user?.email || ""}
                            </p>
                        </div>
                    </div>
                    <div className="flex items-center justify-between">
                        {getRoleBadge()}
                        <button
                            onClick={handleLogout}
                            className="flex items-center gap-1 text-sm text-red-600 hover:text-red-700 font-medium"
                        >
                            <svg
                                className="w-4 h-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
                                />
                            </svg>
                            Salir
                        </button>
                    </div>
                </div>
            </aside>

            {/* Main content */}
            <main className="flex-1 p-6 overflow-auto">
                <Outlet />
            </main>
        </div>
    );
}
