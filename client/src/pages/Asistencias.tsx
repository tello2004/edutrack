import { useEffect, useState } from "react";
import { getAsistencias, type Asistencia } from "../services/asistenciasService";

export default function Asistencias() {
  const [asistencias, setAsistencias] = useState<Asistencia[]>([]);
  const [loading, setLoading] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingAsistencia, setEditingAsistencia] = useState<Asistencia | null>(null);
  const [formData, setFormData] = useState({
    alumno: "",
    matricula: "",
    grupo: "",
    fecha: "",
    presente: true,
  });

  useEffect(() => {
    async function fetchAsistencias() {
      try {
        const data = await getAsistencias();
        setAsistencias(data);
      } catch (error) {
        console.error("Error al cargar asistencias:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchAsistencias();
  }, []);

  const handleAgregar = () => {
    setFormData({ alumno: "", matricula: "", grupo: "", fecha: "", presente: true });
    setEditingAsistencia(null);
    setModalOpen(true);
  };

  const handleEditar = (asistencia: Asistencia) => {
    setFormData({
      alumno: asistencia.alumno,
      matricula: asistencia.matricula,
      grupo: asistencia.grupo,
      fecha: asistencia.fecha,
      presente: asistencia.presente,
    });
    setEditingAsistencia(asistencia);
    setModalOpen(true);
  };

  const handleEliminar = (id: string) => {
    if (confirm("¿Deseas eliminar esta asistencia?")) {
      setAsistencias(asistencias.filter(a => a.id !== id));
    }
  };

  const handleGuardar = () => {
    if (editingAsistencia) {
      setAsistencias(asistencias.map(a => a.id === editingAsistencia.id ? { ...a, ...formData } : a));
    } else {
      const newAsistencia: Asistencia = { id: Date.now().toString(), ...formData };
      setAsistencias([...asistencias, newAsistencia]);
    }
    setModalOpen(false);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Registro de Asistencias</h1>
        <button
          onClick={handleAgregar}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded"
        >
          + Registrar Asistencia
        </button>
      </div>

      {loading ? (
        <p>Cargando asistencias...</p>
      ) : asistencias.length === 0 ? (
        <p>No hay registros de asistencias.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
            <thead className="bg-blue-500 text-white">
              <tr>
                <th className="py-3 px-4 text-left">Matrícula</th>
                <th className="py-3 px-4 text-left">Alumno</th>
                <th className="py-3 px-4 text-left">Grupo</th>
                <th className="py-3 px-4 text-left">Fecha</th>
                <th className="py-3 px-4 text-left">Presente</th>
                <th className="py-3 px-4 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {asistencias.map(a => (
                <tr key={a.id} className="border-b hover:bg-gray-100">
                  <td className="py-2 px-4">{a.matricula}</td>
                  <td className="py-2 px-4">{a.alumno}</td>
                  <td className="py-2 px-4">{a.grupo}</td>
                  <td className="py-2 px-4">{a.fecha}</td>
                  <td className="py-2 px-4">{a.presente ? "Sí" : "No"}</td>
                  <td className="py-2 px-4 space-x-2">
                    <button
                      onClick={() => handleEditar(a)}
                      className="text-blue-500 hover:underline"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleEliminar(a.id)}
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
              {editingAsistencia ? "Editar Asistencia" : "Registrar Asistencia"}
            </h2>

            <div className="flex flex-col gap-3">
              <input
                type="text"
                placeholder="Alumno"
                value={formData.alumno}
                onChange={e => setFormData({ ...formData, alumno: e.target.value })}
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
                type="date"
                placeholder="Fecha"
                value={formData.fecha}
                onChange={e => setFormData({ ...formData, fecha: e.target.value })}
                className="border p-2 rounded"
              />
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formData.presente}
                  onChange={e => setFormData({ ...formData, presente: e.target.checked })}
                  className="w-4 h-4"
                />
                Presente
              </label>
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
