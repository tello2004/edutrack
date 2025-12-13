import { useEffect, useState } from "react";
import {
    getAlumnos,
    addAlumno,
    updateAlumno,
    deleteAlumno,
    type Alumno,
} from "../services/alumnosService";
import { getGrupos, type Grupo } from "../services/gruposService";

export default function Alumnos() {
    const [alumnos, setAlumnos] = useState<Alumno[]>([]);
    const [grupos, setGrupos] = useState<Grupo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const [modalOpen, setModalOpen] = useState(false);
    const [editingAlumno, setEditingAlumno] = useState<Alumno | null>(null);
    const [formData, setFormData] = useState({
        nombre: "",
        matricula: "",
        grupo: "",
        email: "",
        careerId: undefined as number | undefined,
    });
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        fetchData();
    }, []);

    async function fetchData() {
        try {
            setLoading(true);
            setError(null);
            const [alumnosData, gruposData] = await Promise.all([
                getAlumnos(),
                getGrupos(),
            ]);
            setAlumnos(alumnosData);
            setGrupos(gruposData);
        } catch (err) {
            const message =
                err instanceof Error ? err.message : "Error al cargar datos";
            setError(message);
            console.error("Error al cargar datos:", err);
        } finally {
            setLoading(false);
        }
    }

    function handleAgregar() {
        setFormData({
            nombre: "",
            matricula: "",
            grupo: "",
            email: "",
            careerId: undefined,
        });
        setEditingAlumno(null);
        setModalOpen(true);
    }

    function handleEditar(alumno: Alumno) {
        setFormData({
            nombre: alumno.nombre,
            matricula: alumno.matricula,
            grupo: alumno.grupo,
            email: alumno.email || "",
            careerId: alumno.careerId,
        });
        setEditingAlumno(alumno);
        setModalOpen(true);
    }

    async function handleEliminar(id: string) {
        if (!confirm("¿Estás seguro de eliminar este alumno?")) return;

        try {
            await deleteAlumno(id);
            setAlumnos((prev) => prev.filter((a) => a.id !== id));
        } catch (err) {
            const message =
                err instanceof Error ? err.message : "Error al eliminar alumno";
            alert(message);
        }
    }

    async function handleGuardar() {
        if (!formData.nombre.trim() || !formData.matricula.trim()) {
            alert("El nombre y la matrícula son requeridos.");
            return;
        }

        setSaving(true);

        try {
            if (editingAlumno) {
                const actualizado = await updateAlumno({
                    id: editingAlumno.id,
                    nombre: formData.nombre,
                    matricula: formData.matricula,
                    grupo: formData.grupo,
                    email: formData.email,
                    careerId: formData.careerId,
                    accountId: editingAlumno.accountId,
                });
                setAlumnos((prev) =>
                    prev.map((a) =>
                        a.id === editingAlumno.id ? actualizado : a,
                    ),
                );
            } else {
                const nuevo = await addAlumno({
                    nombre: formData.nombre,
                    matricula: formData.matricula,
                    grupo: formData.grupo,
                    email: formData.email,
                    careerId: formData.careerId,
                });
                setAlumnos((prev) => [...prev, nuevo]);
            }
            setModalOpen(false);
        } catch (err) {
            const message =
                err instanceof Error ? err.message : "Error al guardar alumno";
            alert(message);
        } finally {
            setSaving(false);
        }
    }

    function handleGrupoChange(grupoId: string) {
        const grupo = grupos.find((g) => g.id === grupoId);
        setFormData({
            ...formData,
            grupo: grupo?.nombre || "",
            careerId: grupo ? parseInt(grupo.id) : undefined,
        });
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto"></div>
                    <p className="mt-4 text-gray-600">Cargando alumnos...</p>
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
                    onClick={fetchData}
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
                <h1 className="text-2xl font-bold">Gestión de Alumnos</h1>
                <button
                    onClick={handleAgregar}
                    className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded transition"
                >
                    + Agregar Alumno
                </button>
            </div>

            {alumnos.length === 0 ? (
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
                            d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                        />
                    </svg>
                    <p className="mt-2 text-gray-500">
                        No hay alumnos registrados.
                    </p>
                    <button
                        onClick={handleAgregar}
                        className="mt-4 text-blue-600 hover:text-blue-700 font-medium"
                    >
                        Agregar el primer alumno
                    </button>
                </div>
            ) : (
                <div className="overflow-x-auto">
                    <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
                        <thead className="bg-blue-500 text-white">
                            <tr>
                                <th className="py-3 px-4 text-left">
                                    Matrícula
                                </th>
                                <th className="py-3 px-4 text-left">Nombre</th>
                                <th className="py-3 px-4 text-left">
                                    Grupo / Carrera
                                </th>
                                <th className="py-3 px-4 text-left">Email</th>
                                <th className="py-3 px-4 text-left">
                                    Acciones
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {alumnos.map((alumno) => (
                                <tr
                                    key={alumno.id}
                                    className="border-b hover:bg-gray-50 transition-colors"
                                >
                                    <td className="py-3 px-4">
                                        <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
                                            {alumno.matricula}
                                        </span>
                                    </td>
                                    <td className="py-3 px-4 font-medium">
                                        {alumno.nombre}
                                    </td>
                                    <td className="py-3 px-4 text-gray-600">
                                        {alumno.grupo || "—"}
                                    </td>
                                    <td className="py-3 px-4 text-gray-600">
                                        {alumno.email || "—"}
                                    </td>
                                    <td className="py-3 px-4 space-x-2">
                                        <button
                                            onClick={() => handleEditar(alumno)}
                                            className="text-blue-500 hover:text-blue-700 font-medium"
                                        >
                                            Editar
                                        </button>
                                        <button
                                            onClick={() =>
                                                handleEliminar(alumno.id)
                                            }
                                            className="text-red-500 hover:text-red-700 font-medium"
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
                            {editingAlumno ? "Editar Alumno" : "Agregar Alumno"}
                        </h2>

                        <div className="flex flex-col gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Nombre completo *
                                </label>
                                <input
                                    type="text"
                                    placeholder="Ej: Juan Pérez García"
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
                                    Matrícula *
                                </label>
                                <input
                                    type="text"
                                    placeholder="Ej: A001"
                                    value={formData.matricula}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            matricula: e.target.value,
                                        })
                                    }
                                    className="w-full border border-gray-300 p-2 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    disabled={saving || !!editingAlumno}
                                />
                                {editingAlumno && (
                                    <p className="text-xs text-gray-500 mt-1">
                                        La matrícula no puede ser modificada.
                                    </p>
                                )}
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Grupo / Carrera
                                </label>
                                <select
                                    value={formData.careerId?.toString() || ""}
                                    onChange={(e) =>
                                        handleGrupoChange(e.target.value)
                                    }
                                    className="w-full border border-gray-300 p-2 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    disabled={saving}
                                >
                                    <option value="">Seleccionar grupo</option>
                                    {grupos.map((grupo) => (
                                        <option key={grupo.id} value={grupo.id}>
                                            {grupo.nombre}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Email
                                </label>
                                <input
                                    type="email"
                                    placeholder="correo@ejemplo.com"
                                    value={formData.email}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            email: e.target.value,
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
