import {
  getCareers,
  getCareer,
  createCareer,
  updateCareer,
  deleteCareer,
  type Career,
} from "../api";

// Legacy interface for backward compatibility with existing components
export interface Grupo {
  id: string;
  nombre: string;
  descripcion?: string;
  codigo?: string;
  duracion?: number;
  activo?: boolean;
}

// Transform backend Career to frontend Grupo format
function careerToGrupo(career: Career): Grupo {
  return {
    id: career.ID.toString(),
    nombre: career.Name,
    descripcion: career.Description,
    codigo: career.Code,
    duracion: career.Duration,
    activo: career.Active,
  };
}

/**
 * Get all groups (careers)
 */
export async function getGrupos(): Promise<Grupo[]> {
  const careers = await getCareers();
  return careers.map(careerToGrupo);
}

/**
 * Get a single group by ID
 */
export async function getGrupo(id: string): Promise<Grupo> {
  const career = await getCareer(parseInt(id, 10));
  return careerToGrupo(career);
}

/**
 * Add a new group
 */
export async function addGrupo(grupo: Omit<Grupo, "id">): Promise<Grupo> {
  const career = await createCareer({
    name: grupo.nombre,
    code: grupo.codigo || grupo.nombre.toUpperCase().replace(/\s+/g, "-"),
    description: grupo.descripcion,
    duration: grupo.duracion,
  });

  return careerToGrupo(career);
}

/**
 * Update an existing group
 */
export async function updateGrupo(grupo: Grupo): Promise<Grupo> {
  const career = await updateCareer(parseInt(grupo.id, 10), {
    name: grupo.nombre,
    code: grupo.codigo,
    description: grupo.descripcion,
    duration: grupo.duracion,
    active: grupo.activo,
  });

  return careerToGrupo(career);
}

/**
 * Delete a group by ID
 */
export async function deleteGrupo(id: string): Promise<void> {
  await deleteCareer(parseInt(id, 10));
}
