import { useEffect, useState } from "react";
import { getEvaluaciones, type Evaluacion } from "../services/evaluacionesService";

export default function Evaluaciones() {
  const [evaluaciones, setEvaluaciones] = useState<Evaluacion[]>([]);
  const [loading, setLoading] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingEval, setEditingEval] = useState<Evaluacion | null>(null);
  const [formData, setFormData] = useState({
    nombre: "",
    descripcion: "",
    fecha: "",
    peso: 0,
  });

  useEffect(() => {
    async function fetchEvaluaciones() {
      try {
        const data = await getEvaluaciones();
        setEvaluaciones(data);
      } catch (error) {
        console.error("Error al cargar evaluaciones:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchEvaluaciones();
  }, []);

  const handleAgregar = () => {
    setFormData({ nombre: "", descripcion: "", fecha: "", peso: 0 });
    setEditingEval(null);
    setModalOpen(true);
  };

  const handleEditar = (evalItem: Evaluacion) => {
    setFormData({
      nombre: evalItem.nombre,
      descripcion: evalItem.descripcion || "",
      fecha: evalItem.fecha,
      peso: evalItem.peso,
    });
    setEditingEval(evalItem);
    setModalOpen(true);
  };

  const handleEliminar = (id: string) => {
    if (confirm("¿Deseas eliminar esta evaluación?")) {
      setEvaluaciones(evaluaciones.filter(e => e.id !== id));
    }
  };

  const handleGuardar = () => {
    if (editingEval) {
      setEvaluaciones(
        evaluaciones.map(e => (e.id === editingEval.id ? { ...e, ...formData } : e))
      );
    } else {
      const newEval: Evaluacion = { id: Date.now().toString(), ...formData };
      setEvaluaciones([...evaluaciones, newEval]);
    }
    setModalOpen(false);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Gestión de Evaluaciones</h1>
        <button
          onClick={handleAgregar}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded"
        >
          + Agregar Evaluación
        </button>
      </div>

      {loading ? (
        <p>Cargando evaluaciones...</p>
      ) : evaluaciones.length === 0 ? (
        <p>No hay evaluaciones registradas.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
            <thead className="bg-blue-500 text-white">
              <tr>
                <th className="py-3 px-4 text-left">Nombre</th>
                <th className="py-3 px-4 text-left">Descripción</th>
                <th className="py-3 px-4 text-left">Fecha</th>
                <th className="py-3 px-4 text-left">Porcentaje (%)</th>
                <th className="py-3 px-4 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {evaluaciones.map(e => (
                <tr key={e.id} className="border-b hover:bg-gray-100">
                  <td className="py-2 px-4">{e.nombre}</td>
                  <td className="py-2 px-4">{e.descripcion || "—"}</td>
                  <td className="py-2 px-4">{e.fecha}</td>
                  <td className="py-2 px-4">{e.peso}</td>
                  <td className="py-2 px-4 space-x-2">
                    <button
                      onClick={() => handleEditar(e)}
                      className="text-blue-500 hover:underline"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleEliminar(e.id)}
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
              {editingEval ? "Editar Evaluación" : "Agregar Evaluación"}
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
                placeholder="Descripción"
                value={formData.descripcion}
                onChange={e => setFormData({ ...formData, descripcion: e.target.value })}
                className="border p-2 rounded"
              />
              <input
                type="date"
                placeholder="Fecha"
                value={formData.fecha}
                onChange={e => setFormData({ ...formData, fecha: e.target.value })}
                className="border p-2 rounded"
              />
              <input
                type="number"
                placeholder="Peso (%)"
                value={formData.peso}
                onChange={e => setFormData({ ...formData, peso: Number(e.target.value) })}
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
