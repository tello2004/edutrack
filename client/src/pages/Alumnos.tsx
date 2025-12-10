import { useEffect, useState } from "react";
import { getAlumnos, type Alumno } from "../services/alumnosService";

export default function Alumnos() {
  const [alumnos, setAlumnos] = useState<Alumno[]>([]);
  const [loading, setLoading] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingAlumno, setEditingAlumno] = useState<Alumno | null>(null);

  const [formData, setFormData] = useState({
    nombre: "",
    matricula: "",
    grupo: "",
    email: "",
  });

  useEffect(() => {
    async function fetchAlumnos() {
      try {
        const data = await getAlumnos();
        setAlumnos(data);
      } catch (error) {
        console.error("Error al cargar alumnos:", error);
      } finally {
        setLoading(false);
      }
    }

    fetchAlumnos();
  }, []);

  function handleAgregar() {
    setFormData({ nombre: "", matricula: "", grupo: "", email: "" });
    setEditingAlumno(null);
    setModalOpen(true);
  }

  function handleEditar(alumno: Alumno) {
    setFormData({
      nombre: alumno.nombre,
      matricula: alumno.matricula,
      grupo: alumno.grupo,
      email: alumno.email || "",
    });
    setEditingAlumno(alumno);
    setModalOpen(true);
  }

  function handleEliminar(id: string) {
    if (confirm("¿Estás seguro de eliminar este alumno?")) {
      setAlumnos(alumnos.filter(a => a.id !== id));
    }
  }

  function handleGuardar() {
    if (editingAlumno) {
      setAlumnos(alumnos.map(a => a.id === editingAlumno.id ? { ...a, ...formData } : a));
    } else {
      const newAlumno: Alumno = {
        id: Date.now().toString(),
        ...formData,
      };
      setAlumnos([...alumnos, newAlumno]);
    }
    setModalOpen(false);
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Gestión de Alumnos</h1>
        <button
          onClick={handleAgregar}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded"
        >
          + Agregar Alumno
        </button>
      </div>

      {loading ? (
        <p>Cargando alumnos...</p>
      ) : alumnos.length === 0 ? (
        <p>No hay alumnos registrados.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
            <thead className="bg-blue-500 text-white">
              <tr>
                <th className="py-3 px-4 text-left">Matrícula</th>
                <th className="py-3 px-4 text-left">Nombre</th>
                <th className="py-3 px-4 text-left">Grupo</th>
                <th className="py-3 px-4 text-left">Email</th>
                <th className="py-3 px-4 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {alumnos.map(alumno => (
                <tr key={alumno.id} className="border-b hover:bg-gray-100">
                  <td className="py-2 px-4">{alumno.matricula}</td>
                  <td className="py-2 px-4">{alumno.nombre}</td>
                  <td className="py-2 px-4">{alumno.grupo}</td>
                  <td className="py-2 px-4">{alumno.email || "—"}</td>
                  <td className="py-2 px-4 space-x-2">
                    <button
                      onClick={() => handleEditar(alumno)}
                      className="text-blue-500 hover:underline"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleEliminar(alumno.id)}
                      className="text-red-500 hover:underline"
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

      {modalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">
              {editingAlumno ? "Editar Alumno" : "Agregar Alumno"}
            </h2>

            <div className="flex flex-col gap-3">
              <input
                type="text"
                placeholder="Nombre"
                value={formData.nombre}
                onChange={e => setFormData({ ...formData, nombre: e.target.value })}
                className="border p-2 rounded"
              />
              <input
                type="text"
                placeholder="Matrícula"
                value={formData.matricula}
                onChange={e => setFormData({ ...formData, matricula: e.target.value })}
                className="border p-2 rounded"
              />
              <input
                type="text"
                placeholder="Grupo"
                value={formData.grupo}
                onChange={e => setFormData({ ...formData, grupo: e.target.value })}
                className="border p-2 rounded"
              />
              <input
                type="email"
                placeholder="Email"
                value={formData.email}
                onChange={e => setFormData({ ...formData, email: e.target.value })}
                className="border p-2 rounded"
              />
            </div>

            <div className="flex justify-end mt-4 gap-2">
              <button
                onClick={() => setModalOpen(false)}
                className="px-4 py-2 rounded border"
              >
                Cancelar
              </button>
              <button
                onClick={handleGuardar}
                className="px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"
              >
                Guardar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
