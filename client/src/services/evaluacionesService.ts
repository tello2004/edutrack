export interface Evaluacion {
  id: string;
  nombre: string;
  descripcion?: string;
  fecha: string;
  peso: number;
}

export async function getEvaluaciones(): Promise<Evaluacion[]> {
  return [
    { id: '1', nombre: 'Examen Parcial 1', fecha: '2025-12-15', peso: 30 },
    { id: '2', nombre: 'Tarea 1', fecha: '2025-12-10', peso: 40 },
    { id: '3', nombre: 'Tarea 2', fecha: '2025-12-13', peso: 50 },
    { id: '4', nombre: 'Examen Parcial 2', fecha: '2025-12-25', peso: 60 },
    { id: '5', nombre: 'Tarea 3', fecha: '2025-12-20', peso: 20 },
    { id: '6', nombre: 'Tarea 4', fecha: '2025-12-23', peso: 10 },
  ];
}