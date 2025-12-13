import { useEffect, useState } from "react";
import {
    getGrupos,
    addGrupo,
    updateGrupo,
    deleteGrupo,
    type Grupo,
} from "../services/gruposService";

export default function Grupos() {
    const [grupos, setGrupos] = useState<Grupo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const [modalOpen, setModalOpen] = useState(false);
    const [editingGrupo, setEditingGrupo] = useState<Grupo | null>(null);
    const [formData, setFormData] = useState({
        nombre: "",
        descripcion: "",
        codigo: "",
        duracion: 8,
    });
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        fetchGrupos();
    }, []);

    async function fetchGrupos() {
        try {
            setLoading(true);
            setError(null);
            const data = await getGrupos();
            setGrupos(data);
        } catch (err) {
            const message =
                err instanceof Error ? err.message : "Error al cargar grupos";
            setError(message);
            console.error("Error al cargar grupos:", err);
        } finally {
            setLoading(false);
        }
    }

    function handleAgregar() {
        setFormData({ nombre: "", descripcion: "", codigo: "", duracion: 8 });
        setEditingGrupo(null);
        setModalOpen(true);
    }

    function handleEditar(grupo: Grupo) {
        setFormData({
            nombre: grupo.nombre,
            descripcion: grupo.descripcion || "",
            codigo: grupo.codigo || "",
            duracion: grupo.duracion || 8,
        });
        setEditingGrupo(grupo);
        setModalOpen(true);
    }

    async function handleEliminar(id: string) {
        if (!confirm("¿Estás seguro de eliminar este grupo?")) return;

        try {
            await deleteGrupo(id);
            setGrupos((prev) => prev.filter((g) => g.id !== id));
        } catch (err) {
            const message =
                err instanceof Error ? err.message : "Error al eliminar grupo";
            alert(message);
        }
    }

    async function handleGuardar() {
        if (!formData.nombre.trim()) {
            alert("El nombre del grupo es requerido.");
            return;
        }

        setSaving(true);

        try {
            if (editingGrupo) {
                const actualizado = await updateGrupo({
                    id: editingGrupo.id,
                    nombre: formData.nombre,
                    descripcion: formData.descripcion,
                    codigo: formData.codigo,
                    duracion: formData.duracion,
                });
                setGrupos((prev) =>
                    prev.map((g) =>
                        g.id === editingGrupo.id ? actualizado : g,
                    ),
                );
            } else {
                const nuevo = await addGrupo({
                    nombre: formData.nombre,
                    descripcion: formData.descripcion,
                    codigo: formData.codigo,
                    duracion: formData.duracion,
                });
                setGrupos((prev) => [...prev, nuevo]);
            }
            setModalOpen(false);
        } catch (err) {
            const message =
                err instanceof Error ? err.message : "Error al guardar grupo";
            alert(message);
        } finally {
            setSaving(false);
        }
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto"></div>
                    <p className="mt-4 text-gray-600">Cargando grupos...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                <h2 className="text-lg font-semibold text-red-700 mb-2">
                    Error al cargar datos
                </h2>
                <p className="text-red-600 mb-4">{error}</p>
                <button
                    onClick={fetchGrupos}
                    className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition"
                >
                    Reintentar
                </button>
            </div>
        );
    }

    return (
        <div>
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold">
                    Gestión de Grupos / Carreras
                </h1>
                <button
                    className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded transition"
                    onClick={handleAgregar}
                >
                    + Agregar Grupo
                </button>
            </div>

            {grupos.length === 0 ? (
                <div className="text-center py-12 bg-gray-50 rounded-lg">
                    <svg
                        className="mx-auto h-12 w-12 text-gray-400"
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
                    <p className="mt-2 text-gray-500">
                        No hay grupos registrados.
                    </p>
                    <button
                        onClick={handleAgregar}
                        className="mt-4 text-blue-600 hover:text-blue-700 font-medium"
                    >
                        Agregar el primer grupo
                    </button>
                </div>
            ) : (
                <div className="overflow-x-auto">
                    <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
                        <thead className="bg-blue-500 text-white">
                            <tr>
                                <th className="py-3 px-4 text-left">Código</th>
                                <th className="py-3 px-4 text-left">Nombre</th>
                                <th className="py-3 px-4 text-left">
                                    Descripción
                                </th>
                                <th className="py-3 px-4 text-left">
                                    Duración (semestres)
                                </th>
                                <th className="py-3 px-4 text-left">Estado</th>
                                <th className="py-3 px-4 text-left">
                                    Acciones
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {grupos.map((grupo) => (
                                <tr
                                    key={grupo.id}
                                    className="border-b hover:bg-gray-50 transition-colors"
                                >
                                    <td className="py-3 px-4">
                                        <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
                                            {grupo.codigo || "—"}
                                        </span>
                                    </td>
                                    <td className="py-3 px-4 font-medium">
                                        {grupo.nombre}
                                    </td>
                                    <td className="py-3 px-4 text-gray-600">
                                        {grupo.descripcion || "—"}
                                    </td>
                                    <td className="py-3 px-4">
                                        {grupo.duracion || "—"}
                                    </td>
                                    <td className="py-3 px-4">
                                        {grupo.activo !== false ? (
                                            <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">
                                                Activo
                                            </span>
                                        ) : (
                                            <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
                                                Inactivo
                                            </span>
                                        )}
                                    </td>
                                    <td className="py-3 px-4 space-x-2">
                                        <button
                                            className="text-blue-500 hover:text-blue-700 font-medium"
                                            onClick={() => handleEditar(grupo)}
                                        >
                                            Editar
                                        </button>
                                        <button
                                            className="text-red-500 hover:text-red-700 font-medium"
                                            onClick={() =>
                                                handleEliminar(grupo.id)
                                            }
                                        >
                                            Eliminar
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Modal */}
            {modalOpen && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h2 className="text-xl font-bold mb-4">
                            {editingGrupo ? "Editar Grupo" : "Agregar Grupo"}
                        </h2>

                        <div className="flex flex-col gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Nombre *
                                </label>
                                <input
                                    type="text"
                                    placeholder="Ej: Ingeniería en Sistemas"
                                    value={formData.nombre}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            nombre: e.target.value,
                                        })
                                    }
                                    className="w-full border border-gray-300 p-2 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    disabled={saving}
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Código
                                </label>
                                <input
                                    type="text"
                                    placeholder="Ej: ISC-2024"
                                    value={formData.codigo}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            codigo: e.target.value,
                                        })
                                    }
                                    className="w-full border border-gray-300 p-2 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    disabled={saving}
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Descripción
                                </label>
                                <textarea
                                    placeholder="Descripción del grupo o carrera"
                                    value={formData.descripcion}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            descripcion: e.target.value,
                                        })
                                    }
                                    className="w-full border border-gray-300 p-2 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    rows={3}
                                    disabled={saving}
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Duración (semestres)
                                </label>
                                <input
                                    type="number"
                                    min="1"
                                    max="16"
                                    value={formData.duracion}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            duracion:
                                                parseInt(e.target.value) || 8,
                                        })
                                    }
                                    className="w-full border border-gray-300 p-2 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    disabled={saving}
                                />
                            </div>
                        </div>

                        <div className="flex justify-end mt-6 gap-3">
                            <button
                                onClick={() => setModalOpen(false)}
                                className="px-4 py-2 rounded border border-gray-300 hover:bg-gray-50 transition"
                                disabled={saving}
                            >
                                Cancelar
                            </button>
                            <button
                                onClick={handleGuardar}
                                className="px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                                disabled={saving}
                            >
                                {saving ? "Guardando..." : "Guardar"}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
