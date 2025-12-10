export interface Calificacion {
  id: string;
  alumno: string;
  matricula: string;
  evaluacion: string;
  calificacion: number;
}

export async function getCalificaciones(): Promise<Calificacion[]> {
  return [
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', evaluacion: 'Examen Parcial 1', calificacion: 85 },
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', evaluacion: 'Tarea 1', calificacion: 85 },
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', evaluacion: 'Tarea 2', calificacion: 80 },
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', evaluacion: 'Tarea 3', calificacion: 76 },
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', evaluacion: 'Tarea 4', calificacion: 94 },
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', evaluacion: 'Examen Parcial 2', calificacion: 84 },
    { id: '2', alumno: 'María López', matricula: 'A002', evaluacion: 'Examen Parcial 1', calificacion: 95 },
    { id: '2', alumno: 'María López', matricula: 'A002', evaluacion: 'Tarea 1', calificacion: 91 },
    { id: '2', alumno: 'María López', matricula: 'A002', evaluacion: 'Tarea 2', calificacion: 92 },
    { id: '2', alumno: 'María López', matricula: 'A002', evaluacion: 'Tarea 3', calificacion: 97 },
    { id: '2', alumno: 'María López', matricula: 'A002', evaluacion: 'Tarea 4', calificacion: 88 },
    { id: '2', alumno: 'María López', matricula: 'A002', evaluacion: 'TExamen Parcial 2', calificacion: 82 },
  ];
}