// services/evaluacionesCompletas.ts
import {
  getStudents,
  getSubjects,
  getCareers,
  getGrades,
  createGrade,
  updateGrade,
  type Student,
  type Subject,
  type Career,
  type Grade,
  type ListStudentsParams,
  type ListSubjectsParams
} from "../api";

// Interfaces para las funcionalidades
export interface AlumnoParaCalificar {
  id: number;
  student_id: string;
  semester: number;
  account: {
    name: string;
    email: string;
  };
  career: {
    id: number;
    name: string;
  };
  currentGrade?: number;
  materiaId?: number;
}

export interface PromedioAlumno {
  student: Student;
  average: number;
  subjects_count: number;
  details: Array<{
    subject_id: number;
    subject_name: string;
    grade: number;
    semester: string;
  }>;
}

export interface EstadisticasCalificaciones {
  promedioGeneral: number;
  totalAlumnos: number;
  aprobados: number;
  reprobados: number;
  distribucion: Record<string, number>;
}

// Cache para materias
let materiasCache: Subject[] = [];
let materiasCacheTime = 0;
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutos

/**
 * Obtener todas las carreras
 */
export async function obtenerCarreras(): Promise<Career[]> {
  try {
    const carreras = await getCareers();
    return carreras;
  } catch (error) {
    console.error("Error obteniendo carreras:", error);
    return [];
  }
}

/**
 * Obtener grupos disponibles (simulados)
 * Como no tienes grupos en el backend, simularemos grupos A-E
 */
export function obtenerGruposDisponibles(): string[] {
  return ['A', 'B', 'C', 'D', 'E'];
}

/**
 * Obtener materias por carrera
 */
export async function obtenerMateriasPorCarrera(careerId: number): Promise<Subject[]> {
  try {
    const params: ListSubjectsParams = { career_id: careerId.toString() };
    const materias = await getSubjects(params);
    return materias;
  } catch (error) {
    console.error("Error obteniendo materias:", error);
    return [];
  }
}

/**
 * Obtener alumnos para calificar
 */
export async function obtenerAlumnosParaCalificar(
  carreraId: number,
  materiaId?: number
): Promise<AlumnoParaCalificar[]> {
  try {
    // Obtener estudiantes de la carrera
    const params: ListStudentsParams = { career_id: carreraId.toString() };
    const estudiantes = await getStudents(params);

    // Transformar a formato para calificar
    const alumnosParaCalificar: AlumnoParaCalificar[] = await Promise.all(
      estudiantes.map(async (est) => {
        let currentGrade: number | undefined;

        // Si hay materiaId, buscar calificación existente
        if (materiaId) {
          const calificaciones = await getGrades({
            student_id: est.ID,
            subject_id: materiaId
          });

          if (calificaciones.length > 0) {
            currentGrade = calificaciones[0].Value;
          }
        }

        return {
          id: est.ID,
          student_id: est.StudentID,
          semester: est.Semester || 1,
          account: {
            name: est.Account?.Name || '',
            email: est.Account?.Email || ''
          },
          career: {
            id: est.CareerID,
            name: est.Career?.Name || ''
          },
          currentGrade,
          materiaId
        };
      })
    );

    return alumnosParaCalificar;
  } catch (error) {
    console.error("Error obteniendo alumnos para calificar:", error);
    return [];
  }
}

/**
 * Obtener materia por ID (con caché)
 */
async function obtenerMateriaPorId(id: number): Promise<Subject | null> {
  try {
    // Usar caché para evitar llamadas repetidas
    const now = Date.now();
    if (now - materiasCacheTime > CACHE_DURATION) {
      materiasCache = await getSubjects();
      materiasCacheTime = now;
    }

    return materiasCache.find(m => m.ID === id) || null;
  } catch (error) {
    console.error(`Error obteniendo materia ${id}:`, error);
    return null;
  }
}

/**
 * Guardar calificación de un alumno
 */
export async function guardarCalificacionAlumno(
  studentId: number,
  subjectId: number,
  grade: number,
  notas?: string
): Promise<boolean> {
  try {
    // Buscar si ya existe una calificación para este estudiante en esta materia
    const calificacionesExistentes = await getGrades({
      student_id: studentId,
      subject_id: subjectId
    });

    if (calificacionesExistentes.length > 0) {
      // Actualizar calificación existente
      const calificacionId = calificacionesExistentes[0].ID;
      await updateGrade(calificacionId, {
        value: grade,
        notes: notas || `Actualizado: ${new Date().toLocaleDateString()}`
      });
    } else {
      // Crear nueva calificación
      await createGrade({
        value: grade,
        notes: notas || `Creado: ${new Date().toLocaleDateString()}`,
        student_id: studentId,
        subject_id: subjectId,
        teacher_id: 0 // Usar 0 o null según lo permita tu backend
      });
    }

    return true;
  } catch (error) {
    console.error("Error guardando calificación:", error);
    return false;
  }
}

/**
 * Guardar múltiples calificaciones
 */
export async function guardarCalificacionesMasivo(
  calificaciones: Array<{
    studentId: number;
    subjectId: number;
    grade: number;
    notes?: string;
  }>
): Promise<boolean> {
  try {
    // Guardar una por una (no tienes endpoint batch)
    const resultados = await Promise.all(
      calificaciones.map(cal =>
        guardarCalificacionAlumno(cal.studentId, cal.subjectId, cal.grade, cal.notes)
          .catch(() => false)
      )
    );

    return resultados.every(result => result === true);
  } catch (error) {
    console.error("Error guardando calificaciones en lote:", error);
    return false;
  }
}

/**
 * Calcular promedios de alumnos
 */
export async function calcularPromediosAlumnos(
  carreraId?: number
): Promise<PromedioAlumno[]> {
  try {
    // Obtener estudiantes filtrados
    const params: ListStudentsParams = {};
    if (carreraId) params.career_id = carreraId.toString();

    const estudiantes = await getStudents(params);

    // Para cada estudiante, calcular promedio
    const promedios: PromedioAlumno[] = [];

    for (const estudiante of estudiantes) {
      try {
        // Obtener todas las calificaciones del estudiante
        const calificaciones = await getGrades({ student_id: estudiante.ID });

        // Calcular promedio
        let promedio = 0;
        let detalles: Array<{
          subject_id: number;
          subject_name: string;
          grade: number;
          semester: string;
        }> = [];

        if (calificaciones.length > 0) {
          const suma = calificaciones.reduce((total, cal) => total + cal.Value, 0);
          promedio = suma / calificaciones.length;

          // Obtener detalles de cada materia
          for (const cal of calificaciones) {
            try {
              const materia = await obtenerMateriaPorId(cal.SubjectID);
              detalles.push({
                subject_id: cal.SubjectID,
                subject_name: materia?.Name || `Materia ${cal.SubjectID}`,
                grade: cal.Value,
                semester: materia?.Semester?.toString() || 'N/A'
              });
            } catch (error) {
              console.error(`Error obteniendo materia ${cal.SubjectID}:`, error);
            }
          }
        }

        promedios.push({
          student: estudiante,
          average: parseFloat(promedio.toFixed(2)),
          subjects_count: calificaciones.length,
          details: detalles
        });

      } catch (error) {
        console.error(`Error procesando estudiante ${estudiante.ID}:`, error);
      }
    }

    return promedios;
  } catch (error) {
    console.error("Error calculando promedios:", error);
    return [];
  }
}

/**
 * Obtener estadísticas de calificaciones
 */
export async function obtenerEstadisticasCalificaciones(
  carreraId?: number
): Promise<EstadisticasCalificaciones> {
  try {
    const promedios = await calcularPromediosAlumnos(carreraId);

    if (promedios.length === 0) {
      return {
        promedioGeneral: 0,
        totalAlumnos: 0,
        aprobados: 0,
        reprobados: 0,
        distribucion: {}
      };
    }

    // Calcular estadísticas
    const promedioGeneral = promedios.reduce((sum, p) => sum + p.average, 0) / promedios.length;
    const aprobados = promedios.filter(p => p.average >= 6).length;
    const reprobados = promedios.filter(p => p.average < 6).length;

    // Distribución por rangos (0-10)
    const distribucion: Record<string, number> = {
      '9-10': promedios.filter(p => p.average >= 9).length,
      '8-8.9': promedios.filter(p => p.average >= 8 && p.average < 9).length,
      '7-7.9': promedios.filter(p => p.average >= 7 && p.average < 8).length,
      '6-6.9': promedios.filter(p => p.average >= 6 && p.average < 7).length,
      '0-5.9': promedios.filter(p => p.average < 6).length,
    };

    return {
      promedioGeneral: parseFloat(promedioGeneral.toFixed(2)),
      totalAlumnos: promedios.length,
      aprobados,
      reprobados,
      distribucion
    };
  } catch (error) {
    console.error("Error obteniendo estadísticas:", error);
    return {
      promedioGeneral: 0,
      totalAlumnos: 0,
      aprobados: 0,
      reprobados: 0,
      distribucion: {}
    };
  }
}
