import { useEffect, useState } from "react";
import { getAlumnos, type Alumno } from "../services/alumnosService";
import { getAsistencias, type Asistencia } from "../services/asistenciasService";
import { getEvaluaciones, type Evaluacion } from "../services/evaluacionesService";
import { getCalificaciones, type Calificacion } from "../services/calificacionesService";

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
} from "recharts";

export default function Dashboard() {
  const [alumnos, setAlumnos] = useState<Alumno[]>([]);
  const [asistencias, setAsistencias] = useState<Asistencia[]>([]);
  const [evaluaciones, setEvaluaciones] = useState<Evaluacion[]>([]);
  const [calificaciones, setCalificaciones] = useState<Calificacion[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      try {
        const alumnosData = await getAlumnos();
        const asistenciasData = await getAsistencias();
        const evaluacionesData = await getEvaluaciones();
        const calificacionesData = await getCalificaciones();

        setAlumnos(alumnosData);
        setAsistencias(asistenciasData);
        setEvaluaciones(evaluacionesData);
        setCalificaciones(calificacionesData);
      } catch (error) {
        console.error("Error al cargar dashboard:", error);
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  if (loading) return <p>Cargando dashboard...</p>;

  const totalAlumnos = alumnos.length;
  const totalAsistencias = asistencias.length;
  const totalEvaluaciones = evaluaciones.length;
  const promedioCalificaciones =
    calificaciones.length > 0
      ? calificaciones.reduce((sum, c) => sum + c.calificacion, 0) / calificaciones.length
      : 0;

  const asistenciasPorDia = asistencias.reduce((acc: any[], curr) => {
    const fecha = curr.fecha;
    const index = acc.findIndex(a => a.fecha === fecha);
    if (index >= 0) {
      acc[index].cantidad += 1;
    } else {
      acc.push({ fecha, cantidad: 1 });
    }
    return acc;
  }, []);

  const promedioPorEvaluacion = evaluaciones.map(e => {
    const calfs = calificaciones.filter(c => c.evaluacion === e.nombre);
    const promedio =
      calfs.length > 0 ? calfs.reduce((sum, c) => sum + c.calificacion, 0) / calfs.length : 0;
    return { evaluacion: e.nombre, promedio };
  });

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold mb-6">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-blue-500 text-white p-6 rounded shadow">
          <h2 className="text-xl font-semibold">Alumnos</h2>
          <p className="text-3xl mt-2">{totalAlumnos}</p>
        </div>
        <div className="bg-green-500 text-white p-6 rounded shadow">
          <h2 className="text-xl font-semibold">Asistencias</h2>
          <p className="text-3xl mt-2">{totalAsistencias}</p>
        </div>
        <div className="bg-yellow-500 text-white p-6 rounded shadow">
          <h2 className="text-xl font-semibold">Evaluaciones</h2>
          <p className="text-3xl mt-2">{totalEvaluaciones}</p>
        </div>
        <div className="bg-purple-500 text-white p-6 rounded shadow">
          <h2 className="text-xl font-semibold">Promedio Calificaciones</h2>
          <p className="text-3xl mt-2">{promedioCalificaciones.toFixed(1)}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white p-4 rounded shadow">
          <h2 className="text-xl font-semibold mb-4">Asistencias por Día</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={asistenciasPorDia}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="fecha" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="cantidad" fill="#4ade80" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="bg-white p-4 rounded shadow">
          <h2 className="text-xl font-semibold mb-4">Promedio por Evaluación</h2>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={promedioPorEvaluacion}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="evaluacion" />
              <YAxis domain={[0, 100]} />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="promedio" stroke="#3b82f6" strokeWidth={3} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}
