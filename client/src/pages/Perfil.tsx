import { useState } from "react";
import { useAuthStore } from "../stores/authStore";

export default function Perfil() {
    const { user, role } = useAuthStore();
    const [isEditing, setIsEditing] = useState(false);
    const [formData, setFormData] = useState({
        name: user?.name || "",
        email: user?.email || "",
    });

    function handleSave() {
        // TODO: Implement profile update via API
        console.log("Saving profile:", formData);
        setIsEditing(false);
    }

    function getRoleLabel() {
        if (role === "secretary") {
            return "Secretario/a";
        }
        return "Docente";
    }

    function getRoleBadgeClass() {
        if (role === "secretary") {
            return "bg-purple-100 text-purple-700";
        }
        return "bg-green-100 text-green-700";
    }

    return (
        <div className="max-w-2xl mx-auto">
            <h1 className="text-3xl font-bold mb-6">Mi Perfil</h1>

            <div className="bg-white rounded-lg shadow-md p-6">
                {/* Avatar and basic info */}
                <div className="flex items-center gap-4 mb-6 pb-6 border-b">
                    <div className="w-20 h-20 bg-blue-500 rounded-full flex items-center justify-center text-white text-3xl font-bold">
                        {user?.name?.charAt(0).toUpperCase() || "U"}
                    </div>
                    <div>
                        <h2 className="text-xl font-semibold">
                            {user?.name || "Usuario"}
                        </h2>
                        <p className="text-gray-500">{user?.email || ""}</p>
                        <span
                            className={`inline-block mt-2 px-3 py-1 text-sm font-medium rounded-full ${getRoleBadgeClass()}`}
                        >
                            {getRoleLabel()}
                        </span>
                    </div>
                </div>

                {/* Profile details */}
                {isEditing ? (
                    <div className="space-y-4">
                        <div>
                            <label
                                htmlFor="name"
                                className="block text-sm font-medium text-gray-700 mb-1"
                            >
                                Nombre completo
                            </label>
                            <input
                                id="name"
                                type="text"
                                value={formData.name}
                                onChange={(e) =>
                                    setFormData({
                                        ...formData,
                                        name: e.target.value,
                                    })
                                }
                                className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                            />
                        </div>

                        <div>
                            <label
                                htmlFor="email"
                                className="block text-sm font-medium text-gray-700 mb-1"
                            >
                                Correo electrónico
                            </label>
                            <input
                                id="email"
                                type="email"
                                value={formData.email}
                                onChange={(e) =>
                                    setFormData({
                                        ...formData,
                                        email: e.target.value,
                                    })
                                }
                                className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                            />
                        </div>

                        <div className="flex gap-3 pt-4">
                            <button
                                onClick={handleSave}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
                            >
                                Guardar cambios
                            </button>
                            <button
                                onClick={() => {
                                    setIsEditing(false);
                                    setFormData({
                                        name: user?.name || "",
                                        email: user?.email || "",
                                    });
                                }}
                                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition"
                            >
                                Cancelar
                            </button>
                        </div>
                    </div>
                ) : (
                    <div className="space-y-4">
                        <div>
                            <h3 className="text-sm font-medium text-gray-500">
                                Nombre completo
                            </h3>
                            <p className="text-lg">{user?.name || "—"}</p>
                        </div>

                        <div>
                            <h3 className="text-sm font-medium text-gray-500">
                                Correo electrónico
                            </h3>
                            <p className="text-lg">{user?.email || "—"}</p>
                        </div>

                        <div>
                            <h3 className="text-sm font-medium text-gray-500">
                                Rol
                            </h3>
                            <p className="text-lg">{getRoleLabel()}</p>
                        </div>

                        <div>
                            <h3 className="text-sm font-medium text-gray-500">
                                ID de usuario
                            </h3>
                            <p className="text-lg">{user?.id || "—"}</p>
                        </div>

                        <div className="pt-4">
                            <button
                                onClick={() => setIsEditing(true)}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
                            >
                                Editar perfil
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Security section */}
            <div className="bg-white rounded-lg shadow-md p-6 mt-6">
                <h2 className="text-lg font-semibold mb-4">Seguridad</h2>

                <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                        <div>
                            <h3 className="font-medium">Contraseña</h3>
                            <p className="text-sm text-gray-500">
                                Última actualización: hace más de 30 días
                            </p>
                        </div>
                        <button className="px-4 py-2 text-blue-600 border border-blue-600 rounded-lg hover:bg-blue-50 transition">
                            Cambiar contraseña
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
