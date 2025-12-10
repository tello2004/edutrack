import { useEffect, useState } from "react";
import { getCalificaciones, type Calificacion } from "../services/calificacionesService";

export default function Calificaciones() {
  const [calificaciones, setCalificaciones] = useState<Calificacion[]>([]);
  const [loading, setLoading] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingCal, setEditingCal] = useState<Calificacion | null>(null);
  const [formData, setFormData] = useState({
    alumno: "",
    matricula: "",
    evaluacion: "",
    calificacion: 0,
  });

  useEffect(() => {
    async function fetchCalificaciones() {
      try {
        const data = await getCalificaciones();
        setCalificaciones(data);
      } catch (error) {
        console.error("Error al cargar calificaciones:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchCalificaciones();
  }, []);

  const handleAgregar = () => {
    setFormData({ alumno: "", matricula: "", evaluacion: "", calificacion: 0 });
    setEditingCal(null);
    setModalOpen(true);
  };

  const handleEditar = (cal: Calificacion) => {
    setFormData({
      alumno: cal.alumno,
      matricula: cal.matricula,
      evaluacion: cal.evaluacion,
      calificacion: cal.calificacion,
    });
    setEditingCal(cal);
    setModalOpen(true);
  };

  const handleEliminar = (id: string) => {
    if (confirm("¿Deseas eliminar esta calificación?")) {
      setCalificaciones(calificaciones.filter(c => c.id !== id));
    }
  };

  const handleGuardar = () => {
    if (editingCal) {
      setCalificaciones(
        calificaciones.map(c => (c.id === editingCal.id ? { ...c, ...formData } : c))
      );
    } else {
      const newCal: Calificacion = { id: Date.now().toString(), ...formData };
      setCalificaciones([...calificaciones, newCal]);
    }
    setModalOpen(false);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Gestión de Calificaciones</h1>
        <button
          onClick={handleAgregar}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded"
        >
          + Agregar Calificación
        </button>
      </div>

      {loading ? (
        <p>Cargando calificaciones...</p>
      ) : calificaciones.length === 0 ? (
        <p>No hay calificaciones registradas.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
            <thead className="bg-blue-500 text-white">
              <tr>
                <th className="py-3 px-4 text-left">Alumno</th>
                <th className="py-3 px-4 text-left">Matrícula</th>
                <th className="py-3 px-4 text-left">Evaluación</th>
                <th className="py-3 px-4 text-left">Calificación</th>
                <th className="py-3 px-4 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {calificaciones.map(c => (
                <tr key={c.id} className="border-b hover:bg-gray-100">
                  <td className="py-2 px-4">{c.alumno}</td>
                  <td className="py-2 px-4">{c.matricula}</td>
                  <td className="py-2 px-4">{c.evaluacion}</td>
                  <td className="py-2 px-4">{c.calificacion}</td>
                  <td className="py-2 px-4 space-x-2">
                    <button
                      onClick={() => handleEditar(c)}
                      className="text-blue-500 hover:underline"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleEliminar(c.id)}
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
              {editingCal ? "Editar Calificación" : "Agregar Calificación"}
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
                placeholder="Evaluación"
                value={formData.evaluacion}
                onChange={e => setFormData({ ...formData, evaluacion: e.target.value })}
                className="border p-2 rounded"
              />
              <input
                type="number"
                placeholder="Calificación"
                value={formData.calificacion}
                onChange={e => setFormData({ ...formData, calificacion: Number(e.target.value) })}
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