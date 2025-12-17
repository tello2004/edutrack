import { useEffect, useState, useMemo } from "react";
import {
  getAsistencias,
  type Asistencia,
  addAsistencia,
  updateAsistencia,
  deleteAsistencia,
  registrarAsistenciaGrupal
} from "../services/asistenciasService";
import { getAlumnos, type Alumno } from "../services/alumnosService";
import { getGrupos, type Grupo } from "../services/gruposService";
import { getMateriasByCareer, type Materia } from "../services/materiasService";

export default function Asistencias() {
  const [grupos, setGrupos] = useState<Grupo[]>([]);
  const [materias, setMaterias] = useState<Materia[]>([]);
  const [alumnos, setAlumnos] = useState<Alumno[]>([]);
  const [asistencias, setAsistencias] = useState<Asistencia[]>([]);
  const [loading, setLoading] = useState(true);
  const [guardando, setGuardando] = useState(false);
  const [grupoSeleccionado, setGrupoSeleccionado] = useState<string>("");
  const [materiaSeleccionada, setMateriaSeleccionada] = useState<string>("");
  const [fechaSeleccionada, setFechaSeleccionada] = useState<string>(
    new Date().toISOString().split("T")[0]
  );

  const [alumnoBuscado, setAlumnoBuscado] = useState<string>("");
  const [alumnoSeleccionado, setAlumnoSeleccionado] = useState<Alumno | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingAsistencia, setEditingAsistencia] = useState<Asistencia | null>(null);
  const [formData, setFormData] = useState({
    presente: true,
    notas: "",
  });
  useEffect(() => {
    async function fetchDatosIniciales() {
      try {
        const [gruposData, alumnosData] = await Promise.all([
          getGrupos(),
          getAlumnos(),
        ]);
        setGrupos(gruposData);
        setAlumnos(alumnosData);
      } catch (error) {
        console.error("Error al cargar datos iniciales:", error);
      }
    }
    fetchDatosIniciales();
  }, []);

 useEffect(() => {
  async function fetchMaterias() {
    const careerId = parseInt(grupoSeleccionado, 10);
    if (isNaN(careerId)) {
      setMaterias([]);
      setMateriaSeleccionada("");
      return;
    }
    try {
      const materiasData = await getMateriasByCareer(careerId);
      console.log("Materias cargadas:", materiasData);
      setMaterias(materiasData);
    } catch (error) {
      console.error("Error al cargar materias:", error);
      setMaterias([]);
    }
  }
  if (grupoSeleccionado) fetchMaterias();
}, [grupoSeleccionado]);

  const enriquecerAsistenciasConAlumnos = (asistenciasData: Asistencia[]) => {
    return asistenciasData.map(asistencia => {
      const alumno = alumnos.find(a =>
        a.matricula === asistencia.matricula ||
        a.id === asistencia.studentId?.toString()
      );

      const materia = materias.find(m =>
        m.id.toString() === materiaSeleccionada ||
        m.id === asistencia.subjectId
      );

      return {
        ...asistencia,
        alumno: asistencia.alumno || alumno?.nombre || "Desconocido",
        grupo: asistencia.grupo || materia?.nombre || "Sin grupo",
        studentId: asistencia.studentId || (alumno ? parseInt(alumno.id) : 0),
        subjectId: asistencia.subjectId || (materia ? materia.id : 0)
      };
    });
  };

  useEffect(() => {
    async function fetchAsistencias() {
      if (!materiaSeleccionada) {
        setAsistencias([]);
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        const data = await getAsistencias({
          subject_id: parseInt(materiaSeleccionada),
          date: fechaSeleccionada
        });
        const asistenciasEnriquecidas = enriquecerAsistenciasConAlumnos(data);
        setAsistencias(asistenciasEnriquecidas);
      } catch (error) {
        console.error("Error al cargar asistencias:", error);
      } finally {
        setLoading(false);
      }
    }
    if (fechaSeleccionada && materiaSeleccionada) {
      fetchAsistencias();
    }
  }, [materiaSeleccionada, fechaSeleccionada, alumnos, materias]);
  const alumnosDelGrupo = useMemo(() => {
    if (!grupoSeleccionado) return [];
    return alumnos.filter(alumno =>
      alumno.careerId?.toString() === grupoSeleccionado
    );
  }, [alumnos, grupoSeleccionado]);
  const alumnosFiltrados = useMemo(() => {
    if (!alumnoBuscado.trim() || !grupoSeleccionado) return [];

    const searchTerm = alumnoBuscado.toLowerCase();
    return alumnosDelGrupo.filter(alumno =>
      alumno.nombre.toLowerCase().includes(searchTerm) ||
      alumno.matricula.toLowerCase().includes(searchTerm)
    ).slice(0, 5);
  }, [alumnoBuscado, alumnosDelGrupo, grupoSeleccionado]);
  const inicializarAsistenciasParaMateria = () => {
    if (!materiaSeleccionada || !fechaSeleccionada) return;
    const materia = materias.find(m => m.id.toString() === materiaSeleccionada);
    const asistenciasExistentesMatriculas = asistencias.map(a => a.matricula);
    const nuevasAsistencias = alumnosDelGrupo
      .filter(alumno => !asistenciasExistentesMatriculas.includes(alumno.matricula))
      .map(alumno => ({
        id: `temp-${Date.now()}-${alumno.id}`,
        alumno: alumno.nombre,
        matricula: alumno.matricula,
        grupo: materia?.nombre || "Sin nombre",
        fecha: fechaSeleccionada,
        presente: true,
        notas: "",
        studentId: parseInt(alumno.id),
        subjectId: parseInt(materiaSeleccionada)
      } as Asistencia));

    setAsistencias([...asistencias, ...nuevasAsistencias]);
  };
  const handleAgregarAlumno = () => {
    if (!alumnoSeleccionado || !materiaSeleccionada || !fechaSeleccionada) {
      alert("Por favor, complete todos los campos");
      return;
    }
    const yaExiste = asistencias.some(a => a.matricula === alumnoSeleccionado.matricula);
    if (yaExiste) {
      alert("Este alumno ya está en la lista de asistencias");
      return;
    }
    const materia = materias.find(m => m.id.toString() === materiaSeleccionada);
    const nuevaAsistencia: Asistencia = {
      id: `temp-${Date.now()}-${alumnoSeleccionado.id}`,
      alumno: alumnoSeleccionado.nombre,
      matricula: alumnoSeleccionado.matricula,
      grupo: materia?.nombre || "Sin nombre",
      fecha: fechaSeleccionada,
      presente: true,
      notas: "",
      studentId: parseInt(alumnoSeleccionado.id),
      subjectId: parseInt(materiaSeleccionada)
    };

    setAsistencias([...asistencias, nuevaAsistencia]);
    setAlumnoBuscado("");
    setAlumnoSeleccionado(null);
  };
  const handleGuardarGrupo = async () => {
    if (!materiaSeleccionada || !fechaSeleccionada || asistencias.length === 0) {
      alert("No hay asistencias para guardar");
      return;
    }

    try {
      setGuardando(true);
      const asistenciasTemporales = asistencias.filter(a => a.id.startsWith('temp-'));
      const asistenciasExistentes = asistencias.filter(a => !a.id.startsWith('temp-'));
      const nuevasAsistenciasGuardadas: Asistencia[] = [];
      if (asistenciasTemporales.length > 0) {
        for (const asistencia of asistenciasTemporales) {
          try {
            const asistenciaGuardada = await addAsistencia({
              fecha: asistencia.fecha,
              presente: asistencia.presente,
              notas: asistencia.notas || "",
              studentId: asistencia.studentId!,
              subjectId: asistencia.subjectId!
            });
            nuevasAsistenciasGuardadas.push(asistenciaGuardada);
          } catch (error) {
            console.error("Error al guardar asistencia:", error);
          }
        }
      }
      if (asistenciasExistentes.length > 0) {
        await Promise.all(
          asistenciasExistentes.map(async (asistencia) => {
            if (asistencia.studentId && asistencia.subjectId) {
              try {
                await updateAsistencia(asistencia);
              } catch (error) {
                console.error("Error al actualizar asistencia:", error);
              }
            }
          })
        );
      }
      const todasLasAsistencias = [
        ...asistenciasExistentes,
        ...nuevasAsistenciasGuardadas
      ];
      const asistenciasEnriquecidas = enriquecerAsistenciasConAlumnos(todasLasAsistencias);
      setAsistencias(asistenciasEnriquecidas);

      alert("Asistencias guardadas correctamente");

    } catch (error: any) {
      console.error("Error al guardar asistencias:", error);
      alert(`Error al guardar las asistencias: ${error.message || "Verifica los datos"}`);
    } finally {
      setGuardando(false);
    }
  };
  const handleEditar = (asistencia: Asistencia) => {
    setEditingAsistencia(asistencia);
    setFormData({
      presente: asistencia.presente,
      notas: asistencia.notas || "",
    });
    setModalOpen(true);
  };

  const handleGuardarEdicion = async () => {
    if (!editingAsistencia) return;

    try {
      const asistenciaActualizada = {
        ...editingAsistencia,
        ...formData,
        status: formData.presente ? "present" : "absent",
      };

      if (editingAsistencia.id.startsWith('temp-')) {
        setAsistencias(asistencias.map(a =>
          a.id === editingAsistencia.id ? asistenciaActualizada : a
        ));
      } else {
        const updated = await updateAsistencia(asistenciaActualizada);
        const asistenciaEnriquecida = enriquecerAsistenciasConAlumnos([updated])[0];
        setAsistencias(asistencias.map(a =>
          a.id === updated.id ? asistenciaEnriquecida : a
        ));
      }

      setModalOpen(false);
    } catch (error) {
      console.error("Error al actualizar asistencia:", error);
      alert("Error al actualizar la asistencia");
    }
  };

  const handleEliminar = async (asistencia: Asistencia) => {
    if (!confirm("¿Deseas eliminar esta asistencia?")) return;

    try {
      if (!asistencia.id.startsWith('temp-')) {
        await deleteAsistencia(asistencia.id);
      }
      setAsistencias(asistencias.filter(a => a.id !== asistencia.id));
    } catch (error) {
      console.error("Error al eliminar asistencia:", error);
      alert("Error al eliminar la asistencia");
    }
  };

  const handleTogglePresente = (id: string) => {
    setAsistencias(asistencias.map(asistencia =>
      asistencia.id === id
        ? { ...asistencia, presente: !asistencia.presente }
        : asistencia
    ));
  };

  const nombreGrupoSeleccionado = grupos.find(
    g => g.id === grupoSeleccionado
  )?.nombre || "";

  const nombreMateriaSeleccionada = materias.find(
    m => m.id.toString() === materiaSeleccionada
  )?.nombre || "";

  return (
    <div className="p-6">
      <div className="mb-8 p-4 bg-gray-50 rounded-lg">
        <h2 className="text-xl font-semibold mb-4">Registro de Asistencias</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Grupo (Carrera)</label>
            <select
              value={grupoSeleccionado}
              onChange={(e) => {
                setGrupoSeleccionado(e.target.value);
                setMateriaSeleccionada("");
              }}
              className="w-full border rounded-lg p-2"
            >
              <option value="">Seleccionar grupo</option>
              {grupos.map(grupo => (
                <option key={grupo.id} value={grupo.id.toString()}>
                  {grupo.nombre}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Materia</label>
            <select
              value={materiaSeleccionada}
              onChange={(e) => setMateriaSeleccionada(e.target.value)}
              disabled={!grupoSeleccionado}
              className="w-full border rounded-lg p-2 disabled:opacity-50"
            >
              <option value="">Seleccionar materia</option>
              {materias.map(materia => (
                <option key={materia.id} value={materia.id}>
                  {materia.nombre}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Fecha</label>
            <input
              type="date"
              value={fechaSeleccionada}
              onChange={(e) => setFechaSeleccionada(e.target.value)}
              className="w-full border rounded-lg p-2"
            />
          </div>
          <div className="flex items-end gap-2">
            <button
              onClick={inicializarAsistenciasParaMateria}
              disabled={!materiaSeleccionada || !fechaSeleccionada}
              className={`flex-1 py-2 px-4 rounded-lg ${
                materiaSeleccionada && fechaSeleccionada
                  ? "bg-blue-600 hover:bg-blue-700 text-white"
                  : "bg-gray-300 text-gray-500 cursor-not-allowed"
              }`}
            >
              Inicializar Lista
            </button>
            <button
              onClick={handleGuardarGrupo}
              disabled={asistencias.length === 0 || guardando}
              className={`flex-1 py-2 px-4 rounded-lg ${
                asistencias.length > 0 && !guardando
                  ? "bg-green-600 hover:bg-green-700 text-white"
                  : "bg-gray-300 text-gray-500 cursor-not-allowed"
              }`}
            >
              {guardando ? "Guardando..." : "Guardar Todo"}
            </button>
          </div>
        </div>
      </div>
      {materiaSeleccionada && (
        <div className="mb-6 p-4 bg-white border rounded-lg shadow">
          <h3 className="text-lg font-medium mb-3">Agregar Alumno Específico</h3>
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1">
              <label className="block text-sm font-medium mb-1">Buscar Alumno</label>
              <div className="relative">
                <input
                  type="text"
                  value={alumnoBuscado}
                  onChange={(e) => {
                    setAlumnoBuscado(e.target.value);
                    setAlumnoSeleccionado(null);
                  }}
                  placeholder="Escribe nombre o matrícula..."
                  className="w-full border rounded-lg p-2 pr-10"
                  disabled={!grupoSeleccionado}
                />
                {alumnoBuscado && alumnosFiltrados.length > 0 && (
                  <div className="absolute z-10 w-full mt-1 bg-white border rounded-lg shadow-lg max-h-60 overflow-y-auto">
                    {alumnosFiltrados.map(alumno => (
                      <div
                        key={alumno.id}
                        onClick={() => {
                          setAlumnoSeleccionado(alumno);
                          setAlumnoBuscado(`${alumno.nombre} (${alumno.matricula})`);
                        }}
                        className="p-2 hover:bg-gray-100 cursor-pointer border-b last:border-b-0"
                      >
                        <div className="font-medium">{alumno.nombre}</div>
                        <div className="text-sm text-gray-600">{alumno.matricula}</div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
            <div className="flex items-end">
              <button
                onClick={handleAgregarAlumno}
                disabled={!alumnoSeleccionado}
                className={`px-6 py-2 rounded-lg ${
                  alumnoSeleccionado
                    ? "bg-blue-600 hover:bg-blue-700 text-white"
                    : "bg-gray-300 text-gray-500 cursor-not-allowed"
                }`}
              >
                Agregar
              </button>
            </div>
          </div>
          {alumnoSeleccionado && (
            <div className="mt-3 p-3 bg-blue-50 rounded-lg">
              <span className="font-medium">Alumno seleccionado:</span>{" "}
              {alumnoSeleccionado.nombre} ({alumnoSeleccionado.matricula})
            </div>
          )}
        </div>
      )}
      <div className="bg-white rounded-lg shadow">
        <div className="p-4 border-b">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold">
                Asistencias - {nombreGrupoSeleccionado} - {nombreMateriaSeleccionada}
              </h1>
              <p className="text-gray-600">
                {fechaSeleccionada} • {asistencias.length} registros
              </p>
            </div>
            <div className="text-sm text-gray-500">
              {asistencias.filter(a => a.presente).length} presentes •
              {asistencias.filter(a => !a.presente).length} ausentes
            </div>
          </div>
        </div>

        {loading ? (
          <div className="p-8 text-center">
            <p>Cargando asistencias...</p>
          </div>
        ) : !materiaSeleccionada ? (
          <div className="p-8 text-center">
            <p className="text-gray-500">Seleccione una materia para comenzar</p>
          </div>
        ) : asistencias.length === 0 ? (
          <div className="p-8 text-center">
            <p>No hay registros de asistencias para esta materia y fecha.</p>
            <p className="text-gray-500 mt-2">
              Use "Inicializar Lista" para crear la lista o agregue alumnos individualmente.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead className="bg-gray-100">
                <tr>
                  <th className="py-3 px-4 text-left">Matrícula</th>
                  <th className="py-3 px-4 text-left">Alumno</th>
                  <th className="py-3 px-4 text-left">Asistencia</th>
                  <th className="py-3 px-4 text-left">Notas</th>
                  <th className="py-3 px-4 text-left">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {asistencias.map(asistencia => (
                  <tr key={asistencia.id} className="border-b hover:bg-gray-50">
                    <td className="py-3 px-4">{asistencia.matricula}</td>
                    <td className="py-3 px-4">{asistencia.alumno}</td>
                    <td className="py-3 px-4">
                      <button
                        onClick={() => handleTogglePresente(asistencia.id)}
                        className={`px-3 py-1 rounded-full text-sm font-medium ${
                          asistencia.presente
                            ? "bg-green-100 text-green-800"
                            : "bg-red-100 text-red-800"
                        }`}
                      >
                        {asistencia.presente ? "Presente" : "Ausente"}
                      </button>
                    </td>
                    <td className="py-3 px-4">
                      <span className="truncate max-w-xs block">
                        {asistencia.notas || "-"}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <button
                        onClick={() => handleEditar(asistencia)}
                        className="text-blue-600 hover:text-blue-800 mr-3"
                      >
                        Editar
                      </button>
                      <button
                        onClick={() => handleEliminar(asistencia)}
                        className="text-red-600 hover:text-red-800"
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

      {modalOpen && editingAsistencia && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-md">
            <div className="p-6">
              <h2 className="text-xl font-bold mb-4">Editar Asistencia</h2>
              <p className="mb-4">
                <span className="font-medium">Alumno:</span> {editingAsistencia.alumno}
              </p>

              <div className="space-y-4">
                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={formData.presente}
                      onChange={(e) => setFormData({...formData, presente: e.target.checked})}
                      className="w-4 h-4"
                    />
                    <span>Presente</span>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Notas</label>
                  <textarea
                    value={formData.notas}
                    onChange={(e) => setFormData({...formData, notas: e.target.value})}
                    className="w-full border rounded-lg p-2 h-24"
                    placeholder="Observaciones..."
                  />
                </div>
              </div>

              <div className="flex justify-end mt-6 space-x-3">
                <button
                  onClick={() => setModalOpen(false)}
                  className="px-4 py-2 border rounded-lg hover:bg-gray-50"
                >
                  Cancelar
                </button>
                <button
                  onClick={handleGuardarEdicion}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                >
                  Guardar
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
