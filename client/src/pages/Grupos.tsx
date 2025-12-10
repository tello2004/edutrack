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

  useEffect(() => {
    async function fetchGrupos() {
      try {
        const data = await getGrupos();
        setGrupos(data);
      } catch (error) {
        console.error("Error al cargar grupos:", error);
      } finally {
        setLoading(false);
      }
    }

    fetchGrupos();
  }, []);

  const handleAgregar = async () => {
    const nombre = prompt("Nombre del grupo:");
    if (!nombre) return;
    const nuevo: Grupo = { id: Date.now().toString(), nombre };
    await addGrupo(nuevo);
    setGrupos(prev => [...prev, nuevo]);
  };

  const handleEditar = async (grupo: Grupo) => {
    const nombre = prompt("Editar nombre del grupo:", grupo.nombre);
    if (!nombre) return;
    const actualizado = { ...grupo, nombre };
    await updateGrupo(actualizado);
    setGrupos(prev => prev.map(g => (g.id === grupo.id ? actualizado : g)));
  };

  const handleEliminar = async (id: string) => {
    if (!confirm("¿Eliminar este grupo?")) return;
    await deleteGrupo(id);
    setGrupos(prev => prev.filter(g => g.id !== id));
  };

  if (loading) return <p>Cargando grupos...</p>;

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Gestión de Grupos</h1>
        <button
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded"
          onClick={handleAgregar}
        >
          + Agregar Grupo
        </button>
      </div>

      {grupos.length === 0 ? (
        <p>No hay grupos registrados.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden">
            <thead className="bg-blue-500 text-white">
              <tr>
                <th className="py-3 px-4 text-left">Nombre</th>
                <th className="py-3 px-4 text-left">Descripción</th>
                <th className="py-3 px-4 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {grupos.map(grupo => (
                <tr key={grupo.id} className="border-b hover:bg-gray-100">
                  <td className="py-2 px-4">{grupo.nombre}</td>
                  <td className="py-2 px-4">{grupo.descripcion || "—"}</td>
                  <td className="py-2 px-4 space-x-2">
                    <button
                      className="text-blue-500 hover:underline"
                      onClick={() => handleEditar(grupo)}
                    >
                      Editar
                    </button>
                    <button
                      className="text-red-500 hover:underline"
                      onClick={() => handleEliminar(grupo.id)}
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
    </div>
  );
}
