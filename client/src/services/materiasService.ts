import { getSubjects, type Subject } from "../api";

export interface Materia {
  id: number;
  nombre: string;
  descripcion?: string;
  codigo?: string;
  careerIds: number[];
}

function subjectToMateria(subject: Subject): Materia {
  const careerIds = subject.Careers ? subject.Careers.map(c => c.ID) : [];
  return {
    id: subject.ID,
    nombre: subject.Name,
    descripcion: subject.Description,
    codigo: subject.Code,
    careerIds: careerIds,
  };
}

export async function getMaterias(): Promise<Materia[]> {
  try {
    const subjects = await getSubjects();
    const materias = subjects.map(subjectToMateria);
    console.log("Todas las materias:", materias);
    return materias;
  } catch (error) {
    console.error("Error al cargar materias:", error);
    return [];
  }
}

export async function getMateria(id: number): Promise<Materia | null> {
  try {
    const materias = await getMaterias();
    return materias.find(m => m.id === id) || null;
  } catch (error) {
    console.error("Error al cargar materia:", error);
    return null;
  }
}

export async function getMateriasByCareer(careerId: number): Promise<Materia[]> {
  const materias = await getMaterias();
  return materias;
}
